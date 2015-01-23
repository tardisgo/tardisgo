// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build haxe

package math

import "github.com/tardisgo/tardisgo/haxe/hx"

// Abs returns the absolute value of x.
//
// Special cases are:
//	Abs(±Inf) = +Inf
//	Abs(NaN) = NaN
func Abs(x float64) float64 {
	//switch x {
	//case hx.GetFloat("", "Math.NaN"):
	//	return x
	//case hx.GetFloat("", "Math.POSITIVE_INFINITY"),
	//	hx.GetFloat("", "Math.NEGATIVE_INFINITY"):
	//	return hx.GetFloat("", "Math.POSITIVE_INFINITY")
	//default:
	return hx.CallFloat("", "Math.abs", 1, x)
	//}
}

func abs(x float64) float64 {
	switch {
	case x < 0:
		return -x
	case x == 0:
		return 0 // return correctly abs(-0)
	}
	return x
}
