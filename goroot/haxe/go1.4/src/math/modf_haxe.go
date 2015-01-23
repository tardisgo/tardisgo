// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build haxe

package math

import "github.com/tardisgo/tardisgo/haxe/hx"

// Modf returns integer and fractional floating-point numbers
// that sum to f.  Both values have the same sign as f.
//
// Special cases are:
//	Modf(±Inf) = ±Inf, NaN
//	Modf(NaN) = NaN, NaN
func Modf(f float64) (int float64, frac float64) {
	// approach inspired by GopherJS, see that project for copyright etc
	//if f == posInf || f == negInf {
	if !hx.CallBool("", "Math.isFinite", 1, f) {
		return f, hx.GetFloat("", "Math.NaN")
	}
	frac = Mod(f, 1)
	return f - frac, frac
}

func modf(f float64) (int float64, frac float64) {
	if f < 1 {
		if f < 0 {
			int, frac = Modf(-f)
			return -int, -frac
		}
		return 0, f
	}

	x := Float64bits(f)
	e := uint(x>>shift)&mask - bias

	// Keep the top 12+e bits, the integer part; clear the rest.
	if e < 64-12 {
		x &^= 1<<(64-12-e) - 1
	}
	int = Float64frombits(x)
	frac = f - int
	return
}
