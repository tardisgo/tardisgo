// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build haxe

package math

import "github.com/tardisgo/tardisgo/haxe/hx"

// Copysign returns a value with the magnitude
// of x and the sign of y.
func Copysign(x, y float64) float64 {
	// Go libary version:
	//const sign = 1 << 63
	//return Float64frombits(Float64bits(x)&^sign | Float64bits(y)&sign)

	// below adapted from GopherJS see that project for copyright etc
	// original:
	// if (x < 0 || 1/x == negInf) != (y < 0 || 1/y == negInf) {
	//	return -x
	// }
	// return x

	if (x < 0 || (1/x < 0 && !hx.CallBool("", "Math.isFinite", 1, 1/x))) !=
		(y < 0 || (1/y < 0 && !hx.CallBool("", "Math.isFinite", 1, 1/y))) {
		return -x
	}
	return x
}
