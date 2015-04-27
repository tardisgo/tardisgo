// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build haxe

// Copyright 2014 Elliott Stoneham and The tardisgo Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package math

import "github.com/tardisgo/tardisgo/haxe/hx"

// Float32bits returns the IEEE 754 binary representation of f.
//func Float32bits(f float32) uint32 { return *(*uint32)(unsafe.Pointer(&f)) }
func Float32bits(f float32) uint32 {
	return uint32(hx.CodeInt("",
		"Go_haxegoruntime_FFloat32bits.callFromRT(this._goroutine,_a.itemAddr(0).load().val);", f))
}

// Float32frombits returns the floating point number corresponding
// to the IEEE 754 binary representation b.
// func Float32frombits(b uint32) float32 { return *(*float32)(unsafe.Pointer(&b)) }
func Float32frombits(b uint32) float32 {
	return float32(hx.CodeFloat("",
		"Go_haxegoruntime_FFloat32frombits.callFromRT(this._goroutine,_a.itemAddr(0).load().val);", b))
}

// Float64bits returns the IEEE 754 binary representation of f.
//func Float64bits(f float64) uint64 { return *(*uint64)(unsafe.Pointer(&f)) }
func Float64bits(f float64) uint64 {
	return uint64(hx.Int64(hx.CodeDynamic("",
		"Go_haxegoruntime_FFloat64bits.callFromRT(this._goroutine,_a.itemAddr(0).load().val);", f)))
}

// Float64frombits returns the floating point number corresponding
// the IEEE 754 binary representation b.
//func Float64frombits(b uint64) float64 { return *(*float64)(unsafe.Pointer(&b)) }
func Float64frombits(b uint64) float64 {
	return hx.CodeFloat("",
		"Go_haxegoruntime_FFloat64frombits.callFromRT(this._goroutine,_a.itemAddr(0).load().val);", b)
}
