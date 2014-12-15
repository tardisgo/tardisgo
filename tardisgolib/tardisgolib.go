// Copyright 2014 Elliott Stoneham and The tardisgo Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package tardisgolib provides utility library functions for Go code targeting TARDIS Go
package tardisgolib

import "github.com/tardisgo/tardisgo/tardisgolib/hx"

// Host returns the Host language (i.e. "go" or "haxe"), the return value is overridden to give correct host language name
func Host() string { return "go" }

// Platform returns language specific the Platform information, the return value is overridden at runtime
// for "Haxe" as host this returns the target language platform as one of: "flash","js","neko","php","cpp","java","cs"
func Platform() string { return "go" }

// CPos returns a string containing the Go code position in terms of file name and line number
func CPos() string { return "<<go code pos>>" } // the return value is overwridden by the transpiler, here just for Go use

// Zilen returns the runtime native string length of the chinese character "字", meaning "written character", which is pronounced "zi" in Mandarin.
// For UTF8 encoding this value is 3, for UTF16 encoding this value is 1.
func Zilen() uint { return 3 }

// StringsUTF8 returns a boolian answering: Is the native string encoding UTF8?
func StringsUTF8() bool { return Zilen() == 3 }

// StringsUTF16 returns a boolian answering: Is the native string encoding UTF16?
func StringsUTF16() bool { return Zilen() == 1 }

// ThisGoroutine returns a unique number of the current goroutine.
func ThisGoroutine() int {
	return hx.CallInt("", "Scheduler.ThisGoroutine", 0)
}

/*
	Replicant functions of the go "runtime" package, using these rather than the runtime package generates less Haxe code
*/

var gosched_chan = make(chan interface{})

// Gosched schedules other goroutines.
func Gosched() {
	select {
	case <-gosched_chan: // should never happen
		return
	default:
		return
	}
}

// NumGoroutine returns the number of active goroutines (may be more than the number runable).
func NumGoroutine() int {
	return hx.CallInt("", "Scheduler.NumGoroutine", 0)
}
