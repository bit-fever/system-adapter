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

//=============================================================================

const (
	UrlBrokerageAccounts  = "/v3/brokerage/accounts"
	UrlMarketDataSymbols  = "/v3/marketdata/symbols"
	UrlMarketDataBarcharts= "/v3/marketdata/barcharts"
	UrlSymbolsSearch      = "/v2/data/symbols/search"
	UrlSymbolsSuggest     = "/v2/data/symbols/suggest"
)

//=============================================================================
//=== Service: /v3/brokerage/accounts
//=============================================================================

type AccountsResponse struct {
	Accounts []Account
}

//=============================================================================

type Account struct {
	AccountID     string
	Currency      string
	Status        string
	AccountType   string
	AccountDetail AccountDetail
}

//=============================================================================

type AccountDetail struct {
	IsStockLocateEligible      bool
	EnrolledInRegTProgram      bool
	RequiresBuyingPowerWarning bool
	DayTradingQualified        bool
	OptionApprovalLevel        int
	PatternDayTrader           bool
}

//=============================================================================
//=== Service: /v3/brokerage/accounts/XXX/balances
//=============================================================================

type BalancesResponse struct {
	Balances []Balance
	Errors   []interface{}
}

//=============================================================================

type Balance struct {
	AccountID        string
	AccountType      string
	CashBalance      string
	BuyingPower      string
	Equity           string
	MarketValue      string
	TodaysProfitLoss string
	UnclearedDeposit string
	BalanceDetail    BalanceDetail
	CurrencyDetails  []CurrencyDetail
	Commission       string
}

//=============================================================================

type BalanceDetail struct {
	DayTradeExcess           string
	RealizedProfitLoss       string
	UnrealizedProfitLoss     string
	DayTradeOpenOrderMargin  string
	OpenOrderMargin          string
	DayTradeMargin           string
	InitialMargin            string
	MaintenanceMargin        string
	TradeEquity              string
	SecurityOnDeposit        string
	TodayRealTimeTradeEquity string
}

//=============================================================================

type CurrencyDetail struct {
	Currency              string
	Commission            string
	CashBalance           string
	RealizedProfitLoss    string
	UnrealizedProfitLoss  string
	InitialMargin         string
	MaintenanceMargin     string
	AccountConversionRate string
}

//=============================================================================
//=== Service: /v2/data/symbols/search/XXX
//=============================================================================

type SymbolFound struct {
	Name            string
	Description     string
	Exchange        string
	ExchangeID      int
	Category        string
	Country         string
	Root            string
	OptionType      string
	FutureType      string
	ExpirationDate  string
	ExpirationType  string
	StrikePrice     int
	Currency        string
	PointValue      int
	MinMove         float64
	DisplayType     int
	Error           interface{}
}

//=============================================================================
//=== Service: /v2/data/symbols/suggest/XXX
//=============================================================================

type RootFound struct {
	Country        string
	Currency       string
	Description    string
	Exchange       string
	ExpirationDate string
	LotSize        int
	MinMove        float64
	Name           string
	PointValue     float64
	Root           string
	StrikePrice    int
}

//=============================================================================
//=== Service: /v3/marketdata/symbols/XXX
//=============================================================================

type SymbolDetailsResponse struct {
	Symbols []SymbolDetails
	Errors  []interface{}
}

//=============================================================================

type SymbolDetails struct {
	AssetType      string
	Country        string
	Currency       string
	Description    string
	Exchange       string
	FutureType     string
	Symbol         string
	Root           string
	Underlying     string
	PriceFormat    PriceFormat
	QuantityFormat QuantityFormat
}

//=============================================================================

type PriceFormat struct {
	Format         string
	Decimals       string
	IncrementStyle string
	Increment      string
	PointValue     string
}

//=============================================================================

type QuantityFormat struct {
	Format               string
	Decimals             string
	IncrementStyle       string
	Increment            string
	MinimumTradeQuantity string
}

//=============================================================================
//=== Service: /v3/marketdata/barcharts/XXX
//=============================================================================

type BarchartsResponse struct {
	Bars    []Bar
	Error   string
	Message string
}

//=============================================================================

type Bar struct {
	TimeStamp    string
	Epoch        int64
	High         string
	Low          string
	Open         string
	Close        string
	UpVolume     int
	DownVolume   int
	UpTicks      int
	DownTicks    int
	OpenInterest string
}

//=============================================================================
