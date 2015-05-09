// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build haxe

// Copyright 2014 Elliott Stoneham and The tardisgo Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Bitcast elements adapted from https://github.com/gopherjs/gopherjs/blob/master/bitcasts/bitcasts.go
/*
Copyright (c) 2013 Richard Musiol. All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

   * Redistributions of source code must retain the above copyright
notice, this list of conditions and the following disclaimer.
   * Redistributions in binary form must reproduce the above
copyright notice, this list of conditions and the following disclaimer
in the documentation and/or other materials provided with the
distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

// package haxegoruntime is code is always in the runtime. TODO consider how to slim it down...
// this file copied from Package math - provides implementation and documentation of math functions overloaded by TARDIS Go->Haxe transpiler
package haxegoruntime

import (
	"runtime"
	"unsafe"

	"github.com/tardisgo/tardisgo/haxe/hx"
)

const ( // from package math
	uvnan    = 0x7FF8000000000001
	uvinf    = 0x7FF0000000000000
	uvneginf = 0xFFF0000000000000

	MaxFloat32 = 3.40282346638528859811704183484516925440e+38   // 2**127 * (2**24 - 1) / 2**23
	MaxFloat64 = 1.797693134862315708145274237317043567981e+308 // 2**1023 * (2**53 - 1) / 2**52
)

const ( // from http://www.cprogramming.com/tutorial/floating_point/understanding_floating_point_representation.html
	uvinf32    = 0x7f800000
	uvneginf32 = 0xff800000
	uvnan32    = 0x7fc00000
)

var zero float64
var posInf = 1 / zero  //hx.GetFloat("", "Math.POSITIVE_INFINITY") // 1 / zero
var negInf = -1 / zero //hx.GetFloat("", "Math.NEGATIVE_INFINITY") //-1 / zero
var nan = 0 / zero     //hx.GetFloat("", "Math.NaN")                  //0 / zero
var minusZero = zero * -1

const uvminusnan = 0xFFF8000000000001

var MinusNan = Float64frombits(uvminusnan)

func init() { // to avoid DCE
	if false {
		_ = Float32bits(0)
		_ = Float32frombits(0)
		_ = Float64bits(0)
		_ = Float64frombits(0)
	}
}

// Code below adapted from https://github.com/gopherjs/gopherjs/blob/master/bitcasts/bitcasts.go

//commented out code from math package unsafe.go, which is overloaded with the code below
//import "unsafe"

// Float32bits returns the IEEE 754 binary representation of f.
//func Float32bits(f float32) uint32 { return *(*uint32)(unsafe.Pointer(&f)) }
func Float32bits(f float32) uint32 {
	// stop recursion - NOTE also must not compare two float32 in this fn !
	if InF32fb {
		panic("haxegoruntime.Float32bits() InF32fb already set")
	}
	InF32fb = true
	defer func() { InF32fb = false }()

	if hx.GetBool("", "Object.nativeFloats") {
		var t float32 = f
		return *(*uint32)(unsafe.Pointer(&t))
	}
	// TODO cpp/neko/js short-cut
	/*
		switch runtime.GOARCH {
		case "cpp", "neko", "js":
			return uint32(hx.CallInt("", "Force.Float32bits", 1, f))
		}
	*/

	// below logic from math.IsInf
	if float64(f) > float64(MaxFloat32) {
		return uvinf32
	}
	if float64(f) < float64(-MaxFloat32) {
		return uvneginf32
	}
	if float64(f) != float64(f) { // NaN
		if float64(f) < 0 { // -NaN ?
			return uvnan32 | uint32(1)<<31
		}
		return uvnan32
	}
	if float64(f) == 0 {
		if 1/float64(f) < float64(-MaxFloat32) { // dividing by -0 gives -Inf
			return uint32(1) << 31
		}
		return 0
	}

	s := uint32(0)
	if float64(f) < 0 { // must use float64 to avoid recusion on the comparison NOTE ditto below...
		s = 1 << 31
		f = -f
	}

	e := uint32(127 + 23)
	for float64(f) >= 1<<24 {
		f /= 2
		e++
		if e == uint32((1<<8)-1) {
			if float64(f) >= 1<<23 {
				f = float32(posInf)
			}
			break
		}
	}
	for float64(f) < float64(1<<23) {
		e--
		if e == 0 {
			break
		}
		f *= 2
	}

	//r := js.Global.Call("$mod", f, 2).Float()
	// below is code to simulate: r := mth.Mod(float64(f), 2)
	if float64(f) < hx.GetFloat("", "0") {
		panic("haxegoruntime.Float32bits")
	}
	t := float64(f) / 2
	r := float64(f) - (2 * t)
	// end simulate code
	if (r > hx.GetFloat("", "0.5") && r < hx.GetFloat("", "1")) || r >= hx.GetFloat("", "1.5") { // round to nearest even
		f++
	}

	return s | uint32(e)<<23 | (uint32(f) &^ (1 << 23))
}

var InF32fb bool // signal to haxegoruntime Force.toFloat32() to stop re-entry

// Float32frombits returns the floating point number corresponding
// to the IEEE 754 binary representation b.
// func Float32frombits(b uint32) float32 { return *(*float32)(unsafe.Pointer(&b)) }
func Float32frombits(b uint32) float32 {
	if InF32fb {
		panic("haxegoruntime.Float32frombits() InF32fb already set")
	}
	InF32fb = true
	defer func() { InF32fb = false }()
	if hx.GetBool("", "Object.nativeFloats") {
		var t uint32 = b
		return *(*float32)(unsafe.Pointer(&t))
	}
	// TODO cpp/neko/js short-cut
	/*
		switch runtime.GOARCH {
		case "cpp", "neko", "js":
			return float32(hx.CallFloat("", "Force.Float32frombits", 1, b))
		}
	*/

	// first handle the special cases
	switch b {
	case uvnan32:
		return float32(nan)
	case uvnan32 | 1<<31:
		return float32(nan * -1) // -NaN
	case uvinf32:
		return float32(posInf)
	case uvneginf32:
		return float32(negInf)
	case 0:
		return 0
	case 1 << 31:
		return float32(minusZero) // -0
	}

	s := float32(+1)
	if b&(1<<31) != 0 {
		s = -1
	}
	e := (b >> 23) & uint32((1<<8)-1)
	m := b & uint32((1<<23)-1)

	if e == uint32((1<<8)-1) {
		if m == 0 {
			return s / 0 // Inf
		}
		return float32(nan)
	}
	if e != 0 {
		m += 1 << 23
	}
	if e == 0 {
		e = 1
	}

	return float32(Ldexp(float64(m), int(e)-127-23)) * s
}

// Float64bits returns the IEEE 754 binary representation of f.
//func Float64bits(f float64) uint64 { return *(*uint64)(unsafe.Pointer(&f)) }
func Float64bits(f float64) uint64 {

	if hx.GetBool("", "Object.nativeFloats") {
		var t float64 = f
		return *(*uint64)(unsafe.Pointer(&t))
	}

	switch runtime.GOARCH {
	case "cs":
		return uint64(hx.Int64(hx.CallDynamic("", "Force.Float64bits", 1, f)))
		// TODO js/cpp/neko short-cut - using Force.f64byts
		//case "js", "cpp", "neko":
		//	return uint64(hx.Int64(hx.CallDynamic("", "Force.Float64bits", 1, f)))
	}

	// below from math.IsInf
	if f > MaxFloat64 {
		return uvinf
	}
	if f < -MaxFloat64 {
		return uvneginf
	}
	if f != f { // NaN
		if f < 0 { // -NaN ?
			return uvnan | uint64(1)<<63
		}
		//return 9221120237041090561
		return uvnan
	}
	if f == 0 {
		if 1/f < -MaxFloat64 { // dividing by -0 gives -Inf
			return uint64(1) << 63
		}
		return 0
	}

	s := uint64(0)
	if f < hx.GetFloat("", "0") {
		s = 1 << 63
		f = -f
	}

	e := uint32(1023 + 52)
	for f >= float64(int64(1<<53)) {
		f /= hx.GetFloat("", "2")
		e++
		if e == uint32((1<<11)-1) {
			break
		}
	}
	for f < float64(int64(1<<52)) {
		e--
		if e == 0 {
			break
		}
		f *= hx.GetFloat("", "2")
	}

	return s | uint64(e)<<52 | (uint64(f) &^ (1 << 52))
}

// Float64frombits returns the floating point number corresponding
// the IEEE 754 binary representation b.
//func Float64frombits(b uint64) float64 { return *(*float64)(unsafe.Pointer(&b)) }
func Float64frombits(b uint64) float64 {
	if hx.GetBool("", "Object.nativeFloats") {
		var t uint64 = b
		return *(*float64)(unsafe.Pointer(&t))
	}

	switch runtime.GOARCH {
	case "cs":
		return hx.CallFloat("", "Force.Float64frombits", 1, b)
		// TODO js/cpp/neko short-cut
		//case "js", "cpp", "neko":
		//	return hx.CallFloat("", "Force.Float64frombits", 1, b)
	}

	// first handle the special cases
	switch b {
	case uvnan:
		return nan
	case uvnan | 1<<63:
		return nan * -1 // -NaN
	case uvinf:
		return posInf
	case uvneginf:
		return negInf
	case 0:
		return 0
	case 1 << 63:
		return zero * -1 // -0
	}

	// below from GopherJS
	s := hx.GetFloat("", "1")
	if b&(1<<63) != 0 {
		s = hx.GetFloat("", "-1")
	}
	e := (b >> 52) & uint64((1<<11)-1)
	m := b & uint64((1<<52)-1)

	if e == uint64((1<<11)-1) {
		if m == 0 {
			return s / hx.GetFloat("", "0")
		}
		return nan
	}
	if e != 0 {
		m += 1 << 52
	}
	if e == 0 {
		e = 1
	}

	return Ldexp(float64(m), int(e)-1023-52) * s
}

func Ldexp(frac float64, exp int) float64 { // adapted from GopherJS
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
