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
	"bytes"
	"crypto/tls"
	"github.com/bit-fever/core/req"
	"github.com/bit-fever/system-adapter/pkg/adapter"
	"github.com/bit-fever/system-adapter/pkg/business"
	"github.com/gin-gonic/gin"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
)

//=============================================================================

const InstanceCode = "InstanceCode"

//=============================================================================

func proxyLoginRequests(c *gin.Context) {
	cookie,err := c.Request.Cookie("InstanceCode")
	if err != nil {
		slog.Warn("Called without cookie", "url", c.Request.URL.String())
		return
	}

	code := cookie.Value

	ctx := business.GetConnectionContextByInstanceCode(code)
	if ctx == nil {
		req.ReturnError(c, req.NewBadRequestError("Connection context not found : "+ code))
		return
	}

	authUrl := ctx.Adapter.GetAuthUrl()
	target, err := url.Parse(authUrl)
	if err != nil {
		req.ReturnError(c, req.NewBadRequestError("Bad authentication url : "+ authUrl))
		return
	}

	proxy := buildProxy(c, target, c.Request.URL.Path, ctx)
	proxy.ServeHTTP(c.Writer, c.Request)
}

//=============================================================================

func buildProxy(c *gin.Context, target *url.URL, forwardPath string, ctx *adapter.ConnectionContext) *httputil.ReverseProxy {
	proxy := &httputil.ReverseProxy{}

	proxy.Rewrite = func (r *httputil.ProxyRequest) {
		out := r.Out

		sb := strings.Builder{}
		sb.WriteString("\n====== REQUEST : "+ out.Method +" "+out.URL.String()+" --> "+target.Host+forwardPath+" =========\n")
		sb.WriteString(dumpHeader(&out.Header, "Original header:"))

		out.URL.Scheme = target.Scheme
		out.URL.Host   = target.Host
		out.URL.Path   = forwardPath
		out.Host       = target.Host

		remapHeader(&out.Header, c.Request.Host, target, false)

		sb.WriteString(dumpHeader(&out.Header, "Modified header:"))
		sb.WriteString(remapCookies(r.Out.Cookies(), &r.Out.Header, c.Request.Host, target.Host, false))
		sb.WriteString("==============================================================\n")
		slog.Info("Proxy request", "data", sb.String())
	}

	proxy.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	proxy.ModifyResponse = func(res *http.Response) error {
		r := res.Request
		sb := strings.Builder{}
		sb.WriteString("\n====== RESPONSE : "+ r.Method +" "+r.URL.String()+" --> "+target.Host+forwardPath+" =========\n")
		sb.WriteString(dumpHeader(&res.Header, "Response header"))
		remapHeader(&res.Header, c.Request.Host, target, true)
		sb.WriteString(dumpHeader(&res.Header, "Modified header:"))
		sb.WriteString(remapCookies(res.Cookies(), &res.Header, c.Request.Host, target.Host, true))
		sb.WriteString("==============================================================\n")
		slog.Info("Proxy response", "data", sb.String())

		if ctx.Adapter.IsWebLoginCompleted(res.StatusCode, res.Request.URL.Path) {
			err := ctx.Adapter.InitFromWebLogin(&res.Request.Header, res.Cookies())
			defer res.Body.Close()

			message := htmlfy("Success", "This page can be closed")

			if err != nil {
				message = htmlfy("Authentication failed", "Cause: "+ err.Error())
				slog.Error("Adapter authentication failed", "adapter", ctx.Adapter.GetInfo().Name, "error", err.Error())
			}

			res.Body = io.NopCloser(bytes.NewReader([]byte(message)))
			res.StatusCode    = http.StatusOK
			res.ContentLength = int64(len(message))
			res.Header.Set("Content-Length", strconv.Itoa(len(message)))
		}
		return nil
	}

	return proxy
}

//=============================================================================

func remapHeader(header *http.Header, source string, target *url.URL, isResponse bool) {
	//--- Request

	if !isResponse {
		if header.Get("Origin") != "" {
			header.Set("Origin",  target.Scheme+ "://" +target.Host)
		}

		if header.Get("Referer") != "" {
			header.Set("Referer", target.String())
		}
		header.Set("Host", target.Host)
	}

	//--- Response

	if isResponse {
		if header.Get("Origin") != "" {
			header.Set("Origin",  target.Scheme+ "://" +source)
		}

		header.Set("Host", source)
		header.Del("Link")
	}
}

//=============================================================================

func remapCookies(cookies []*http.Cookie, header *http.Header, source, destin string, isResponse bool) string {
	sb := strings.Builder{}
	sb.WriteString("Cookies:\n")

	host := destin

	if isResponse {
		host = source
	}

	domain := extractDomain(host)

	for _, cookie := range cookies {
		sb.WriteString("   "+ cookie.Name+": "+ cookie.Value)

		if cookie.Domain != "" {
			sb.WriteString(" (remapped: "+ cookie.Domain +" --> "+ domain +")")
			cookie.Domain = domain
			header.Set("Set-Cookie", cookie.String())
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

//=============================================================================

func extractDomain(host string) string {
	index := strings.Index(host, ":")
	if index == -1 {
		return host
	}

	return host[0:index]
}

//=============================================================================

func dumpHeader(header *http.Header, message string) string {
	sb := strings.Builder{}
	sb.WriteString(message +"\n")

	for x,y := range *header {
		sb.WriteString("   "+ x + " : ")
		for _,s := range y {
			sb.WriteString(s +" | ")
		}
		sb.WriteString("\n")
	}

	print(sb.String())
	return sb.String()
}

//=============================================================================

func htmlfy(title, message string) string {
	var sb strings.Builder
	sb.WriteString("<html><body><h1>")
	sb.WriteString(title)
	sb.WriteString("</h1><br/>")
	sb.WriteString(message)
	sb.WriteString("</body></html>")

	return sb.String()
}

//=============================================================================
