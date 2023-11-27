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

package main

import (
	"github.com/bit-fever/core/boot"
	"github.com/bit-fever/core/req"
	"github.com/bit-fever/system-adapter/pkg/app"
	"github.com/bit-fever/system-adapter/pkg/service"
	"log/slog"
)

//=============================================================================

const component = "system-adapter"

//=============================================================================

func main() {
	cfg := &app.Config{}
	boot.ReadConfig(component, cfg)
	logger := boot.InitLogger(component, &cfg.Application)
	engine := boot.InitEngine(logger,    &cfg.Application)
	initClients()
	service.Init(engine, cfg, logger)
	boot.RunHttpServer(engine, &cfg.Application)
}

//=============================================================================

func initClients() {
	slog.Info("Initializing clients...")
	req.AddClient("bf", "ca.crt", "server.crt", "server.key")
}

//=============================================================================
