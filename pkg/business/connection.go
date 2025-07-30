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
	"github.com/bit-fever/core/msg"
	"github.com/bit-fever/core/req"
	"github.com/bit-fever/system-adapter/pkg/adapter"
	"sync"
)

//=============================================================================

var userConnections = struct {
	sync.RWMutex
	m map[string]*UserConnections
}{m: make(map[string]*UserConnections)}

//=============================================================================
//===
//=== Public methods
//===
//=============================================================================

func GetConnections(c *auth.Context, filter map[string]any, offset int, limit int) *[]*ConnectionInfo {
	userConnections.Lock()
	defer userConnections.Unlock()

	us := c.Session
	uc,found := userConnections.m[us.Username]

	var list []*ConnectionInfo

	if found {
		for _, ctx := range uc.contexts {
			ci := ConnectionInfo{
				Username      : ctx.Username,
				ConnectionCode: ctx.ConnectionCode,
				SystemCode    : ctx.Adapter.GetInfo().Code,
				SystemName    : ctx.Adapter.GetInfo().Name,
			}
			list = append(list, &ci)
		}
	}

	return &list
}

//=============================================================================

func GetConnectionContextByInstanceCode(instanceCode string) *adapter.ConnectionContext {
	userConnections.Lock()
	defer userConnections.Unlock()

	//TODO
	//for _,uc := range userConnections.m {
	//	for _,ctx := range uc.contexts {
	//		if ctx.InstanceCode == instanceCode {
	//			return ctx
	//		}
	//	}
	//}

	return nil
}

//=============================================================================

func Connect(c *auth.Context, connectionCode string, cs *ConnectionSpec) (*ConnectionResult, error) {
	user := c.Session.Username

	userConnections.Lock()
	defer userConnections.Unlock()
	uc,found := userConnections.m[user]

	//--- Add entry if it is the first time

	if !found {
		uc = NewUserConnections()
		userConnections.m[user] = uc
	}

	ctx,found := uc.contexts[connectionCode]
	if found {
		if ctx.Status == adapter.ContextStatusConnected {
			return &ConnectionResult{
				Status : ConnectionStatusConnected,
				Message: "Already connected",
			}, nil
		}

		if ctx.Status == adapter.ContextStatusConnecting {
			return &ConnectionResult{
				Status : ConnectionStatusConnecting,
				Message: "Still connecting",
			}, nil
		}
	}

	ad,ok := adapters[cs.SystemCode]
	if !ok {
		return nil, req.NewNotFoundError("System not found: %v", cs.SystemCode)
	}

	err := validateParameters(ad.GetInfo().ConfigParams, cs.ConfigParams)
	if err != nil {
		return nil, err
	}

	err = validateParameters(ad.GetInfo().ConnectParams, cs.ConnectParams)
	if err != nil {
		return nil, err
	}

	ctx = &adapter.ConnectionContext{
		ConnectionCode: connectionCode,
		Username      : c.Session.Username,
		Host          : c.Gin.Request.Host,
		Adapter       : ad.Clone(cs.ConfigParams, cs.ConnectParams),
		Status        : adapter.ContextStatusDisconnected,
	}

	//--- It is better to store again the context even if it is already there: the user could use the
	//--- same connection code but with a different adapter

	uc.contexts[connectionCode] = ctx

	res := &ConnectionResult{
		Status : ConnectionStatusError,
		Message: err.Error(),
		Action: ConnectionActionNone,
	}

	cr,err := ctx.Adapter.Connect(ctx)
	if err != nil {
		return res,nil
	}

	err = sendConnectionChangeMessage(c, ctx)
	if err != nil {
		return &ConnectionResult{
			Status : ConnectionStatusError,
			Message: err.Error(),
			Action: ConnectionActionNone,
		}, nil
	}

	switch cr {
		case adapter.ConnectionResultConnected:
			res.Status = ConnectionStatusConnected
			res.Action = ConnectionActionNone

		case adapter.ConnectionResultOpenUrl:
			res.Status = ConnectionStatusConnecting
			res.Action = ConnectionActionOpenUrl
			res.Message = ctx.Adapter.GetAuthUrl()

		//TODO: to review: hardcoded url
		case adapter.ConnectionResultProxyUrl:
			res.Status  = ConnectionStatusConnecting
			res.Action  = ConnectionActionOpenUrl
			res.Message = "https://bitfever-server:8449/api/system/v1/weblogin/"+ user +"/"+ connectionCode +"/login"
	}

	return res, nil
}

//=============================================================================

func Disconnect(c *auth.Context, connectionCode string) error {
	user := c.Session.Username

	userConnections.Lock()
	defer userConnections.Unlock()

	uc, ok := userConnections.m[user]
	if !ok {
		return req.NewNotFoundError("Connection not found for user: %v", user)
	}

	ctx, found := uc.contexts[connectionCode]
	if !found {
		return req.NewNotFoundError("Connection not found: %v", connectionCode)
	}

	if ctx.Status == adapter.ContextStatusDisconnected {
		return nil
	}

	err := sendConnectionChangeMessage(c, ctx)
	if err != nil {
		return req.NewServerErrorByError(err)
	}

	_ = ctx.Adapter.Disconnect(ctx)
	ctx.Status = adapter.ContextStatusDisconnected
	ctx.Host   = ""

	return nil
}

//=============================================================================
//===
//=== Private methods
//===
//=============================================================================

func validateParameters(params []*adapter.ParamDef, values map[string]any) error {
	for _, p := range params {
		err := p.Validate(values)
		if err != nil {
			return err
		}
	}

	return nil
}

//=============================================================================

func sendConnectionChangeMessage(c *auth.Context, ctx *adapter.ConnectionContext) error {
	ccm := ConnectionChangeSystemMessage{
		Username      : ctx.Username,
		ConnectionCode: ctx.ConnectionCode,
		SystemCode    : ctx.Adapter.GetInfo().Code,
		Status        : ctx.Status,
	}
	err := msg.SendMessage(msg.ExSystem, msg.SourceConnection, msg.TypeChange, &ccm)

	if err != nil {
		c.Log.Error("sendConnectionChangeMessage: Could not publish the change message", "error", err.Error())
		return err
	}

	return nil
}

//=============================================================================
