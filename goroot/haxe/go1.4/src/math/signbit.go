// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math

import "github.com/tardisgo/tardisgo/haxe/hx"

// Signbit returns true if x is negative or negative zero.
func Signbit(x float64) bool {
	//return Float64bits(x)&(1<<63) != 0

	// below approach copyright GopherJS, see that project for Copyright etc
	//return x < 0 || 1/x == negInf

	return x < 0 || (1/x < 0 && !hx.CallBool("", "Math.isFinite", 1, 1/x))
}
