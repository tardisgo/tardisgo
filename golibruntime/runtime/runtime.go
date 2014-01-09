// Copyright 2014 Elliott Stoneham and The tardisgo Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Runtime functions for the Go "runtime" standard library package when used by TARDIS Go
package runtime

// THE GOLANG RUNTIME PACKAGE IS NOT CURRENTLY ALL USABLE

import "github.com/tardisgo/tardisgo/tardisgolib"

func Gosched()          { tardisgolib.Gosched() }
func NumGoroutine() int { return tardisgolib.NumGoroutine() }

//	EVERYTHING BELOW NOT YET IMPLEMENTED

func Goexit()                            { panic("runtime.Goexit() not yet implemented") }
func Callers(skip int, pc []uintptr) int { panic("runtime.Callers() not yet implemented"); return 0 }
func FuncForPC(pc uintptr) *uintptr      { panic("runtime.FuncForPC() not yet implemented"); return nil }
func SetFinalizer(x, f interface{})      { panic("runtime.SetFinalizer() not yet implemented") }

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
	panic("runtime.getgoroot() not yet implemented")
	return ""
}
