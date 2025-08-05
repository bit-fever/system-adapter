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

package service

import (
	"github.com/bit-fever/core/auth"
	"github.com/bit-fever/core/auth/roles"
	"github.com/bit-fever/core/req"
	"github.com/bit-fever/system-adapter/pkg/app"
	"github.com/gin-gonic/gin"
	"log/slog"
)

//=============================================================================

func Init(router *gin.Engine, cfg *app.Config, logger *slog.Logger) {

	ctrl := auth.NewOidcController(cfg.Authentication.Authority, req.GetClient("bf"), logger, cfg)

	router.GET   ("/api/system/v1/adapters",                   ctrl.Secure(getAdapters,    roles.Admin_User))
	router.GET   ("/api/system/v1/adapters/:code",             ctrl.Secure(getAdapter,     roles.Admin_User))
	router.GET   ("/api/system/v1/connections",                ctrl.Secure(getConnections, roles.Admin_User))
	router.PUT   ("/api/system/v1/connections/:code",          ctrl.Secure(connect,        roles.Admin_User))
	router.DELETE("/api/system/v1/connections/:code",          ctrl.Secure(disconnect,     roles.Admin_User))

	//--- Adapter services

	router.GET   ("/api/system/v1/connections/:code/roots",                    ctrl.Secure(getRootSymbols, roles.Admin_User))
	router.GET   ("/api/system/v1/connections/:code/roots/:root",              ctrl.Secure(getRootSymbol,  roles.Admin_User))
	router.GET   ("/api/system/v1/connections/:code/roots/:root/instruments",  ctrl.Secure(getInstruments, roles.Admin_User))
	router.GET   ("/api/system/v1/connections/:code/prices",                   ctrl.Secure(getPrices,      roles.Admin_User))
	router.GET   ("/api/system/v1/connections/:code/accounts",                 ctrl.Secure(getAccounts,    roles.Admin_User))
	router.GET   ("/api/system/v1/connections/:code/orders",                   ctrl.Secure(getOrders,      roles.Admin_User))
	router.GET   ("/api/system/v1/connections/:code/positions",                ctrl.Secure(getPositions,   roles.Admin_User))
	router.POST  ("/api/system/v1/connections/:code/test",                     ctrl.Secure(testAdapter,    roles.Admin_User))

	//TODO: To review
	//router.GET   ("/api/system/v1/connections/:code/login",   webLogin)
	//router.Use   (proxyLoginRequests)
}

//=============================================================================

