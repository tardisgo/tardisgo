package reflect

import (
	"unsafe"

	"github.com/tardisgo/tardisgo/haxe/hx"
)

// the haxe-specific parts

var haxeIDmap = make(map[int]Type)

func haxeInterfaceUnpack(i interface{}) *emptyInterface {
	if i == nil {
		panic("reflect.haxeInterfaceUnpack() nil value")
	}
	typInfo := hx.CodeDynamic("", "TypeInfo.typesByID[_a.itemAddr(0).load().typ];", i)

	/*
		println("haxeInterfaceUnpack DEBUG trace:",
			hx.CodeInt("", "_a.itemAddr(0).load().typ;", i),
			hx.CodeDynamic("", "_a.itemAddr(0).load().val;", i),
			hx.FgetBool("", typInfo, "", "isValid"),
			hx.FgetInt("", typInfo, "", "size"),
			hx.FgetInt("", typInfo, "", "align"),
			hx.FgetInt("", typInfo, "", "kind"),
			hx.FgetString("", typInfo, "", "name"),
			hx.FgetString("", typInfo, "", "stringForm"))
	*/

	if !hx.FgetBool("", typInfo, "", "isValid") {
		panic("reflect.haxeInterfaceUnpack() invalid type")
		return nil
	}
	ret := new(emptyInterface)
	ret.typ = &rtype{
		size:   uintptr(hx.FgetInt("", typInfo, "", "size")),
		align:  uint8(hx.FgetInt("", typInfo, "", "align")),
		kind:   uint8(hx.FgetInt("", typInfo, "", "kind")),
		string: (*string)(hx.Malloc(unsafe.Sizeof(""))),
	}
	*(*string)(ret.typ.string) = hx.FgetString("", typInfo, "", "stringForm")

	if hx.FgetString("", typInfo, "", "name") != "" {
		// TODO
	}

	ret.word = hx.Malloc(ret.typ.size)

	switch ret.typ.Kind() {
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
		*(*int64)(ret.word) = int64(hx.CodeDynamic("", "_a.itemAddr(0).load().val;", i))
	case Uint:
		*(*uint)(ret.word) = uint(hx.CodeInt("", "_a.itemAddr(0).load().val;", i))
	case Uint8:
		*(*uint8)(ret.word) = uint8(hx.CodeInt("", "_a.itemAddr(0).load().val;", i))
	case Uint16:
		*(*uint16)(ret.word) = uint16(hx.CodeInt("", "_a.itemAddr(0).load().val;", i))
	case Uint32:
		*(*uint32)(ret.word) = uint32(hx.CodeInt("", "_a.itemAddr(0).load().val;", i))
	case Uint64:
		*(*uint64)(ret.word) = uint64(hx.CodeDynamic("", "_a.itemAddr(0).load().val;", i))
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
	case String:
		*(*string)(ret.word) = hx.CodeString("", "_a.itemAddr(0).load().val;", i)
	case UnsafePointer, Ptr:
		*(*unsafe.Pointer)(ret.word) = unsafe.Pointer(hx.CodeDynamic("", "_a.itemAddr(0).load().val;", i))

	case Array:
		panic("reflect.haxeInterfaceUnpack() not yet programmed Array ")
	case Chan:
		panic("reflect.haxeInterfaceUnpack() not yet programmed Chan ")
	case Func:
		panic("reflect.haxeInterfaceUnpack() not yet programmed Func ")
	case Interface:
		panic("reflect.haxeInterfaceUnpack() not yet programmed Interface ")
	case Map:
		panic("reflect.haxeInterfaceUnpack() not yet programmed Map ")
	case Slice:
		panic("reflect.haxeInterfaceUnpack() not yet programmed Slice ")
	case Struct:
		panic("reflect.haxeInterfaceUnpack() not yet programmed Struct ")
	}

	return ret
}

func haxeInterfacePack(*emptyInterface) interface{} {
	panic("reflect.haxeInterfacePack() not yet implemented")
	return interface{}(nil)
}
