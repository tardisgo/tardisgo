// Copyright 2014 Elliott Stoneham and The tardisgo Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Documentation of math functions overloaded by TARDIS Go->Haxe transpiler
package math

// this package currently for documentation purposes only

//emulated in Go standard math package:
/*
func Atan2(y, x float64) float64                 { return atan2(x, y) }
func Frexp(f float64) (frac float64, exp int)    { return frexp(f) }
func Hypot(p, q float64) float64                 { return hypot(p, q) }
func Ldexp(frac float64, exp int) float64        { return ldexp(frac, exp) }
func Log1p(x float64) float64                    { return log1p(x) }
func Mod(x, y float64) float64                   { return mod(x, y) }
func Modf(f float64) (int float64, frac float64) { return modf(f) }
func Sincos(x float64) (sin, cos float64)        { return sincos(x) }
*/

//Go Math functions
//mapped into Haxe:
/*
	"math_Abs":   "Math.abs",
	"math_Acos":  "Math.acos",
	"math_Asin":  "Math.asin",
	"math_Atan":  "Math.atan",
	"math_Ceil":  "Math.fceil",
	"math_Cos":   "Math.cos",
	"math_Exp":   "Math.exp",
	"math_Floor": "Math.ffloor",
	"math_Log":   "Math.log",
	"math_Sin":   "Math.sin",
	"math_Sqrt":  "Math.sqrt",
	"math_Tan":   "Math.tan",
*/
