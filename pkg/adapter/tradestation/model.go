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

package tradestation

import (
	"github.com/bit-fever/system-adapter/pkg/adapter"
	"net/http"
)

//=============================================================================
//--- Config parameters

const (
	ParamClientId    = "clientId"
	ParamLiveAccount = "liveAccount"
)

//=============================================================================

const (
	LiveAPI = "https://api.tradestation.com"
	DemoAPI = "https://sim-api.tradestation.com"

	LoginPageUrl       = "https://my.tradestation.com/api/auth/login?returnTo=%2F"
	LoginPostUrl       = "https://signin.tradestation.com/usernamepassword/login"
	LoginCallbackUrl   = "https://signin.tradestation.com/login/callback"
	LoginTwoFAUrl      = "https://signin.tradestation.com/u/mfa-otp-challenge"
	LoginTwoFAPath     = "/u/mfa-otp-challenge"
	LoginDashboardPath = "/dashboard"
	RefreshTokenUrl    = "https://my.tradestation.com/api/auth/token"
)

//=============================================================================

var configParams = []*adapter.ParamDef {
	{
		Name     : ParamClientId,
		Type     : adapter.ParamTypeString,
		DefValue : "TDoILDTVZjp0k5J0xsWbXS1yEUncnj08",
		Nullable : false,
		MinValue : 0,
		MaxValue : 64,
		GroupName: "",
	},
	{
		Name     : ParamLiveAccount,
		Type     : adapter.ParamTypeBool,
		DefValue : "false",
		Nullable : false,
		MinValue : 0,
		MaxValue : 0,
		GroupName: "",
	},
}

//-----------------------------------------------------------------------------

var connectParams = []*adapter.ParamDef {
	{
		Name     : adapter.ParamUsername,
		Type     : adapter.ParamTypeString,
		DefValue : "",
		Nullable : false,
		MinValue : 0,
		MaxValue : 64,
		GroupName: "",
	},
	{
		Name     : adapter.ParamPassword,
		Type     : adapter.ParamTypePassword,
		DefValue : "",
		Nullable : false,
		MinValue : 0,
		MaxValue : 64,
		GroupName: "",
	},
	{
		Name     : adapter.ParamTwoFACode,
		Type     : adapter.ParamTypeString,
		DefValue : "",
		Nullable : false,
		MinValue : 0,
		MaxValue : 64,
		GroupName: "",
	},
}

//-----------------------------------------------------------------------------

var info = adapter.Info{
	Code                : "TS",
	Name                : "Tradestation",
	ConfigParams        : configParams,
	ConnectParams       : connectParams,
	SupportsData        : true,
	SupportsBroker      : true,
	SupportsMultipleData: false,
	SupportsInventory   : true,
}

//=============================================================================

type ConfigParams struct {
	ClientId    string
	LiveAccount bool
}

//=============================================================================

type ConnectParams struct {
	Username  string
	Password  string
	TwoFACode string
}

//=============================================================================

type tradestation struct {
	configParams  *ConfigParams
	connectParams *ConnectParams
	client        *http.Client
	header        *http.Header
	accessToken    string
	refreshToken   string
	clientId       string
	apiUrl         string
}

//=============================================================================
//===
//=== Tradestation REST API structures
//===
//=============================================================================


//=============================================================================
