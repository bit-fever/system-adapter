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
)

//=============================================================================

type ParamType string

const ParamTypeString   ParamType = "string"
const ParamTypePassword ParamType = "password"
const ParamTypeBool     ParamType = "bool"
const ParamTypeInt      ParamType = "int"

//=============================================================================

type Param struct {
	Name      string       `json:"name"`
	Type      ParamType    `json:"type"`      // string|int|bool|group|password
	DefValue  string       `json:"defValue"`
	Nullable  bool         `json:"nullable"`
	MinValue  int          `json:"minValue"`
	MaxValue  int          `json:"maxValue"`
	GroupName string       `json:"groupName"` // links this param to a group whose type is group
}

//-----------------------------------------------------------------------------

func (p *Param) Validate(values map[string]any) error {
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
	Code                 string   `json:"code"`
	Name                 string   `json:"name"`
	SupportsData         bool     `json:"supportsData"`
	SupportsBroker       bool     `json:"supportsBroker"`
	SupportsMultipleData bool     `json:"supportsMultipleData"`
	SupportsInventory    bool     `json:"supportsInventory"`
	Params               []*Param `json:"params"`
}

//=============================================================================

type Adapter interface {
	GetInfo()    *Info
	GetAuthUrl() string
	Clone        (config map[string]any)  Adapter
	Connect      (ctx *ConnectionContext) *ConnectionResult
	Disconnect   (ctx *ConnectionContext) error

	IsWebLoginCompleted(httpCode int, path string) bool
	InitFromWebLogin(reqHeader *http.Header, resCookies []*http.Cookie) error
}

//=============================================================================

type ConnectionContext struct {
	InstanceCode string
	Username     string
	Host         string
	Adapter      Adapter
}

//=============================================================================

type ConnectionInfo struct {
	InstanceCode string `json:"instanceCode"`
	Username     string `json:"username"`
	SystemCode   string `json:"systemCode"`
	SystemName   string `json:"systemName"`
}

//=============================================================================

const ConnectionStatusConnected    = "connected"
const ConnectionStatusDisconnected = "disconnected"
const ConnectionStatusError        = "error"
const ConnectionStatusOpenUrl      = "open-url"

//-----------------------------------------------------------------------------

type ConnectionResult struct {
	InstanceCode string `json:"instanceCode"`
	Status       string `json:"status"`
	Message      string `json:"message"`
}

//=============================================================================
