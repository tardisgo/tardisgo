// Copyright 2014 Elliott Stoneham and The tardisgo Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package tgoruntime provdies functions for the Go "runtime" standard library package when used by TARDIS Go.
// Please note, this is an incomplete implementation.
package tgoruntime

// THE GOLANG RUNTIME PACKAGE IS NOT CURRENTLY ALL USABLE

import "github.com/tardisgo/tardisgo/tardisgolib"

func init() { // make calls in here to protect against Dead Code Elimination
	// NOTE: only working code included here for now
	//	Gosched()
	//	NumGoroutine()
	//memeq(nil, nil, 0)
}

// Gosched implements runtime.Goshed
func Gosched() { tardisgolib.Gosched() }

// NumGoroutine emulates runtime.NumGoroutine
func NumGoroutine() int { return tardisgolib.NumGoroutine() }

func GOMAXPROCS(n int) int { return 1 }

// functions for Go 1.4
/*
const verybig = 1 << 32 // the largest object size allowable

func memeq(p, q unsafe.Pointer, size uintptr) bool {
	a := (*[verybig]byte)(p)
	b := (*[verybig]byte)(q)
	for i := uintptr(0); i < size; i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
*/
//	EVERYTHING BELOW NOT YET IMPLEMENTED

// TEST TEST this is a kludge
//var sizeof_C_MStats int

// Goexit unimplemented
//func Goexit() { panic("runtime.Goexit() not yet implemented") }

// FuncForPC not implemented
//func FuncForPC(pc uintptr) (uip *uintptr) { panic("runtime.FuncForPC() not yet implemented") }

// SetFinalizer NoOp
//func SetFinalizer(x, f interface{}) {
//panic("runtime.SetFinalizer() not yet implemented")
// used in init process, so must be NoOp for now
//}

// implemented in symtab.c in GC runtime package

//func funcline_go(*uintptr /* should be *runtime.Func*/, uintptr) (s string, i int) {
//	panic("runtime.funcline_go() not yet implemented")
//}
//func funcname_go(*uintptr /* should be *runtime.Func*/) (s string) {
//	panic("runtime.funcname_go() not yet implemented")
//}
//func funcentry_go(*uintptr /* should be *runtime.Func*/) (uip uintptr) {
//	panic("runtime.funcentry_go() not yet implemented")
//}

////

//func cstringToGo(uintptr) (s string) {
//	panic("runtime.cstringToGo() not yet implemented")
//}

//func getgoroot() string {
//	//panic("runtime.getgoroot() not yet implemented")
//	return "" // GOROOT is empty in TARDIS Go
//}

// Caller reports file and line number information about function invocations on
// the calling goroutine's stack.  The argument skip is the number of stack frames
// to ascend, with 0 identifying the caller of Caller.  (For historical reasons the
// meaning of skip differs between Caller and Callers.) The return values report the
// program counter, file name, and line number within the file of the corresponding
// call.  The boolean ok is false if it was not possible to recover the information.
//func Caller(skip int) (pc uintptr, file string, line int, ok bool) {
//	panic("runtime.Caller() not yet implemented")
//}

// Callers fills the slice pc with the program counters of function invocations
// on the calling goroutine's stack.  The argument skip is the number of stack frames
// to skip before recording in pc, with 0 identifying the frame for Callers itself and
// 1 identifying the caller of Callers.
// It returns the number of entries written to pc.
//func Callers(skip int, pc []uintptr) (i int) {
//	panic("runtime.Callers() not yet implemented")
//}
