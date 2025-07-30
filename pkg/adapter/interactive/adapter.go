//=============================================================================
/*
Copyright © 2024 Andrea Carboni andrea.carboni71@gmail.com

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

package interactive

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"github.com/bit-fever/core/req"
	"github.com/bit-fever/system-adapter/pkg/adapter"
	"log/slog"
	"net/http"
	"time"
)

//=============================================================================

func NewAdapter() adapter.Adapter {
	return &ib{}
}

//=============================================================================

func (a *ib) GetInfo() *adapter.Info {
	return &info
}

//=============================================================================

func (a *ib) GetAuthUrl() string {
	return a.configParams.AuthUrl
}

//=============================================================================

func (a *ib) Clone(configParams map[string]any, connectParams map[string]any) adapter.Adapter {
	b := *a
	b.configParams = retrieveParams(configParams)
	return &b
}

//=============================================================================

func (a *ib) Connect(ctx *adapter.ConnectionContext) (adapter.ConnectionResult,error) {
	if a.configParams.NoAuth {
		//TODO: we should check if the connection actually works...
		//---   connection to the gateway

		return adapter.ConnectionResultConnected,nil
	}

	return adapter.ConnectionResultProxyUrl,nil
}

//=============================================================================

func (a *ib) Disconnect(ctx *adapter.ConnectionContext) error {
	return nil
}

//=============================================================================

func (a *ib) IsWebLoginCompleted(httpCode int, path string) bool {
	return httpCode == http.StatusFound && path == "/sso/Dispatcher"
}
//=============================================================================

func (a *ib) InitFromWebLogin(reqHeader *http.Header, resCookies []*http.Cookie) error {
	header, err := buildHttpHeader(reqHeader, resCookies)
	if err != nil {
		return err
	}

	a.header = header
	a.client = &http.Client{
		Timeout: time.Minute * 3,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
			},
		},
	}

	res, err := a.ssoValidate()

	if err != nil {
		return err
	}

	if !res.Result {
		return errors.New("session is invalid")
	}

//	orders,err := a.getAccountOrders()
//	pnl,err := a.getAccountProfitAndLoss()
//	tic,err := a.tickle()
//	fmt.Println("RES:"+tic.Session)
	return nil
}

//=============================================================================
//===
//=== Private functions
//===
//=============================================================================

func retrieveParams(values map[string]any) *Params {
	return &Params{
		AuthUrl: values[ParamAuthUrl] .(string),
		ApiUrl : values[ParamApiUrl]  .(string),
	}
}

//=============================================================================

func buildHttpHeader(reqHeader *http.Header, resCookies []*http.Cookie) (*http.Header, error) {
	userId, err := findUserId(resCookies)
	if err != nil {
		return nil, err
	}

	cookies := reqHeader.Get("Cookie")
	cookies += "; "+ userId.Name+"="+ userId.Value

	h := make(http.Header)
	h.Set("Accept",         "*/*")
	h.Set("Cache-Control",  "no-cache")
	h.Set("Pragma",         "no-cache")
	h.Set("Sec-Fetch-Dest", "empty")
	h.Set("Sec-Fetch-Mode", "cors")

	h.Set("Cookie",          cookies)
	h.Set("Accept-Encoding", reqHeader.Get("Accept-Encoding"))
	h.Set("Accept-Language", reqHeader.Get("Accept-Language"))

	return &h, nil
}

//=============================================================================

func findUserId(cookies []*http.Cookie) (*http.Cookie, error){
	for _, cookie := range cookies {
		if cookie.Name == "USERID" {
			return cookie, nil
		}
	}

	return nil, errors.New("cookie USERID was not found")
}

//=============================================================================

func (a *ib) doGet(url string, output any) error {
	rq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		slog.Error("Error creating a GET request", "error", err.Error())
		return err
	}

	rq.Header = *a.header
	res, err := a.client.Do(rq)
	return req.BuildResponse(res, err, &output)
}

//=============================================================================

func (a *ib) doPost(url string, params any, output any) error {
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
	rq.Header = *a.header
	rq.Header.Set("Content-Type", "Application/json")

	res, err := a.client.Do(rq)
	return req.BuildResponse(res, err, &output)
}

//=============================================================================
//===
//=== IBKR services
//===
//=============================================================================

func (a *ib) ssoValidate() (*Validate, error) {
	apiUrl := a.configParams.ApiUrl +"/v1/api/sso/validate"
	var res Validate
	err := a.doGet(apiUrl, &res)

	return &res, err
}

//=============================================================================

func (a *ib) getAccountOrders() (*OrdersResponse, error) {
	apiUrl := a.configParams.ApiUrl +"/v1/api/iserver/account/orders?force=true"
	var res OrdersResponse
	err := a.doGet(apiUrl, &res)

	return &res, err
}

//=============================================================================

func (a *ib) getAccountProfitAndLoss() (*AccountPnLResponse, error) {
	apiUrl := a.configParams.ApiUrl +"/v1/api/iserver/account/pnl/partitioned"
	var res AccountPnLResponse
	err := a.doGet(apiUrl, &res)

	return &res, err
}

//=============================================================================

func (a *ib) tickle() (*TickleResponse, error) {
	apiUrl := a.configParams.ApiUrl +"/v1/api/tickle"
	var res TickleResponse
	err := a.doPost(apiUrl, "{}", &res)

	return &res, err
}

//=============================================================================
