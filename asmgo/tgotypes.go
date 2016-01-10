package asmgo

import (
	"fmt"
	"reflect"
	"unicode"
	"unicode/utf8"

	"go/types"
	//"golang.org/x/tools/go/ssa"
)

const ( // from reflect package
	kindDirectIface = 1 << 5
	kindGCProg      = 1 << 6 // Type.gc points to GC program
	kindNoPointers  = 1 << 7
	kindMask        = (1 << 5) - 1
)

func escapedTypeString(s string) string {
	buf := []byte(s)
	r := ""
	for _, ch := range buf {
		r += fmt.Sprintf("\\x%x", ch)
	}
	return r
}

func synthTypesFor(t types.Type) {}

func getTypeInfo(t types.Type, tname string) (kind reflect.Kind, name string) {
	if t == nil {
		return reflect.Invalid, ""
	}
	if tname != "" {
		name = tname
	}
	switch t.(type) {
	case *types.Basic:
		tb := t.(*types.Basic)
		switch tb.Kind() {
		case types.Bool:
			kind = reflect.Bool
		case types.Int:
			kind = reflect.Int
		case types.Int8:
			kind = reflect.Int8
		case types.Int16:
			kind = reflect.Int16
		case types.Int32:
			kind = reflect.Int32
		case types.Int64:
			kind = reflect.Int64
		case types.Uint:
			kind = reflect.Uint
		case types.Uint8:
			kind = reflect.Uint8
		case types.Uint16:
			kind = reflect.Uint16
		case types.Uint32:
			kind = reflect.Uint32
		case types.Uint64:
			kind = reflect.Uint64
		case types.Uintptr:
			kind = reflect.Uintptr
		case types.Float32:
			kind = reflect.Float32
		case types.Float64:
			kind = reflect.Float64
		case types.Complex64:
			kind = reflect.Complex64
		case types.Complex128:
			kind = reflect.Complex128
		case types.UnsafePointer:
			kind = reflect.UnsafePointer
		case types.String:
			kind = reflect.String
		case types.UntypedBool, types.UntypedComplex, types.UntypedFloat, types.UntypedInt,
			types.UntypedNil, types.UntypedRune, types.UntypedString, types.Invalid:
			kind = reflect.Invalid
		default:
			panic(fmt.Sprintf("haxe.getTypeinfo() unhandled basic kind: %s", tb.String()))
		}

	case *types.Array:
		kind = reflect.Array
	case *types.Chan:
		kind = reflect.Chan
	case *types.Signature:
		kind = reflect.Func
	case *types.Interface:
		kind = reflect.Interface
	case *types.Map:
		kind = reflect.Map
	case *types.Pointer:
		kind = reflect.Ptr
	case *types.Slice:
		kind = reflect.Slice
	case *types.Struct:
		kind = reflect.Struct
	case *types.Named:
		if tname == "" {
			tname = t.(*types.Named).Obj().Name() // only do this for the top-level type name
		}
		return getTypeInfo(t.Underlying(), tname)
	case *types.Tuple:
		kind = reflect.Invalid
	default:
		panic(fmt.Sprintf("haxe.getTypeinfo() unhandled type: %T", t))

	}

	switch kind {
	case reflect.UnsafePointer, reflect.Ptr,
		reflect.Map, reflect.Chan, reflect.Func: // TODO not sure about these three
	default:
		kind |= kindNoPointers
	}

	// TODO work out when/if to set kindDirect
	switch kind & kindMask {
	case reflect.UnsafePointer, reflect.Ptr,
		reflect.Map, reflect.Chan, reflect.Func: // TODO not sure about these three
		kind |= kindDirectIface
	default:
	}

	return
}

func (l langType) BuildTypeHaxe() string {

	l.buildTBI()
	for i, t := range l.hc.typesByID {
		if i > 0 {
			synthTypesFor(t)
		}
	}
	l.buildTBI()

	ret := "class Tgotypes {\n"

	for i, t := range l.hc.typesByID {
		if i > 0 {
			ret += l.typeBuild(i, t)
		}
	}

	ret += "public static function setup() {\nvar a=Go.haxegoruntime_TTypeTTable.load();\n"
	ret += "var b=a.baseArray.obj;\nvar f=a.baseArray.off+a.itemOff(0);\nvar s=a.itemOff(1)-a.itemOff(0);\n"
	for i := range l.hc.typesByID {
		if i > 0 {
			//fmt.Println("DEBUG setup",i,t)
			ret += fmt.Sprintf(
				"b.set((%d*s)+f,type%d());\n",
				i, i)
		}
	}

	ret += "}\n" + "}\n"

	l.PogoComp().WriteAsClass("Tgotypes", ret)

	//fmt.Println("DEBUG generated Haxe code:", ret)

	return ret
}

func (l langType) typeBuild(i int, t types.Type) string {
	sizes := &haxeStdSizes
	ret := fmt.Sprintf( // sizeof largest struct (funcType) is 76
		"private static var type%dptr:Pointer=null; // %s\npublic static function type%d():Pointer { if(type%dptr==null) { type%dptr=Pointer.make(Object.make(80));",
		i, t.String(), i, i, i)
	ret += ""

	name := ""
	if namedT, named := t.(*types.Named); named {
		name = namedT.Obj().Name()
	}
	if basic, isBasic := t.(*types.Basic); isBasic {
		name = basic.Name()
	}
	rtype, kind := l.rtypeBuild(i, sizes, t, name)

	switch t.(type) {
	case *types.Named:
		t = t.(*types.Named).Underlying()
	}

	switch kind & kindMask {
	case reflect.Invalid, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.String, reflect.UnsafePointer:
		ret += fmt.Sprintf("Go_haxegoruntime_fillRRtype.callFromRT(0,type%dptr,%s)", i, rtype)

	case reflect.Ptr:
		ret += fmt.Sprintf("Go_haxegoruntime_fillPPtrTType.callFromRT(0,type%dptr,\n/*rtype:*/ ", i) + rtype + ",\n"
		if l.hc.pte.At(t.(*types.Pointer).Elem()) == nil {
			ret += fmt.Sprintf("/*elem:*/ nil,\n")
		} else {
			ret += fmt.Sprintf("/*elem:*/ type%d()\n",
				l.hc.pte.At(t.(*types.Pointer).Elem()).(int))
		}
		ret += ")"

	case reflect.Array:
		ret += fmt.Sprintf("Go_haxegoruntime_fillAArrayTType.callFromRT(0,type%dptr,\n/*rtype:*/ ", i) + rtype + ",\n"
		ret += fmt.Sprintf("/*elem:*/ type%d(),\n",
			l.hc.pte.At(t.(*types.Array).Elem()).(int))
		asl := "null" // slice type
		for _, tt := range l.hc.pte.Keys() {
			slt, isSlice := tt.(*types.Slice)
			if isSlice {
				if l.hc.pte.At(slt.Elem()) == l.hc.pte.At(t.(*types.Array).Elem()) {
					asl = fmt.Sprintf("type%d()",
						l.hc.pte.At(slt).(int))
					break
				}
			}
		}
		// TODO make missing slice types before we start outputting types to avoid not having one?
		ret += fmt.Sprintf("/*slice:*/ %s,\n", asl)
		ret += fmt.Sprintf("/*len:*/ %d\n", t.(*types.Array).Len())
		ret += ")"

	case reflect.Slice:
		ret += fmt.Sprintf("Go_haxegoruntime_fillSSliceTType.callFromRT(0,type%dptr,\n/*rtype:*/ ", i) + rtype + ",\n"
		ret += fmt.Sprintf("/*elem:*/ type%d()\n", l.hc.pte.At(t.(*types.Slice).Elem()).(int))
		ret += ")"

	case reflect.Struct:
		fields := []*types.Var{}
		for fld := 0; fld < t.(*types.Struct).NumFields(); fld++ {
			fldInfo := t.(*types.Struct).Field(fld)
			//if fldInfo.IsField() {
			fields = append(fields, fldInfo)
			//}
		}
		offs := sizes.Offsetsof(fields)

		ret += fmt.Sprintf("Go_haxegoruntime_fillSStructTType.callFromRT(0,type%dptr,\n/*rtype:*/ ", i) + rtype + ",\n/*fields:*/ "
		fret := "Go_haxegoruntime_newSStructFFieldSSlice.callFromRT(0)"
		numFlds := t.(*types.Struct).NumFields()
		for fld := 0; fld < numFlds; fld++ {
			fldInfo := t.(*types.Struct).Field(fld)
			name := fldInfo.Name()
			path := fldInfo.Pkg().Path()
			if fldInfo.Exported() {
				path = ""
			}
			if fldInfo.Anonymous() {
				name = ""
			}

			fret = "\tGo_haxegoruntime_addSStructFFieldSSlice.callFromRT(0," + fret + ","
			fret += "\n\t\t/*name:*/ \"" + name + "\",\n"
			fret += "\t\t/*pkgPath:*/ \"" + path + "\",\n"
			fret += fmt.Sprintf("\t\t/*typ:*/ type%d(),// %s\n", l.hc.pte.At(fldInfo.Type()), fldInfo.Type().String())
			fret += "\t\t/*tag:*/ \"" + escapedTypeString(t.(*types.Struct).Tag(fld)) + "\", // " + t.(*types.Struct).Tag(fld) + "\n"
			fret += fmt.Sprintf("\t\t/*offset:*/ %d\n", offs[fld])

			fret += "\t)"
		}

		ret += fret + ")"

	case reflect.Interface:
		ret += fmt.Sprintf("Go_haxegoruntime_fillIInterfaceTType.callFromRT(0,type%dptr,\n/*rtype:*/ ", i) + rtype + ",\n/*methods:*/ "
		mret := "Go_haxegoruntime_newIImethodSSlice.callFromRT(0)"
		for m := 0; m < t.(*types.Interface).NumMethods(); m++ {
			meth := t.(*types.Interface).Method(m)
			mret = "Go_haxegoruntime_addIImethodSSlice.callFromRT(0," + mret + ","
			mret += "\t\t/*name:*/ \"" + meth.Name() + "\",\n"
			path := "\"\""
			if !meth.Exported() {
				path = "\"" + meth.Pkg().Path() + "\""
			}
			mret += "\t\t/*pkgPath:*/ " + path + ",\n"
			typ := "null"
			iface := l.hc.pte.At(meth.Type())
			if iface != interface{}(nil) {
				typ = fmt.Sprintf("type%d()", iface.(int))
			}
			mret += fmt.Sprintf("\t\t/*typ:*/ %s // %s\n", typ, meth.String())
			mret += "\t)\n"
		}
		ret += mret + ")"

	case reflect.Map:
		ret += fmt.Sprintf("Go_haxegoruntime_fillMMapTType.callFromRT(0,type%dptr,\n/*rtype:*/ ", i) + rtype + ",\n"
		ret += fmt.Sprintf("/*key:*/ type%d(),\n",
			l.hc.pte.At(t.(*types.Map).Key()).(int))
		ret += fmt.Sprintf("/*elem:*/ type%d()\n",
			l.hc.pte.At(t.(*types.Map).Elem()).(int))
		ret += ")"

	case reflect.Func:
		ret += fmt.Sprintf("Go_haxegoruntime_fillFFuncTType.callFromRT(0,type%dptr,\n/*rtype:*/ ", i) + rtype + ",\n"
		ret += fmt.Sprintf("/*dotdotdot:*/ %v,\n", t.(*types.Signature).Variadic())
		ret += "/*in:*/ "
		iret := "Go_haxegoruntime_newPPtrTToRRtypeSSlice.callFromRT(0)"
		for i := 0; i < t.(*types.Signature).Params().Len(); i++ {
			iret = fmt.Sprintf("Go_haxegoruntime_addPPtrTToRRtypeSSlice.callFromRT(0,%s,\n\ttype%d())", iret,
				l.hc.pte.At((t.(*types.Signature).Params().At(i).Type())).(int))
		}
		ret += iret + ",\n/*out:*/  "
		oret := "Go_haxegoruntime_newPPtrTToRRtypeSSlice.callFromRT(0)"
		for o := 0; o < t.(*types.Signature).Results().Len(); o++ {
			oret = fmt.Sprintf("Go_haxegoruntime_addPPtrTToRRtypeSSlice.callFromRT(0,%s,\n\ttype%d())", oret,
				l.hc.pte.At((t.(*types.Signature).Results().At(o).Type())).(int))
		}
		ret += oret + " )\n"

	case reflect.Chan:
		ret += fmt.Sprintf("Go_haxegoruntime_fillCChanTType.callFromRT(0,type%dptr,\n/*rtype:*/ ", i) + rtype + ",\n"
		ret += fmt.Sprintf("/*elem:*/ type%d(),\n",
			l.hc.pte.At(t.(*types.Chan).Elem()).(int))
		reflectDir := reflect.ChanDir(0)
		switch t.(*types.Chan).Dir() {
		case types.SendRecv:
			reflectDir = reflect.BothDir
		case types.SendOnly:
			reflectDir = reflect.SendDir
		case types.RecvOnly:
			reflectDir = reflect.RecvDir
		}
		ret += fmt.Sprintf("/*dir:*/ %d\n", reflectDir)
		ret += ")"

	default:
		panic("unhandled reeflect.type")
	}

	ret += ";"
	ret += fmt.Sprintf("}; return type%dptr; }\n", i)
	return ret
}
func (l langType) rtypeBuild(i int, sizes types.Sizes, t types.Type, name string) (string, reflect.Kind) {
	var kind reflect.Kind
	kind, name = getTypeInfo(t, name)
	sof := int64(4)
	aof := int64(4)
	if kind != reflect.Invalid {
		sof = sizes.Sizeof(t)
		aof = sizes.Alignof(t)
	}

	ret := "Go_haxegoruntime_newRRtype.callFromRT(0,\n"
	ret += fmt.Sprintf("\t/*size:*/ %d,\n", sof)
	ret += fmt.Sprintf("\t/*align:*/ %d,\n", aof)
	ret += fmt.Sprintf("\t/*fieldAlign:*/ %d,\n", aof) // TODO check correct for fieldAlign
	ret += fmt.Sprintf("\t/*kind:*/ %d, // %s\n", kind, (kind & ((1 << 5) - 1)).String())
	alg := "false"
	if types.Comparable(t) {
		alg = "true"
	}
	ret += fmt.Sprintf("\t/*comprable:*/ %s,\n", alg) // TODO change this to be the actual function
	ret += fmt.Sprintf("\t/*string:*/ \"%s\", // %s\n", escapedTypeString(t.String()), t.String())
	ret += fmt.Sprintf("\t/*uncommonType:*/ %s,\n", l.uncommonBuild(i, sizes, name, t))
	ptt := "null"
	for pti, pt := range l.hc.typesByID {
		_, isPtr := pt.(*types.Pointer)
		if isPtr {
			ele := l.hc.pte.At(pt.(*types.Pointer).Elem())
			if ele != nil {
				if i == ele.(int) {
					ptt = fmt.Sprintf("type%d()", pti)
				}
			}
		}
	}
	ret += fmt.Sprintf("\t/*ptrToThis:*/ %s", ptt)
	ret += ")"
	return ret, kind
}

func (l langType) uncommonBuild(i int, sizes types.Sizes, name string, t types.Type) string {
	pkgPath := ""
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
	}

	var methods *types.MethodSet
	numMethods := 0
	methods = l.PogoComp().MethodSetFor(t)
	numMethods = methods.Len()
	if name != "" || numMethods > 0 {
		ret := "Go_haxegoruntime_newPPtrTToUUncommonTType.callFromRT(0,\n"
		ret += "\t\t/*name:*/ \"" + name + "\",\n"
		ret += "\t\t/*pkgPath:*/ \"" + pkgPath + "\",\n"
		ret += "\t\t/*methods:*/ "
		meths := "Go_haxegoruntime_newMMethodSSlice.callFromRT(0)"
		//_, isIF := t.Underlying().(*types.Interface)
		//if !isIF {
		for m := 0; m < numMethods; m++ {
			sel := methods.At(m)
			ssaFn := l.PogoComp().RootProgram().MethodValue(sel)
			if l.PogoComp().FnIsCalled(ssaFn) {
				fn := "null"
				fnToCall := "null"
				var name, str, path string
				fid, haveFn := l.hc.pte.At(sel.Obj().Type()).(int)
				if haveFn {
					fn = fmt.Sprintf("type%d()", fid)
				}
				name = sel.Obj().Name()
				str = sel.String()
				funcObj, ok := sel.Obj().(*types.Func)
				if ok {
					pn := "unknown"
					if funcObj.Pkg() != nil {
						pn = sel.Obj().Pkg().Name()
						path = sel.Obj().Pkg().Path()
					}
					fnToCall = `Go_` + l.LangName(
						pn+":"+sel.Recv().String(),
						funcObj.Name())
				}

				// now write out the method information
				meths = "Go_haxegoruntime_addMMethod.callFromRT(0," + meths + ",\n"
				meths += fmt.Sprintf("\n\t\t\t/*name:*/ \"%s\", // %s\n", name, str)
				rune1, _ := utf8.DecodeRune([]byte(name))
				if unicode.IsUpper(rune1) {
					path = ""
				}

				meths += fmt.Sprintf("\t\t\t/*pkgPath:*/ \"%s\",\n", path)
				// TODO should the two lines below be different?
				meths += fmt.Sprintf("\t\t\t/*mtyp:*/ %s,\n", fn)
				meths += fmt.Sprintf("\t\t\t/*typ:*/ %s,\n", fn)
				// add links to the functions ...

				if l.hc.funcNamesUsed[fnToCall] {
					fnToCall += ".call"
				} else {
					//println("DEBUG uncommonBuild function name not found: ", fnToCall)
					fnToCall = "null /* " + fnToCall + " */ "
				}
				meths += "\t\t\t" + fnToCall + "," + fnToCall + ")"
			}
		}
		//}
		ret += meths
		return ret + "\t)"
	}
	return "null"
}
