// Copyright 2014 Elliott Stoneham and The tardisgo Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package reflect is not implemented in TARDIS Go, this code a set of non-working TEST routines
package reflect

import "unsafe"

// An iword is the word that would be stored in an
// interface to represent a given value v.  Specifically, if v is
// bigger than a pointer, its word is a pointer to v's data.
// Otherwise, its word holds the data stored
// in its leading bytes (so is not a pointer).
// Because the value sometimes holds a pointer, we use
// unsafe.Pointer to represent it, so that if iword appears
// in a struct, the garbage collector knows that might be
// a pointer.
type iword unsafe.Pointer

// Method on non-interface type
type method struct {
	name    *string        // name of method
	pkgPath *string        // nil for exported Names; otherwise import path
	mtyp    *rtype         // method type (without receiver)
	typ     *rtype         // .(*FuncType) underneath (with receiver)
	ifn     unsafe.Pointer // fn used in interface call (one-word receiver)
	tfn     unsafe.Pointer // fn used for normal method call
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

// rtype is the common implementation of most values.
// It is embedded in other, public struct types, but always
// with a unique tag like `reflect:"array"` or `reflect:"ptr"`
// so that code cannot convert from, say, *arrayType to *ptrType.
type rtype struct {
	size          uintptr        // size in bytes
	hash          uint32         // hash of type; avoids computation in hash tables
	_             uint8          // unused/padding
	align         uint8          // alignment of variable with this type
	fieldAlign    uint8          // alignment of struct field with this type
	kind          uint8          // enumeration for C
	alg           *uintptr       // algorithm table (../runtime/runtime.h:/Alg)
	gc            unsafe.Pointer // garbage collection data
	string        *string        // string form; unnecessary but undeniably useful
	*uncommonType                // (relatively) uncommon fields
	ptrToThis     *rtype         // type for pointer to this type, if used in binary or has methods
}

func methodValueCall() {
	panic("reflect.methodValueCall() NOT IMPLEMENTED")
}

func methodFuncStub() {
	panic("reflect.methodFuncStub() NOT IMPLEMENTED")
}

// implemented in ../pkg/runtime
func chancap(ch iword) int {
	panic("reflect.chancap() NOT IMPLEMENTED")
	return 0
}
func chanclose(ch iword) {
	panic("reflect.chanclose() NOT IMPLEMENTED")
}
func chanlen(ch iword) int {
	panic("reflect.chanlen() NOT IMPLEMENTED")
	return 0
}
func chanrecv(t *rtype, ch iword, nb bool) (val iword, selected, received bool) {
	panic("reflect.chanrecv() NOT IMPLEMENTED")
}
func chansend(t *rtype, ch iword, val iword, nb bool) bool {
	panic("reflect.chansend() NOT IMPLEMENTED")
	return false
}

func makechan(typ *rtype, size uint64) (ch iword) {
	panic("reflect.makechan() NOT IMPLEMENTED")
}
func makemap(t *rtype) (m iword) {
	panic("reflect.makemap() NOT IMPLEMENTED")
}
func mapaccess(t *rtype, m iword, key iword) (val iword, ok bool) {
	panic("reflect.mapaccess() NOT IMPLEMENTED")
}
func mapassign(t *rtype, m iword, key, val iword, ok bool) {
	panic("reflect.mapassign() NOT IMPLEMENTED")
}
func mapiterinit(t *rtype, m iword) *byte {
	panic("reflect.mapiterint() NOT IMPLEMENTED")
}
func mapiterkey(it *byte) (key iword, ok bool) {
	panic("reflect.mapiterkey() NOT IMPLEMENTED")
}
func mapiternext(it *byte) {
	panic("reflect.mapiternext() NOT IMPLEMENTED")
}
func maplen(m iword) int {
	panic("reflect.maplen() NOT IMPLEMENTED")
	return 0
}

func call(fn, arg unsafe.Pointer, n uint32) {
	panic("reflect.call() NOT IMPLEMENTED")
}
func ifaceE2I(t *rtype, src interface{}, dst unsafe.Pointer) {
	panic("reflect.ifaceE2I() NOT IMPLEMENTED")

}

// implemented in runtime
func ismapkey(*rtype) bool {
	panic("reflect.ismapkey() NOT IMPLEMENTED")
	return false
}

// implemented in package runtime
func unsafe_New(*rtype) unsafe.Pointer {
	panic("reflect.unsafe_New() NOT IMPLEMENTED")
	return nil
}
func unsafe_NewArray(*rtype, int) unsafe.Pointer {
	panic("reflect.unsafe_NewArray() NOT IMPLEMENTED")
	return nil
}

// typelinks is implemented in package runtime.
// It returns a slice of all the 'typelink' information in the binary,
// which is to say a slice of known types, sorted by string.
// Note that strings are not unique identifiers for types:
// there can be more than one with a given string.
// Only types we might want to look up are included:
// channels, maps, slices, and arrays.
func typelinks() []*rtype {
	panic("reflect.typelinks() NOT IMPLEMENTED")
	return nil
}

// A runtimeSelect is a single case passed to rselect.
// This must match ../runtime/chan.c:/runtimeSelect
type runtimeSelect struct {
	dir uintptr // 0, SendDir, or RecvDir
	typ *rtype  // channel type
	ch  iword   // interface word for channel
	val iword   // interface word for value (for SendDir)
}

// rselect runs a select. It returns the index of the chosen case,
// and if the case was a receive, the interface word of the received
// value and the conventional OK bool to indicate whether the receive
// corresponds to a sent value.
func rselect([]runtimeSelect) (chosen int, recv iword, recvOK bool) {
	panic("reflect.rselect() NOT IMPLEMENTED")
}
