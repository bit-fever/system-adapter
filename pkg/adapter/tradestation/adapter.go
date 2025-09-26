//=============================================================================
/*
Copyright © 2025 Andrea Carboni andrea.carboni71@gmail.com

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
//=============================================================================

package tradestation

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"net/http/cookiejar"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/bit-fever/core/datatype"
	"github.com/bit-fever/core/req"
	"github.com/bit-fever/system-adapter/pkg/adapter"
)

//=============================================================================

func NewAdapter() adapter.Adapter {
	return &tradestation{}
}

//=============================================================================

func (a *tradestation) GetInfo() *adapter.Info {
	return &info
}

//=============================================================================

func (a *tradestation) GetAuthUrl() string {
	return ""
}

//=============================================================================

func (a *tradestation) Clone(configParams map[string]any, connectParams map[string]any) adapter.Adapter {
	b := *a
	b.configParams  = retrieveConfigParams (configParams)
	b.connectParams = retrieveConnectParams(connectParams)
	return &b
}

//=============================================================================

func (a *tradestation) Connect(ctx *adapter.ConnectionContext) (adapter.ConnectionResult,error) {
	jar,_:= cookiejar.New(nil)

	a.client = &http.Client{
		Jar: jar,
		Timeout: time.Minute * 3,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
			},
		},
	}

	loginInfo,err := a.createLoginInfo()
	if err != nil {
		return adapter.ConnectionResultError, err
	}

	a.clientId = loginInfo.Client

	loginRes,err := a.login(loginInfo)
	if err != nil {
		return adapter.ConnectionResultError, err
	}

	newState, err := a.callCallback(loginRes)
	if err != nil {
		return adapter.ConnectionResultError, err
	}

	err = a.submitTwoFACode(newState)
	if err != nil {
		return adapter.ConnectionResultError, err
	}

	a.apiUrl = DemoAPI
	if a.configParams.LiveAccount {
		a.apiUrl = LiveAPI
	}

	//--- Test tokens & accounts
	err = a.testToken()
	if err != nil {
		return adapter.ConnectionResultError, err
	}

	return adapter.ConnectionResultConnected, nil
}

//=============================================================================

func (a *tradestation) Disconnect(ctx *adapter.ConnectionContext) error {
	return nil
}

//=============================================================================

func (a *tradestation) IsWebLoginCompleted(httpCode int, path string) bool {
	return false
}
//=============================================================================

func (a *tradestation) InitFromWebLogin(reqHeader *http.Header, resCookies []*http.Cookie) error {
	return nil
}

//=============================================================================

func (a *tradestation) GetTokenExpSeconds() int {
	return 20*60;
}

//=============================================================================

func (a *tradestation) RefreshToken() error {
	payload := bytes.NewBufferString("")

	rq, err := http.NewRequest("POST", RefreshTokenUrl, payload)
	if err != nil {
		return err
	}

	setupCommonHeader(&rq.Header)
	rq.Header.Set("Accept",         "*/*")
	rq.Header.Set("Origin",         "https://my.tradestation.com")
	rq.Header.Set("Sec-Fetch-Dest", "empty")
	rq.Header.Set("Sec-Fetch-Mode", "cors")
	rq.Header.Set("Sec-Fetch-Site", "same-origin")

	res, err := a.client.Do(rq)

	var out TokenRefreshResponse
	err = req.BuildResponse(res, err, &out)

	if err == nil {
		a.accessToken = out.AccessToken
		a.refreshToken= out.IdToken

		if a.accessToken == "" {
			err = errors.New("Got an empty access token (refresh token is not working)")
		}
	}

	if res != nil && res.Body != nil {
		_=res.Body.Close()
	}

	return err
}

//=============================================================================
//===
//=== Services
//===
//=============================================================================

func (a *tradestation) GetRootSymbols(filter string) ([]*adapter.RootSymbol,error) {
	//--- Category=FU   (Category=Futures)
	//--- $top=1000     (returns first 1000 results)

	apiUrl := a.apiUrl + UrlSymbolsSuggest +"/"+ filter +"?$filter=Category%20eq%20%27FU%27&$top=1000"

	var res []RootFound
	err := a.doGet(apiUrl, &res)
	if err != nil {
		return nil, err
	}

	var rootsMap = make(map[string]*adapter.RootSymbol)

	for _,rf := range res {
		r,ok := rootsMap[rf.Root]
		if !ok {
			r = &adapter.RootSymbol{
				Code      : rf.Root,
				Instrument: rf.Description,
				Country   : rf.Country,
				Currency  : rf.Currency,
				Exchange  : rf.Exchange,
				PointValue: rf.PointValue,
			}

			rootsMap[r.Code] = r
		}
	}

	return slices.Collect(maps.Values(rootsMap)), nil
}

//=============================================================================

func (a *tradestation) GetRootSymbol(root string) (*adapter.RootSymbol,error) {
	apiUrl := a.apiUrl + UrlMarketDataSymbols +"/"+ buildInstrumentList(root)

	var res SymbolDetailsResponse
	err := a.doGet(apiUrl, &res)
	if err != nil {
		return nil, err
	}

	for _,sd := range res.Symbols {
		if sd.AssetType == "FUTURE" {
			rs := &adapter.RootSymbol{
				Code      : sd.Root,
				Instrument: sd.Description,
				Exchange  : sd.Exchange,
				PointValue: toFloat64(sd.PriceFormat.PointValue),
				Increment : toFloat64(sd.PriceFormat.Increment),
				Country   : sd.Country,
				Currency  : sd.Currency,
			}

			return rs,nil
		}
	}

	return nil, req.NewNotFoundError("Root symbol not found: %v", root)
}

//=============================================================================

func (a *tradestation) GetInstruments(root string) ([]*adapter.Instrument,error) {
	//--- C=FU     (Category=Futures)
	//--- Exp=true (shows all synbols, expired or not)
	//--- R=xxx    (root symbol)

	apiUrl := a.apiUrl + UrlSymbolsSearch +"/C=FU&Exp=true&R="+root

	var res []SymbolFound
	err := a.doGet(apiUrl, &res)
	if err != nil {
		return nil, req.NewServiceUnavailableError("Cannot get instruments: %v", err)
	}

	var instruments []*adapter.Instrument

	for _,sf := range res {
		if sf.Category == "Future" {
			expDate,err := convertExpirationDate(sf.ExpirationDate)
			if err != nil {
				return nil, req.NewServerErrorByError(err)
			}

			i := adapter.Instrument{
				Name          : sf.Name,
				Description   : sf.Description,
				Exchange      : sf.Exchange,
				Country       : sf.Country,
				Root          : sf.Root,
				ExpirationDate: expDate,
				PointValue    : sf.PointValue,
				MinMove       : sf.MinMove,
				Continuous    : sf.Name[0:1]=="@",
				Month         : extractMonth(sf.Name),
			}

			//--- Tradestation reports very future contracts (like 2036) without an expiration date.
			//--- We have to skip those because as time passes the expiration date should be set

			hasExpDate := i.ExpirationDate != nil || i.Continuous

			//--- Before 2008, contracts seem not very good in quality. Let's skip them

			isRecent := i.ExpirationDate != nil && i.ExpirationDate.Year() >= 2006

			if  hasExpDate && isRecent {
				instruments = append(instruments,&i)
			}
		}
	}

	return instruments, nil
}

//=============================================================================

func (a *tradestation) GetPriceBars(symbol string, date datatype.IntDate) (*adapter.PriceBars,error) {
	//--- Last time set to 23:59:50 (and not 59) as it seems that Tradestation somethimes returns 1 extra bar
	query := "unit=minute&interval=1&firstdate="+ date.String() +"T00%3A00%3A00Z&lastdate="+ date.String() +"T23%3A59%3A00Z"
	apiUrl := a.apiUrl + UrlMarketDataBarcharts +"/"+ symbol +"?"+ query

	priceBars := adapter.PriceBars{
		Symbol: symbol,
		Date  : int(date),
	}

	var res BarchartsResponse
	rres,err := a.doGetWithResponse(apiUrl)
	if err != nil {
		return nil, err
	}

	//--- Handle special cases

	if rres.StatusCode == http.StatusNotFound {
		priceBars.NoData = true
		return &priceBars, nil
	}

	if rres.StatusCode == http.StatusGatewayTimeout {
		priceBars.Timeout = true
		return &priceBars, nil
	}

	//--- Read response

	err = req.BuildResponse(rres, err, &res)
	if err != nil {
		return nil, err
	}

	for _,bar := range res.Bars {
		pb := adapter.PriceBar{
			TimeStamp   : time.UnixMilli(bar.Epoch),
			High        : toFloat64(bar.High),
			Low         : toFloat64(bar.Low),
			Open        : toFloat64(bar.Open),
			Close       : toFloat64(bar.Close),
			UpVolume    : bar.UpVolume,
			DownVolume  : bar.DownVolume,
			UpTicks     : bar.UpTicks,
			DownTicks   : bar.DownTicks,
			OpenInterest: toInt(bar.OpenInterest),
		}

		priceBars.Bars = append(priceBars.Bars, &pb)
	}

	return &priceBars, nil
}

//=============================================================================

func (a *tradestation) GetAccounts() ([]*adapter.Account,error) {
	apiUrl := a.apiUrl + UrlBrokerageAccounts

	var res AccountsResponse
	err := a.doGet(apiUrl, &res)
	if err != nil {
		return nil, err
	}

	accounts := convertAccounts(&res)

	for _,acc := range accounts {
		apiUrl = a.apiUrl + UrlBrokerageAccounts +"/"+ acc.Code +"/balances"
		var bres BalancesResponse
		err = a.doGet(apiUrl, &bres)
		if err != nil {
			return nil, err
		}

		if len(bres.Balances) != 1 {
			return nil, errors.New(fmt.Sprintf("Incorrect number of balances returned: %d",len(bres.Balances)))
		}

		b := bres.Balances[0]
		acc.CashBalance          = toFloat64(b.CashBalance)
		acc.Equity               = toFloat64(b.Equity)
		acc.RealizedProfitLoss   = toFloat64(b.BalanceDetail.RealizedProfitLoss)
		acc.UnrealizedProfitLoss = toFloat64(b.BalanceDetail.UnrealizedProfitLoss)
		acc.OpenOrderMargin      = toFloat64(b.BalanceDetail.OpenOrderMargin)
		acc.InitialMargin        = toFloat64(b.BalanceDetail.InitialMargin)
		acc.MaintenanceMargin    = toFloat64(b.BalanceDetail.MaintenanceMargin)
	}

	return accounts,nil
}

//=============================================================================

func (a *tradestation) GetOrders() (any,error) {
	return nil, nil
}

//=============================================================================

func (a *tradestation) GetPositions() (any,error) {
	return nil, nil
}

//=============================================================================

func (a *tradestation) TestService(path,param string) (string,error) {
	slog.Info("TestService: Testing service", "path", path, "param", param)
	url := a.apiUrl + path
	if param != "" {
		url = url + "?" + param
	}

	rq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		slog.Error("Error creating a GET request", "error", err.Error())
		return "",err
	}

	rq.Header.Set("Authorization", "Bearer "+ a.accessToken)
	rq.Header.Set("Content-Type", "application/json")

	res, err := a.client.Do(rq)
	if err != nil {
		return "",err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		slog.Error("TestService: Error reading response", "path", path, "param", param, "error", err.Error())
		return "",err
	}

	return string(body), nil
}

//=============================================================================
//===
//=== Private functions
//===
//=============================================================================

func (a *tradestation) doGet(url string, output any) error {
	res, err := a.doGetWithResponse(url)
	return req.BuildResponse(res, err, &output)
}

//=============================================================================

func (a *tradestation) doGetWithResponse(url string) (*http.Response, error) {
	rq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		slog.Error("Error creating a GET request", "error", err.Error())
		return nil,err
	}

	rq.Header.Set("Authorization", "Bearer "+ a.accessToken)
	rq.Header.Set("Content-Type", "application/json")

	return a.client.Do(rq)
}

//=============================================================================

func (a *tradestation) doPost(url string, params any, output any) error {
	body, err := json.Marshal(&params)
	if err != nil {
		slog.Error("Error marshalling POST parameter", "error", err.Error())
		return err
	}

	reader := bytes.NewReader(body)

	rq, err := http.NewRequest("POST", url, reader)
	if err != nil {
		slog.Error("Error creating a POST request", "error", err.Error())
		return err
	}

	rq.Header.Set("Authorization", "Bearer "+ a.accessToken)
	rq.Header.Set("Content-Type", "application/json")

	res, err := a.client.Do(rq)
	return req.BuildResponse(res, err, &output)
}

//=============================================================================

func convertAccounts(ar *AccountsResponse) []*adapter.Account {
	var list []*adapter.Account

	for _,acc := range ar.Accounts {
		if acc.AccountType == "Futures" && acc.Status == "Active" {
			var aa adapter.Account
			aa.Code         = acc.AccountID
			aa.Type         = adapter.AccountTypeFutures
			aa.CurrencyCode = acc.Currency

			list = append(list, &aa)
		}
	}

	return list
}

//=============================================================================

func toFloat64(value string) float64 {
	val, err := strconv.ParseFloat(value, 64)
	if err != nil {
		slog.Warn("Tradestation: Error converting value to float64", "value", value)
		return 0
	}

	return val
}

//=============================================================================

func toInt(value string) int {
	val, err := strconv.Atoi(value)
	if err != nil {
		slog.Warn("Tradestation: Error converting value to int", "value", value)
		return 0
	}

	return val
}

//=============================================================================

func convertExpirationDate(date string) (*time.Time,error) {
	fIdx := strings.Index(date, "(")
	lIdx := strings.Index(date, ")")

	if fIdx == -1 || lIdx == -1 {
		return nil,errors.New("Invalid expiration date : "+ date)
	}

	date = date[fIdx+1 : lIdx]
	ts,err := strconv.ParseInt(date, 10, 64)
	if err != nil {
		return nil,errors.New("Invalid expiration date : "+ date)
	}

	if ts < 0 {
		return nil,nil
	}

	res  := time.UnixMilli(ts)
	return &res, nil
}

//=============================================================================

func buildInstrumentList(root string) string {
	year := strconv.Itoa(time.Now().Year() % 100)

	return "@"+root+
			","+root +"F"+year+
			","+root +"G"+year+
			","+root +"H"+year+
			","+root +"J"+year+
			","+root +"K"+year+
			","+root +"M"+year+
			","+root +"N"+year+
			","+root +"Q"+year+
			","+root +"U"+year+
			","+root +"V"+year+
			","+root +"X"+year+
			","+root +"Z"+year
}

//=============================================================================

func extractMonth(symbol string) string {
	size := len(symbol)
	if size <4 {
		return ""
	}

	year := symbol[size-2 : size]
	_,err := strconv.Atoi(year)
	if err != nil {
		return ""
	}

	//--- Ok, last 2 digits are a number (the year). Let's get the month

	return symbol[size-3 : size-2]
}

//=============================================================================
