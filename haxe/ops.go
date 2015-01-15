// Copyright 2014 Elliott Stoneham and The TARDIS Go Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package haxe

import (
	"fmt"

	"github.com/tardisgo/tardisgo/pogo"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/types"
)

func (l langType) codeUnOp(op string, v interface{}, CommaOK bool, errorInfo string) string {
	useInt64 := false
	lt := l.LangType(v.(ssa.Value).Type().Underlying(), false, errorInfo)
	if lt == "GOint64" {
		useInt64 = true
	}

	// neko target platform requires special handling because in makes whole-number Float into Int without asking
	// see: https://github.com/HaxeFoundation/haxe/issues/1282 which was marked as closed, but not fixed as at 2013.9.6

	switch op {
	case "<-":
		pogo.LogError(errorInfo, "Haxe", fmt.Errorf("codeUnOp(): impossible to reach <- code"))
		return ""
	case "*":
		goTyp := v.(ssa.Value).Type().Underlying().(*types.Pointer).Elem().Underlying()

		//lt = l.LangType(goTyp, false, errorInfo)
		iVal := "" + l.IndirectValue(v, errorInfo) + "" // need to cast it to pointer, when using -dce full and closures
		//switch lt {
		//case "Int":
		//	return "(" + iVal + ".load()|0)" + fmt.Sprintf("/* %v %s */", goTyp, loadStoreSuffix(goTyp)) // force to Int for js, compiled platforms should optimize this away
		//default:
		//if strings.HasPrefix(lt, "Pointer") {
		//	return "({var _v:PointerIF=" + iVal + `.load(); _v;})` // Ensure Haxe can work out that it is a pointer being returned
		//}
		return "Pointer.check(" + iVal + ").load" + loadStoreSuffix(goTyp, false) + ")" + fmt.Sprintf("/* %v */", goTyp)
		//}
	case "-":
		if l.LangType(v.(ssa.Value).Type().Underlying(), false, errorInfo) == "Complex" {
			return "Complex.neg(" + l.IndirectValue(v, errorInfo) + ")"
		}
		fallthrough
	default:
		if useInt64 {
			switch op { // roughly in the order of the GOint64 api spec
			case "-":
				return l.intTypeCoersion(v.(ssa.Value).Type().Underlying(),
					"GOint64.neg("+l.IndirectValue(v, errorInfo)+")", errorInfo)
			case "^":
				return l.intTypeCoersion(v.(ssa.Value).Type().Underlying(),
					"GOint64.xor("+l.IndirectValue(v, errorInfo)+",GOint64.make(-1,-1))", errorInfo)
			default:
				pogo.LogError(errorInfo, "Haxe", fmt.Errorf("codeUnOp(): unhandled Int64 op: %s", op))
				return ""
			}
		} else {
			valStr := l.IndirectValue(v, errorInfo)
			switch v.(ssa.Value).Type().Underlying().(*types.Basic).Kind() {
			case types.Uintptr: // although held as Dynamic, uintpointers are integers when doing calculations
				valStr = "Force.toInt(" + valStr + ")"
			case types.Float32, types.Float64:
				valStr = "Force.toFloat(" + valStr + ")"
			}
			switch op {
			case "^":
				// Haxe has a different operator for bit-wise complement
				return l.intTypeCoersion(v.(ssa.Value).Type().Underlying(),
					"(~"+valStr+")", errorInfo)
			case "-": //both negation and bit-complement can overflow
				return l.intTypeCoersion(v.(ssa.Value).Type().Underlying(),
					"(-"+valStr+")", errorInfo)
			default: //no overflow issues
				return "(" + op + valStr + ")"
			}
		}
	}
}

func (l langType) UnOp(register, op string, v interface{}, CommaOK bool, errorInfo string) string {
	if op == "<-" { // wait for a channel to be ready
		return l.Select(false, register, v, CommaOK, errorInfo)
	}
	return register + "=" + l.codeUnOp(op, v, CommaOK, errorInfo) + ";"
}

func (l langType) codeBinOp(op string, v1, v2 interface{}, errorInfo string) string {
	ret := ""
	useInt64 := false
	v1LangType := l.LangType(v1.(ssa.Value).Type().Underlying(), false, errorInfo)
	v2LangType := l.LangType(v2.(ssa.Value).Type().Underlying(), false, errorInfo)
	v1string := l.IndirectValue(v1, errorInfo)
	v2string := l.IndirectValue(v2, errorInfo)
	if v1LangType == "GOint64" {
		useInt64 = true
	}

	// neko target platform requires special handling because in makes whole-number Float into Int without asking
	// see: https://github.com/HaxeFoundation/haxe/issues/1282 which was marked as closed, but not fixed as at 2013.9.6
	switch v1LangType {
	case "Float":
		v1string = "Force.toFloat(" + v1string + ")"
	case "Dynamic": // assume it is a uintptr, so integer arithmetic is required
		v1string = "(" + v1string + "|0)"
	}
	switch v2LangType {
	case "Float":
		v2string = "Force.toFloat(" + v2string + ")"
	case "Dynamic": // assume it is a uintptr, so integer arithmetic is required
		v2string = "(" + v2string + "|0)"
	}

	if v1LangType == "Complex" {
		switch op {
		case "+":
			return "Complex.add(" + v1string + "," + v2string + ")"
		case "/": // TODO review divide by zero error handling for this case (currently in Haxe Complex class)
			return "Complex.div(" + v1string + "," + v2string + ")"
		case "*":
			return "Complex.mul(" + v1string + "," + v2string + ")"
		case "-":
			return "Complex.sub(" + v1string + "," + v2string + ")"
		case "==":
			return "Complex.eq(" + v1string + "," + v2string + ")"
		case "!=":
			return "Complex.neq(" + v1string + "," + v2string + ")"
		default:
			pogo.LogError(errorInfo, "Haxe", fmt.Errorf("codeBinOp(): unhandled Complex op: %s", op))
			return ""
		}

	} else if v1LangType == "String" {
		//switch op {
		//case ">", "<", "<=", ">=":
		//	return "(Go_haxegoruntime_StringCompare.callFromRT(this._goroutine," + v1string + "," + v2string +
		//		")" + op + "0)"
		//default:
		return "(" + v1string + op + v2string + ")"
		//}

	} else if v1LangType == "Interface" {
		switch op {
		case "==":
			return "Interface.isEqual(" + v1string + "," + v2string + ")"
		case "!=":
			return "!Interface.isEqual(" + v1string + "," + v2string + ")"
		default:
			pogo.LogError(errorInfo, "Haxe", fmt.Errorf("codeBinOp(): unhandled Interface op: %s", op))
			return ""
		}

	} else if v1LangType == "Pointer" {
		switch op {
		case "==":
			return "Pointer.isEqual(" + v1string + "," + v2string + ")"
		case "!=":
			return "!Pointer.isEqual(" + v1string + "," + v2string + ")"
		default:
			pogo.LogError(errorInfo, "Haxe", fmt.Errorf("codeBinOp(): unhandled Pointer op: %s", op))
			return ""
		}

	} else {
		if useInt64 { // explicitly enumerate all of the Int64 functions
			isSignedStr := "true"
			if (v1.(ssa.Value).Type().Underlying().(*types.Basic).Info() & types.IsUnsigned) != 0 {
				isSignedStr = "false"
			}

			switch op { // roughly in the order of the GOint64 api spec
			case "+":
				ret = l.intTypeCoersion(v1.(ssa.Value).Type().Underlying(),
					"GOint64.add("+v1string+","+v2string+")", errorInfo)
			case "&":
				ret = l.intTypeCoersion(v1.(ssa.Value).Type().Underlying(),
					"GOint64.and("+v1string+","+v2string+")", errorInfo)
			case "/":
				ret = l.intTypeCoersion(v1.(ssa.Value).Type().Underlying(),
					"GOint64.div("+v1string+","+v2string+","+isSignedStr+")", errorInfo)
			case "%":
				ret = l.intTypeCoersion(v1.(ssa.Value).Type().Underlying(),
					"GOint64.mod("+v1string+","+v2string+","+isSignedStr+")", errorInfo)
			case "*":
				ret = l.intTypeCoersion(v1.(ssa.Value).Type().Underlying(),
					"GOint64.mul("+v1string+","+v2string+")", errorInfo)
			case "|":
				ret = l.intTypeCoersion(v1.(ssa.Value).Type().Underlying(),
					"GOint64.or("+v1string+","+v2string+")", errorInfo)
			case "<<":
				if v2LangType == "GOint64" {
					v2string = "GOint64.toInt(" + v2string + ")"
				}
				ret = l.intTypeCoersion(v1.(ssa.Value).Type().Underlying(),
					"GOint64.shl("+v1string+","+v2string+")", errorInfo)
			case ">>":
				if v2LangType == "GOint64" {
					v2string = "GOint64.toInt(" + v2string + ")"
				}
				if isSignedStr == "true" {
					ret = l.intTypeCoersion(v1.(ssa.Value).Type().Underlying(),
						"GOint64.shr("+v1string+","+v2string+")", errorInfo) // GOint64.shr does sign extension
				} else {
					ret = l.intTypeCoersion(v1.(ssa.Value).Type().Underlying(),
						"GOint64.ushr("+v1string+","+v2string+")", errorInfo) // GOint64.ushr does not do sign extension
				}
			case "-":
				ret = l.intTypeCoersion(v1.(ssa.Value).Type().Underlying(),
					"GOint64.sub("+v1string+","+v2string+")", errorInfo)
			case "^":
				ret = l.intTypeCoersion(v1.(ssa.Value).Type().Underlying(),
					"GOint64.xor("+v1string+","+v2string+")", errorInfo)
			case "&^":
				v2string = "GOint64.xor(" + v2string + ",GOint64.make(-1,-1))"
				ret = l.intTypeCoersion(v1.(ssa.Value).Type().Underlying(),
					"GOint64.and("+v1string+","+v2string+")", errorInfo)
			case "==", "!=", "<", ">", "<=", ">=":
				compFunc := "GOint64.compare("
				if (v1.(ssa.Value).Type().Underlying().(*types.Basic).Info() & types.IsUnsigned) != 0 {
					compFunc = "GOint64.ucompare("
				}
				ret = "(" + compFunc + v1string + "," + v2string + ")" + op + "0)"
			default:
				pogo.LogError(errorInfo, "Haxe", fmt.Errorf("codeBinOp(): unhandled 64-bit op: %s", op))
				return ""
			}

		} else {
			switch op { // standard case, use Haxe operators
			case "==", "!=", "<", ">", "<=", ">=": // no integer coersion, boolian results
				switch v1.(ssa.Value).Type().Underlying().(type) {
				case *types.Basic:
					if (v1.(ssa.Value).Type().Underlying().(*types.Basic).Info() & types.IsUnsigned) != 0 {
						ret = "(Force.uintCompare(" + v1string + "," + v2string + ")" + op + "0)"
					} else {
						ret = "(" + v1string + op + v2string + ")"
					}
				default:
					ret = "(" + v1string + op + v2string + ")"
				}
			case ">>", "<<":
				if v2LangType == "GOint64" {
					v2string = "GOint64.toInt(" + v2string + ")"
				}
				switch v1.(ssa.Value).Type().Underlying().(*types.Basic).Kind() {
				case types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uintptr: // unsigned bit shift
					if op == ">>" {
						op = ">>>" // logical right shift if unsigned
					}
				}
				ret = "({var _v1:Int=" + v1string + "; var _v2:Int=" + v2string + "; _v2==0?_v1:_v1" + op + "_v2;})" //NoOp if v2==0
			case "/":
				switch v1.(ssa.Value).Type().Underlying().(*types.Basic).Kind() {
				case types.Int8:
					ret = "Force.intDiv(" + v1string + "," + v2string + ",1)" // 1 byte special processing
				case types.Int16:
					ret = "Force.intDiv(" + v1string + "," + v2string + ",2)" // 2 byte special processing
				case types.UntypedInt, types.Int, types.Int32: // treat all unknown int types as int 32
					ret = "Force.intDiv(" + v1string + "," + v2string + ",4)" // 4 byte special processing
				case types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uintptr: // unsigned division
					ret = "Force.intDiv(" + v1string + "," + v2string + ",0)" // spec does not require special processing, but is unsigned
				case types.UntypedFloat, types.Float32, types.Float64:
					ret = "Force.floatDiv(" + v1string + "," + v2string + ")"
				default:
					pogo.LogError(errorInfo, "Haxe", fmt.Errorf("codeBinOp(): unhandled divide type"))
					ret = "(ERROR)"
				}
			case "%":
				switch v1.(ssa.Value).Type().Underlying().(*types.Basic).Kind() {
				case types.Int8:
					ret = "Force.intMod(" + v1string + "," + v2string + ", 1)"
				case types.Int16:
					ret = "Force.intMod(" + v1string + "," + v2string + ", 2)"
				case types.UntypedInt, types.Int, types.Int32: // treat all unknown int types as int 32
					ret = "Force.intMod(" + v1string + "," + v2string + ", 4)"
				case types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uintptr: // unsigned mod
					ret = "Force.intMod(" + v1string + "," + v2string + ", 0)"
				case types.UntypedFloat, types.Float32, types.Float64:
					ret = "Force.floatMod(" + v1string + "," + v2string + ")"
				default:
					pogo.LogError(errorInfo, "Haxe", fmt.Errorf("codeBinOp(): unhandled divide type"))
					ret = "(ERROR)"
				}

			case "&^":
				op = "&~" // Haxe has a different operator for bit-wise complement
				fallthrough
			default:
				innerCode := "(" + v1string + op + v2string + ")"
				ret = l.intTypeCoersion(
					v1.(ssa.Value).Type().Underlying(),
					innerCode, errorInfo)
			}
		}
		return ret
	}
}

func (l langType) BinOp(register, op string, v1, v2 interface{}, errorInfo string) string {
	return register + "=" + l.codeBinOp(op, v1, v2, errorInfo) + ";"
}
