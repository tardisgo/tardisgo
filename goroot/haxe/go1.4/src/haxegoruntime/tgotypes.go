package haxegoruntime

import (
	"unsafe"

	"github.com/tardisgo/tardisgo/haxe/hx"
)

// addrString is an utility function
func addrString(s string) *string {
	return &s
}

// NOTE types below MUST be clones of the reflect package structs

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
	kind         uint8             // enumeration for C
	alg          *unsafe.Pointer   // was: typeAlg          // algorithm table (../runtime/runtime.h:/Alg)
	gc           [2]unsafe.Pointer // garbage collection data
	string       *string           // string form; unnecessary but undeniably useful
	uncommonType unsafe.Pointer    //*R_uncommonType   // (relatively) uncommon fields
	ptrToThis    unsafe.Pointer    // type for pointer to this type, if used in binary or has methods
	zero         unsafe.Pointer    // pointer to zero value
}

func newRtype(size uintptr, align, fieldAlign, kind uint8, str string, uncommonType unsafe.Pointer, ptrToThis unsafe.Pointer) rtype {
	return rtype{
		size: size, align: align, fieldAlign: fieldAlign, kind: kind, string: addrString(str),
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
	ifn     unsafe.Pointer // fn used in interface call (one-word receiver)
	tfn     unsafe.Pointer // fn used for normal method call
}

func addMethod(meths []method, name, pkgPath string, mtyp, typ, ifn, tfn unsafe.Pointer) []method {
	return append(meths, method{
		name: addrString(name), pkgPath: addrString(pkgPath), mtyp: mtyp, typ: typ, ifn: ifn, tfn: tfn,
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
	return &uncommonType{name: addrString(name), pkgPath: addrString(pkgPath), methods: methods}
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
	return append(imeths, imethod{addrString(name), addrString(pkgPath), typ})
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
		name:   addrString(name),
		typ:    typ,
		tag:    addrString(tag),
		offset: offset,
	}

	// VERY important that pkgPath = nil rather than "" for reflection to work!
	if pkgPath != "" {
		sf.pkgPath = addrString(pkgPath)
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

func typetest() {
	for i, tp := range TypeTable {
		if tp != nil {
			println(i, *tp.string)
		}
	}
}

func getTypeString(id int) string {
	if id < 1 || id >= len(TypeTable) { // entry 0 is always nil
		return ""
	}
	rt := TypeTable[id]
	return *(rt.string)
}
