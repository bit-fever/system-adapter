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

package tokenrefresh

import (
	"github.com/bit-fever/core/msg"
	"github.com/bit-fever/system-adapter/pkg/adapter"
	"github.com/bit-fever/system-adapter/pkg/app"
	"github.com/bit-fever/system-adapter/pkg/business"
	"log/slog"
	"time"
)

//=============================================================================

func InitRefresh(cfg *app.Config) *time.Ticker {
	ticker := time.NewTicker(10 * time.Second)

	go func() {
		time.Sleep(10 * time.Second)
		run()

		for range ticker.C {
			run()
		}
	}()

	return ticker
}

//=============================================================================

func run() {
	list := business.GetConnectionsToRefresh()

	for _, ctx := range list {
		err := ctx.RefreshToken()
		if err != nil {
			slog.Error("TokenRefresher: Cannot refresh token. Disconnecting", "username", ctx.Username, "connection", ctx.ConnectionCode, "error", err.Error())
			err = sendConnectionChangeMessage(ctx)
			if err != nil {
				slog.Error("TokenRefresher:  Could not publish the disconnection message (!)", "username", ctx.Username, "connection", ctx.ConnectionCode, "error", err.Error())
			}
		} else {
			slog.Info("TokenRefresher: Refreshed token complete", "username", ctx.Username, "connection", ctx.ConnectionCode)
		}
	}
}

//=============================================================================

func sendConnectionChangeMessage(ctx *adapter.ConnectionContext) error {
	ccm := business.ConnectionChangeSystemMessage{
		Username      : ctx.Username,
		ConnectionCode: ctx.ConnectionCode,
		SystemCode    : ctx.GetAdapterInfo().Code,
		Status        : ctx.GetStatus(),
	}

	return msg.SendMessage(msg.ExSystem, msg.SourceConnection, msg.TypeChange, &ccm)
}

//=============================================================================
