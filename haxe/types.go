// Copyright 2014 Elliott Stoneham and The TARDIS Go Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package haxe

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/tardisgo/tardisgo/pogo"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/types"
	"golang.org/x/tools/go/types/typeutil"
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
				if retInitVal {
					return "0"
				}
				return "UNTYPED_INT" // NOTE: if this value were ever to be used, it would cause a Haxe compilation error
			case types.UnsafePointer:
				if retInitVal {
					return "null" // NOTE ALL pointers are unsafe
				}
				return "Pointer"
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
			haxeName := getHaxeClass(t.(*types.Named).String())
			//fmt.Println("DEBUG Go named type -> Haxe type :", t.(*types.Named).String(), "->", haxeName)
			if haxeName != "" {
				if retInitVal {
					return `null` // NOTE code to the right does not work in openfl/flash: `Type.createEmptyInstance(` + haxeName + ")"
				}
				return haxeName
			}
			return l.LangType(t.(*types.Named).Underlying(), retInitVal, errorInfo)
		case *types.Chan:
			if retInitVal {
				return "new Channel(1)" //waa: <" + l.LangType(t.(*types.Chan).Elem(), false, errorInfo) + ">(1)"
			}
			return "Channel" //was: <" + l.LangType(t.(*types.Chan).Elem(), false, errorInfo) + ">"
		case *types.Map:
			if retInitVal {
				return "new GOmap(" +
					l.LangType(t.(*types.Map).Key(), true, errorInfo) + "," +
					l.LangType(t.(*types.Map).Elem(), true, errorInfo) + ")"
			}
			return "GOmap"
		case *types.Slice:
			if retInitVal {
				return "new Slice(new Pointer(" +
					"new Object(0)" +
					"),0,0,0," + "1" + arrayOffsetCalc(t.(*types.Slice).Elem().Underlying()) + ")"
			}
			return "Slice"
		case *types.Array:
			if retInitVal {
				return fmt.Sprintf("new Object(%d)", haxeStdSizes.Sizeof(t))
			}
			return "Object"
		case *types.Struct:
			if retInitVal {
				return fmt.Sprintf("new Object(%d)", haxeStdSizes.Sizeof(t.(*types.Struct).Underlying()))
			}
			return "Object"
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
					ret += pogo.MakeID("r"+fmt.Sprintf("%d", ele)) + ":"
					if !retInitVal {
						ret += "Null<"
					}
					ret += l.LangType(tup.At(ele).Type().Underlying(), retInitVal, errorInfo)
					if !retInitVal {
						ret += ">"
					}
				}
				return ret + "}"
			}
		case *types.Pointer:
			if retInitVal {
				// NOTE pointer declarations create endless recursion for self-referencing structures unless initialized with null
				return "null" //rather than: + l.LangType(t.(*types.Pointer).Elem(), retInitVal, errorInfo) + ")"
			}
			return "Pointer"
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
	if srcTyp == langType && langType != "Float" && langType != "Int" { // no cast required because the Haxe type is the same
		return register + "=" + l.IndirectValue(v, errorInfo) + ";"
	}
	switch langType { // target Haxe type
	case "Dynamic": // no cast allowed for dynamic variables
		vInt := l.IndirectValue(v, errorInfo)
		// but some Go code uses uintptr as just another integer, so ensure it is unsigned
		switch srcTyp {
		case "GOint64":
			vInt = "Force.toUint32(GOint64.toInt(" + vInt + "))"
		case "Float":
			vInt = "Force.toUint32({var _f:Float=" + vInt + ";_f>=0?Math.floor(_f):Math.ceil(_f);})" // same as signed
		case "Int":
			vInt = "Force.toUint32(" + vInt + ")"
		}
		return register + "=" + vInt + ";"
	case "Pointer":
		if srcTyp == "Dynamic" {
			return register + "=" + l.IndirectValue(v, errorInfo) + "==null?null:Pointer.check(" + l.IndirectValue(v, errorInfo) + ");"
		}
		pogo.LogError(errorInfo, "Haxe", fmt.Errorf("haxe.Convert() - can only convert uintptr to unsafe.Pointer"))
		return ""
	case "String":
		switch srcTyp {
		case "Slice":
			switch v.(ssa.Value).Type().Underlying().(*types.Slice).Elem().Underlying().(*types.Basic).Kind() {
			case types.Rune: // []rune
				return register +
					"=Force.toRawString(this._goroutine,Go_haxegoruntime_RRunesTToUUTTFF8.callFromRT(this._goroutine," +
					l.IndirectValue(v, errorInfo) + "));"
			case types.Byte: // []byte
				return register + "=Force.toRawString(this._goroutine," + l.IndirectValue(v, errorInfo) + ");"
			default:
				pogo.LogError(errorInfo, "Haxe", fmt.Errorf("haxe.Convert() - Unexpected slice type to convert to String"))
				return ""
			}
		case "Int": // make a string from a single rune
			return "{var _r:Slice=Go_haxegoruntime_RRune2RRaw.callFromRT(this._goroutine," + l.IndirectValue(v, errorInfo) + ");" +
				register + "=\"\";for(_i in 0..._r.len())" +
				register + "+=String.fromCharCode(_r.itemAddr(_i).load_int32(" + "));};"
		case "GOint64": // make a string from a single rune (held in 64 bits)
			return "{var _r:Slice=Go_haxegoruntime_RRune2RRaw.callFromRT(this._goroutine,GOint64.toInt(" + l.IndirectValue(v, errorInfo) + "));" +
				register + "=\"\";for(_i in 0..._r.len())" +
				register + "+=String.fromCharCode(_r.itemAddr(_i).load_int32(" + "));};"
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
			//return register + "=" + newSliceCode("Int", "0",
			//	l.IndirectValue(v, errorInfo)+".length",
			//	l.IndirectValue(v, errorInfo)+".length", errorInfo, "4 /*len(rune)*/") + ";" +
			//	"for(_i in 0..." + l.IndirectValue(v, errorInfo) + ".length)" +
			//	register + ".itemAddr(_i).store_int32(({var _c:Null<Int>=" + l.IndirectValue(v, errorInfo) +
			//	`.charCodeAt(_i);(_c==null)?0:Std.int(_c)&0xff;})` + ");" +
			//	register + "=Go_haxegoruntime_Raw2Runes.callFromRT(this._goroutine," + register + ");"
			return register +
				"=Go_haxegoruntime_UUTTFF8toRRunes.callFromRT(this._goroutine,Force.toUTF8slice(this._goroutine," +
				l.IndirectValue(v, errorInfo) + "));"
		case types.Byte:
			return register + "=Force.toUTF8slice(this._goroutine," + l.IndirectValue(v, errorInfo) + ");"
		default:
			pogo.LogError(errorInfo, "Haxe", fmt.Errorf("haxe.Convert() - Unexpected slice elementto convert to %s ([]rune/[]byte): %s",
				langType, srcTyp))
			return ""
		}
	case "Int":
		vInt := ""
		switch srcTyp {
		case "Int":
			vInt = l.IndirectValue(v, errorInfo) // to get the type coercion below
		case "GOint64":
			vInt = "GOint64.toInt(" + l.IndirectValue(v, errorInfo) + ")" // un/signed OK as just truncates
		case "Float":
			vInt = "{var _f:Float=" + l.IndirectValue(v, errorInfo) + ";_f>=0?Math.floor(_f):Math.ceil(_f);}"
		case "Dynamic":
			vInt = "Force.toInt(" + l.IndirectValue(v, errorInfo) + ")" // Dynamic == uintptr
		default:
			pogo.LogError(errorInfo, "Haxe", fmt.Errorf("haxe.Convert() - unhandled convert to u/int from: %s", srcTyp))
			return ""
		}
		return register + "=" + l.intTypeCoersion(destType, vInt, errorInfo) + ";"
	case "GOint64":
		switch srcTyp {
		case "Int":
			if v.(ssa.Value).Type().Underlying().(*types.Basic).Info()&types.IsUnsigned != 0 {
				return register + "=GOint64.ofUInt(" + l.IndirectValue(v, errorInfo) + ");"
			}
			return register + "=GOint64.ofInt(" + l.IndirectValue(v, errorInfo) + ");"
		case "Float":
			if destType.Underlying().(*types.Basic).Info()&types.IsUnsigned != 0 {
				return register + "=GOint64.ofUFloat(" + l.IndirectValue(v, errorInfo) + ");" 
			}
			return register + "=GOint64.ofFloat(" + l.IndirectValue(v, errorInfo) + ");"
		case "Dynamic": // uintptr
			return register + "=GOint64.ofUInt(Force.toInt(" + l.IndirectValue(v, errorInfo) + "));" // let Haxe work out how to do the cast
		default:
			pogo.LogError(errorInfo, "Haxe", fmt.Errorf("haxe.Convert() - unhandled convert to u/int64 from: %s", srcTyp))
			return ""
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
			return register + "=Force.toFloat(" + l.IndirectValue(v, errorInfo) + ");" // just the default conversion to float required
		case "Dynamic":
			return register + "=GOint64.toUFloat(GOint64.ofUInt(Force.toInt(" + l.IndirectValue(v, errorInfo) + ")));"
		case "Float":
			if destType.Underlying().(*types.Basic).Kind() == types.Float32 {
				return register + "=Force.toFloat32(" +
					l.IndirectValue(v, errorInfo) + ");" // need to truncate to float32
			}
			return register + "=Force.toFloat(" + l.IndirectValue(v, errorInfo) + ");"
		default:
			pogo.LogError(errorInfo, "Haxe", fmt.Errorf("haxe.Convert() - unhandled convert to float from: %s", srcTyp))
			return ""
		}
	case "UnsafePointer":
		//pogo.LogWarning(errorInfo, "Haxe", fmt.Errorf("converting a pointer to an Unsafe Pointer"))
		return register + "=" + l.IndirectValue(v, errorInfo) + ";" // ALL Pointers are unsafe ?
	default:
		if strings.HasPrefix(srcTyp, "Array<") {
			pogo.LogError(errorInfo, "Haxe", fmt.Errorf("haxe.Convert() - No way to convert to %s from %s ", langType, srcTyp))
			return ""
		}
		pogo.LogError(errorInfo, "Haxe", fmt.Errorf("haxe.Convert() - Unhandled convert to %s from %s ", langType, srcTyp))
		//return register + "=cast(" + l.IndirectValue(v, errorInfo) + "," + langType + ");"
		return ""
	}
}

func (l langType) MakeInterface(register string, regTyp types.Type, v interface{}, errorInfo string) string {
	ret := `new Interface(` + pogo.LogTypeUse(v.(ssa.Value).Type() /*NOT underlying()*/) + `,` +
		l.IndirectValue(v, errorInfo) + ")"
	if getHaxeClass(regTyp.String()) != "" {
		ret = "Force.toHaxeParam(" + ret + ")" // as interfaces are not native to haxe, so need to convert
		// TODO optimize when stable
	}
	return register + `=` + ret + ";"
}

func (l langType) ChangeInterface(register string, regTyp types.Type, v interface{}, errorInfo string) string {
	pogo.LogTypeUse(regTyp) // make sure it is in the DB
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
		/*
			rx := v.(*ssa.Function).Signature.Recv()
			pf := fmt.Sprintf("fn%d", v.(*ssa.Function).Pos())
			if rx != nil { // it is not the name of a normal function, but that of a method, so append the method description
				pf = rx.Type().String() // NOTE no underlying()
			} else {
				if v.(*ssa.Function).Pkg != nil {
					pf = v.(*ssa.Function).Pkg.Object.Path() // was .Name()
				}
			}
			return register + "=" +
				"new Closure(Go_" + l.LangName(pf, v.(*ssa.Function).Name()) + ".call,[]);"
		*/
		return register + "=" +
			"new Closure(Go_" + l.LangName(pogo.FuncPathName(v.(*ssa.Function))) + ".call,[]);"
	default:
		hType := getHaxeClass(regTyp.(types.Type).String())
		if hType != "" {
			switch v.(ssa.Value).Type().Underlying().(type) {
			case *types.Interface:
				return register + "=" + l.IndirectValue(v, errorInfo) + ".val;"
			default:
				return register + "=cast " + l.IndirectValue(v, errorInfo) + ";" // unsafe cast!
			}
		}
		switch v.(ssa.Value).Type().Underlying().(type) {
		case *types.Basic:
			if v.(ssa.Value).Type().Underlying().(*types.Basic).Kind() == types.UnsafePointer {
				/* from https://groups.google.com/forum/#!topic/golang-dev/6eDTDZPWvoM
				   	Treat unsafe.Pointer -> *T conversions by returning new(T).
				   	This is incorrect but at least preserves type-safety...
					TODO decide if the above method is better than just copying the value as below
				*/
				return register + "=" + l.LangType(regTyp.(types.Type), true, errorInfo) + ";"
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

func getHaxeClass(fullname string) string { // NOTE capital letter de-doubling not handled here
	if fullname[0] != '*' { // pointers can't be Haxe types
		bits := strings.Split(fullname, "/")
		s := bits[len(bits)-1] // right-most bit contains the package name & type name
		// fmt.Println("DEBUG bit to consider", s)
		if s[0] == '_' { // leading _ on the package name means a haxe type
			//fmt.Println("DEBUG non-pointer goType", goType)
			splits := strings.Split(s, ".")
			if len(splits) == 2 { // we have a package and type
				goType := splits[1][1:] // type part only, without the leading Restrictor
				haxeType := strings.Replace(goType, "_", ".", -1)
				haxeType = strings.Replace(haxeType, "...", ".", -1)
				// fmt.Println("DEBUG go->haxe found", goType, "->", haxeType)
				return haxeType
			}
		}
	}
	return ""
}

func preprocessTypeName(v string) string {
	s := ""
	hadbackslash := false
	content := strings.Trim(v, `"`)
	for _, c := range content {
		if hadbackslash {
			hadbackslash = false
			s += string(c)
		} else {
			switch c {
			case '"': // the reason we are here - orphan ""
				s += "\\\""
			case '\\':
				hadbackslash = true
				s += string(c)
			default:
				s += string(c)
			}
		}
	}
	return s
}

func getTypeInfo(t types.Type, tname string) (kind reflect.Kind, name string) {
	return GetTypeInfo(t, tname)
}

func notInterface(t types.Type) bool {
	isNamed := false
	tt := t
	named1, isNamed1 := tt.(*types.Named)
	if isNamed1 {
		tt = named1.Underlying()
		isNamed = true
	}
	ptr, isPtr := tt.(*types.Pointer)
	if isPtr {
		tt = ptr.Elem()
	}
	named2, isNamed2 := tt.(*types.Named)
	if isNamed2 {
		tt = named2.Underlying()
		isNamed = true
	}
	_, isInterface := tt.(*types.Interface)
	if !isInterface && isNamed {
		return true
	}
	return false
}

var typesByID []types.Type
var pte typeutil.Map
var pteKeys []types.Type

func buildTBI() {
	pte = pogo.TypesEncountered
	pteKeys = pogo.TypesEncountered.Keys()
	sort.Sort(pogo.TypeSorter(pteKeys))
	typesByID = make([]types.Type, pogo.NextTypeID)
	for k := range pteKeys {
		v := pte.At(pteKeys[k]).(int)
		typesByID[v] = pteKeys[k]
	}
}

func (l langType) EmitTypeInfo() string {

	BuildTypeHaxe() // generate the code to emulate compiler reflect data output

	var ret string = ""

	/*if false { // was: pogo.UsingPackage("reflect")

		ret += "class PtrTypeInfo{\n"

		buildTBI()

		ret += "public static var ptrByID:Map<Int,Int> = [ 0=> 0,"
		for id, t := range typesByID {
			switch t.(type) {
			case *types.Named:
				t = t.(*types.Named).Underlying()
			}
			switch t.(type) {
			case *types.Pointer:
				ret += fmt.Sprintf("\t%d=>%s,\n", id, pogo.LogTypeUse(t.(*types.Pointer).Elem()))
			}
		}
		ret += "];\n"
		ret += "}\n"

		pogo.WriteAsClass("PtrTypeInfo", ret)
		ret = ""

		ret += "\nclass SliceTypeInfo{\n"

		buildTBI()

		ret += "public static var sliceByID:Map<Int,Int> = [ 0=> 0,\n"
		for id, t := range typesByID {
			switch t.(type) {
			case *types.Named:
				t = t.(*types.Named).Underlying()
			}
			switch t.(type) {
			case *types.Slice:
				ret += fmt.Sprintf("\t%d=>%s,\n", id, pogo.LogTypeUse(t.(*types.Slice).Elem()))
			}
		}
		ret += "];\n"
		ret += "}\n"

		pogo.WriteAsClass("SliceTypeInfo", ret)
		ret = ""

		ret += "\nclass StructTypeInfo{\n"

		buildTBI()

		ret += "// mirrors reflect.structType / reflect.structField\n"
		ret += "//[0] name    *string // nil for embedded fields\n"
		ret += "//[1] pkgPath *string // nil for exported Names; otherwise import path\n"
		ret += "//[2] typ     *rtype  // type of field\n"
		ret += "//[3] tag     *string // nil if no tag\n"
		ret += "//[4] offset  uintptr // byte offset of field within struct\n"
		ret += "public static var structByID:Array<{id:Int,flds:Array<Array<Dynamic>>}> = "
		ret += "haxe.Json.parse(\"[{ \\\"id\\\":0, \\\"flds\\\":[] }"
		for id, t := range typesByID {
			switch t.(type) {
			case *types.Named:
				t = t.(*types.Named).Underlying()
			}
			switch t.(type) {
			case *types.Struct:
				fields := []*types.Var{}
				for fld := 0; fld < t.(*types.Struct).NumFields(); fld++ {
					fldInfo := t.(*types.Struct).Field(fld)
					if fldInfo.IsField() {
						fields = append(fields, fldInfo)
					}
				}
				offs := haxeStdSizes.Offsetsof(fields)
				ret += fmt.Sprintf(",{\\\"id\\\": %d, \\\"flds\\\":[ ", id)
				is1st := true
				for fld := 0; fld < t.(*types.Struct).NumFields(); fld++ {
					fldInfo := t.(*types.Struct).Field(fld)
					if fldInfo.IsField() {
						name := fldInfo.Name()
						path := ""
						rune1, _ := utf8.DecodeRune([]byte(name))
						if unicode.IsLower(rune1) {
							path = fldInfo.Pkg().Path()
						}
						if is1st {
							is1st = false
						} else {
							ret += ","
						}
						ret += "[ "
						ret += "\\\"" + strings.Trim(haxeStringConst(`"`+name+`"`, "CompilerInternal:haxe.EmitTypeInfo()"), `"`) + "\\\", "
						ret += "\\\"" + strings.Trim(haxeStringConst(`"`+path+`"`, "CompilerInternal:haxe.EmitTypeInfo()"), `"`) + "\\\", "
						ret += fmt.Sprintf("%s, ", pogo.LogTypeUse(fldInfo.Type()))
						ret += "\\\"" + strings.Trim(haxeStringConst(`"`+t.(*types.Struct).Tag(fld)+`"`, "CompilerInternal:haxe.EmitTypeInfo()"), `"`) + "\\\", "
						ret += fmt.Sprintf("%d ", offs[fld])
						ret += "]"
					}
				}
				ret += "]}"
			}
		}
		ret += "]\");\n"
		ret += "}\n"

		pogo.WriteAsClass("StructTypeInfo", ret)
		ret = ""

		ret += "\nclass IfaceTypeInfo{\n"

		buildTBI()

		ret += "public static var ifaceByID:Map<Int,Array<Array<Dynamic>>> = [ 0=> [],\n"
		ret += "// mirrors reflect.interfaceType / reflect.imethod\n"
		ret += "//[0] name    *string // name of method\n"
		ret += "//[1] pkgPath *string // nil for exported Names; otherwise import path\n"
		ret += "//[2] typ     *rtype  //  .(*FuncType) underneath\n"
		for id, t := range typesByID {
			switch t.(type) {
			case *types.Named:
				t = t.(*types.Named).Underlying()
			}
			switch t.(type) {
			case *types.Interface:
				ret += fmt.Sprintf("\t%d=>[ \n", id)
				for mth := 0; mth < t.(*types.Interface).NumMethods(); mth++ {
					methInfo := t.(*types.Interface).Method(mth)
					name := methInfo.Name()
					path := ""
					rune1, _ := utf8.DecodeRune([]byte(name))
					if unicode.IsLower(rune1) {
						path = methInfo.Pkg().Path()
					}
					ret += "\t\t[ "
					ret += haxeStringConst(`"`+name+`"`, "CompilerInternal:haxe.EmitTypeInfo()") + ", "
					ret += haxeStringConst(`"`+path+`"`, "CompilerInternal:haxe.EmitTypeInfo()") + ", "
					ret += fmt.Sprintf("%s, ", pogo.LogTypeUse(methInfo.Type()))
					ret += "],\n"
				}
				ret += "\t],\n"
			}
		}
		ret += "];\n"
		ret += "}\n"

		pogo.WriteAsClass("IfaceTypeInfo", ret)
		ret = ""

		ret += "\nclass MethTypeInfo{\n"

		buildTBI()

		ret += "// mirrors reflect.ucommonType / reflect.method\n"
		ret += "public static var methByID:Array<{\n" // using a map makes it too big for Java
		ret += "\tid:Int,          // id of type\n"
		ret += "\tname:String,     // name of method\n"
		ret += "\tpkgPath:String,  // nil for exported Names; otherwise import path\n"
		ret += "\tmtyp:Int,        // method type (without receiver)\n"
		ret += "\ttyp:Int,         // .(*FuncType) underneath\n"
		//ret += "\tifn:Dynamic,     // fn used in interface call (one-word receiver)\n"
		//ret += "\ttfn:Dynamic,     // fn used for normal method call\n"
		ret += "\n}> = haxe.Json.parse(\"[ "
		firstTime := true
		for id, t := range typesByID {
			ms := types.NewMethodSet(t)
			if ms.Len() > 0 {
				for mth := 0; mth < ms.Len(); mth++ {
					methInfo := ms.At(mth)
					name := methInfo.Obj().Name()
					path := ""
					rune1, _ := utf8.DecodeRune([]byte(name))
					if unicode.IsLower(rune1) {
						path = methInfo.Obj().Pkg().Path()
					}
					if firstTime {
						firstTime = false
					} else {
						ret += ", "
					}
					ret += "{ "
					ret += fmt.Sprintf("\\\"id\\\": %d, ", id)
					ret += "\\\"name\\\": \\\"" + strings.Trim(haxeStringConst(`"`+name+`"`, "CompilerInternal:haxe.EmitTypeInfo()"), `"`) + "\\\", "
					ret += "\\\"pkgPath\\\": \\\"" + strings.Trim(haxeStringConst(`"`+path+`"`, "CompilerInternal:haxe.EmitTypeInfo()"), `"`) + "\\\", "
					ret += fmt.Sprintf("\\\"mtyp\\\": %s, ", pogo.LogTypeUse(methInfo.Type())) // TODO should be without Receiver???
					ret += fmt.Sprintf("\\\"typ\\\": %s ", pogo.LogTypeUse(methInfo.Type()))
					//if notInterface(t) {
					//	fnToCall := `Go_` + l.LangName(
					//		methInfo.Obj().Pkg().String()+":"+methInfo.Recv().String(),
					//		name) + `.call`
					//	ret += fmt.Sprintf("ifn: %s, ", fnToCall) // TODO check ok
					//	ret += fmt.Sprintf("tfn: %s, ", fnToCall) // TODO check ok
					//} else {
					//	ret += "ifn: null, tfn: null, "
					//}
					ret += "} "
				}
			}
		}
		ret += "]\");\n"
		ret += "}\n"

		pogo.WriteAsClass("MethTypeInfo", ret)
		ret = ""

		ret += "\nclass MapTypeInfo{\n"

		buildTBI()

		ret += "public static var mapByID:Map<Int,{key:Int,elem:Int}> = [ 0=>{key:0,elem:0}, \n"
		for id, t := range typesByID {
			switch t.(type) {
			case *types.Named:
				t = t.(*types.Named).Underlying()
			}
			switch t.(type) {
			case *types.Map:
				ret += fmt.Sprintf("\t%d=>{key:%s,elem:%s},\n", id,
					pogo.LogTypeUse(t.(*types.Map).Key()),
					pogo.LogTypeUse(t.(*types.Map).Elem()))
			}
		}
		ret += "];\n"
		ret += "}\n"

		pogo.WriteAsClass("MapTypeInfo", ret)
		ret = ""

		ret += "\nclass ChanTypeInfo{\n"

		buildTBI()

		ret += "public static var chanByID:Map<Int,{elem:Int,dir:Int}> = [ 0=>{elem:0,dir:0}, \n"
		for id, t := range typesByID {
			switch t.(type) {
			case *types.Named:
				t = t.(*types.Named).Underlying()
			}
			switch t.(type) {
			case *types.Chan:
				reflectDir := reflect.ChanDir(0)
				switch t.(*types.Chan).Dir() {
				case types.SendRecv:
					reflectDir = reflect.BothDir
				case types.SendOnly:
					reflectDir = reflect.SendDir
				case types.RecvOnly:
					reflectDir = reflect.RecvDir
				}
				ret += fmt.Sprintf("\t%d=>{elem:%s,dir:%d},\n", id,
					pogo.LogTypeUse(t.(*types.Chan).Elem()),
					reflectDir)
			}
		}
		ret += "];\n"
		ret += "}\n"

		pogo.WriteAsClass("ChanTypeInfo", ret)
		ret = ""

		ret += "\nclass FuncTypeInfo{\n"

		buildTBI()

		ret += "public static var funcByID:Map<Int,{ddd:Bool,pin:Array<Int>,pout:Array<Int>}> = [ 0=>{ddd:false,pin:[],pout:[]}, \n"
		for id, t := range typesByID {
			switch t.(type) {
			case *types.Named:
				t = t.(*types.Named).Underlying()
			}
			switch t.(type) {
			case *types.Signature:
				ret += fmt.Sprintf("\t%d=>{ddd:%v,pin:[", id, t.(*types.Signature).Variadic())
				for i := 0; i < t.(*types.Signature).Params().Len(); i++ {
					ret += fmt.Sprintf("%s,", pogo.LogTypeUse(t.(*types.Signature).Params().At(i).Type()))
				}
				ret += "],pout:["
				for o := 0; o < t.(*types.Signature).Results().Len(); o++ {
					ret += fmt.Sprintf("%s,", pogo.LogTypeUse(t.(*types.Signature).Results().At(o).Type()))
				}
				ret += fmt.Sprintf("]}, // %s \n", t.(*types.Signature).String())
			}
		}
		ret += "];\n"
		ret += "}\n"

		pogo.WriteAsClass("FuncTypeInfo", ret)
		ret = ""

		ret += "\nclass ArrayTypeInfo{\n"

		buildTBI()

		ret += "public static var arrayByID:Map<Int,{elem:Int,slice:Int,len:Int}> = [ 0=>{elem:0,slice:0,len:0}, \n"
		for id, t := range typesByID {
			var err error
			switch t.(type) {
			case *types.Named:
				t = t.(*types.Named).Underlying()
			}
			switch t.(type) {
			case *types.Array:
				slT := 0
				for ids, ts := range typesByID {
					switch ts.(type) {
					case *types.Slice:
						if pogo.LogTypeUse(ts.(*types.Slice).Elem()) == pogo.LogTypeUse(t.(*types.Array).Elem()) {
							slT = ids
							goto slTfound
						}
					}
				}
				//println("DEBUG making new slice for ", t.String())
				slT, err = strconv.Atoi(pogo.LogTypeUse(types.NewSlice(t.(*types.Array).Elem())))
				if err != nil {
					panic("haxe.EmitTypeInfo() correct slice type not created for array:" + t.String())
				}
			slTfound:
				ret += fmt.Sprintf("\t%d=>{elem:%s,slice:%d,len:%d},\n", id,
					pogo.LogTypeUse(t.(*types.Array).Elem()),
					slT, t.(*types.Array).Len())
			}
		}
		ret += "];\n"

		ret += "}\n"

		pogo.WriteAsClass("ArrayTypeInfo", ret)
		ret = ""

	}*/ // end of Reflect-releated type info

	/*if false {
		buildTBI()

		ret += "\nclass TypeInfoIDs{\n\n"

		ret += "// mirrors reflect.rtype etc\n"
		ret += "//\t[0] = isValid:Bool,\n"
		ret += "//\t[1] = size:Int,\n"
		ret += "//\t[2] = align:Int,\n"
		ret += "//\t[3] = fieldAlign:Int,\n"
		ret += "//\t[4] = kind:Int,\n"
		ret += "//\t[5] = stringForm:String,\n"
		ret += "//\t[6] = name:String,\n"
		ret += "//\t[7] = ptrToThis:Int,\n"
		ret += "//\t[8] = pkgPath:String,\n"
		ret += "//\t[9] = numMethods:Int,\n"
		ret += "public static var typesByID:Array<Array<Dynamic>>=haxe.Json.parse(\" [ "
		for _, t := range typesByID {
			kind, name := getTypeInfo(t, "")
			if t == nil || kind == reflect.Invalid {
				ret += "[false,0,0,0,0,\\\"\\\",\\\"\\\",0,\\\"\\\",0] " // the first one
			} else {
				ptrT := 0
				for idp, tp := range typesByID {
					switch tp.(type) {
					case *types.Pointer:
						if pogo.LogTypeUse(tp.(*types.Pointer).Elem()) == pogo.LogTypeUse(t) {
							ptrT = idp
							goto ptrTfound
						}
					}
				}
				//ret += "\t// correct pointer to type not found, set to 0\n"
			ptrTfound:
				pkgPath := ""
				numMethods := 0
				tt := t
				switch tt.(type) {
				case *types.Pointer:
					el, ok := tt.(*types.Pointer).Elem().(*types.Named)
					if ok {
						tt = el
					}
				}
				switch tt.(type) {
				case *types.Named:
					obj := tt.(*types.Named).Obj()
					if obj != nil {
						pkg := obj.Pkg()
						if pkg != nil {
							pkgPath = pkg.Path()
						}
					}
					numMethods = types.NewMethodSet(t).Len()
				}
				ret += ", ["
				ret += fmt.Sprintf(" true,")
				ret += fmt.Sprintf(" %d,", haxeStdSizes.Sizeof(t))
				ret += fmt.Sprintf(" %d,", haxeStdSizes.Alignof(t))
				ret += fmt.Sprintf(" %d,", haxeStdSizes.Alignof(t)) // TODO check correct for fieldAlign
				ret += fmt.Sprintf(" %d,", kind)
				ret += fmt.Sprintf(" \\\"%s\\\",",
					strings.Trim(haxeStringConst(`"`+preprocessTypeName(t.String())+`"`, "CompilerInternal:haxe.EmitTypeInfo()"), `"`))
				ret += fmt.Sprintf(" \\\"%s\\\",",
					strings.Trim(haxeStringConst(`"`+name+`"`, "CompilerInternal:haxe.EmitTypeInfo()"), `"`))
				ret += fmt.Sprintf(" %d,", ptrT)
				ret += fmt.Sprintf(" \\\"%s\\\",",
					strings.Trim(haxeStringConst(`"`+pkgPath+`"`, "CompilerInternal:haxe.EmitTypeInfo()"), `"`))
				ret += fmt.Sprintf(" %d", numMethods)
				ret += fmt.Sprintf("]")
			}
		}
		ret += "]\");\n}\n"

		pogo.WriteAsClass("TypeInfoIDs", ret)
		ret = ""

		buildTBI()
	}*/
	ret += "\nclass TypeInfo{\n\n"

	// TODO review if this is required
	ret += "public static function isHaxeClass(id:Int):Bool {\nswitch(id){" + "\n"
	for k := range pteKeys {
		v := pte.At(pteKeys[k])
		goType := pteKeys[k].String()
		//fmt.Println("DEBUG full goType", goType)
		haxeClass := getHaxeClass(goType)
		if haxeClass != "" {
			ret += "case " + fmt.Sprintf("%d", v) + `: return true; // ` + goType + "\n"
		}
	}
	ret += `default: return false;}}` + "\n"

	ret += "public static function getName(id:Int):String {\n"
	ret += "\tif(id<0||id>=nextTypeID)return \"reflect.CREATED\"+Std.string(id);\n"
	//for k, v := range typesByID {
	//	ret += "case " + fmt.Sprintf("%d", k) + `: return ` +
	//		haxeStringConst(`"`+preprocessTypeName(v.String())+`"`, "CompilerInternal:haxe.EmitTypeInfo()") +
	//		`;` + "\n"
	//}
	ret += "\tif(id==0)return \"(haxeTypeID=0)\";" + "\n"
	ret += "\t#if (js || php || node) if(id==null)return \"(haxeTypeID=null)\"; #end\n"
	//ret += "\t" + `return TypeInfoIDs.typesByID[id][5];` + "\n}\n"
	ret += "\t" + `return Go_haxegoruntime_getTTypeSString.hx(id);` + "\n}\n"
	ret += "public static function typeString(i:Interface):String {\nreturn getName(i.typ);\n}\n"

	ret += "static var typIDs:Map<String,Int> = ["
	deDup := make(map[string]bool)
	for k := range pteKeys {
		v := pte.At(pteKeys[k])
		nam := haxeStringConst("`"+preprocessTypeName(pteKeys[k].String())+"`", "CompilerInternal:haxe.EmitTypeInfo()")
		if len(nam) != 0 {
			if deDup[nam] { // have one already!!
				nam = fmt.Sprintf("%s (duplicate type name! this id=%d)\"", nam[:len(nam)-1], v)
			} else {
				deDup[nam] = true
			}
			ret += ` ` + nam + ` => ` + fmt.Sprintf("%d", v) + `,` + "\n"
		}
	}
	ret += "];\n"
	ret += "public static function getId(name:String):Int {\n"
	ret += "\tvar t:Int;\n"
	ret += "\ttry { t=typIDs[name];\n"
	ret += "\t} catch(x:Dynamic) { Scheduler.panicFromHaxe(\"TraceInfo.getId() not found:\"+name+x); t=-1; } ;\n"
	ret += "\treturn t;\n}\n"

	//emulation of: func IsAssignableTo(V, T Type) bool
	ret += "public static function isAssignableTo(v:Int,t:Int):Bool {\n\tif(v==t) return true;\n"
	ret += "\tfor(ae in isAsssignableToArray) if(ae==(v<<16|t)) return true;\n"
	ret += "\treturn false;\n}\n"

	ret += "static var isAsssignableToArray:Array<Int> = ["
	for V := range pteKeys {
		v := pte.At(pteKeys[V])
		for T := range pteKeys {
			t := pte.At(pteKeys[T])
			if v != t && types.AssignableTo(pteKeys[V], pteKeys[T]) {
				//ret0 += "\t" + `case ` + fmt.Sprintf("%d", t) + `: return true;` + "\n"
				ret += fmt.Sprintf("%d,", v.(int)<<16|t.(int))
			}
		}
		///if ret0 != "" {
		//	ret += `case ` + fmt.Sprintf("%d", v) + `: switch(t){` + "\n"
		//	ret += ret0
		//	ret += "default: return false;}\n"
		//}
		ret += "\n"
	}
	//ret += "default: return false;}}\n"
	ret += "];\n"

	//emulation of: func IsIdentical(x, y Type) bool
	ret += "public static function isIdentical(v:Int,t:Int):Bool {\nif(v==t) return true;\nswitch(v){" + "\n"
	for V := range pteKeys {
		v := pte.At(pteKeys[V])
		ret0 := ""
		for T := range pteKeys {
			t := pte.At(pteKeys[T])
			if v != t && types.Identical(pteKeys[V], pteKeys[T]) {
				ret0 += `case ` + fmt.Sprintf("%d", t) + `: return true;` + "\n"
			}
		}
		if ret0 != "" {
			ret += `case ` + fmt.Sprintf("%d", v) + `: switch(t){` + "\n"
			ret += ret0
			ret += "default: return false;}\n"
		}
	}
	ret += "default: return false;}}\n"

	//emulation of: func type.AsertableTo(V *Interface, T Type) bool
	ret += "public static function assertableTo(v:Int,t:Int):Bool {\n"
	//ret += "trace(\"DEBUG assertableTo()\",v,t);\n"
	ret += "\treturn isAssertableToMap[(v<<16)|t];\n"
	ret += "}\n"
	//	ret += "if(v==t) return true;\nswitch(v){" + "\n"
	ret += "static var isAssertableToMap:Map<Int,Bool> = [ 0 => false, "
	for tid, typ := range typesByID {
		ret0 := ""
		if typ != nil {
			for iid, ityp := range typesByID {
				named, isNamed := ityp.(*types.Named)
				if isNamed {
					iface, isIface := named.Underlying().(*types.Interface)
					if isIface {
						if tid != iid && types.AssertableTo(iface, typ) {
							ret0 += fmt.Sprintf("0x%X => true, ", (tid<<16)|iid)
							//ret0 += "\t" + `case ` + fmt.Sprintf("%d", iid) + `: return true; // ` + named.String() + "\n"
							//println("DEBUG type ", tid, typ.String(), " Implemets interface ", iid, named.String(), iface.String())
						}
					}
				}
			}
		}
		if ret0 != "" {
			//	ret += `case ` + fmt.Sprintf("%d", tid) + `: switch(t){ // ` + typ.String() + "\n"
			ret += ret0
			ret += "\n"
			//	ret += "\tdefault: return false;}\n"
		}
	}
	//	ret += "default: return false;}}\n"
	ret += "];\n"

	//function to answer the question is the type a concrete value?
	ret += "public static function isConcrete(t:Int):Bool {\nswitch(t){" + "\n"
	for T := range pteKeys {
		t := pte.At(pteKeys[T])
		switch pteKeys[T].Underlying().(type) {
		case *types.Interface:
			ret += `case ` + fmt.Sprintf("%d", t) + `: return false;` + "\n"
		default:
			//ret += `case ` + fmt.Sprintf("%d", t) + `: return true;` + "\n"
		}
	}
	ret += "default: return true;}}\n"

	// function to give the zero value for each type
	ret += "public static function zeroValue(t:Int):Dynamic {\nswitch(t){" + "\n"
	for T := range pteKeys {
		t := pte.At(pteKeys[T])
		z := l.LangType(pteKeys[T], true, "EmitTypeInfo()")
		if z == "" {
			z = "null"
		}
		if z != "null" {
			ret += `case ` + fmt.Sprintf("%d", t) + `: return `
			ret += z + ";\n"
		}
	}
	ret += "default: return null;}}\n"

	ret += fmt.Sprintf("public static var nextTypeID=%d;\n", pogo.NextTypeID) // must be last as will change during processing

	ret += "}\n"

	pogo.WriteAsClass("TypeInfo", ret)

	ret = "class MethodTypeInfo {"

	ret += "public static function method(t:Int,m:String):Dynamic {\nswitch(t){" + "\n"

	tta := pogo.TypesWithMethodSets() //[]types.Type
	sort.Sort(pogo.TypeSorter(tta))
	for T := range tta {
		t := pte.At(tta[T])
		if t != nil { // it is used?
			ret += `case ` + fmt.Sprintf("%d", t) + `: switch(m){` + "\n"
			ms := types.NewMethodSet(tta[T])
			msNames := []string{}
			for m := 0; m < ms.Len(); m++ {
				msNames = append(msNames, ms.At(m).String())
			}
			sort.Strings(msNames)
			deDup := make(map[string][]int) // TODO check this logic, required for non-public methods
			for pass := 1; pass <= 2; pass++ {
				for _, msString := range msNames {
					for m := 0; m < ms.Len(); m++ {
						if ms.At(m).String() == msString { // ensure we do this in a repeatable order
							funcObj, ok := ms.At(m).Obj().(*types.Func)
							pkgName := "unknown"
							if ok && funcObj.Pkg() != nil && ms.At(m).Recv() == tta[T] {
								line := ""
								ss := strings.Split(funcObj.Pkg().Name(), "/")
								pkgName = ss[len(ss)-1]
								if strings.HasPrefix(pkgName, "_") { // exclude functions in haxe for now
									// TODO NoOp for now... so haxe types cant be "Involked" when held in interface types
									// *** need to deal with getters and setters
									// *** also with calling parameters which are different for a Haxe API
								} else {
									switch pass {
									case 1:
										idx, exists := deDup[funcObj.Name()]
										if exists {
											if len(idx) > len(ms.At(m).Index()) {
												deDup[funcObj.Name()] = ms.At(m).Index()
											}
										} else {
											deDup[funcObj.Name()] = ms.At(m).Index()
										}
									case 2:
										idx, _ := deDup[funcObj.Name()]
										if len(idx) != len(ms.At(m).Index()) {
											line += "// Duplicate unused: "
										}
										line += `case "` + funcObj.Name() + `": return `
										fnToCall := l.LangName(
											ms.At(m).Obj().Pkg().Name()+":"+ms.At(m).Recv().String(),
											funcObj.Name())
										line += `Go_` + fnToCall + `.call` + "; "
									}
								}
								ret += line
							}
							if pass == 2 {
								ret += fmt.Sprintf("// %v %v %v %v\n",
									ms.At(m).Obj().Name(),
									ms.At(m).Kind(),
									ms.At(m).Index(),
									ms.At(m).Indirect())
							}
						}
					}
				}
			}
			ret += "default:}\n"
		}
	}

	// TODO look for overloaded types at this point

	ret += "default:}\n Scheduler.panicFromHaxe( " + `"no method found!"` + "); return null;}\n" // TODO improve error

	pogo.WriteAsClass("MethodTypeInfo", ret+"}\n")

	return ""
}

func fixKeyWds(w string) string {
	switch w {
	case "new":
		return w + "_"
	default:
		return w
	}
}

func loadStoreSuffix(T types.Type, hasParameters bool) string {
	if bt, ok := T.Underlying().(*types.Basic); ok {
		switch bt.Kind() {
		case types.Bool,
			types.Int8,
			types.Int16,
			types.Int64,
			types.Uint16,
			types.Uint64,
			types.Uintptr,
			types.Float32,
			types.Float64,
			types.Complex64,
			types.Complex128,
			types.String:
			return "_" + types.TypeString(nil, T) + "("
		case types.Uint8: // to avoid "byte"
			return "_uint8("
		case types.Int, types.Int32: // for int and to avoid "rune"
			return "_int32("
		case types.Uint, types.Uint32:
			return "_uint32("
		}
	}
	if _, ok := T.Underlying().(*types.Array); ok {
		ret := fmt.Sprintf("_object(%d", haxeStdSizes.Sizeof(T))
		if hasParameters {
			ret += ","
		}
		return ret
	}
	if _, ok := T.Underlying().(*types.Struct); ok {
		ret := fmt.Sprintf("_object(%d", haxeStdSizes.Sizeof(T))
		if hasParameters {
			ret += ","
		}
		return ret
	}
	return "(" // no suffix, so some dynamic type
}

// Type definitions are only carried through to Haxe to allow access to objects as if they were native Haxe classes.
// TODO consider renaming
func (l langType) TypeStart(nt *types.Named, err string) string {
	typName := "GoType" + l.LangName("", nt.String())
	hxTyp := l.LangType(nt.Obj().Type(), false, nt.String())
	ret := ""
	switch hxTyp {
	case "Object":
		ret += "class " + typName
		ret += " extends " + hxTyp + " {\n"
	default:
		ret += "abstract " + typName + "(" + hxTyp + ") from " + hxTyp + " to " + hxTyp + " {\n"
	}
	switch nt.Underlying().(type) {
	case *types.Struct:
		str := nt.Underlying().(*types.Struct)
		ret += "inline public function new(){ super new(" + strconv.Itoa(int(haxeStdSizes.Sizeof(nt.Obj().Type()))) + "); }\n"
		flds := []string{}
		for f := 0; f < str.NumFields(); f++ {
			fName := str.Field(f).Name()
			if len(fName) > 0 {
				if unicode.IsUpper(rune(fName[0])) {
					flds = append(flds, fName)
				}
			}
		}
		sort.Strings(flds) // make sure the fields are always in the same order in the file
		for _, fName := range flds {
			for f := 0; f < str.NumFields(); f++ {
				if fName == str.Field(f).Name() {
					haxeTyp := l.LangType(str.Field(f).Type(), false, nt.String())
					fOff := fieldOffset(str, f)
					sfx := loadStoreSuffix(str.Field(f).Type(), true)
					ret += fmt.Sprintf("public var _%s(get,set):%s;\n", fName, haxeTyp)
					ret += fmt.Sprintf("function get__%s():%s { return get%s%d); }\n",
						fName, haxeTyp, sfx, fOff)
					ret += fmt.Sprintf("function set__%s(v:%s):%s { return set%s%d,v); }\n",
						fName, haxeTyp, haxeTyp, sfx, fOff)
					break
				}
			}
		}
	case *types.Array:
		ret += "inline public function new(){ super new(" + strconv.Itoa(int(haxeStdSizes.Sizeof(nt.Obj().Type()))) + "); }\n"
	default: // TODO not yet sure how to handle named types that are not structs
		ret += "inline public function new(v:" + hxTyp + ") { this = v; }\n"
	}

	meths := []string{}
	for m := 0; m < nt.NumMethods(); m++ {
		mName := nt.Method(m).Name()
		if len(mName) > 0 {
			if unicode.IsUpper(rune(mName[0])) {
				meths = append(meths, mName)
			}
		}
	}
	sort.Strings(meths) // make sure the methods always appear in the same order in the file
	for _, mName := range meths {
		for m := 0; m < nt.NumMethods(); m++ {
			meth := nt.Method(m)
			if mName == meth.Name() {
				sig := meth.Type().(*types.Signature)
				ret += "// " + mName + " " + sig.String() + "\n"
				ret += "public function _" + mName + "("
				for p := 0; p < sig.Params().Len(); p++ {
					if p > 0 {
						ret += ","
					}
					ret += "_" + sig.Params().At(p).Name() + ":" + l.LangType(sig.Params().At(p).Type(), false, nt.String())
				}
				ret += ")"
				switch sig.Results().Len() {
				case 0:
					ret += ":Void "
				case 1:
					ret += ":" + l.LangType(sig.Results().At(0).Type(), false, nt.String())
				default:
					ret += ":{"
					for r := 0; r < sig.Results().Len(); r++ {
						if r > 0 {
							ret += ","
						}
						ret += fmt.Sprintf("r%d:%s", r, l.LangType(sig.Results().At(r).Type(), false, nt.String()))
					}
					ret += "}"
				}
				ret += "{\n\t"
				if sig.Results().Len() > 0 {
					ret += "return "
				}
				fnToCall := l.LangName(
					nt.Obj().Pkg().Name()+":"+sig.Recv().Type().String(),
					meth.Name())
				ret += `Go_` + fnToCall + `.hx(this`
				for p := 0; p < sig.Params().Len(); p++ {
					ret += ", _" + sig.Params().At(p).Name()
				}
				ret += ");\n}\n"
			}
		}
	}

	pogo.WriteAsClass(typName, ret+"}\n")

	return "" //ret
}

//func (langType) TypeEnd(nt *types.Named, err string) string {
//	return "" //"}"
//}
