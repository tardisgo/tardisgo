// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build haxe

package math

import "github.com/tardisgo/tardisgo/haxe/hx"

// Ldexp is the inverse of Frexp.
// It returns frac × 2**exp.
//
// Special cases are:
//	Ldexp(±0, exp) = ±0
//	Ldexp(±Inf, exp) = ±Inf
//	Ldexp(NaN, exp) = NaN
func Ldexp(frac float64, exp int) float64 {

	// this inspired by the GopherJS project, see there for copyright etc
	if frac == 0 {
		return frac
	}
	if exp >= 1024 {
		return frac * hx.CallFloat("", "Math.pow", 2, 2, 1023) * hx.CallFloat("", "Math.pow", 2, 2, exp-1023)
	}
	if exp <= -1024 {
		return frac * hx.CallFloat("", "Math.pow", 2, 2, -1023) * hx.CallFloat("", "Math.pow", 2, 2, exp+1023)
	}
	return frac * hx.CallFloat("", "Math.pow", 2, 2, exp)
}

func ldexp(frac float64, exp int) float64 {

	// special cases
	switch {
	case frac == 0:
		return frac // correctly return -0
	case IsInf(frac, 0) || IsNaN(frac):
		return frac
	}
	frac, e := normalize(frac)
	exp += e
	x := Float64bits(frac)
	exp += int(x>>shift)&mask - bias
	if exp < -1074 {
		return Copysign(0, frac) // underflow
	}
	if exp > 1023 { // overflow
		if frac < 0 {
			return Inf(-1)
		}
		return Inf(1)
	}
	var m float64 = 1
	if exp < -1022 { // denormal
		exp += 52
		m = 1.0 / (1 << 52) // 2**-52
	}
	x &^= mask << shift
	x |= uint64(exp+bias) << shift
	return m * Float64frombits(x)
}
