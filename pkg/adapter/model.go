//=============================================================================
/*
Copyright © 2023 Andrea Carboni andrea.carboni71@gmail.com

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

package adapter

import (
	"errors"
	"log/slog"
	"net/http"
	"reflect"
	"strconv"
	"time"
)

//=============================================================================
//=== Common connection parameters

const (
	ParamUsername = "username"
	ParamPassword = "password"
	ParamTwoFACode= "twoFACode"
)

//=============================================================================

type ParamType string

const (
	ParamTypeString   ParamType = "string"
	ParamTypePassword ParamType = "password"
	ParamTypeBool     ParamType = "bool"
	ParamTypeInt      ParamType = "int"
)

//=============================================================================

type ParamDef struct {
	Name      string      `json:"name"`
	Type      ParamType   `json:"type"`      // string|int|bool|group|password
	DefValue  string      `json:"defValue"`
	Nullable  bool        `json:"nullable"`
	MinValue  int         `json:"minValue"`
	MaxValue  int         `json:"maxValue"`
	GroupName string      `json:"groupName"` // links this param to a group whose type is group
}

//-----------------------------------------------------------------------------

func (p *ParamDef) Validate(values map[string]any) error {
	value,ok := values[p.Name]

	if !ok {
		//--- Check default value

		if p.DefValue != "" {
			switch p.Type {
				case ParamTypeBool:
					if p.DefValue != "true" && p.DefValue != "false" {
						return errors.New("invalid default value for a boolean parameter : "+ p.Name)
					}
					break;

				case ParamTypeInt:
					v, err := strconv.Atoi(p.DefValue)
					if err != nil {
						return errors.New("invalid value for an integer parameter : "+ p.Name)
					}
					if v<p.MinValue || v>p.MaxValue {
						return errors.New("invalid range for this integer parameter : "+ p.Name)
					}
					break;
			}
			return nil
		}

		if !p.Nullable {
			return errors.New("missing mandatory value for parameter : "+ p.Name)
		}
	} else {
		//--- Check provided value

		t := reflect.TypeOf(value)
		slog.Info("Param Type is", "type", t.Name())
		switch t.Name() {
			case "string":
				if p.Type == ParamTypeString || p.Type == ParamTypePassword {
					return nil
				}
				break;

			case "bool":
				if p.Type == ParamTypeBool {
					return nil
				}
				break;

			case "int":
				if p.Type == ParamTypeInt {
					return nil
				}

			default:
				return errors.New("unknown parameter type : "+ p.Name)
		}

		return errors.New("invalid parameter value : "+ p.Name)
	}

	return nil
}

//=============================================================================

type Info struct {
	Code                 string       `json:"code"`
	Name                 string       `json:"name"`
	SupportsData         bool         `json:"supportsData"`
	SupportsBroker       bool         `json:"supportsBroker"`
	SupportsMultipleData bool         `json:"supportsMultipleData"`
	SupportsInventory    bool         `json:"supportsInventory"`
	ConfigParams         []*ParamDef  `json:"configParams"`
	ConnectParams        []*ParamDef  `json:"connectParams"`
}

//=============================================================================

type ConnectionResult int

const (
	ConnectionResultError     = -1
	ConnectionResultConnected =  0
	ConnectionResultOpenUrl   =  1
	ConnectionResultProxyUrl  =  2
)

//-----------------------------------------------------------------------------

type Adapter interface {
	GetInfo() *Info
	GetAuthUrl() string
	Clone(configParams map[string]any, connectParams map[string]any)  Adapter
	Connect(ctx *ConnectionContext) (ConnectionResult, error)
	Disconnect(ctx *ConnectionContext) error
	IsWebLoginCompleted(httpCode int, path string) bool
	InitFromWebLogin(reqHeader *http.Header, resCookies []*http.Cookie) error
	GetTokenExpSeconds() int
	RefreshToken() error

	//--- Services

	GetRootSymbols(filter string) ([]*RootSymbol,error)
	GetRootSymbol(root string) (*RootSymbol,error)
	GetInstruments(root string) ([]*Instrument,error)
	GetPrices() (any,error)
	GetAccounts() ([]*Account,error)
	GetOrders() (any,error)
	GetPositions() (any,error)
	TestService(path,param string) (string,error)
}

//=============================================================================
//===
//=== API model
//===
//=============================================================================

type AccountType int
const (
	AccountTypeFutures = 0
	AccountTypeCrypto  = 1
)

//-----------------------------------------------------------------------------

type Account struct {
	Code                 string      `json:"code"`
	Type                 AccountType `json:"type"`
	CurrencyCode         string      `json:"currencyCode"`
	CashBalance          float64     `json:"cashBalance"`
	Equity               float64     `json:"equity"`
	RealizedProfitLoss   float64     `json:"realizedProfitLoss"`
	UnrealizedProfitLoss float64     `json:"unrealizedProfitLoss"`
	OpenOrderMargin      float64     `json:"openOrderMargin"`
	InitialMargin        float64     `json:"initialMargin"`
	MaintenanceMargin    float64     `json:"maintenanceMargin"`
}

//=============================================================================

type Instrument struct {
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	Exchange        string     `json:"exchange"`
	Country         string     `json:"country"`
	Root            string     `json:"root"`
	ExpirationDate  *time.Time `json:"expirationDate"`
	PointValue      int        `json:"pointValue"`
	MinMove         int        `json:"minMove"`
	Continuous      bool       `json:"continuous"`
}

//=============================================================================

type RootSymbol struct {
	Code        string  `json:"code"`
	Instrument  string  `json:"instrument"`
	Exchange    string  `json:"exchange"`
	PointValue  float64 `json:"pointValue"`
	Increment   float64 `json:"increment"`
	Country     string  `json:"country"`
	Currency    string  `json:"currency"`
}

//=============================================================================
