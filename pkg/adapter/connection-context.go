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

package adapter

import (
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/bit-fever/core/datatype"
)

//=============================================================================

const (
	RefreshRetries   = 5
	PriceBarsRetries = 5
)

//=============================================================================
//===
//=== ConnectionContext
//===
//=============================================================================

type ContextStatus int

const (
	ContextStatusDisconnected = 0
	ContextStatusConnecting   = 1
	ContextStatusConnected    = 2
)

//=============================================================================

type ConnectionContext struct {
	Username        string
	ConnectionCode  string
	Host            string
	ConnectedSince  time.Time
	//-------------------------------
	status          ContextStatus
	lastRefreshTime time.Time
	refreshRetries  int
	adapter         Adapter
	sync.RWMutex
}

//=============================================================================

func NewConnectionContext(username string, connectionCode string, host string, a Adapter, configParams, connectParams map[string]any) (*ConnectionContext,error) {
	err := validateParameters(a.GetInfo().ConfigParams, configParams)
	if err != nil {
		return nil, err
	}

	err = validateParameters(a.GetInfo().ConnectParams, connectParams)
	if err != nil {
		return nil, err
	}

	return &ConnectionContext{
		Username      : username,
		ConnectionCode: connectionCode,
		Host          : host,
		adapter       : a.Clone(configParams, connectParams),
		status        : ContextStatusDisconnected,
		refreshRetries: RefreshRetries,
	},nil
}

//=============================================================================

func (cc *ConnectionContext) GetAdapterInfo() *Info {
	return cc.adapter.GetInfo()
}

//=============================================================================

func (cc *ConnectionContext) GetAdapterAuthUrl() string {
	return cc.adapter.GetAuthUrl()
}

//=============================================================================

func (cc *ConnectionContext) GetStatus() ContextStatus {
	return cc.status
}

//=============================================================================

func (cc *ConnectionContext) IsConnected() bool {
	return cc.status == ContextStatusConnected
}

//=============================================================================

func (cc *ConnectionContext) IsConnecting() bool {
	return cc.status == ContextStatusConnecting
}

//=============================================================================

func (cc *ConnectionContext) IsDisconnected() bool {
	return cc.status == ContextStatusDisconnected
}

//=============================================================================

func (cc *ConnectionContext) Connect() (ConnectionResult, error) {
	cr,err := cc.adapter.Connect(cc)

	if err == nil {
		switch cr {
			case ConnectionResultConnected:
				cc.status          = ContextStatusConnected
				cc.ConnectedSince  = time.Now()
				cc.lastRefreshTime = cc.ConnectedSince

			case ConnectionResultOpenUrl:
				cc.status = ContextStatusConnecting

			case ConnectionResultProxyUrl:
				cc.status = ContextStatusConnecting
		}
	}

	return cr,err
}

//=============================================================================

func (cc *ConnectionContext) Disconnect() error {
	cc.status = ContextStatusDisconnected
	return cc.adapter.Disconnect(cc)
}

//=============================================================================

func (cc *ConnectionContext) NeedsRefresh() bool {
	now := time.Now()
	sec := cc.adapter.GetTokenExpSeconds()

	if sec == 0 || !cc.IsConnected(){
		return false
	}

	//--- We remove the last 2 mins to allow a few retries and to avoid to be borderline
	if sec >= 150 {
		sec -= 2*60
	}

	delta := int(now.Sub(cc.lastRefreshTime) / time.Second)

	return delta > sec
}

//=============================================================================

func (cc *ConnectionContext) IsWebLoginCompleted(httpCode int, path string) bool {
	return cc.adapter.IsWebLoginCompleted(httpCode, path)
}

//=============================================================================

func (cc *ConnectionContext) InitFromWebLogin(reqHeader *http.Header, resCookies []*http.Cookie) error {
	return cc.adapter.InitFromWebLogin(reqHeader, resCookies)
}

//=============================================================================
//===
//=== Services
//===
//=============================================================================

func (cc *ConnectionContext) RefreshToken() error {
	cc.Lock()
	defer cc.Unlock()

	err := cc.adapter.RefreshToken()

	if err == nil {
		cc.lastRefreshTime = time.Now()
		cc.refreshRetries  = RefreshRetries
	} else {
		cc.refreshRetries--
		if cc.refreshRetries > 0 {
			slog.Warn("RefreshToken: The adapter cannot refresh the token. Retrying...", "username", cc.Username, "connection", cc.ConnectionCode, "error", err.Error())
			return nil
		} else {
			cc.status = ContextStatusDisconnected
		}
	}

	return err
}

//=============================================================================

func (cc *ConnectionContext) GetRootSymbols(filter string) ([]*RootSymbol,error) {
	cc.RLock()
	defer cc.RUnlock()

	return cc.adapter.GetRootSymbols(filter)
}

//=============================================================================

func (cc *ConnectionContext) GetRootSymbol(root string) (*RootSymbol,error) {
	cc.RLock()
	defer cc.RUnlock()

	return cc.adapter.GetRootSymbol(root)
}

//=============================================================================

func (cc *ConnectionContext) GetInstruments(root string) ([]*Instrument,error) {
	cc.RLock()
	defer cc.RUnlock()

	return cc.adapter.GetInstruments(root)
}

//=============================================================================

func (cc *ConnectionContext) GetPriceBars(symbol string, date datatype.IntDate) (*PriceBars,error) {
	cc.RLock()
	defer cc.RUnlock()

	counter := 0

	for {
		pc,err := cc.adapter.GetPriceBars(symbol, date)

		if err != nil || !pc.Timeout {
			return pc,err
		}

		counter++
		slog.Warn("GetPriceBars: Got timeout from adapter", "adapter", cc.adapter.GetInfo().Name, "counter", counter)

		if counter == PriceBarsRetries {
			return nil,errors.New("Maximum number of retries exceeded: "+ cc.adapter.GetInfo().Name)
		}
	}
}

//=============================================================================

func (cc *ConnectionContext) GetAccounts() ([]*Account,error) {
	cc.RLock()
	defer cc.RUnlock()

	return cc.adapter.GetAccounts()
}

//=============================================================================

func (cc *ConnectionContext) GetOrders() (any,error) {
	cc.RLock()
	defer cc.RUnlock()

	return cc.adapter.GetOrders()
}

//=============================================================================

func (cc *ConnectionContext) GetPositions() (any,error) {
	cc.RLock()
	defer cc.RUnlock()

	return cc.adapter.GetPositions()
}

//=============================================================================

func (cc *ConnectionContext) TestAdapter(service, query string) (string,error) {
	cc.RLock()
	defer cc.RUnlock()

	return cc.adapter.TestService(service, query)
}

//=============================================================================
//===
//=== Functions
//===
//=============================================================================

func validateParameters(params []*ParamDef, values map[string]any) error {
	for _, p := range params {
		err := p.Validate(values)
		if err != nil {
			return err
		}
	}

	return nil
}

//=============================================================================
