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

/* removed by new implementation of rtype info
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
*/

func createHaxeType(id int) *rtype {
	if id <= 0 || id >= hx.GetInt("", "TypeInfo.nextTypeID") {
		//panic("reflect.createHaxeType() invalid haxe id: " + hx.CallString("", "Std.string", 1, id))
		return nil
	}

	// new version of type infomation:
	return (*rtype)(unsafe.Pointer(haxegoruntime.TypeTable[id]))
	//return createHaxeTypeA(id)
}

/* remove old version
func createHaxeTypeA(id int) *rtype {

	//println("createHaxeType:", id)
	tptr, ok := haxeIDmap[id]
	if ok {
		//println("createHaxeType already in map")
		return (*rtype)(tptr)
	}

	debug := (*rtype)(unsafe.Pointer(haxegoruntime.TypeTable[id]))

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
	if basicT.size != debug.size {
		panic("bad size")
	}
	basicT.align = uint8(hx.CodeInt("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", typInfo, TBIDalign))
	if basicT.align != debug.align {
		panic("bad align")
	}
	basicT.fieldAlign = uint8(hx.CodeInt("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", typInfo, TBIDfieldAlign))
	if basicT.fieldAlign != debug.fieldAlign {
		panic("bad fieldalign")
	}
	basicT.kind = uint8(hx.CodeInt("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", typInfo, TBIDkind))
	if basicT.kind != debug.kind {
		panic("bad kind")
	}
	basicT.string = (*string)(hx.Malloc(unsafe.Sizeof("")))
	basicT.zero = hx.Malloc(basicT.size) // assuming zeroes is always the right zero value
	for z := 0; z < int(basicT.size); z++ {
		if basicT.size < 256 {
			if ((*[256]byte)(basicT.zero))[z] != ((*[256]byte)(debug.zero))[z] {
				panic("bad zero value")
			}
		}
	}
	haxeIDmap[id] = space // NOTE must bag a spot in the map before we recurse

	*basicT.string = hx.CodeString("",
		"_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", typInfo, TBIDstringForm)
	if *basicT.string != *debug.string {
		panic("bad string")
	}

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
	if !(ptt == 0 && debug.ptrToThis == nil) {
		if ptt != typeIdFromPtr(debug.ptrToThis) {
			panic("bad pointer to type")
		}
	}

	typeName := hx.CodeString("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", typInfo, TBIDname)
	pkgPath := hx.CodeString("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", typInfo, TBIDpkgPath)
	numMethods := hx.CodeInt("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", typInfo, TBIDnumMethods)
	if typeName != "" || numMethods > 0 {
		basicT.uncommonType = &uncommonType{
			name:    &typeName,
			pkgPath: &pkgPath,
		}
		if *debug.uncommonType.name != *basicT.uncommonType.name {
			panic("bad name")
		}
		if *debug.uncommonType.pkgPath != *basicT.uncommonType.pkgPath {
			panic("bad pkgPath")
		}
		basicT.uncommonType.methods = make([]method, numMethods)
		if len(basicT.uncommonType.methods) != len(debug.uncommonType.methods) {
			panic("bad number of methods")
		}
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
					if *(basicT.uncommonType.methods[mFound].name) != *(debug.uncommonType.methods[mFound].name) {
						panic("bad method name")
					}
					if *(basicT.uncommonType.methods[mFound].pkgPath) != *(debug.uncommonType.methods[mFound].pkgPath) {
						panic("bad method pkgPath")
					}
					if mt0 != typeIdFromPtr(debug.uncommonType.methods[mFound].mtyp) {
						panic("bad method mtyp")
					}
					if t0 != typeIdFromPtr(debug.uncommonType.methods[mFound].typ) {
						panic("bad method typ")
					}
					mFound++
					//println("DEBUG createHaxeType() uncommonType method", id, mFound, m, mname)
				}
			}
		}
	} else {
		if debug.uncommonType != nil {
			panic("uncommon type set when it should be nil")
		}
	}

	switch basicT.Kind() & kindMask {
	case Invalid, Bool, Int, Int8, Int16, Int32, Int64, Uint, Uint8, Uint16, Uint32, Uint64, Uintptr,
		Float32, Float64, Complex64, Complex128, String, UnsafePointer:
		// NoOp - basicT already in the map

	case Ptr:
		elemID := hx.CodeInt("", "PtrTypeInfo.ptrByID[_a.itemAddr(0).load().val];", id)
		ptrT := (*ptrType)(space)
		ptrT.elem = createHaxeType(elemID)
		if typeIdFromPtr((*ptrType)(unsafe.Pointer(debug)).elem) != elemID {
			panic("bad ptr elem")
		}

			//if ptrT.elem.Kind() == Map {
			//	///////////
			//	m := (*mapType)(unsafe.Pointer(ptrT.elem))
			//	println("DEBUG pointer-to-map ", id, elemID, m.String(), m.key, m.elem, m)
			//}

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
		arrayT.elem = el //*rtype // array element type
		if typeIdFromPtr((*arrayType)(unsafe.Pointer(debug)).elem) != hx.FgetInt("", arrayInfo, "", "elem") {
			panic("bad array elem")
		}
		arrayT.slice = sl // *rtype // slice type
		if typeIdFromPtr((*arrayType)(unsafe.Pointer(debug)).slice) != hx.FgetInt("", arrayInfo, "", "slice") {
			panic("bad array slice")
		}
		arrayT.len = hx.FgetDynamic("", arrayInfo, "", "len") //  uintptr
		if int((*arrayType)(unsafe.Pointer(debug)).len) != hx.FgetInt("", arrayInfo, "", "len") {
			panic("bad array len")
		}

	case Slice:
		elemID := hx.CodeInt("", "Force.toInt(SliceTypeInfo.sliceByID[_a.itemAddr(0).load().val]);", id)
		sliceT := (*sliceType)(space)
		sliceT.elem = createHaxeType(elemID)
		if typeIdFromPtr((*sliceType)(unsafe.Pointer(debug)).elem) != elemID {
			panic("bad slice elem")
		}

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
		structDebug := (*structType)(unsafe.Pointer(debug))
		if len(structDebug.fields) != numFlds {
			panic("struct wrong number of fields")
		}
		for f := 0; f < numFlds; f++ {
			fldInfo := hx.CodeDynamic("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", haxeFlds, f)
			fStr := structField{}
			name := hx.CodeString("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", fldInfo, 0)
			//if name != "" {
			fStr.name = &name
			if *structDebug.fields[f].name != *fStr.name {
				panic("bad struct field name")
			}
			//}
			pkgPath := hx.CodeString("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", fldInfo, 1)
			if pkgPath != "" { // very important for reflection to work!
				fStr.pkgPath = &pkgPath
				if *structDebug.fields[f].pkgPath != *fStr.pkgPath {
					panic("bad struct field pkgPath")
				}
			} else {
				if structDebug.fields[f].pkgPath != nil {
					panic("bad struct field pkgPath should be nil")
				}
			}
			tid := hx.CodeInt("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", fldInfo, 2)
			fStr.typ = createHaxeType(tid)
			if tid != typeIdFromPtr(structDebug.fields[f].typ) {
				panic("bad struct field typ")
			}
			tag := hx.CodeString("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", fldInfo, 3)
			//if tag != "" {
			fStr.tag = &tag
			if *structDebug.fields[f].tag != *fStr.tag {
				panic("bad struct field tag")
			}
			//}
			fStr.offset = uintptr(hx.CodeInt("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", fldInfo, 4))
			if structDebug.fields[f].offset != fStr.offset {
				panic("bad struct field offset")
			}

			flds[f] = fStr
			//println("DEBUG struct.field ", basicT.Name(), *fStr.name, fStr.typ.Kind().String(), fStr.typ.Name(), fStr.typ.NumMethod())
		}
		structT := (*structType)(space)
		structT.fields = flds

	case Interface:
		interfaceT := (*interfaceType)(space)
		debugIface := (*interfaceType)(unsafe.Pointer(debug))
		haxeMeths := hx.CodeDynamic("", "IfaceTypeInfo.ifaceByID[_a.itemAddr(0).load().val];", id)
		if hx.IsNull(haxeMeths) {
			panic("reflect.createHaxeType() can't find named type info for type: " + hx.CallString("", "Std.string", 1, id))
		}
		numMeths := hx.FgetInt("", haxeMeths, "", "length")
		interfaceT.methods = make([]imethod, numMeths)
		if numMethods != len(debugIface.methods) {
			panic("interface wrong number of methods")
		}
		for m := 0; m < numMeths; m++ {
			methInfo := hx.CodeDynamic("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", haxeMeths, m)
			iStr := imethod{}
			name := hx.CodeString("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", methInfo, 0)
			//if name != "" {
			iStr.name = &name
			if *iStr.name != *debugIface.methods[m].name {
				panic("interface bad method name")
			}
			//}
			pkgPath := hx.CodeString("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", methInfo, 1)
			//if pkgPath != "" {
			iStr.pkgPath = &pkgPath
			if *iStr.pkgPath != *debugIface.methods[m].pkgPath {
				panic("interface bad method pkgPath")
			}
			//}
			typid := hx.CodeInt("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", methInfo, 2)
			iStr.typ = createHaxeType(typid)
			if typid != typeIdFromPtr(debugIface.methods[m].typ) {
				panic("interface bad method id")
			}
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
		mapT.key = ky // *rtype // map key type
		if typeIdFromPtr((*mapType)(unsafe.Pointer(debug)).key) != hx.FgetInt("", mapInfo, "", "key") {
			panic("bad map key")
		}
		mapT.elem = el // *rtype // map element (value) type
		if typeIdFromPtr((*mapType)(unsafe.Pointer(debug)).elem) != hx.FgetInt("", mapInfo, "", "elem") {
			panic("bad map elem")
		}
		//println("DEBUG mapT=", mapT, ky, el)

	case Func:
		funcT := (*funcType)(space)
		debugFunc := (*funcType)(unsafe.Pointer(debug))
		haxeFunc := hx.CodeDynamic("", "FuncTypeInfo.funcByID[_a.itemAddr(0).load().val];", id)
		if hx.IsNull(haxeFunc) {
			panic("reflect.createHaxeType() can't find func type info for type: " + hx.CallString("", "Std.string", 1, id))
		}
		funcT.dotdotdot = hx.FgetBool("", haxeFunc, "", "ddd")
		if funcT.dotdotdot != debugFunc.dotdotdot {
			panic("func bad dotdotdot")
		}
		pin := hx.FgetDynamic("", haxeFunc, "", "pin")
		pout := hx.FgetDynamic("", haxeFunc, "", "pout")
		pinL := hx.FgetInt("", pin, "", "length")
		poutL := hx.FgetInt("", pout, "", "length")
		funcT.in = make([]*rtype, pinL)
		if pinL != len(debugFunc.in) {
			panic("func bad len(in)")
		}
		funcT.out = make([]*rtype, poutL)
		if poutL != len(debugFunc.out) {
			panic("func bad len(out)")
		}
		for i := 0; i < pinL; i++ {
			ht := hx.CodeInt("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", pin, i)
			funcT.in[i] = createHaxeType(ht)
			if ht != typeIdFromPtr(debugFunc.in[i]) {
				panic("func bad in")
			}
		}
		for o := 0; o < poutL; o++ {
			ht := hx.CodeInt("", "_a.itemAddr(0).load().val[_a.itemAddr(1).load().val];", pout, o)
			funcT.out[o] = createHaxeType(ht)
			if ht != typeIdFromPtr(debugFunc.out[o]) {
				panic("func bad out")
			}
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
		if (*chanType)(unsafe.Pointer(debug)).dir != dir {
			panic("bad chan dir")
		}
		chanT.elem = el
		if typeIdFromPtr((*chanType)(unsafe.Pointer(debug)).elem) != hx.FgetInt("", chanInfo, "", "elem") {
			panic("bad chan elem")
		}
		//println("DEBUG chanT=", chanT, el, dir)

	}

	return (*rtype)(haxeIDmap[id])

}*/

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
		/*
			htyp := "null"
			if !hx.IsNull(val) {
				htyp = hx.CallString("", "Type.getClassName", 1, val)
			}
			println("DEBUG unpack haxe type=", htyp, " Go type=", ret.typ.Kind().String(), "val=", val)
		*/
		*(*uintptr)(ret.word) = val
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

	// TODO new version of type infomation
	//if useTgotypes {
	ityp := hx.CodeInt("", "_a.itemAddr(0).load().typ;", i)
	if unsafe.Pointer(ei.typ) != unsafe.Pointer(haxegoruntime.TypeTable[ityp]) {
		typ := typeIdFromPtr(ei.typ)
		//println("DEBUG typ(new)!=ityp(old) ", typ, ityp)
		hx.Code("",
			"var x=cast(_a.itemAddr(0).load(),Interface); x.typ=_a.itemAddr(1).load().val; _a.itemAddr(2).store(x); ",
			i, typ, &i)
		//println("DEBUG amended to ", hx.CodeInt("", "_a.itemAddr(0).load().typ;", i))
	}
	//}

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
		//println("DEBUG pack haxe ptr type=", hx.CallString("", "Type.getClassName", 1, ei.word),
		//	" Go type=", ei.typ.Kind().String(), "PtrVal=", ei.word)
		val := uintptr(ei.word) //?????? TODO this must be wrong, but needed for fmt to pass
		if ei.typ.Kind() != Map {
			val = *(*uintptr)(unsafe.Pointer(ei.word))
		}
		return hx.CodeIface("", ei.typ.String(), "_a.itemAddr(0).load().val;", val)

	}

	panic("reflect.haxeInterfacePack() not yet implemented for " + ei.typ.String() +
		" Kind= " + ei.typ.Kind().String())

	return interface{}(nil)
}

func haxeUnsafeNew(rtp *rtype) unsafe.Pointer {
	return hx.Malloc(rtp.size)
}
