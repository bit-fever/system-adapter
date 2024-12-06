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
	"github.com/bit-fever/system-adapter/pkg/adapter"
	"net/http"
)

//=============================================================================

const ParamAuthUrl  = "authUrl"
const ParamApiUrl   = "apiUrl"

//=============================================================================

var params = []*adapter.Param{
	{
		Name     : ParamAuthUrl,
		Type     : adapter.ParamTypeString,
		DefValue : "https://www.interactivebrokers.co.uk/sso/Login",
		Nullable : false,
		MinValue : 0,
		MaxValue : 64,
		GroupName: "",
	},
	{
		Name     : ParamApiUrl,
		Type     : adapter.ParamTypeString,
		DefValue : "https://api.ibkr.com",
		Nullable : false,
		MinValue : 0,
		MaxValue : 64,
		GroupName: "",
	},
}

//-----------------------------------------------------------------------------

var info = adapter.Info{
	Code                : "IBKR",
	Name                : "Interactive Brokers",
	Params              : params,
	SupportsData        : true,
	SupportsBroker      : true,
	SupportsMultipleData: false,
	SupportsInventory   : true,
}

//=============================================================================

type Params struct {
	AuthUrl  string
	ApiUrl   string
}

//=============================================================================

type ib struct {
	params *Params
	client *http.Client
	header *http.Header
}

//=============================================================================
//===
//=== IBKR REST API structures
//===
//=============================================================================

type Validate struct {
	UserId      int      `json:"USER_ID"`
	UserName    string   `json:"USER_NAME"`
	Result      bool     `json:"RESULT"`
	AuthTime    int64    `json:"AUTH_TIME"`
	IsFreeTrial bool     `json:"IS_FREE_TRIAL"`
	Credential  string   `json:"CREDENTIAL"`
	Ip          string   `json:"IP"`
	Expires     int      `json:"EXPIRES"`
}

//=============================================================================

type Orders struct {
	Orders   []*Order `json:"orders"`
	Snapshot bool     `json:"snapshot"`
}

//=============================================================================

type Order struct {
	AccountId          string  `json:"acct"`
	ContractIdEx       string  `json:"conidex"`
	ContractId         int     `json:"conid"`
	OrderId            int     `json:"orderId"`
	CashCurrency       string  `json:"cashCcy"`
	SizeAndFills       string  `json:"sizeAndFills"`
	OrderDesc          string  `json:"orderDesc"`
	Description1       string  `json:"description1"`
	Ticker             string  `json:"ticker"`
	SecurityType       string  `json:"secType"`
	ListingExchange    string  `json:"listingExchange"`
	RemainingQuantity  float32 `json:"remainingQuantity"`
	FilledQuantity     float32 `json:"filledQuantity"`
	CompanyName        string  `json:"companyName"`
	Status             string  `json:"status"`            // Filled
	OrderCcpStatus     string  `json:"order_ccp_status"`  // Filled
	OrigOrderType      string  `json:"origOrderType"`     // MARKET
	SupportsTaxOpt     string  `json:"supportsTaxOpt"`
	LastExecutionTime  string  `json:"lastExecutionTime"` // Format is YYMMDDHHmmss in UTC
	OrderType          string  `json:"orderType"`         // Market
	OrderRef           string  `json:"order_ref"`
	TimeInForce        string  `json:"timeInForce"`       // GTC
	AveragePrice       string  `json:"avgPrice"`
	LastExecutionTimeR int64   `json:"lastExecutionTime_r"`
	Side               string  `json:"side"`              // SELL

	Account          string  `json:"account"`
	TotalSize        float32 `json:"totalSize"`
}

//=============================================================================
