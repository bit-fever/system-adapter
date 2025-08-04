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
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/bit-fever/system-adapter/pkg/adapter"
	"golang.org/x/net/html"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

//=============================================================================
//===
//=== Model
//===
//=============================================================================

type LoginInfo struct {
	State               string
	Client              string
	Protocol            string
	Scope               string
	ResponseType        string
	RedirectUri         string
	Audience            string
	Nonce               string
	CodeChallengeMethod string
	CodeChallenge       string
	Auth0Config         *Auth0Config
}

//=============================================================================

type LoginRequest struct {
	Audience     string `json:"audience"`
	ClientId     string `json:"client_id"`
	Connection   string `json:"connection"`
	Nonce        string `json:"nonce"`
	Password     string `json:"password"`
	RedirectUri  string `json:"redirect_uri"`
	ResponseType string `json:"response_type"`
	Scope        string `json:"scope"`
	State        string `json:"state"`
	Tenant       string `json:"tenant"`
	Username     string `json:"username"`
	Csrf	     string `json:"_csrf"`
	IntState     string `json:"_intstate"`
}

//=============================================================================

type Auth0Config struct {
	ClientId     string `json:"clientID"`
	Auth0Domain  string `json:"auth0Domain"`
	Auth0Tenant  string `json:"auth0Tenant"`
	InternalOptions struct {
		Protocol string `json:"protocol"`
		Csrf     string `json:"_csrf"`
		Intstate string `json:"_intstate"`
	} `json:"internalOptions"`
}

//=============================================================================

type LoginResult struct {
	Wa      string `json:"wa"`
	Wresult string `json:"wresult"`
	Wctx    string `json:"wctx"`
}

//=============================================================================

type TokenRefreshResponse struct {
	AccessToken  string `json:"accessToken"`
	IdToken      string `json:"idToken"`
	Expiry       int    `json:"expiry"`
}

//=============================================================================
//===
//=== Methods
//===
//=============================================================================

func (a *tradestation) createLoginInfo() (*LoginInfo, error){
	rq, err := http.NewRequest("GET", LoginPageUrl, nil)

	if err != nil {
		slog.Error("createLoginInfo: Error creating a GET request", "error", err.Error())
		return nil, err
	}

	//--- Let's try to replicate a basic header sent by Chrome
	//--- If we don't do that. Tradestation will redirect to wwww.tradestation.com

	h := http.Header{}
	setupCommonHeader(&h)
	h.Add("Accept",            "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	h.Add("Priority",          "u=0, i")
	h.Add("Sec-Fetch-Dest",    "document")
	h.Add("Sec-Fetch-Mode",    "navigate")
	h.Add("Sec-Fetch-Site",    "none")
	h.Add("Sec-Fetch-User",    "?1")
	h.Add("Upgrade-Insecure-Requests","1")
	rq.Header = h

	res, err := a.client.Do(rq)

	defer res.Body.Close()
	doc, err := html.Parse(res.Body)
	if err != nil {
		slog.Error("createLoginInfo: Error reading from body", "error", err.Error())
		return nil, err
	}

	scripts       := extractScripts(doc, nil)
	encodedConfig := extractEncodedConfig(scripts)

	if encodedConfig == "" {
		return nil, errors.New("Can't extract config from node")
	}

	bytes, err := base64.StdEncoding.DecodeString(encodedConfig)
	if err != nil {
		return nil, errors.New("Can't decode config from base64 : "+encodedConfig)
	}

	var authConfig Auth0Config
	err = json.Unmarshal(bytes, &authConfig)
	if err != nil {
		slog.Error("Bad JSON response from server", "error", err.Error())
		return nil,err
	}

	//--- The original GET request gets many refirects. We setup some common headers for later use
	h = res.Header
	setupCommonHeader(&h)
	a.header = &h

	q := res.Request.URL.Query()

	return &LoginInfo{
		State              : q.Get("state"),
		Client             : q.Get("client"),
		Protocol           : q.Get("protocol"),
		Scope              : q.Get("scope"),
		ResponseType       : q.Get("response_type"),
		RedirectUri        : q.Get("redirect_uri"),
		Audience           : q.Get("audience"),
		Nonce              : q.Get("nonce"),
		CodeChallengeMethod: q.Get("code_challenge_method"),
		CodeChallenge      : q.Get("code_challenge"),
		Auth0Config        : &authConfig,
	}, nil
}

//=============================================================================

func (a *tradestation) login(info *LoginInfo) (*LoginResult, error) {
	lr := LoginRequest{
		Audience    : info.Audience,
		ClientId    : info.Client,
		Connection  : "auth0-api-connection",
		Nonce       : info.Nonce,
		Password    : a.connectParams.Password,
		RedirectUri : info.RedirectUri,
		ResponseType: info.ResponseType,
		Scope       : info.Scope,
		State       : info.State,
		Tenant      : info.Auth0Config.Auth0Tenant,
		Username    : a.connectParams.Username,
		Csrf        : info.Auth0Config.InternalOptions.Csrf,
		IntState    : info.Auth0Config.InternalOptions.Intstate,
	}

	body, err := json.Marshal(&lr)
	if err != nil {
		slog.Error("login: Error marshalling POST parameter", "error", err.Error())
		return nil,err
	}

	reader := bytes.NewReader(body)

	rq, err := http.NewRequest("POST", LoginPostUrl, reader)
	if err != nil {
		slog.Error("login: Error creating a POST request", "error", err.Error())
		return nil,err
	}

	rq.Header = *a.header
	setupHeader(&rq.Header)
	res, err := a.client.Do(rq)

	defer res.Body.Close()
	doc, err := html.Parse(res.Body)
	if err != nil {
		slog.Error("login: Error reading from body", "error", err.Error())
		return nil,err
	}

	lres := LoginResult{}
	extractLoginResult(doc, &lres)
	if lres.Wa == "" {
		slog.Error("login: Can't login to Tradestation", "result", toString(doc))
		return nil,errors.New("Can't login to Tradestation")
	}

	return &lres, nil
}

//=============================================================================

func (a *tradestation) callCallback(lr *LoginResult) (string,error) {
	var params = url.Values{}
	params.Set("wa"     , lr.Wa)
	params.Set("wresult", lr.Wresult)
	params.Set("wctx"   , lr.Wctx)
	payload := bytes.NewBufferString(params.Encode())

	rq, err := http.NewRequest("POST", LoginCallbackUrl, payload)
	if err != nil {
		slog.Error("callCallback: Error creating a POST request", "error", err.Error())
		return "",err
	}

	rq.Header = *a.header
	setupHeader(&rq.Header)
	rq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := a.client.Do(rq)

	defer res.Body.Close()
	doc, err := html.Parse(res.Body)
	if err != nil {
		slog.Error("callCallback: Error reading from body", "error", err.Error())
		return "",err
	}

	if res.Request.URL.Path != LoginTwoFAPath {
		slog.Error("callCallback: Didn't get the 2FA page", "result", toString(doc))
		return "",errors.New("Didn't get the 2FA page")
	}

	newState := res.Request.URL.Query().Get("state")
	return newState,nil
}

//=============================================================================

func (a *tradestation) submitTwoFACode(state string) error {
	var params = url.Values{}
	params.Set("state" , state)
	params.Set("code"  , a.connectParams.TwoFACode)
	params.Set("action", "default")
	payload := bytes.NewBufferString(params.Encode())

	rq, err := http.NewRequest("POST", LoginTwoFAUrl+"?state="+state, payload)
	if err != nil {
		slog.Error("submitTwoFACode: Error creating a POST request", "error", err.Error())
		return err
	}

	rq.Header = *a.header
	setupCommonHeader(&rq.Header)
	rq.Header.Set("Accept",       "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	rq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := a.client.Do(rq)

	defer res.Body.Close()
	doc, err := html.Parse(res.Body)
	if err != nil {
		slog.Error("submitTwoFACode: Error reading from body", "error", err.Error())
		return err
	}

	if res.Request.URL.Path != LoginDashboardPath {
		slog.Error("submitTwoFACode: Didn't get the dashboard path", "result", toString(doc))
		return errors.New("Didn't get the dashboard page")
	}

	a.accessToken  = res.Header.Get("X-Authorization")
	a.refreshToken = res.Header.Get("X-Id-Token")

	return nil
}

//=============================================================================

func (a *tradestation) testToken() error {
	accounts,err := a.GetAccounts()
	if err != nil {
		return err
	}

	if len(accounts.Accounts) == 0 {
		return errors.New("Futures account not found or not active")
	}

	return nil
}

//=============================================================================
//===
//=== Functions
//===
//=============================================================================

func retrieveConfigParams(values map[string]any) *ConfigParams {
	return &ConfigParams{
		ClientId    : values[ParamClientId]   .(string),
		LiveAccount : values[ParamLiveAccount].(bool),
	}
}

//=============================================================================

func retrieveConnectParams(values map[string]any) *ConnectParams {
	return &ConnectParams{
		Username  : values[adapter.ParamUsername] .(string),
		Password  : values[adapter.ParamPassword] .(string),
		TwoFACode : values[adapter.ParamTwoFACode].(string),
	}
}

//=============================================================================

func extractScripts(node *html.Node, list []*html.Node) []*html.Node {
	if node.Type == html.ElementNode && node.Data == "script" {
		return append(list, node)
	} else {
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			subList := extractScripts(c, nil)
			list = append(list, subList...)
		}

		return list
	}
}

//=============================================================================

func extractEncodedConfig(scripts []*html.Node) string {
	for _, node := range scripts {
		if node.Attr == nil {
			text := node.FirstChild.Data
			idx  := strings.Index(text, "'")
			size := len(text)
			return text[idx+1 : size-6]
		}
	}

	return ""
}

//=============================================================================

func extractLoginResult(node *html.Node, lr *LoginResult) {
	if node.Type == html.ElementNode && node.Data == "input" {
		extractLoginResultAttribute(node, lr)
	} else {
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			extractLoginResult(c, lr)
		}
	}
}

//=============================================================================

func extractLoginResultAttribute(node *html.Node, lr *LoginResult) {
	name  := getAttributeValue(node, "name")
	value := getAttributeValue(node, "value")

	switch name {
	case "wa"     : lr.Wa      = value
	case "wresult": lr.Wresult = value
	case "wctx"   : lr.Wctx    = value
	}
}

//=============================================================================

func getAttributeValue(node *html.Node, name string) string {
	for _, a := range node.Attr {
		if a.Key == name {
			return a.Val
		}
	}

	return ""
}

//=============================================================================

func setupHeader(h *http.Header) {
	h.Set("Accept",         "*/*")
	h.Add("Priority",       "u=0, i")
	h.Set("Origin",         "https://signin.tradestation.com")
	h.Add("Sec-Fetch-Dest", "empty")
	h.Add("Sec-Fetch-Mode", "cors")
	h.Add("Sec-Fetch-Site", "same-origin")
	h.Set("Content-Type",   "Application/json")
	h.Set("Auth0-Client",   "eyJuYW1lIjoiYXV0aDAuanMtdWxwIiwidmVyc2lvbiI6IjkuMTYuNCJ9")
}

//=============================================================================

func setupCommonHeader(h *http.Header) {
	h.Add("Accept-Encoding",   "gzip, deflate, br, zstd")
	h.Add("Accept-Language",   "en-US,en;q=0.9")
	h.Add("Cache-Control",     "no-cache")
	h.Add("Pragma",            "no-cache")
	h.Add("Sec-Ch-Ua",         "\"Not)A;Brand\";v=\"8\", \"Chromium\";v=\"138\", \"Google Chrome\";v=\"138\"")
	h.Add("Sec-Ch-Ua-mobile",  "?0")
	h.Add("Sec-Ch-Ua-platform","\"Linux\"")
	h.Add("User-Agent",        "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36")
}

//=============================================================================

func toString(node *html.Node) string {
	var b bytes.Buffer
	err  := html.Render(&b, node)
	if err != nil {
		return err.Error()
	}

	return b.String()
}

//=============================================================================
