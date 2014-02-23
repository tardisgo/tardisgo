// Copyright 2014 Elliott Stoneham and The tardisgo Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package runtime provdies functions for the Go "runtime" standard library package when used by TARDIS Go.
// Please note, this is an incomplete implementation.
package runtime

// THE GOLANG RUNTIME PACKAGE IS NOT CURRENTLY ALL USABLE

import "github.com/tardisgo/tardisgo/tardisgolib"

// Gosched implements runtime.Goshed
func Gosched() { tardisgolib.Gosched() }

// NumGoroutine emulates runtime.NumGoroutine
func NumGoroutine() int { return tardisgolib.NumGoroutine() }

//	EVERYTHING BELOW NOT YET IMPLEMENTED

// TEST TEST this is a kludge
//var sizeof_C_MStats int

// Goexit unimplemented
func Goexit() { panic("runtime.Goexit() not yet implemented") }

// FuncForPC not implemented
func FuncForPC(pc uintptr) *uintptr { panic("runtime.FuncForPC() not yet implemented"); return nil }

// SetFinalizer NoOp
func SetFinalizer(x, f interface{}) {
	//panic("runtime.SetFinalizer() not yet implemented")
	// used in init process, so must be NoOp for now
}

// implemented in symtab.c in GC runtime package
func funcline_go(*uintptr /* should be *runtime.Func*/, uintptr) (string, int) {
	panic("runtime.funcline_go() not yet implemented")
	return "", 0
}
func funcname_go(*uintptr /* should be *runtime.Func*/) string {
	panic("runtime.funcname_go() not yet implemented")
	return ""
}
func funcentry_go(*uintptr /* should be *runtime.Func*/) uintptr {
	panic("runtime.funcentry_go() not yet implemented")
	return 0
}

////

func cstringToGo(uintptr) string {
	panic("runtime.cstringToGo() not yet implemented")
	return ""
}

func getgoroot() string {
	//panic("runtime.getgoroot() not yet implemented")
	return "" // GOROOT is empty in TARDIS Go
}

// Caller reports file and line number information about function invocations on
// the calling goroutine's stack.  The argument skip is the number of stack frames
// to ascend, with 0 identifying the caller of Caller.  (For historical reasons the
// meaning of skip differs between Caller and Callers.) The return values report the
// program counter, file name, and line number within the file of the corresponding
// call.  The boolean ok is false if it was not possible to recover the information.
func Caller(skip int) (pc uintptr, file string, line int, ok bool) {
	panic("runtime.Caller() not yet implemented")
	return 0, "", 0, false
}

// Callers fills the slice pc with the program counters of function invocations
// on the calling goroutine's stack.  The argument skip is the number of stack frames
// to skip before recording in pc, with 0 identifying the frame for Callers itself and
// 1 identifying the caller of Callers.
// It returns the number of entries written to pc.
func Callers(skip int, pc []uintptr) int {
	panic("runtime.Callers() not yet implemented")
	return 0
}
