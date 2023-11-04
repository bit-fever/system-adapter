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
	"github.com/bit-fever/core/auth"
	"github.com/bit-fever/core/auth/roles"
	"github.com/bit-fever/core/req"
	"github.com/bit-fever/system-adapter/pkg/adapter"
	"github.com/bit-fever/system-adapter/pkg/adapter/interactive"
	"github.com/bit-fever/system-adapter/pkg/adapter/local"
	"github.com/google/uuid"
	"sync"
)

//=============================================================================

var adapters map[string]adapter.Adapter
var infos    []*adapter.AdapterInfo

var connections = struct {
	sync.RWMutex
	m map[string]*adapter.ConnectionContext
}{m: make(map[string]*adapter.ConnectionContext)}

//=============================================================================
//===
//=== Public methods
//===
//=============================================================================

func GetSystems() *[]*adapter.AdapterInfo {
	return &infos
}

//=============================================================================

func Connect(us *auth.UserSession, params *ConnectionParams) (*ConnectionResponse, error) {
	connections.Lock()
	defer connections.Unlock()

	ad,ok := adapters[params.SystemCode]
	if !ok {
		return nil, req.NewRequestError("System not found: %v", params.SystemCode)
	}

	ctx := &adapter.ConnectionContext{
		Code: uuid.New().String(),
		Username: us.Username,
		Config: params.Config,
		Adapter: ad,
	}

	err := ad.Connect(ctx)
	connections.m[ctx.Code] = ctx

	return &ConnectionResponse{
		Code: ctx.Code,
		Error: messageFor(err),
	}, nil
}

//=============================================================================

func Disconnect(us *auth.UserSession, code string) (*ConnectionResponse, error) {
	connections.Lock()
	defer connections.Unlock()

	ctx, ok := connections.m[code]
	if !ok {
		return nil, req.NewRequestError("Connection not found: %v", code)
	}

	if (ctx.Username != us.Username) || ( ! us.IsUserInRole(roles.Admin_Service)) {
		return nil, req.NewAccessError("Connection not owned by user: %v", code)
	}

	err := ctx.Adapter.Disconnect(ctx)
	delete(connections.m, code)

	return &ConnectionResponse{
		Code: code,
		Error: messageFor(err),
	}, nil
}

//=============================================================================
//===
//=== Private methods
//===
//=============================================================================

func init() {
	register(interactive.NewAdapter())
	register(local      .NewAdapter())
}

//=============================================================================

func register(a adapter.Adapter) {
	info := a.GetInfo()
	adapters[info.Code] = a
	infos = append(infos, info)
}

//=============================================================================

func messageFor(e error) string {
	if e != nil {
		return e.Error()
	}

	return ""
}

//=============================================================================
