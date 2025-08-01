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

package business

import (
	"github.com/bit-fever/system-adapter/pkg/adapter"
	"sync"
)

//=============================================================================

type ConnectionSpec struct {
	SystemCode     string         `json:"systemCode"     binding:"required"`
	ConfigParams   map[string]any `json:"configParams"   binding:"required"`
	ConnectParams  map[string]any `json:"connectParams"  binding:"required"`
}

//=============================================================================

type ConnectionInfo struct {
	Username       string `json:"username"`
	ConnectionCode string `json:"connectionCode"`
	SystemCode     string `json:"systemCode"`
	SystemName     string `json:"systemName"`
	Status         string `json:"status"`
}

//=============================================================================

type UserConnections struct {
	sync.RWMutex
	username string
	contexts map[string]*adapter.ConnectionContext
}

//-----------------------------------------------------------------------------

func NewUserConnections() *UserConnections {
	uc := &UserConnections{}
	uc.contexts = make(map[string]*adapter.ConnectionContext)

	return uc
}

//=============================================================================

type ConnectionChangeSystemMessage struct {
	Username       string                `json:"username"`
	ConnectionCode string                `json:"connectionCode"`
	SystemCode     string                `json:"systemCode"`
	Status         adapter.ContextStatus `json:"status"`
}

//=============================================================================

const (
	ConnectionStatusConnecting = "connecting"
	ConnectionStatusConnected  = "connected"
	ConnectionStatusError      = "error"

	ConnectionActionNone       = "none"
	ConnectionActionOpenUrl    = "open-url"
)

//-----------------------------------------------------------------------------

type ConnectionResult struct {
	Status  string `json:"status"`
	Action  string `json:"action"`
	Message string `json:"message"`
}

//=============================================================================
