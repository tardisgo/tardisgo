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
	TBIDpkgPath
	TBIDnumMethods
)

func createHaxeType(id int) *rtype {
	if id <= 0 || id >= hx.GetInt("", "TypeInfo.nextTypeID") {
		//panic("reflect.createHaxeType() invalid haxe id: " + hx.CallString("", "Std.string", 1, id))
		return nil
	}
	//println("createHaxeType:", id)
	tptr, ok := haxeIDmap[id]
	if ok {
		//println("createHaxeType already in map")
		return (*rtype)(tptr)
	}

	//println("DEBUG createHaxeType() id=", id)
	typInfo := hx.CodeDynamic("", "TypeInfoIDs.typesByID[_a.itemAddr(0).load().val];", id)
	//println("DEBUG createHaxeType() typInfo=", typInfo)

	if hx.IsNull(typInfo) {
		panic("reflect.createHaxeType() null type info for id: " + hx.CallString("", "Std.string", 1, id))
		return nil
	}
	if !hx.CodeBool("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", typInfo, TBIDisValid) {
		//panic("reflect.createHaxeType() invalid type id: " + hx.CallString("", "Std.string", 1, id))
		return nil
	}

	//println("DEBUG Sizes:", unsafe.Sizeof(mapType{}), unsafe.Sizeof(interfaceType{}), unsafe.Sizeof(funcType{}))
	space := hx.Malloc(unsafe.Sizeof(funcType{})) // funcType is the largest of the types

	basicT := (*rtype)(space)
	basicT.size = uintptr(hx.CodeInt("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", typInfo, TBIDsize))
	basicT.align = uint8(hx.CodeInt("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", typInfo, TBIDalign))
	basicT.fieldAlign = uint8(hx.CodeInt("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", typInfo, TBIDfieldAlign))
	basicT.kind = uint8(hx.CodeInt("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", typInfo, TBIDkind))
	basicT.string = (*string)(hx.Malloc(unsafe.Sizeof("")))
	basicT.zero = hx.Malloc(basicT.size) // assuming zeroes is always the right zero value
	haxeIDmap[id] = space                // NOTE must bag a spot in the map before we recurse

	*basicT.string = hx.CodeString("",
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

	typeName := hx.CodeString("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", typInfo, TBIDname)
	pkgPath := hx.CodeString("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", typInfo, TBIDpkgPath)
	numMethods := hx.CodeInt("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", typInfo, TBIDnumMethods)
	if typeName != "" || numMethods > 0 {
		basicT.uncommonType = &uncommonType{
			name:    &typeName,
			pkgPath: &pkgPath,
		}
		basicT.uncommonType.methods = make([]method, numMethods)
		//println("DEBUG uncommonType id,name,numMethods=", id, typeName, numMethods)
		if numMethods > 0 {
			//println("DEBUG createHaxeType() Methods id=", id)
			//methArray := hx.CodeDynamic("", "MethTypeInfo.methByID[_a.itemAddr(0).load().val];", id)
			methArray := hx.CodeDynamic("", "MethTypeInfo.methByID;")
			//println("DEBUG createHaxeType() Methods array=", methArray)
			if hx.IsNull(methArray) {
				panic("reflect.createHaxeType() can't find methods info for type: " + hx.CallString("", "Std.string", 1, id))
			}
			mFound := 0
			for m := 0; m < hx.FgetInt("", methArray, "", "length"); m++ {
				if id == hx.CodeInt("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val].id;", methArray, m) {
					mname := hx.CodeString("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val].name;", methArray, m)
					pp := hx.CodeString("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val].pkgPath;", methArray, m)
					mt0 := hx.CodeInt("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val].mtyp;", methArray, m)
					t0 := hx.CodeInt("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val].typ;", methArray, m)
					//ifnDynamic := hx.CodeDynamic("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val].ifn;", methArray, m)
					//tfnDynamic := hx.CodeDynamic("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val].tfn;", methArray, m)
					basicT.uncommonType.methods[mFound] = method{
						name:    &mname,
						pkgPath: &pp,
						mtyp:    createHaxeType(mt0), // *rtype         // method type (without receiver)
						typ:     createHaxeType(t0),  // *rtype         // .(*FuncType) underneath (with receiver)
						//ifn:     unsafe.Pointer(&ifnDynamic), // unsafe.Pointer // fn used in interface call (one-word receiver)
						//tfn:     unsafe.Pointer(&tfnDynamic), // unsafe.Pointer // fn used for normal method call
					}
					mFound++
					//println("DEBUG createHaxeType() uncommonType method", id, mFound, m, mname)
				}
			}
		}
	}

	switch basicT.Kind() {
	case Invalid, Bool, Int, Int8, Int16, Int32, Int64, Uint, Uint8, Uint16, Uint32, Uint64, Uintptr,
		Float32, Float64, Complex64, Complex128, String, UnsafePointer:
		// NoOp - basicT already in the map

	case Ptr:
		elemID := hx.CodeInt("", "PtrTypeInfo.ptrByID[_a.itemAddr(0).load().val];", id)
		ptrT := (*ptrType)(space)
		ptrT.elem = createHaxeType(elemID)
		/*
			if ptrT.elem.Kind() == Map {
				///////////
				m := (*mapType)(unsafe.Pointer(ptrT.elem))
				println("DEBUG pointer-to-map ", id, elemID, m.String(), m.key, m.elem, m)
			}
		*/

	case Array:
		arrayInfo := hx.CodeDynamic("", "ArrayTypeInfo.arrayByID[_a.itemAddr(0).load().val];", id)
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
		arrayT := (*arrayType)(space)
		arrayT.elem = el                                      //*rtype // array element type
		arrayT.slice = sl                                     // *rtype // slice type
		arrayT.len = hx.FgetDynamic("", arrayInfo, "", "len") //  uintptr

	case Slice:
		elemID := hx.CodeInt("", "Force.toInt(SliceTypeInfo.sliceByID[_a.itemAddr(0).load().val]);", id)
		sliceT := (*sliceType)(space)
		sliceT.elem = createHaxeType(elemID)

	case Struct:
		var haxeFlds uintptr
		for idFound := 0; idFound < hx.CodeInt("", "StructTypeInfo.structByID.length;"); idFound++ {
			if id == hx.CodeInt("", "StructTypeInfo.structByID[_a.itemAddr(0).load().val].id;", idFound) {
				haxeFlds = hx.CodeDynamic("", "StructTypeInfo.structByID[_a.itemAddr(0).load().val].flds;", idFound)
				break
			}
		}
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
			//println("DEBUG struct.field ", basicT.Name(), *fStr.name, fStr.typ.Kind().String(), fStr.typ.Name(), fStr.typ.NumMethod())
		}
		structT := (*structType)(space)
		structT.fields = flds

	case Interface:
		interfaceT := (*interfaceType)(space)
		haxeMeths := hx.CodeDynamic("", "IfaceTypeInfo.ifaceByID[_a.itemAddr(0).load().val];", id)
		if hx.IsNull(haxeMeths) {
			panic("reflect.createHaxeType() can't find named type info for type: " + hx.CallString("", "Std.string", 1, id))
		}
		numMeths := hx.FgetInt("", haxeMeths, "", "length")
		interfaceT.methods = make([]imethod, numMeths)
		for m := 0; m < numMeths; m++ {
			methInfo := hx.CodeDynamic("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", haxeMeths, m)
			iStr := imethod{}
			name := hx.CodeString("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", methInfo, 0)
			if name != "" {
				iStr.name = &name
			}
			pkgPath := hx.CodeString("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", methInfo, 1)
			if pkgPath != "" {
				iStr.pkgPath = &pkgPath
			}
			iStr.typ = createHaxeType(
				hx.CodeInt("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", methInfo, 2))
			interfaceT.methods[m] = iStr
		}
		//println("DEBUG Interface name, methods=", basicT.string, interfaceT.methods)

	case Map:
		//panic("reflect.createHaxeType() not yet programmed Map ")
		mapInfo := hx.CodeDynamic("", "MapTypeInfo.mapByID[_a.itemAddr(0).load().val];", id)
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

		mapT := (*mapType)(space)
		mapT.key = ky  // *rtype // map key type
		mapT.elem = el // *rtype // map element (value) type
		//println("DEBUG mapT=", mapT, ky, el)

	case Func:
		funcT := (*funcType)(space)
		haxeFunc := hx.CodeDynamic("", "FuncTypeInfo.funcByID[_a.itemAddr(0).load().val];", id)
		if hx.IsNull(haxeFunc) {
			panic("reflect.createHaxeType() can't find func type info for type: " + hx.CallString("", "Std.string", 1, id))
		}
		funcT.dotdotdot = hx.FgetBool("", haxeFunc, "", "ddd")
		pin := hx.FgetDynamic("", haxeFunc, "", "pin")
		pout := hx.FgetDynamic("", haxeFunc, "", "pout")
		pinL := hx.FgetInt("", pin, "", "length")
		poutL := hx.FgetInt("", pout, "", "length")
		funcT.in = make([]*rtype, pinL)
		funcT.out = make([]*rtype, poutL)
		for i := 0; i < pinL; i++ {
			ht := hx.CodeInt("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", pin, i)
			funcT.in[i] = createHaxeType(ht)
		}
		for o := 0; o < poutL; o++ {
			ht := hx.CodeInt("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", pout, o)
			funcT.out[o] = createHaxeType(ht)
		}
		//println("DEBUG func:", funcT.string, funcT.dotdotdot, funcT.in, funcT.out)

	case Chan:
		//panic("reflect.createHaxeType() not yet programmed Chan ")
		chanInfo := hx.CodeDynamic("", "ChanTypeInfo.chanByID[_a.itemAddr(0).load().val];", id)
		if hx.IsNull(chanInfo) {
			panic("reflect.createHaxeType() no chan information for id: " +
				hx.CallString("", "Std.string", 1, id))
		}
		el := createHaxeType(hx.FgetInt("", chanInfo, "", "elem"))
		if el == nil {
			panic("reflect.createHaxeType() no map element information for id: " +
				hx.CallString("", "Std.string", 1, id))
		}
		dir := uintptr(hx.FgetInt("", chanInfo, "", "dir"))

		chanT := (*chanType)(space)
		chanT.dir = dir
		chanT.elem = el
		//println("DEBUG chanT=", chanT, el, dir)

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

	haxe2go(ret, i)
	return ret
}

func haxe2go(ret *emptyInterface, i interface{}) {
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

	case Slice, Interface, Map, Func, Chan:
		//println("DEBUG ret.word", ret.word, hx.MethString("", uintptr(ret.word), "Pointer", "toUniqueVal", 0))
		hx.Code("",
			"_a.itemAddr(1).load().val.store(_a.itemAddr(0).load().val);",
			i, ret.word)

	}

}

func haxeInterfacePack(ei *emptyInterface) interface{} {

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

	case Slice:
		stp := (*sliceType)(unsafe.Pointer(ei.typ))
		r := hx.CodeIface("", stp.String(),
			"_a.itemAddr(0).load().val.load();",
			ei.word)
		//println("DEBUG slice pack r=", r)
		return r

	case Map:
		mtp := (*mapType)(unsafe.Pointer(ei.typ))
		r := hx.CodeIface("", mtp.String(),
			"_a.itemAddr(0).load().val.load();",
			ei.word)
		//println("DEBUG map pack r=", r)
		return r

	case Interface:
		return hx.CodeIface("", ei.typ.String(),
			"_a.itemAddr(0).load().val.load();",
			ei.word)

	case Chan:
	case Func:
	}

	panic("reflect.haxeInterfacePack() not yet implemented for " + ei.typ.String() +
		" Kind= " + ei.typ.Kind().String())

	return interface{}(nil)
}

func haxeUnsafeNew(rtp *rtype) unsafe.Pointer {
	return hx.Malloc(rtp.size)
}
