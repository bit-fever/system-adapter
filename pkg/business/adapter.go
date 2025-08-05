//=============================================================================
/*
Copyright © 2024 Andrea Carboni andrea.carboni71@gmail.com

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
	"github.com/bit-fever/system-adapter/pkg/adapter"
	"github.com/bit-fever/system-adapter/pkg/adapter/local"
	"github.com/bit-fever/system-adapter/pkg/adapter/tradestation"
)

//=============================================================================

var adapters map[string]adapter.Adapter
var infos    []*adapter.Info

//=============================================================================
//===
//=== Init
//===
//=============================================================================

func init() {
	adapters = map[string]adapter.Adapter{}

	register(local       .NewAdapter())
	register(tradestation.NewAdapter())
//	register(interactive .NewAdapter())
}

//=============================================================================

func register(a adapter.Adapter) {
	info := a.GetInfo()
	adapters[info.Code] = a
	infos = append(infos, info)
}

//=============================================================================
//===
//=== Public methods
//===
//=============================================================================

func GetAdapters() *[]*adapter.Info {
	return &infos
}

//=============================================================================

func GetAdapter(code string) *adapter.Info {
	for _,a := range infos {
		if a.Code == code {
			return a
		}
	}

	return nil
}

//=============================================================================
