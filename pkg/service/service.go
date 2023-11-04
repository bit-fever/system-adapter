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
	"github.com/bit-fever/system-adapter/pkg/model/config"
	"github.com/gin-gonic/gin"
	"log"
)

//=============================================================================

func Init(router *gin.Engine, cfg *config.Config) {

	ctrl, err := auth.NewOidcController(cfg.Authentication.Authority, req.GetClient("bf"))
	if err != nil {
		log.Fatal(err)
	}

	router.GET   ("/api/system/v1/systems",         ctrl.Secure(getSystems,     roles.Admin_User))
//	router.GET   ("/api/system/v1/connections",     ctrl.Secure(getConnections, roles.Admin_User))
	router.POST  ("/api/system/v1/connections",     ctrl.Secure(connect,        roles.Admin_User))
	router.DELETE("/api/system/v1/connections/:id", ctrl.Secure(disconnect,     roles.Admin_User))
}

//=============================================================================

