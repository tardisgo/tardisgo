// +build haxe

package reflect

import (
	"haxegoruntime"
	"unsafe"

	"github.com/tardisgo/tardisgo/haxe/hx"
)

// the haxe-specific parts

var haxeIDmap = make(map[int]unsafe.Pointer)

func findHaxeID(p unsafe.Pointer) int {
	for k, v := range haxeIDmap {
		if p == v {
			return k
		}
	}
	panic("reflect.findHaxeID - type not found")
	return 0
}

func createHaxeType(id int) *rtype {
	if id <= 0 || id >= hx.GetInt("", "TypeInfo.nextTypeID") {
		//panic("reflect.createHaxeType() invalid haxe id: " + hx.CallString("", "Std.string", 1, id))
		return nil
	}

	// new version of type infomation:
	return (*rtype)(unsafe.Pointer(haxegoruntime.TypeTable[id]))
}

func haxeInterfaceUnpack(i interface{}) *emptyInterface {
	ret := new(emptyInterface)
	if i == nil {
		//panic("reflect.haxeInterfaceUnpack() nil value")
		return ret
	}
	haxeTypeID := hx.CodeInt("", "_a.itemAddr(0).load().typ;", i)
	ret.typ = createHaxeType(haxeTypeID)
	//if ret.typ == nil {
	//	panic("reflect.haxeInterfaceUnpack() nil type pointer for HaxeTypeID " +
	//		hx.CallString("", "Std.string", 1, haxeTypeID))
	//}

	haxe2go(ret, i)
	return ret
}

func haxe2go(ret *emptyInterface, i interface{}) {
	/*
		if ret.typ == nil { // HELP! assume dynamic?
			panic("DEBUG reflect.haxe2go() nil Go type")
			//println("DEBUG reflect.haxe2go() nil Go type, using uintptr")
			//uip := hx.CodeIface("", "uintptr", "null;")
			//ret.typ = createHaxeType(hx.CodeInt("", "_a.itemAddr(0).load().typ;", uip))
		}
	*/
	ret.word = hx.Malloc(ret.typ.size)

	switch ret.typ.Kind() & kindMask {
	case Bool:
		*(*bool)(ret.word) = hx.CodeBool("", "_a.itemAddr(0).load().val;", i)
	case Int:
		*(*int)(ret.word) = hx.CodeInt("", "_a.itemAddr(0).load().val;", i)
	case Int8:
		*(*int8)(ret.word) = int8(hx.CodeInt("", "_a.itemAddr(0).load().val;", i))
	case Int16:
		*(*int16)(ret.word) = int16(hx.CodeInt("", "_a.itemAddr(0).load().val;", i))
	case Int32:
		*(*int32)(ret.word) = int32(hx.CodeInt("", "_a.itemAddr(0).load().val;", i))
	case Int64:
		*(*int64)(ret.word) = hx.Int64(hx.CodeDynamic("", "_a.itemAddr(0).load().val;", i))
	case Uint:
		*(*uint)(ret.word) = uint(hx.CodeInt("", "_a.itemAddr(0).load().val;", i))
	case Uint8:
		*(*uint8)(ret.word) = uint8(hx.CodeInt("", "_a.itemAddr(0).load().val;", i))
	case Uint16:
		*(*uint16)(ret.word) = uint16(hx.CodeInt("", "_a.itemAddr(0).load().val;", i))
	case Uint32:
		*(*uint32)(ret.word) = uint32(hx.CodeInt("", "_a.itemAddr(0).load().val;", i))
	case Uint64:
		*(*uint64)(ret.word) = uint64(hx.Int64(hx.CodeDynamic("", "_a.itemAddr(0).load().val;", i)))
	case Uintptr:
		*(*uintptr)(ret.word) = uintptr(hx.CodeDynamic("", "_a.itemAddr(0).load().val;", i))
	case Float32:
		*(*float32)(ret.word) = float32(hx.CodeFloat("", "_a.itemAddr(0).load().val;", i))
	case Float64:
		*(*float64)(ret.word) = float64(hx.CodeFloat("", "_a.itemAddr(0).load().val;", i))
	case Complex64:
		*(*complex64)(ret.word) = complex64(hx.Complex(hx.CodeDynamic("", "_a.itemAddr(0).load().val;", i)))
	case Complex128:
		*(*complex128)(ret.word) = hx.Complex(hx.CodeDynamic("", "_a.itemAddr(0).load().val;", i))
	case UnsafePointer, Ptr:
		//*(*unsafe.Pointer)(ret.word) = unsafe.Pointer(hx.CodeDynamic("", "_a.itemAddr(0).load().val;", i))
		ret.word = unsafe.Pointer(hx.CodeDynamic("", "_a.itemAddr(0).load().val;", i))

	case String:
		*(*string)(ret.word) = hx.CodeString("", "_a.itemAddr(0).load().val;", i)

	case Array, Struct:
		hx.Code("",
			"_a.itemAddr(1).load().val.store_object(_a.itemAddr(2).load().val,_a.itemAddr(0).load().val);",
			i, ret.word, ret.typ.size)

	case Slice, Interface, Map, Func, Chan:
		val := hx.CodeDynamic("", "_a.itemAddr(0).load().val;", i)
		*(*uintptr)(ret.word) = val

		/*
			htyp := "null"
			if !hx.IsNull(val) {
				htyp = hx.CallString("", "Type.getClassName", 1, val)
			}
			println("DEBUG unpack haxe type=", htyp, " Go type=", ret.typ.Kind().String(), "val=", val, "encoded=", ret)
		*/
	}

}

func typeIdFromPtr(ptr *rtype) int {
	for typ := 0; typ < len(haxegoruntime.TypeTable); typ++ {
		if unsafe.Pointer(ptr) == unsafe.Pointer(haxegoruntime.TypeTable[typ]) {
			return typ
		}
	}
	return 0
}
func haxeInterfacePack(ei *emptyInterface) interface{} {
	i := haxeInterfacePackB(ei)

	ityp := hx.CodeInt("", "_a.itemAddr(0).load().typ;", i)
	if unsafe.Pointer(ei.typ) != unsafe.Pointer(haxegoruntime.TypeTable[ityp]) {
		typ := typeIdFromPtr(ei.typ)
		//println("DEBUG typ(new)!=ityp(old) ", typ, ityp)
		hx.Code("",
			"var x=cast(_a.itemAddr(0).load(),Interface); x.typ=_a.itemAddr(1).load().val; _a.itemAddr(2).store(x); ",
			i, typ, &i)
		//println("DEBUG amended to ", hx.CodeInt("", "_a.itemAddr(0).load().typ;", i))
	}

	return i
}

func haxeInterfacePackB(ei *emptyInterface) interface{} {

	// TODO deal with created types, if any ?

	switch ei.typ.Kind() {
	case Bool:
		return hx.CodeIface("", ei.typ.String(), "_a.itemAddr(0).load().val;", *(*bool)(ei.word))
	case Int:
		return hx.CodeIface("", ei.typ.String(), "_a.itemAddr(0).load().val;", *(*int)(ei.word))
	case Int8:
		return hx.CodeIface("", ei.typ.String(), "_a.itemAddr(0).load().val;", *(*int8)(ei.word))
	case Int16:
		return hx.CodeIface("", ei.typ.String(), "_a.itemAddr(0).load().val;", *(*int16)(ei.word))
	case Int32:
		return hx.CodeIface("", ei.typ.String(), "_a.itemAddr(0).load().val;", *(*int32)(ei.word))
	case Int64:
		return hx.CodeIface("", ei.typ.String(), "_a.itemAddr(0).load().val;", *(*int64)(ei.word))
	case Uint:
		return hx.CodeIface("", ei.typ.String(), "_a.itemAddr(0).load().val;", *(*uint)(ei.word))
	case Uint8:
		return hx.CodeIface("", ei.typ.String(), "_a.itemAddr(0).load().val;", *(*uint8)(ei.word))
	case Uint16:
		return hx.CodeIface("", ei.typ.String(), "_a.itemAddr(0).load().val;", *(*uint16)(ei.word))
	case Uint32:
		return hx.CodeIface("", ei.typ.String(), "_a.itemAddr(0).load().val;", *(*uint32)(ei.word))
	case Uint64:
		return hx.CodeIface("", ei.typ.String(), "_a.itemAddr(0).load().val;", *(*uint64)(ei.word))
	case Uintptr:
		return hx.CodeIface("", ei.typ.String(), "_a.itemAddr(0).load().val;", *(*uintptr)(ei.word))
	case Float32:
		return hx.CodeIface("", ei.typ.String(), "_a.itemAddr(0).load().val;", *(*float32)(ei.word))
	case Float64:
		return hx.CodeIface("", ei.typ.String(), "_a.itemAddr(0).load().val;", *(*float64)(ei.word))
	case Complex64:
		return hx.CodeIface("", ei.typ.String(), "_a.itemAddr(0).load().val;", *(*complex64)(ei.word))
	case Complex128:
		return hx.CodeIface("", ei.typ.String(), "_a.itemAddr(0).load().val;", *(*complex128)(ei.word))
	case UnsafePointer:
		return hx.CodeIface("", ei.typ.String(), "_a.itemAddr(0).load().val;", ei.word)

	case String:
		//println("DEBUG string pack=", *(*string)(ei.word))
		return hx.CodeIface("", ei.typ.String(), "_a.itemAddr(0).load().val;", *(*string)(ei.word))

	case Ptr:
		p := hx.CodeIface("", ei.typ.String(), "_a.itemAddr(0).load().val;", ei.word)
		//println("DEBUG ptr pack p=", p)
		return p

	case Array, Struct:
		return hx.CodeIface("", ei.typ.String(),
			"_a.itemAddr(0).load().val.load_object(_a.itemAddr(1).load().val);",
			ei.word, ei.typ.size)

	case Slice, Map, Interface, Func, Chan:
		if ei.word == nil {
			return hx.CodeIface("", ei.typ.String(), "null;")
		}
		/*
			htyp := "null"
			if !hx.IsNull(uintptr(ei.word)) {
				htyp = hx.CallString("", "Type.getClassName", 1, ei.word)
			}
			println("DEBUG pack haxe type=", htyp, " Go type=", ei.typ.Kind().String(), "val=", ei.word, "encoded=", ei)
		*/
		val := *(*uintptr)(unsafe.Pointer(ei.word))
		r := hx.CodeIface("", ei.typ.String(), "_a.itemAddr(0).load().val;", val)
		//println("DEBUG pack haxe encoded=", ei, "type=", hx.CallString("", "Type.getClassName", 1, ei.word),
		//	"Go type=", ei.typ.Kind().String(), "PtrVal=", ei.word, "Return=", r)
		return r
	}

	panic("reflect.haxeInterfacePack() not yet implemented for " + ei.typ.String() +
		" Kind= " + ei.typ.Kind().String())

	return interface{}(nil)
}

func haxeUnsafeNew(rtp *rtype) unsafe.Pointer {
	return hx.Malloc(rtp.size)
}
