package reflect

// this code relies on "unsafe" being included in the runtime - see Ptr in haxeInterfacePack
import (
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

const ( // keep in line with haxe/types.go
	TBIDisValid int = iota
	TBIDsize
	TBIDalign
	TBIDfieldAlign
	TBIDkind
	TBIDstringForm
	TBIDname
	TBIDptrToThis
	TBID_len
)

func createHaxeType(id int) *rtype {
	if id <= 0 || id >= hx.GetInt("", "TypeInfo.nextTypeID") {
		return nil
	}
	//println("createHaxeType:", id)
	tptr, ok := haxeIDmap[id]
	if ok {
		//println("createHaxeType already in map")
		return (*rtype)(tptr)
	}

	typInfo := hx.CodeDynamic("", "TypeInfo.typesByID[_a.itemAddr(0).load().val];", id)

	if hx.IsNull(typInfo) {
		panic("reflect.createHaxeType() null type info: " + hx.CallString("", "Std.string", 1, id))
		return nil
	}
	if !hx.CodeBool("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", typInfo, TBIDisValid) {
		//panic("reflect.createHaxeType() invalid type id: " + hx.CallString("", "Std.string", 1, id))
		return nil
	}

	basicT := rtype{
		size:   uintptr(hx.CodeInt("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", typInfo, TBIDsize)),
		align:  uint8(hx.CodeInt("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", typInfo, TBIDalign)),
		kind:   uint8(hx.CodeInt("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", typInfo, TBIDkind)),
		string: (*string)(hx.Malloc(unsafe.Sizeof(""))),
	}
	haxeIDmap[id] = unsafe.Pointer(&basicT) // NOTE must bag a spot in the map before we recurse

	*(*string)(basicT.string) = hx.CodeString("",
		"_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", typInfo, TBIDstringForm)

	ptt := hx.CodeInt("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", typInfo, TBIDptrToThis)
	pttHt := (*rtype)(nil)
	if ptt == 0 {
		// TODO create pointer to type?
	} else {
		if ptt == id {
			panic("createHaxeType pointer-to-type == id")
		}
		pttHt = createHaxeType(ptt)
	}
	basicT.ptrToThis = pttHt

	if hx.CodeString("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", typInfo, TBIDname) != "" {
		// TODO
	}

	switch basicT.Kind() {
	case Invalid, Bool, Int, Int8, Int16, Int32, Int64, Uint, Uint8, Uint16, Uint32, Uint64, Uintptr,
		Float32, Float64, Complex64, Complex128, String, UnsafePointer:
		// NoOp - basicT already in the map

	case Ptr:
		elemID := hx.CodeInt("", "TypeInfo.ptrByID[_a.itemAddr(0).load().val];", id)
		ptrT := ptrType{
			rtype: basicT,
			elem:  createHaxeType(elemID),
		}
		haxeIDmap[id] = unsafe.Pointer(&ptrT)

	case Array:
		arrayInfo := hx.CodeDynamic("", "TypeInfo.arrayByID[_a.itemAddr(0).load().val];", id)
		if hx.IsNull(arrayInfo) {
			panic("reflect.createHaxeType() no array information for id: " +
				hx.CallString("", "Std.string", 1, id))
		}
		el := createHaxeType(hx.FgetInt("", arrayInfo, "", "elem"))
		if el == nil {
			panic("reflect.createHaxeType() no array element information for id: " +
				hx.CallString("", "Std.string", 1, id))
		}
		sl := createHaxeType(hx.FgetInt("", arrayInfo, "", "slice"))
		if sl == nil {
			panic("reflect.createHaxeType() no array slice information for id: " +
				hx.CallString("", "Std.string", 1, id))
		}
		arrayT := arrayType{
			rtype: basicT,
			elem:  el,                                       //*rtype // array element type
			slice: sl,                                       // *rtype // slice type
			len:   hx.FgetDynamic("", arrayInfo, "", "len"), //  uintptr
		}
		haxeIDmap[id] = unsafe.Pointer(&arrayT)

	case Slice:
		elemID := hx.CodeInt("", "Force.toInt(TypeInfo.sliceByID[_a.itemAddr(0).load().val]);", id)
		sliceT := sliceType{
			rtype: basicT,
			elem:  createHaxeType(elemID),
		}
		haxeIDmap[id] = unsafe.Pointer(&sliceT)

	case Struct:
		haxeFlds := hx.CodeDynamic("", "TypeInfo.structByID[_a.itemAddr(0).load().val];", id)
		if hx.IsNull(haxeFlds) {
			panic("reflect.createHaxeType() can't find struct type info for type: " + hx.CallString("", "Std.string", 1, id))
		}
		numFlds := hx.FgetInt("", haxeFlds, "", "length")
		flds := make([]structField, numFlds)
		for f := 0; f < numFlds; f++ {
			fldInfo := hx.CodeDynamic("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", haxeFlds, f)
			fStr := structField{}
			name := hx.CodeString("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", fldInfo, 0)
			if name != "" {
				fStr.name = &name
			}
			pkgPath := hx.CodeString("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", fldInfo, 1)
			if pkgPath != "" {
				fStr.pkgPath = &pkgPath
			}
			fStr.typ = createHaxeType(
				hx.CodeInt("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", fldInfo, 2))
			tag := hx.CodeString("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", fldInfo, 3)
			if tag != "" {
				fStr.tag = &tag
			}
			fStr.offset = uintptr(hx.CodeInt("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", fldInfo, 4))
			flds[f] = fStr
		}
		structT := structType{
			rtype:  basicT,
			fields: flds,
		}
		haxeIDmap[id] = unsafe.Pointer(&structT)

	case Interface:
		interfaceT := interfaceType{
			rtype:   basicT,
			methods: make([]imethod, 0), // TODO
		}
		haxeIDmap[id] = unsafe.Pointer(&interfaceT)

	case Map:
		//panic("reflect.createHaxeType() not yet programmed Map ")
		mapInfo := hx.CodeDynamic("", "TypeInfo.mapByID[_a.itemAddr(0).load().val];", id)
		if hx.IsNull(mapInfo) {
			panic("reflect.createHaxeType() no map information for id: " +
				hx.CallString("", "Std.string", 1, id))
		}
		el := createHaxeType(hx.FgetInt("", mapInfo, "", "elem"))
		if el == nil {
			panic("reflect.createHaxeType() no map element information for id: " +
				hx.CallString("", "Std.string", 1, id))
		}
		ky := createHaxeType(hx.FgetInt("", mapInfo, "", "key"))
		if ky == nil {
			panic("reflect.createHaxeType() no map key information for id: " +
				hx.CallString("", "Std.string", 1, id))
		}

		mapT := mapType{
			rtype: basicT,
			key:   ky, // *rtype // map key type
			elem:  el, // *rtype // map element (value) type
		}
		haxeIDmap[id] = unsafe.Pointer(&mapT)

	case Chan:
		panic("reflect.createHaxeType() not yet programmed Chan ")
	case Func:
		panic("reflect.createHaxeType() not yet programmed Func ")
	}

	return (*rtype)(haxeIDmap[id])

}

func haxeInterfaceUnpack(i interface{}) *emptyInterface {
	ret := new(emptyInterface)
	if i == nil {
		//panic("reflect.haxeInterfaceUnpack() nil value")
		return ret
	}
	ret.typ = createHaxeType(hx.CodeInt("", "_a.itemAddr(0).load().typ;", i))
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
	case UnsafePointer, Ptr:
		*(*unsafe.Pointer)(ret.word) = unsafe.Pointer(hx.CodeDynamic("", "_a.itemAddr(0).load().val;", i))

	case String:
		*(*string)(ret.word) = hx.CodeString("", "_a.itemAddr(0).load().val;", i)

	case Array, Struct:
		hx.Code("",
			"_a.itemAddr(1).load().val.store_object(_a.itemAddr(2).load().val,_a.itemAddr(0).load().val);",
			i, ret.word, ret.typ.size)

	case Slice, Interface, Map:
		hx.Code("",
			"_a.itemAddr(1).load().val.store(_a.itemAddr(0).load().val);",
			i, ret.word)

	case Chan:
		panic("reflect.haxeInterfaceUnpack() not yet programmed Chan ")
	case Func:
		panic("reflect.haxeInterfaceUnpack() not yet programmed Func ")
		//case Map:
		//	panic("reflect.haxeInterfaceUnpack() not yet programmed Map ")
	}

	return ret
}

func haxeInterfacePack(ei *emptyInterface) interface{} {

	switch ei.typ.Kind() {
	case Bool:
		return hx.CodeIface("", "bool", "_a.itemAddr(0).load().val;", *(*bool)(ei.word))
	case Int:
		return hx.CodeIface("", "int", "_a.itemAddr(0).load().val;", *(*int)(ei.word))
	case Int8:
		return hx.CodeIface("", "int8", "_a.itemAddr(0).load().val;", *(*int8)(ei.word))
	case Int16:
		return hx.CodeIface("", "int16", "_a.itemAddr(0).load().val;", *(*int16)(ei.word))
	case Int32:
		return hx.CodeIface("", "int32", "_a.itemAddr(0).load().val;", *(*int32)(ei.word))
	case Int64:
		return hx.CodeIface("", "int64", "_a.itemAddr(0).load().val;", *(*int64)(ei.word))
	case Uint:
		return hx.CodeIface("", "uint", "_a.itemAddr(0).load().val;", *(*uint)(ei.word))
	case Uint8:
		return hx.CodeIface("", "uint8", "_a.itemAddr(0).load().val;", *(*uint8)(ei.word))
	case Uint16:
		return hx.CodeIface("", "uint16", "_a.itemAddr(0).load().val;", *(*uint16)(ei.word))
	case Uint32:
		return hx.CodeIface("", "uint32", "_a.itemAddr(0).load().val;", *(*uint32)(ei.word))
	case Uint64:
		return hx.CodeIface("", "uint64", "_a.itemAddr(0).load().val;", *(*uint64)(ei.word))
	case Uintptr:
		return hx.CodeIface("", "uintptr", "_a.itemAddr(0).load().val;", *(*uintptr)(ei.word))
	case Float32:
		return hx.CodeIface("", "float32", "_a.itemAddr(0).load().val;", *(*float32)(ei.word))
	case Float64:
		return hx.CodeIface("", "float64", "_a.itemAddr(0).load().val;", *(*float64)(ei.word))
	case Complex64:
		return hx.CodeIface("", "complex64", "_a.itemAddr(0).load().val;", *(*complex64)(ei.word))
	case Complex128:
		return hx.CodeIface("", "complex128", "_a.itemAddr(0).load().val;", *(*complex128)(ei.word))
	case UnsafePointer, Ptr: // NOTE unsafe package must always be present for this modelling of Ptr
		return hx.CodeIface("", "unsafe.Pointer", "_a.itemAddr(0).load().val;", ei.word)

	case String:
		return hx.CodeIface("", "string", "_a.itemAddr(0).load().val;", *(*string)(ei.word))

	case Array, Struct:
		return hx.CodeIface("", ei.typ.String(),
			"_a.itemAddr(0).load().val.load_object(_a.itemAddr(1).load().val);",
			ei.word, ei.typ.size)

	case Slice:
		stp := (*sliceType)(unsafe.Pointer(ei.typ))
		r := hx.CodeIface("", stp.String(),
			"_a.itemAddr(0).load().val.load();",
			ei.word)
		//println("DEBUG slice pack r=", r)
		return r

	case Chan:
	case Func:
	case Interface:
	case Map:
	}

	panic("reflect.haxeInterfacePack() not yet implemented for " + ei.typ.String() +
		" Kind= " + ei.typ.Kind().String())

	return interface{}(nil)
}

func haxeUnsafeNew(rtp *rtype) unsafe.Pointer {
	// TODO feedback more fully into the Haxe TypeInfo structures ...

	switch rtp.Kind() {
	case Bool, Int, Int8, Int16, Int32, Int64, Uint, Uint8, Uint16, Uint32, Uint64, Uintptr,
		Float32, Float64, Complex64, Complex128, String, UnsafePointer:
		n := new(rtype)
		*n = *rtp
		return unsafe.Pointer(n)

	case Slice:
		stp := (*sliceType)(unsafe.Pointer(rtp))
		n := new(sliceType)
		*n = *stp
		return unsafe.Pointer(n)

		/*
			case Ptr:
				elemID := hx.CodeInt("", "TypeInfo.ptrByID[_a.itemAddr(0).load().val];", id)
				ptrT := ptrType{
					rtype: basicT,
					elem:  createHaxeType(elemID),
				}
				haxeIDmap[id] = unsafe.Pointer(&ptrT)

			case Array:
				arrayInfo := hx.CodeDynamic("", "TypeInfo.arrayByID[_a.itemAddr(0).load().val];", id)
				if hx.IsNull(arrayInfo) {
					panic("reflect.createHaxeType() no array information for id: " +
						hx.CallString("", "Std.string", 1, id))
				}
				arrayT := arrayType{
					rtype: basicT,
					elem:  createHaxeType(hx.FgetInt("", arrayInfo, "", "elem")),  //*rtype // array element type
					slice: createHaxeType(hx.FgetInt("", arrayInfo, "", "slice")), // *rtype // slice type
					len:   hx.FgetDynamic("", arrayInfo, "", "len"),               //  uintptr
				}
				haxeIDmap[id] = unsafe.Pointer(&arrayT)

			case Slice:
				elemID := hx.CodeInt("", "TypeInfo.sliceByID[_a.itemAddr(0).load().val];", id)
				sliceT := sliceType{
					rtype: basicT,
					elem:  createHaxeType(elemID),
				}
				haxeIDmap[id] = unsafe.Pointer(&sliceT)

			case Chan:
				panic("reflect.createHaxeType() not yet programmed Chan ")
			case Func:
				panic("reflect.createHaxeType() not yet programmed Func ")
			case Interface:
				panic("reflect.createHaxeType() not yet programmed Interface ")
			case Map:
				panic("reflect.createHaxeType() not yet programmed Map ")
			case Struct:
				panic("reflect.createHaxeType() not yet programmed Struct ")
		*/
	}
	panic("reflect.haxeUnsafeNew() unhandled type: " + rtp.String())
	return nil

}
