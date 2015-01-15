// Copyright 2014 Elliott Stoneham and The TARDIS Go Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package haxe

import (
	"go/ast"

	"golang.org/x/tools/go/ssa"
	//"fmt"
)

var builtinOverloadMap = map[string]string{
//Go Math functions
//built into Haxe:
//"math_Abs":  "Math.abs",
//"math_Acos": "Math.acos",
//"math_Asin": "Math.asin",
//"math_Atan": "Math.atan",
//"math_Ceil": "Math.fceil",
//"math_Cos": "Math.cos",
//"math_Exp":   "Math.exp", // use Go version to make tests work
//"math_Floor": "Math.ffloor",
//"math_Log":  "Math.log",
//"math_Sin":  "Math.sin",
//"math_Sqrt": "Math.sqrt",
//"math_Tan":  "Math.tan",

//Type of an interface value
//runtime
//"runtime_typestring": "TypeInfo.typeString",
}

var fnOverloadMap = map[string]string{
//Go Math functions
//emulated in Go standard maths package:
//"math_Frexp":     "Go_math_frexp.call",
//"math_Modf":   "Go_math_modf.call",
//"math_Mod":    "Go_math_mod.call",
//"math_Sincos": "Go_math_sincos.call",
//"math_Log1p":  "Go_math_log1p.call",
//"math_Ldexp":  "Go_math_ldexp.call",
//"math_Hypot":  "Go_math_hypot.call",
//"math_Atan2":     "Go_math_atan2.call",
//"math_Max":       "Go_math_max.call",
//"math_Min":       "Go_math_min.call",
//"math_Log2":  "Go_math_log2.call",
//"math_Log10": "Go_math_log10.call",
//"math_Expm1": "Go_math_expm1.call",
//"math_Trunc":     "Go_math_trunc.call",
//"math_Remainder": "Go_math_remainder.call",
//"math_Dim":       "Go_math_dim.call",
//"math_Exp": "Go_math_exp.call",

//emulated in golibruntime/math
//"math_Float32bits":     "Go_tgoaddmath_glrFloat32bits.call",
//"math_Float32frombits": "Go_tgoaddmath_glrFloat32frombits.call",
//"math_Float64bits":     "Go_tgoaddmath_glrFloat64bits.call",
//"math_Float64frombits": "Go_tgoaddmath_glrFloat64frombits.call",
}

var fnToVarOverloadMap = map[string]string{
//built into Haxe as variables:
//Go Math functions
//"math_NaN": "Math.NaN",
}

func (l langType) FunctionOverloaded(pkg, fun string) bool {
	//fmt.Printf("DEBUG fn ov :%s:%s:\n", pkg, fun)
	_, ok := fnOverloadMap[pkg+"_"+fun]
	if ok {
		return true
	}
	_, ok = fnToVarOverloadMap[pkg+"_"+fun]
	if ok {
		return true
	}
	_, ok = builtinOverloadMap[pkg+"_"+fun]
	return ok
}

func (l langType) FuncName(fnx *ssa.Function) string {
	pn := ""
	if fnx.Signature.Recv() != nil {
		pn = fnx.Signature.Recv().Type().String() // NOTE no use of underlying here
	} else {
		pn = "unknown"
		fn := ssa.EnclosingFunction(fnx.Package(), []ast.Node{fnx.Syntax()})
		if fn == nil {
			fn = fnx
		}
		if fn.Pkg != nil {
			if fn.Pkg.Object != nil {
				pn = fn.Pkg.Object.Name()
			}
		} else {
			if fn.Object() != nil {
				if fn.Object().Pkg() != nil {
					pn = fn.Object().Pkg().Name()
				}
			}
		}
	}
	return l.LangName(pn, fnx.Name())
}
