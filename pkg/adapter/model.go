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

//=============================================================================

type ParamType string

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
	GetInfo() *Info
	Connect   (ctx *ConnectionContext) error
	Disconnect(ctx *ConnectionContext) error
}

//=============================================================================

type ConnectionContext struct {
	Code     string
	Username string
	Config   map[string]string
	Adapter  Adapter
}

//=============================================================================

type ConnectionInfo struct {
	Code       string `json:"code"`
	Username   string `json:"username"`
	SystemCode string `json:"systemCode"`
	SystemName string `json:"systemName"`
}

//=============================================================================
