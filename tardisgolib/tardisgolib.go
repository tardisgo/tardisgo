// Copyright 2014 Elliott Stoneham and The tardisgo Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Utility library functions for TARDIS Go
package tardisgolib

// Insert the given Haxe code at this point.
// BEWARE! It is very easy to write code that will break the system.
// code string must be a constant containing a well-formed Haxe statement, probably terminated with a ";"
// ret is a Haxe Dynamic value mapped into Go as a uintptr
func HAXE(code string) (ret uintptr) { return }

// returns the Host language (i.e. "go" or "haxe"), the return value is overridden to give correct host language name
func Host() string { return "go" }

// Return language specific the Platform information, the return value is overridden at runtime
// for "Haxe" as host this returns the target language platform as one of: "flash","js","neko","php","cpp","java","cs"
func Platform() string { return "go" }

// Return a string containing the Go code position in terms of file name and line number
func CPos() string { return "<<go code pos>>" } // the return value is overwridden by the transpiler, here just for Go use

// Zilen() returns the runtime native string length of the chinese character "字", meaning "written character", which is pronounced "zi" in Mandarin.
// For UTF8 encoding this value is 3, for UTF16 encoding this value is 1.
func Zilen() uint { return 3 }

// Is the native string encoding UTF8?
func StringsUTF8() bool { return Zilen() == 3 }

// Is the native string encoding UTF16?
func StringsUTF16() bool { return Zilen() == 1 }

/*
	Replicant functions of the go "runtime" package, using these rather than the runtime package generates less Haxe code
*/

// Schedule other goroutines.
func Gosched() {} // an empty function here works fine to enable other goroutines to be scheduled

// Number of goroutines (may be more than the number runable).
func NumGoroutine() int { return int(HAXE("Scheduler.NumGoroutine();")) }
