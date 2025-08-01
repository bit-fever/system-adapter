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

package local

import (
	"github.com/bit-fever/system-adapter/pkg/adapter"
	"net/http"
)

//=============================================================================

var configParams  []*adapter.ParamDef
var connectParams []*adapter.ParamDef

//-----------------------------------------------------------------------------

var info = adapter.Info{
	Code                : "LOCAL",
	Name                : "Local system",
	ConfigParams        : configParams,
	ConnectParams       : connectParams,
	SupportsData        : true,
	SupportsBroker      : true,
	SupportsMultipleData: true,
	SupportsInventory   : false,
}

//=============================================================================

func NewAdapter() adapter.Adapter {
	return &local{}
}

//=============================================================================

type local struct {
}

//=============================================================================

func (a *local) GetInfo() *adapter.Info {
	return &info
}

//=============================================================================

func (a *local) GetAuthUrl() string {
	return ""
}

//=============================================================================

func (a *local) Clone(configParams map[string]any, connectParams map[string]any) adapter.Adapter {
	b := *a
	return &b
}

//=============================================================================

func (a *local) Connect(ctx *adapter.ConnectionContext) (adapter.ConnectionResult,error) {
	return adapter.ConnectionResultConnected,nil
}

//=============================================================================

func (a *local) Disconnect(ctx *adapter.ConnectionContext) error {
	return nil
}

//=============================================================================

func (a *local) IsWebLoginCompleted(httpCode int, path string) bool {
	return true
}

//=============================================================================

func (a *local) InitFromWebLogin(reqHeader *http.Header, resCookies []*http.Cookie) error {
	return nil
}

//=============================================================================
