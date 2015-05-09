package haxegoruntime

import (
	"unsafe"

	"github.com/tardisgo/tardisgo/haxe/hx"
)

var stringList = make([]*string, 0, 128) // this may be too small a start size

// addrString is an utility function, strings with the same content must have the same address
// very simple implementation TODO optimize
func addrString(s string) *string {
	for i := range stringList {
		if s == *stringList[i] {
			return stringList[i]
		}
	}
	stringList = append(stringList, &s)
	return &s
}

// nilIfEmpty is an utiltiy to substitute nil pointer if the pointed at string is ""
func nilIfEmpty(sp *string) *string {
	if *sp == "" {
		return nil
	}
	return sp
}

// NOTE types below MUST be clones of the reflect package structs
// A Kind represents the specific kind of type that a Type represents.
// The zero Kind is not a valid kind.
type Kind uint8 // Haxe change, was uint

const (
	Invalid Kind = iota
	Bool
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Uintptr
	Float32
	Float64
	Complex64
	Complex128
	Array
	Chan
	Func
	Interface
	Map
	Ptr
	Slice
	String
	Struct
	UnsafePointer
)
const kindMask = (1 << 5) - 1

// rtype is the common implementation of most values.
// It is embedded in other, public struct types, but always
// with a unique tag like reflect:"array" or reflect:"ptr"
// so that code cannot convert from, say, *arrayType to *ptrType.
type rtype struct {
	size         uintptr
	hash         uint32            // hash of type; avoids computation in hash tables
	_            uint8             // unused/padding
	align        uint8             // alignment of variable with this type
	fieldAlign   uint8             // alignment of struct field with this type
	kind         Kind              // enumeration for C
	alg          *typeAlg          // algorithm table (../runtime/runtime.h:/Alg)
	gc           [2]unsafe.Pointer // garbage collection data
	string       *string           // string form; unnecessary but undeniably useful
	uncommonType unsafe.Pointer    //*R_uncommonType   // (relatively) uncommon fields
	ptrToThis    unsafe.Pointer    // type for pointer to this type, if used in binary or has methods
	zero         unsafe.Pointer    // pointer to zero value
}

type typeAlg struct {
	// function for hashing objects of this type
	// (ptr to object, size, seed) -> hash
	hash func(unsafe.Pointer, uintptr, uintptr) uintptr
	// function for comparing objects of this type
	// (ptr to object A, ptr to object B, size) -> ==?
	equal func(unsafe.Pointer, unsafe.Pointer, uintptr) bool
}

func equal(unsafe.Pointer, unsafe.Pointer, uintptr) bool {
	panic("typeAlg.equal not yet implemented")
	return false
}

func newRtype(size uintptr, align, fieldAlign, kind uint8, comprable bool, str string, uncommonType unsafe.Pointer, ptrToThis unsafe.Pointer) rtype {
	var alg *typeAlg
	if comprable {
		alg = &typeAlg{equal: equal}
	}
	return rtype{
		size: size, align: align, fieldAlign: fieldAlign, kind: Kind(kind), alg: alg, string: addrString(str),
		uncommonType: uncommonType, ptrToThis: ptrToThis, zero: hx.Malloc(size),
	}
}
func fillRtype(prt *rtype, rt rtype) {
	*prt = rt
}

// Method on non-interface type
type method struct {
	name    *string        // name of method
	pkgPath *string        // nil for exported Names; otherwise import path
	mtyp    unsafe.Pointer // *rtype         // method type (without receiver)
	typ     unsafe.Pointer // *rtype         // .(*FuncType) underneath (with receiver)
	ifn     uintptr        //unsafe.Pointer // fn used in interface call (one-word receiver)
	tfn     uintptr        //unsafe.Pointer // fn used for normal method call
}

func addMethod(meths []method, name, pkgPath string, mtyp, typ unsafe.Pointer, ifn, tfn uintptr) []method {
	return append(meths, method{
		name: addrString(name), pkgPath: nilIfEmpty(addrString(pkgPath)), mtyp: mtyp, typ: typ, ifn: ifn, tfn: tfn,
	})
}

func newMethodSlice() []method {
	return []method{}
}

// uncommonType is present only for types with names or methods
// (if T is a named type, the uncommonTypes for T and *T have methods).
// Using a pointer to this struct reduces the overall size required
// to describe an unnamed type with no methods.
type uncommonType struct {
	name    *string  // name of type
	pkgPath *string  // import path; nil for built-in types like int, string
	methods []method // methods associated with type
}

func newPtrToUncommonType(name, pkgPath string, methods []method) *uncommonType {
	return &uncommonType{name: addrString(name), pkgPath: nilIfEmpty(addrString(pkgPath)), methods: methods}
}

// arrayType represents a fixed array type.
type arrayType struct {
	rtype
	elem  *rtype // array element type
	slice *rtype // slice type
	len   uintptr
}

func fillArrayType(pat *arrayType, rtype rtype, elem, slice *rtype, len uintptr) {
	*pat = arrayType{rtype, elem, slice, len}
}

// chanType represents a channel type.
type chanType struct {
	rtype
	elem *rtype  // channel element type
	dir  uintptr // channel direction (ChanDir)
}

func fillChanType(pct *chanType, rtype rtype, elem *rtype, dir uintptr) {
	*pct = chanType{rtype, elem, dir}
}

// funcType represents a function type.
type funcType struct {
	rtype
	dotdotdot bool     // last input parameter is ...
	in        []*rtype // input parameter types
	out       []*rtype // output parameter types
}

func newPtrToRtypeSlice() []*rtype {
	return []*rtype{}
}
func addPtrToRtypeSlice(sl []*rtype, rtp *rtype) []*rtype {
	return append(sl, rtp)
}

func fillFuncType(fp *funcType, rtype rtype, dotdotdot bool, in, out []*rtype) {
	*fp = funcType{rtype, dotdotdot, in, out}
}

// imethod represents a method on an interface type
type imethod struct {
	name    *string // name of method
	pkgPath *string // nil for exported Names; otherwise import path
	typ     *rtype  // .(*FuncType) underneath
}

func newImethodSlice() []imethod {
	return []imethod{}
}

func addImethodSlice(imeths []imethod, name, pkgPath string, typ *rtype) []imethod {
	return append(imeths, imethod{addrString(name), nilIfEmpty(addrString(pkgPath)), typ})
}

// interfaceType represents an interface type.
type interfaceType struct {
	rtype
	methods []imethod // sorted by hash
}

func fillInterfaceType(ip *interfaceType, rtype rtype, methods []imethod) {
	*ip = interfaceType{rtype, methods}
}

// mapType represents a map type.
type mapType struct {
	rtype
	key           *rtype // map key type
	elem          *rtype // map element (value) type
	bucket        *rtype // internal bucket structure
	hmap          *rtype // internal map header
	keysize       uint8  // size of key slot
	indirectkey   uint8  // store ptr to key instead of key itself
	valuesize     uint8  // size of value slot
	indirectvalue uint8  // store ptr to value instead of value itself
	bucketsize    uint16 // size of bucket
}

func fillMapType(mp *mapType, rtype rtype, key, elem *rtype) {
	*mp = mapType{rtype: rtype, key: key, elem: elem}
}

// ptrType represents a pointer type.
type ptrType struct {
	rtype
	elem *rtype // pointer element (pointed at) type
}

func fillPtrType(p *ptrType, rtype rtype, elem *rtype) {
	*p = ptrType{rtype, elem}
}

// sliceType represents a slice type.
type sliceType struct {
	rtype
	elem *rtype // slice element type
}

func fillSliceType(sp *sliceType, rtype rtype, elem *rtype) {
	*sp = sliceType{rtype, elem}
}

// Struct field
type structField struct {
	name    *string // nil for embedded fields
	pkgPath *string // nil for exported Names; otherwise import path
	typ     *rtype  // type of field
	tag     *string // nil if no tag
	offset  uintptr // byte offset of field within struct
}

func newStructFieldSlice() []structField {
	return []structField{}
}

func addStructFieldSlice(sl []structField, name, pkgPath string, typ *rtype, tag string, offset uintptr) []structField {
	sf := structField{
		name:    nilIfEmpty(addrString(name)),    // This must be nil for "" for .Anonymous() to work
		pkgPath: nilIfEmpty(addrString(pkgPath)), // VERY important that pkgPath = nil rather than "" for reflection to work!
		typ:     typ,
		tag:     nilIfEmpty(addrString(tag)),
		offset:  offset,
	}
	return append(sl, sf)
}

// structType represents a struct type.
type structType struct {
	rtype
	fields []structField // sorted by offset
}

func fillStructType(stp *structType, rtype rtype, fields []structField) {
	*stp = structType{rtype, fields}
}

// NOTE end of reflect type clones

//type ErrorInterface interface {
//	Error() string
//}

// TypeTable provides a mapping between used type numbers (the index) to their reflect definitions
// but reflect is not used here in order to avoid pulling it into every build
var TypeTable = make([]*rtype, hx.GetInt("", "TypeInfo.nextTypeID"))

func init() {
	hx.Call("", "Tgotypes.setup", 0) // to set-up the TypeTable above
	//typetest()
	//println("DEBUG sizeof(funcType{}) = ", unsafe.Sizeof(funcType{}))
}

func AddHaxeType(ptr unsafe.Pointer) {
	TypeTable = append(TypeTable, (*rtype)(ptr))
	hx.SetInt("", "TypeInfo.nextTypeID", len(TypeTable))
}

func typetest() {
	for i, tp := range TypeTable {
		if tp != nil {
			println(i, *tp.string)
		}
	}
}

func getTypeString(id int) string {
	if id < 1 || id >= len(TypeTable) { // entry 0 is always nil
		return "<Type Not Found!>"
	}
	rt := TypeTable[id]
	return *(rt.string)
}

func getTypeID(s string) int {
	for id := 1; id < len(TypeTable); id++ { // TODO optimise this runtime loop to use a map
		if s == *(TypeTable[id].string) {
			return id
		}
	}
	return 0
}

func getMethod(tid int, path, name string) uintptr {
	//println("DEBUG getMethod:", tid, path, name)
	if tid < 1 || tid >= len(TypeTable) { // entry 0 is always nil
		hx.Call("", "Scheduler.panicFromHaxe", 1, "haxegoruntime.method() type id out of range")
	}
	rt := TypeTable[tid]
	ut := (*uncommonType)(rt.uncommonType)
	if ut == nil {
		hx.Call("", "Scheduler.panicFromHaxe", 1, "haxegoruntime.method() type has no uncommonType record")
	}
	if name == "" {
		hx.Call("", "Scheduler.panicFromHaxe", 1, "haxegoruntime.method() no method name")
	}
	exported := false
	if name[0] >= 'A' && name[0] <= 'Z' {
		exported = true
		// exported name, so path should be ""
		path = ""
	}
	if path != "" {
		for _, m := range ut.methods { // TODO these should be in order, so could use sort.search to speed up?
			mname := ""
			if m.name != nil {
				mname = *m.name
			}
			mpath := ""
			if m.pkgPath != nil {
				mpath = *m.pkgPath
			}
			//println("DEBUG check:", getTypeString(tid), mname, name, mpath, path)
			if mname == name && mpath == path {
				ret := m.ifn // is ifn the right one?
				if hx.IsNull(ret) {
					hx.Call("", "Scheduler.panicFromHaxe", 1, "haxegoruntime.method() null method for "+
						getTypeString(tid)+"."+name+" called from "+path)
				}
				//println("DEBUG found:", getTypeString(tid), path, name, ret)
				return ret
			}
		}
	} else {
		// no path so find the method without the path, but only if it is unique when not exported
		found := -1
		for f, m := range ut.methods { // TODO these should be in order, so could use sort.search to speed up?
			mname := ""
			if m.name != nil {
				mname = *m.name
			}
			//println("DEBUG check (ignore path):", getTypeString(tid), mname, name)
			if mname == name {
				//println("DEBUG found (ignore path):", getTypeString(tid), name)
				if found == -1 {
					found = f
					if exported {
						break
					}
				} else {
					hx.Call("", "Scheduler.panicFromHaxe", 1,
						"haxegoruntime.method() duplicates method found (ignoring path) for "+
							getTypeString(tid)+"."+name+" called from "+path)
				}
			}
		}
		if found >= 0 {
			ret := ut.methods[found].ifn // is ifn the right one?
			if hx.IsNull(ret) {
				hx.Call("", "Scheduler.panicFromHaxe", 1, "haxegoruntime.method() null method for "+
					getTypeString(tid)+"."+name+" called from "+path)
			}
			return ret
		}
	}
	//println("DEBUG not found:", getTypeString(tid), path, name)
	hx.Call("", "Scheduler.panicFromHaxe", 1, "haxegoruntime.method() no method found for "+
		getTypeString(tid)+"."+name+" called from "+path)
	return 0
}

func assertableTo(vid, tid int) bool {
	// id equality test done in Haxe
	if vid < 1 || vid >= len(TypeTable) { // entry 0 is always nil
		hx.Call("", "Scheduler.panicFromHaxe", 1, "haxegoruntime.assertableTo() tested type id out of range")
	}
	V := TypeTable[vid]
	if tid < 1 || tid >= len(TypeTable) { // entry 0 is always nil
		hx.Call("", "Scheduler.panicFromHaxe", 1, "haxegoruntime.assertableTo() interface type id out of range")
	}
	T := TypeTable[tid]
	ret := implements(T, V, false)
	/*
		// NOTE may not work properly in all circumstances yet, so test code retained
		ans := hx.CallBool("", "TypeAssert.assertableTo", 2, vid, tid)
		if ret != ans {
			println("DEBUG assertableTo wrong answer for T,V= ",
				*(T.string), T.kind&kindMask == Interface, tid,
				*(V.string), V.kind&kindMask == Interface, vid)
			implements(T, V, true)
			return ans
		}
	*/
	return ret
}

// implements copied from reflect package

// implements returns true if the type V implements the interface type T.
func implements(T, V *rtype, debug bool) bool {
	if debug {
		println("DEBUG implements T,V=", *(T.string), *(V.string))
	}
	if (T.kind & kindMask) /*.Kind()*/ != Interface {
		if debug {
			println("DEBUG imlements not an interface type", T.kind, T.kind&kindMask, Interface)
		}
		return false
	}
	t := (*interfaceType)(unsafe.Pointer(T))
	if len(t.methods) == 0 {
		if debug {
			println("DEBUG imlements no methods for interface type")
		}
		return true
	}

	// The same algorithm applies in both cases, but the
	// method tables for an interface type and a concrete type
	// are different, so the code is duplicated.
	// In both cases the algorithm is a linear scan over the two
	// lists - T's methods and V's methods - simultaneously.
	// Since method tables are stored in a unique sorted order
	// (alphabetical, with no duplicate method names), the scan
	// through V's methods must hit a match for each of T's
	// methods along the way, or else V does not implement T.
	// This lets us run the scan in overall linear time instead of
	// the quadratic time  a naive search would require.
	// See also ../runtime/iface.c.
	if (V.kind & kindMask) /*.Kind()*/ == Interface {
		if debug {
			println("DEBUG V is an interface type")
		}
		v := (*interfaceType)(unsafe.Pointer(V))
		i := 0
		for j := 0; j < len(v.methods); j++ {
			tm := &t.methods[i]
			vm := &v.methods[j]
			if vm.name == tm.name && vm.pkgPath == tm.pkgPath && vm.typ == tm.typ {
				if i++; i >= len(t.methods) {
					return true
				}
			}
		}
		if debug {
			println("DEBUG not found")
		}
		return false
	}
	if debug {
		println("DEBUG V not an interface type")
	}

	v := (*uncommonType)(V.uncommonType) //V.uncommon()
	if v == nil {
		if debug {
			println("DEBUG V has no commonType")
		}
		return false
	}
	i := 0
	for j := 0; j < len(v.methods); j++ {
		tm := &t.methods[i]
		vm := &v.methods[j]
		if debug {
			println("DEBUG compare i,j= ", i, j,
				"names=", ptrString(vm.name), uint(uintptr(unsafe.Pointer(vm.name))),
				ptrString(tm.name), uint(uintptr(unsafe.Pointer(tm.name))),
				"paths=", ptrString(vm.pkgPath), uint(uintptr(unsafe.Pointer(vm.pkgPath))),
				ptrString(tm.pkgPath), uint(uintptr(unsafe.Pointer(tm.pkgPath))),
				"types=", uint(uintptr(vm.mtyp)), uint(uintptr(unsafe.Pointer(tm.typ))))
		}
		if vm.name == tm.name && vm.pkgPath == tm.pkgPath && vm.mtyp == unsafe.Pointer(tm.typ) {
			if debug {
				println("DEBUG equal!!")
			}
			if i++; i >= len(t.methods) {
				if debug {
					println("DEBUG run-out of methods")
				}
				return true
			}
		}
	}
	if debug {
		println("DEBUG not found")
	}
	return false
}

func ptrString(ps *string) string {
	if ps == nil {
		return "<nil>"
	}
	return *ps
}
