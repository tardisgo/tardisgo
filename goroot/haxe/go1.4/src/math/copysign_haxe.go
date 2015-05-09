// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build haxe

package math

// Copysign returns a value with the magnitude
// of x and the sign of y.
func Copysign(x, y float64) float64 {
	// Go libary version:
	//const sign = 1 << 63
	//return Float64frombits(Float64bits(x)&^sign | Float64bits(y)&sign)

	const sign = 1 << 63
	xBits := Float64bits(x)
	yBits := Float64bits(y)
	dataBits := xBits &^ sign
	signBits := yBits & sign
	ret := Float64frombits(dataBits | signBits)
	/*
		if x == 0 && y < 0 {
			retBits := Float64bits(ret)
			println("DEBUG math.Copysign -0 x y xBits yBits dataBits signBits ret retBits=",
				x, y, xBits, yBits, dataBits, signBits, ret, retBits)
		}
	*/
	return ret

	/*
		rx := 1 / x
		ry := 1 / y

		if (x < 0 || (rx < 0 && !hx.CallBool("", "Math.isFinite", 1, rx))) !=
			(y < 0 || (ry < 0 && !hx.CallBool("", "Math.isFinite", 1, ry))) {
			return -x
		}
		return x
	*/
}
