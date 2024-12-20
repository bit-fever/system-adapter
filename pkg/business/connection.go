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
	"github.com/bit-fever/core/req"
	"github.com/bit-fever/system-adapter/pkg/adapter"
	"github.com/google/uuid"
	"sync"
)

//=============================================================================

var connections = struct {
	sync.RWMutex
	m map[string]*adapter.ConnectionContext
}{m: make(map[string]*adapter.ConnectionContext)}

//=============================================================================
//===
//=== Public methods
//===
//=============================================================================

func GetConnections(c *auth.Context, filter map[string]any, offset int, limit int) *[]*adapter.ConnectionInfo {
	connections.Lock()
	defer connections.Unlock()

	us := c.Session

	var list []*adapter.ConnectionInfo

	for _, cc := range connections.m {
		if us.IsAdmin() || us.Username == cc.Username {
			ci := adapter.ConnectionInfo{
				InstanceCode: cc.InstanceCode,
				Username    : cc.Username,
				SystemCode  : cc.Adapter.GetInfo().Code,
				SystemName  : cc.Adapter.GetInfo().Name,
			}
			list = append(list, &ci)
		}
	}

	return &list
}

//=============================================================================

func GetConnectionContextByInstanceCode(instanceCode string) *adapter.ConnectionContext {
	connections.Lock()
	defer connections.Unlock()

	ctx, ok := connections.m[instanceCode]
	if !ok {
		return nil
	}

	return ctx
}

//=============================================================================

func Connect(c *auth.Context, cs *ConnectionSpec) (*adapter.ConnectionResult, error) {
	connections.Lock()
	defer connections.Unlock()

	ad,ok := adapters[cs.SystemCode]
	if !ok {
		return nil, req.NewNotFoundError("System not found: %v", cs.SystemCode)
	}

	err := validateParameters(ad.GetInfo().Params, cs.Config)
	if err != nil {
		return nil, err
	}

	ctx := &adapter.ConnectionContext{
		InstanceCode: uuid.New().String(),
		Username    : c.Session.Username,
		Host        : c.Gin.Request.Host,
		Adapter     : ad.Clone(cs.Config),
	}

	res := ctx.Adapter.Connect(ctx)

	switch res.Status {
		case adapter.ConnectionStatusConnected:
			connections.m[ctx.InstanceCode] = ctx

		case adapter.ConnectionStatusOpenUrl:
			link := "https://bitfever-server:8449/api/system/v1/connections/"+ctx.InstanceCode +"/login"
			res.Message = link
			connections.m[ctx.InstanceCode] = ctx
	}

	return res, nil
}

//=============================================================================

func Disconnect(c *auth.Context, code string) error {
	connections.Lock()
	defer connections.Unlock()

	ctx, ok := connections.m[code]
	if !ok {
		return req.NewNotFoundError("Connection not found: %v", code)
	}

	us := c.Session

	if ! us.IsAdmin() {
		if ctx.Username != us.Username {
			return req.NewForbiddenError("Connection not owned by user: %v", code)
		}
	}

	err := ctx.Adapter.Disconnect(ctx)
	delete(connections.m, code)

	return err
}

//=============================================================================
//===
//=== Private methods
//===
//=============================================================================

func validateParameters(params []*adapter.Param, values map[string]any) error {
	for _, p := range params {
		err := p.Validate(values)
		if err != nil {
			return err
		}
	}

	return nil
}

//=============================================================================
