// Copyright 2014 Elliott Stoneham and The TARDIS Go Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package haxe

import (
	"fmt"
	"reflect"
	"strings"

	"code.google.com/p/go.tools/go/ssa"
	"code.google.com/p/go.tools/go/types"
	"github.com/tardisgo/tardisgo/pogo"
)

func (l langType) LangType(t types.Type, retInitVal bool, errorInfo string) string {
	if pogo.IsValidInPogo(t, errorInfo) {
		switch t.(type) {
		case *types.Basic:
			switch t.(*types.Basic).Kind() {
			case types.Bool, types.UntypedBool:
				if retInitVal {
					return "false"
				}
				return "Bool"
			case types.String, types.UntypedString:
				if retInitVal {
					return `""`
				}
				return "String"
			case types.Float64, types.Float32, types.UntypedFloat:
				if retInitVal {
					return "0.0"
				}
				return "Float"
			case types.Complex64, types.Complex128, types.UntypedComplex:
				if retInitVal {
					return "new Complex(0.0,0.0)"
				}
				return "Complex"
			case types.Int, types.Int8, types.Int16, types.Int32, types.UntypedRune,
				types.Uint8, types.Uint16, types.Uint, types.Uint32: // NOTE: untyped runes default to Int without a warning
				if retInitVal {
					return "0"
				}
				return "Int"
			case types.Int64, types.Uint64:
				if retInitVal {
					return "GOint64.make(0,0)"
				}
				return "GOint64"
			case types.UntypedInt: // TODO: investigate further the situations in which this warning is generated
				pogo.LogWarning(errorInfo, "Haxe", fmt.Errorf("haxe.LangType() types.UntypedInt is ambiguous"))
				return "UNTYPED_INT" // NOTE: if this value were ever to be used, it would cause a Haxe compilation error
			case types.UnsafePointer:
				if retInitVal {
					return "new UnsafePointer(null)" // TODO is this correct? or should it be null
				}
				return "UnsafePointer"
			case types.Uintptr: // Uintptr sometimes used as an integer type, sometimes as a container for another type
				if retInitVal {
					return "null"
				}
				return "Dynamic"
			default:
				pogo.LogWarning(errorInfo, "Haxe", fmt.Errorf("haxe.LangType() unrecognised basic type, Dynamic assumed"))
				if retInitVal {
					return "null"
				}
				return "Dynamic"
			}
		case *types.Interface:
			if retInitVal {
				return `null`
			}
			return "Interface"
		case *types.Named:
			return l.LangType(t.(*types.Named).Underlying(), retInitVal, errorInfo)
		case *types.Chan:
			if retInitVal {
				return "new Channel<" + l.LangType(t.(*types.Chan).Elem(), false, errorInfo) + ">(1)"
			}
			return "Channel<" + l.LangType(t.(*types.Chan).Elem(), false, errorInfo) + ">"
		case *types.Map:
			if retInitVal {
				return "new Map<" + l.LangType(t.(*types.Map).Key(), false, errorInfo) + "," +
					l.LangType(t.(*types.Map).Elem(), false, errorInfo) + ">()"
			}
			return "Map<" + l.LangType(t.(*types.Map).Key(), false, errorInfo) + "," +
				l.LangType(t.(*types.Map).Elem(), false, errorInfo) + ">"
		case *types.Slice:
			if retInitVal {
				return "new Slice(new Pointer<" + l.LangType(t.(*types.Slice).Elem(), false, errorInfo) +
					">(new Array<" + l.LangType(t.(*types.Slice).Elem(), false, errorInfo) +
					">()),0,0" + ")"
			}
			return "Slice"
		case *types.Array: // TODO consider using Vector rather than Array, if faster and can be made to work
			if retInitVal {
				return fmt.Sprintf("new Make<%s>().array(%s,%d)",
					l.LangType(t.(*types.Array).Elem(), false, errorInfo),
					l.LangType(t.(*types.Array).Elem(), true, errorInfo),
					t.(*types.Array).Len())
			}
			return "Array<" + l.LangType(t.(*types.Array).Elem(), false, errorInfo) + ">"
		case *types.Struct:
			ret := "{"
			for ele := 0; ele < t.(*types.Struct).NumFields(); ele++ {
				if ele != 0 {
					ret += ","
				}
				ret += `f_` + t.(*types.Struct).Field(ele).Name() + `: `
				ret += fmt.Sprintf("%s", // "new BoxedVar<%s>(%s)",
					//l.LangType(t.(*types.Struct).Field(ele).Type().Underlying(), false, errorInfo),
					l.LangType(t.(*types.Struct).Field(ele).Type().Underlying(), retInitVal, errorInfo))
			}
			return ret + "}"
		case *types.Tuple: // what is returned by a call and some other instructions, not in the Go language spec!
			tup := t.(*types.Tuple)
			switch tup.Len() {
			case 0:
				return ""
			case 1:
				return l.LangType(tup.At(0).Type().Underlying(), retInitVal, errorInfo)
			default:
				ret := "{"
				for ele := 0; ele < tup.Len(); ele++ {
					if ele != 0 {
						ret += ","
					}
					ret += pogo.MakeID("r"+fmt.Sprintf("%d", ele)) +
						":" + l.LangType(tup.At(ele).Type().Underlying(), retInitVal, errorInfo)
				}
				return ret + "}"
			}
		case *types.Pointer:
			if retInitVal {
				// NOTE pointer declarations create endless recursion for self-referencing structures unless initialized with null
				return "null" //rather than: + l.LangType(t.(*types.Pointer).Elem(), retInitVal, errorInfo) + ")"
			}
			return "PointerIF"
		case *types.Signature:
			if retInitVal {
				return "null"
			}
			ret := "Closure"
			return ret
		default:
			rTyp := reflect.TypeOf(t).String()
			if rTyp == "*ssa.opaqueType" { // NOTE the type for map itterators, not in the Go language spec!
				if retInitVal { // use dynamic type, brief tests seem OK, but may not always work on static hosts
					return "null"
				}
				return "Dynamic"
			}
			pogo.LogError(errorInfo, "Haxe",
				fmt.Errorf("haxe.LangType() internal error, unhandled non-basic type: %s", rTyp))
		}
	}
	return "UNKNOWN_LANGTYPE" // this should generate a Haxe compiler error
}

func (l langType) Convert(register, langType string, destType types.Type, v interface{}, errorInfo string) string {
	srcTyp := l.LangType(v.(ssa.Value).Type().Underlying(), false, errorInfo)
	if srcTyp == langType { // no cast required because the Haxe type is the same
		return register + "=" + l.IndirectValue(v, errorInfo) + ";"
	}
	switch langType { // target Haxe type
	case "Dynamic": // no cast allowed for dynamic variables
		return register + "=" + l.IndirectValue(v, errorInfo) + ";" // TODO review if this is correct for Int64
	case "String":
		switch srcTyp {
		case "Slice":
			switch v.(ssa.Value).Type().Underlying().(*types.Slice).Elem().Underlying().(*types.Basic).Kind() {
			case types.Rune: // []rune
				return "{var _r:Slice=Go_haxegoruntime_Runes2Raw.callFromRT(this._goroutine," + l.IndirectValue(v, errorInfo) + ");" +
					register + "=\"\";for(_i in 0..._r.len())" +
					register + "+=String.fromCharCode(_r.getAt(_i" + "));};"
			case types.Byte: // []byte
				return register + "=Force.toRawString(this._goroutine," + l.IndirectValue(v, errorInfo) + ");"
			default:
				pogo.LogError(errorInfo, "Haxe", fmt.Errorf("haxe.Convert() - Unexpected slice type to convert to String"))
				return ""
			}
		case "Int": // make a string from a single rune
			return "{var _r:Slice=Go_haxegoruntime_Rune2Raw.callFromRT(this._goroutine," + l.IndirectValue(v, errorInfo) + ");" +
				register + "=\"\";for(_i in 0..._r.len())" +
				register + "+=String.fromCharCode(_r.getAt(_i" + "));};"
		case "GOint64": // make a string from a single rune (held in 64 bits)
			return "{var _r:Slice=Go_haxegoruntime_Rune2Raw.callFromRT(this._goroutine,GOint64.toInt(" + l.IndirectValue(v, errorInfo) + "));" +
				register + "=\"\";for(_i in 0..._r.len())" +
				register + "+=String.fromCharCode(_r.getAt(_i" + "));};"
		case "Dynamic":
			return register + "=cast(" + l.IndirectValue(v, errorInfo) + ",String);"
		default:
			pogo.LogError(errorInfo, "Haxe", fmt.Errorf("haxe.Convert() - Unexpected type to convert to String: %s", srcTyp))
			return ""
		}
	case "Slice": // []rune or []byte
		if srcTyp != "String" {
			pogo.LogError(errorInfo, "Haxe", fmt.Errorf("haxe.Convert() - Unexpected type to convert to %s ([]rune or []byte): %s",
				langType, srcTyp))
			return ""
		}
		switch destType.Underlying().(*types.Slice).Elem().Underlying().(*types.Basic).Kind() {
		case types.Rune:
			return register + "=" + newSliceCode("Int", "0",
				l.IndirectValue(v, errorInfo)+".length", l.IndirectValue(v, errorInfo)+".length", errorInfo) + ";" +
				"for(_i in 0..." + l.IndirectValue(v, errorInfo) + ".length)" +
				register + ".setAt(_i,({var _c:Null<Int>=" + l.IndirectValue(v, errorInfo) +
				`.charCodeAt(_i);(_c==null)?0:cast(_c,Int);})` + ");" +
				register + "=Go_haxegoruntime_Raw2Runes.callFromRT(this._goroutine," + register + ");"
		case types.Byte:
			return register + "=Force.toUTF8slice(this._goroutine," + l.IndirectValue(v, errorInfo) + ");"
		default:
			pogo.LogError(errorInfo, "Haxe", fmt.Errorf("haxe.Convert() - Unexpected slice elementto convert to %s ([]rune/[]byte): %s",
				langType, srcTyp))
			return ""
		}
	case "Int": //TODO check that unsigned handelled correctly here
		vInt := ""
		switch srcTyp {
		case "GOint64":
			vInt = "GOint64.toInt(" + l.IndirectValue(v, errorInfo) + ")"
		case "Float":
			vInt = "{var _f:Float=" + l.IndirectValue(v, errorInfo) + ";_f>=0?Math.floor(_f):Math.ceil(_f);}"
		default:
			vInt = "cast(" + l.IndirectValue(v, errorInfo) + "," + langType + ")"
		}
		return register + "=" + l.intTypeCoersion(destType, vInt, errorInfo) + ";"
	case "GOint64":
		switch srcTyp {
		case "Int":
			return register + "=GOint64.ofInt(" + l.IndirectValue(v, errorInfo) + ");"
		case "Float":
			if destType.Underlying().(*types.Basic).Info()&types.IsUnsigned != 0 {
				return register + "=GOint64.ofUFloat(" + l.IndirectValue(v, errorInfo) + ");"
			}
			return register + "=GOint64.ofFloat(" + l.IndirectValue(v, errorInfo) + ");"
		case "Dynamic": // uintptr
			return register + "=" + l.IndirectValue(v, errorInfo) + ";" // let Haxe work out how to do the cast
		default:
			return register + "=cast(" + l.IndirectValue(v, errorInfo) + "," + langType + ");" //TODO unreliable in Java from Dynamic?
		}
	case "Float":
		switch srcTyp {
		case "GOint64":
			if v.(ssa.Value).Type().Underlying().(*types.Basic).Info()&types.IsUnsigned != 0 {
				return register + "=GOint64.toUFloat(" + l.IndirectValue(v, errorInfo) + ");"
			}
			return register + "=GOint64.toFloat(" + l.IndirectValue(v, errorInfo) + ");"
		case "Int":
			if v.(ssa.Value).Type().Underlying().(*types.Basic).Info()&types.IsUnsigned != 0 {
				return register + "=GOint64.toUFloat(GOint64.make(0," + l.IndirectValue(v, errorInfo) + "));"
			}
			return register + "=" + l.IndirectValue(v, errorInfo) + ";" // just the default conversion to float required
		default:
			return register + "=cast(" + l.IndirectValue(v, errorInfo) + "," + langType + ");"
		}
	case "UnsafePointer":
		pogo.LogWarning(errorInfo, "Haxe", fmt.Errorf("attempt to convert a value to be an Unsafe Pointer, which is unsupported"))
		return register + "=new UnsafePointer(" + l.IndirectValue(v, errorInfo) + ");" // this will generate a runtime exception if called
	default:
		if strings.HasPrefix(srcTyp, "Array<") {
			pogo.LogError(errorInfo, "Haxe", fmt.Errorf("haxe.Convert() - No way to convert to %s from %s ", langType, srcTyp))
			return ""
		}
		return register + "=cast(" + l.IndirectValue(v, errorInfo) + "," + langType + ");"
	}
}

func (l langType) MakeInterface(register string, regTyp types.Type, v interface{}, errorInfo string) string {
	return register + `=new Interface(` + pogo.LogTypeUse(v.(ssa.Value).Type() /*NOT underlying()*/) + `,` +
		l.IndirectValue(v, errorInfo) + ");"
}

func (l langType) ChangeInterface(register string, regTyp types.Type, v interface{}, errorInfo string) string {
	return register + `=Interface.change(` + pogo.LogTypeUse(v.(ssa.Value).Type() /*NOT underlying()*/) + `,` +
		l.IndirectValue(v, errorInfo) + ");"
}

/* from the SSA documentation:
The ChangeType instruction applies to X a value-preserving type change to Type().

Type changes are permitted:

- between a named type and its underlying type.
- between two named types of the same underlying type.
- between (possibly named) pointers to identical base types.
- between f(T) functions and (T) func f() methods.
- from a bidirectional channel to a read- or write-channel,
  optionally adding/removing a name.
*/
func (l langType) ChangeType(register string, regTyp interface{}, v interface{}, errorInfo string) string {
	//fmt.Printf("DEBUG CHANGE TYPE: %v -- %v\n", regTyp, v)
	switch v.(ssa.Value).(type) {
	case *ssa.Function:
		rx := v.(*ssa.Function).Signature.Recv()
		pf := ""
		if rx != nil { // it is not the name of a normal function, but that of a method, so append the method description
			pf = rx.Type().String() // NOTE no underlying()
		} else {
			if v.(*ssa.Function).Pkg != nil {
				pf = v.(*ssa.Function).Pkg.Object.Name()
			}
		}
		return register + "=" +
			"new Closure(Go_" + l.LangName(pf, v.(*ssa.Function).Name()) + ".call,[]);"
	default:
		switch v.(ssa.Value).Type().Underlying().(type) {
		case *types.Basic:
			if v.(ssa.Value).Type().Underlying().(*types.Basic).Kind() == types.UnsafePointer {
				/* from https://groups.google.com/forum/#!topic/golang-dev/6eDTDZPWvoM
				   	Treat unsafe.Pointer -> *T conversions by returning new(T).
				   	This is incorrect but at least preserves type-safety...
					TODO decide how UnsafePointer should fail!
				*/
				return register + "=new UnsafePointer(" + l.LangType(regTyp.(types.Type), true, errorInfo) + ");"
			}
		}
	}
	return register + `=` + l.IndirectValue(v, errorInfo) + ";" // usually, this is a no-op as far as Haxe is concerned

}
func (l langType) TypeAssert(register string, v ssa.Value, AssertedType types.Type, CommaOk bool, errorInfo string) string {
	if register == "" {
		return ""
	}
	if CommaOk {
		return register + `=Interface.assertOk(` + pogo.LogTypeUse(AssertedType) + `,` + l.IndirectValue(v, errorInfo) + ");"
	}
	return register + `=Interface.assert(` + pogo.LogTypeUse(AssertedType) + `,` + l.IndirectValue(v, errorInfo) + ");"
}

func (l langType) EmitTypeInfo() string {
	ret := "class TypeInfo{\n"
	pte := pogo.TypesEncountered
	pteKeys := pogo.TypesEncountered.Keys()

	ret += "public static function getName(id:Int):String {\nswitch(id){" + "\n"
	for k := range pteKeys {
		v := pte.At(pteKeys[k])
		ret += "case " + fmt.Sprintf("%d", v) + `: return "` + pteKeys[k].String() + `";` + "\n"
	}
	ret += `default: return "UNKNOWN";}}` + "\n"

	ret += "public static function typeString(i:Interface):String {\nreturn getName(i.typ);\n}\n"

	ret += "public static function getId(name:String):Int {\nswitch(name){" + "\n"
	for k := range pteKeys {
		v := pte.At(pteKeys[k])
		ret += `case "` + pteKeys[k].String() + `": return ` + fmt.Sprintf("%d", v) + `;` + "\n"
	}
	ret += "default: return -1;}}\n"

	//emulation of: func IsAssignableTo(V, T Type) bool
	ret += "public static function isAssignableTo(v:Int,t:Int):Bool {\nif(v==t) return true;\nswitch(v){" + "\n"
	for V := range pteKeys {
		v := pte.At(pteKeys[V])
		ret += `case ` + fmt.Sprintf("%d", v) + `: switch(t){` + "\n"
		for T := range pteKeys {
			t := pte.At(pteKeys[T])
			if v != t && types.AssignableTo(pteKeys[V], pteKeys[T]) {
				ret += `case ` + fmt.Sprintf("%d", t) + `: return true;` + "\n"
			}
		}
		ret += "default: return false;}\n"
	}
	ret += "default: return false;}}\n"

	//emulation of: func IsIdentical(x, y Type) bool
	ret += "public static function isIdentical(v:Int,t:Int):Bool {\nif(v==t) return true;\nswitch(v){" + "\n"
	for V := range pteKeys {
		v := pte.At(pteKeys[V])
		ret += `case ` + fmt.Sprintf("%d", v) + `: switch(t){` + "\n"
		for T := range pteKeys {
			t := pte.At(pteKeys[T])
			if v != t && types.Identical(pteKeys[V], pteKeys[T]) {
				ret += `case ` + fmt.Sprintf("%d", t) + `: return true;` + "\n"
			}
		}
		ret += "default: return false;}\n"
	}
	ret += "default: return false;}}\n"

	//function to answer the question is the type a concrete value?
	ret += "public static function isConcrete(t:Int):Bool {\nswitch(t){" + "\n"
	for T := range pteKeys {
		t := pte.At(pteKeys[T])
		switch pteKeys[T].Underlying().(type) {
		case *types.Interface:
			ret += `case ` + fmt.Sprintf("%d", t) + `: return false;` + "\n"
		default:
			ret += `case ` + fmt.Sprintf("%d", t) + `: return true;` + "\n"
		}
	}
	ret += "default: return false;}}\n"

	// function to give the zero value for each type
	ret += "public static function zeroValue(t:Int):Dynamic {\nswitch(t){" + "\n"
	for T := range pteKeys {
		t := pte.At(pteKeys[T])
		ret += `case ` + fmt.Sprintf("%d", t) + `: return `
		ret += l.LangType(pteKeys[T], true, "EmitTypeInfo()") + ";\n"
	}
	ret += "default: return null;}}\n"

	ret += "public static function method(t:Int,m:String):Dynamic {\nswitch(t){" + "\n"

	tta := pogo.TypesWithMethodSets() //[]types.Type

	for T := range tta {
		t := pte.At(tta[T])
		if t != nil { // it is used?
			ret += `case ` + fmt.Sprintf("%d", t) + `: switch(m){` + "\n"
			ms := types.NewMethodSet(tta[T])
			for m := 0; m < ms.Len(); m++ {
				funcObj, ok := ms.At(m).Obj().(*types.Func)
				pkgName := "unknown"
				if ok && funcObj.Pkg() != nil {
					line := ""
					ss := strings.Split(funcObj.Pkg().Name(), "/")
					pkgName = ss[len(ss)-1]
					if strings.HasPrefix(pkgName, "_") { // exclude functions in haxe for now
						// TODO NoOp for now... so haxe types cant be "Involked" when held in interface types
						// *** need to deal with getters and setters
						// *** also with calling parameters which are different for a Haxe API
					} else {
						line = `case "` + funcObj.Name() + `": return `
						fnToCall := l.LangName(ms.At(m).Recv().String(),
							funcObj.Name())
						ovPkg, _, isOv := l.PackageOverloaded(funcObj.Pkg().Name())
						if isOv {
							fnToCall = strings.Replace(fnToCall, "_"+funcObj.Pkg().Name()+"_", "_"+ovPkg+"_", -1) // NOTE this is not a fool-proof method
						}
						line += `Go_` + fnToCall + `.call` + "; "
					}
					ret += line
				}
				ret += fmt.Sprintf("// %v %v %v %v\n",
					ms.At(m).Obj().Name(),
					ms.At(m).Kind(),
					ms.At(m).Index(),
					ms.At(m).Indirect())
			}
			ret += "default:}\n"
		}
	}
	ret += "default:}\n Scheduler.panicFromHaxe( " + `"no method found!"` + "); return null;}\n" // TODO improve error

	return ret + "}"
}
