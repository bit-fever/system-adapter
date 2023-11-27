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

type AdapterParam struct {
	Name      string       `json:"name"`
	Type      ParamType    `json:"type"`      // string|int|bool|group
	Label     string       `json:"label"`
	Nullable  bool         `json:"nullable"`
	MinValue  int          `json:"minValue"`
	MaxValue  int          `json:"maxValue"`
	Tooltip   string       `json:"tooltip"`
	GroupName string       `json:"groupName"` // links this param to a group whose type is group
}

//=============================================================================

type AdapterInfo struct {
	Code                  string         `json:"code"`
	Name                  string         `json:"name"`
	SupportsFeed          bool           `json:"supportsFeed"`
	SupportsBroker        bool           `json:"supportsBroker"`
	SupportsMultipleFeeds bool           `json:"supportsMultipleFeeds"`
	SupportsInventory     bool           `json:"supportsInventory"`
	Params                []AdapterParam `json:"params"`
}

//=============================================================================

type Adapter interface {
	GetInfo() *AdapterInfo
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
