// Copyright 2014 Elliott Stoneham and The tardisgo Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package tgoruntime provdies functions for the Go "runtime" standard library package when used by TARDIS Go.
// Please note, this is an incomplete implementation.
package runtime

// THE GOLANG RUNTIME PACKAGE IS NOT CURRENTLY ALL USABLE

import (
	"github.com/tardisgo/tardisgo/haxe/hx"
)

func init() {
}

// Haxe specific
func UnzipTestFS() {} // this will be overwritten by the compiler

// Constant values

const Compiler = "TARDISgo" // this is checked by the proper runtime, so might need to be "gc"

var GOARCH string = hx.CallString("", "Go.Platform", 0) // this is a const in the main Go installation

const GOOS string = "nacl" // of course it is only an emulation of nacl...

var MemProfileRate int = 512 * 1024 // TODO not currently used

// NOT YET IMPLEMENTED

type BlockProfileRecord struct {
	Count  int64
	Cycles int64
	StackRecord
}

func BlockProfile(p []BlockProfileRecord) (n int, ok bool) {
	panic("TODO:runtime.BlockProfile")
	return
}

func Breakpoint() {} // this will be overwritten by the compiler

func CPUProfile() []byte {
	panic("TODO:runtime.CPUProfile")
}

func GoroutineProfile(p []StackRecord) (n int, ok bool) {
	panic("TODO:runtime.GoroutineProfile")
	return
}

func MemProfile(p []MemProfileRecord, inuseZero bool) (n int, ok bool) {
	panic("TODO:runtime.MemProfile")
	return
}

func ReadMemStats(m *MemStats) {
	//panic("TODO:runtime.ReadMemStats")
	// bytes testing calls this, so no-op
}
func ThreadCreateProfile(p []StackRecord) (n int, ok bool) {
	panic("TODO:runtime.ThreadCreateProfile")
	return
}

type MemProfileRecord struct {
	AllocBytes, FreeBytes     int64       // number of bytes allocated, freed
	AllocObjects, FreeObjects int64       // number of objects allocated, freed
	Stack0                    [32]uintptr // stack trace for this record; ends at first 0 entry
}

func (r *MemProfileRecord) InUseBytes() int64   { return 0 }
func (r *MemProfileRecord) InUseObjects() int64 { return 0 }
func (r *MemProfileRecord) Stack() []uintptr    { return nil }

func Goexit() {
	panic("TODO:runtime.Goexit")
}

type MemStats struct {
	// General statistics.
	Alloc      uint64 // bytes allocated and still in use
	TotalAlloc uint64 // bytes allocated (even if freed)
	Sys        uint64 // bytes obtained from system (sum of XxxSys below)
	Lookups    uint64 // number of pointer lookups
	Mallocs    uint64 // number of mallocs
	Frees      uint64 // number of frees

	// Main allocation heap statistics.
	HeapAlloc    uint64 // bytes allocated and still in use
	HeapSys      uint64 // bytes obtained from system
	HeapIdle     uint64 // bytes in idle spans
	HeapInuse    uint64 // bytes in non-idle span
	HeapReleased uint64 // bytes released to the OS
	HeapObjects  uint64 // total number of allocated objects

	// Low-level fixed-size structure allocator statistics.
	//	Inuse is bytes used now.
	//	Sys is bytes obtained from system.
	StackInuse  uint64 // bytes used by stack allocator
	StackSys    uint64
	MSpanInuse  uint64 // mspan structures
	MSpanSys    uint64
	MCacheInuse uint64 // mcache structures
	MCacheSys   uint64
	BuckHashSys uint64 // profiling bucket hash table
	GCSys       uint64 // GC metadata
	OtherSys    uint64 // other system allocations

	// Garbage collector statistics.
	NextGC       uint64 // next collection will happen when HeapAlloc ≥ this amount
	LastGC       uint64 // end time of last collection (nanoseconds since 1970)
	PauseTotalNs uint64
	PauseNs      [256]uint64 // circular buffer of recent GC pause durations, most recent at [(NumGC+255)%256]
	PauseEnd     [256]uint64 // circular buffer of recent GC pause end times
	NumGC        uint32
	EnableGC     bool
	DebugGC      bool

	// Per-size allocation statistics.
	// 61 is NumSizeClasses in the C code.
	BySize [61]struct {
		Size    uint32
		Mallocs uint64
		Frees   uint64
	}
}

type StackRecord struct {
	Stack0 [32]uintptr // stack trace for this record; ends at first 0 entry
}

func (r *StackRecord) Stack() []uintptr { return nil }

type TypeAssertionError struct {
	// contains filtered or unexported fields
}

func (e *TypeAssertionError) Error() string { return "TODO:runtime.TypeAssertionError.Error" }
func (*TypeAssertionError) RuntimeError()   {}

// NO-OP functions

func SetBlockProfileRate(rate int) {}

func SetCPUProfileRate(hz int) {}

func NumCPU() int { return 1 }

func GOMAXPROCS(n int) int { return 1 }

func NumCgoCall() int64 { return 0 }

func GOROOT() string { return "" } // TODO set as compile time value

func Version() string { return "go1.4" } // TODO automate this

func SetFinalizer(obj interface{}, finalizer interface{}) {}

func GC() {}

func LockOSThread()   {}
func UnlockOSThread() {}

type Error interface {
	error

	// RuntimeError is a no-op function but
	// serves to distinguish types that are runtime
	// errors from ordinary errors: a type is a
	// runtime error if it has a RuntimeError method.
	RuntimeError()
}

// Part-Implemented

func Callers(skip int, pc []uintptr) int {
	limit := hx.CallInt("", "Scheduler.getNumCallers", 1, hx.GetInt("", "this._goroutine"))
	for i := 0; i < limit; i++ {
		if i > skip {
			pc = append(pc, uintptr(hx.CallInt("", "Scheduler.getCallerX", 2, hx.GetInt("", "this._goroutine"), i)))
		}
	}
	return limit - skip
}

func Caller(skip int) (pc uintptr, file string, line int, ok bool) {
	pc = uintptr(hx.CallInt("", "Scheduler.getCallerX", 2, hx.GetInt("", "this._goroutine"), 1+skip))
	fnc := FuncForPC(pc)
	file, line = fnc.FileLine(pc)
	ok = true
	if file == "" || line == 0 {
		ok = false
	}
	return
}

type Func struct {
	// contains filtered or unexported fields
}

func FuncForPC(pc uintptr) *Func {
	//FuncForPC returns a *Func describing the function that contains the given program counter address, or else nil.
	return nil
}

func (f *Func) Entry() uintptr {
	//Entry returns the entry address of the function.
	return 0 // TODO
}

func (f *Func) FileLine(pc uintptr) (file string, line int) {
	//FileLine returns the file name and line number of the source code corresponding to the program counter pc. The result will not be accurate if pc is not a program counter within f.
	detail := hx.CallString("", "Go.CPos", 1, pc)
	if len(detail) == 0 {
		return "", 0
	}
	if detail[0] == '(' { // error return
		return "", 0
	}
	if detail[0:5] == "near " {
		detail = detail[5:]
	}
	for i, c := range detail {
		if c == ':' {
			file = detail[:i]
			x := hx.CallIface("", "int", "Std.parseInt", 1, detail[i+1:])
			if x == nil {
				return "", 0
			}
			line = x.(int)
			return file, line
		}
	}
	// should never get here
	return "", 0
}

func (f *Func) Name() string {
	return "TODO:runtime.Func.Name"
}

var gosched_chan = make(chan interface{})

// Gosched schedules other goroutines.
// gets run a great deal, so need it to be as small as possible
func Gosched() {
	//select {
	//case <-gosched_chan: // should never happen
	//	return
	//default:
	//	return
	//}
	_ = gosched_chan // NOTE referencing a channel will mean this function cannot be optimized to not use goroutines
}

// NumGoroutine returns the number of active goroutines (may be more than the number runable).
func NumGoroutine() int {
	return hx.CallInt("", "Scheduler.NumGoroutine", 0)
}

func Stack(buf []byte, all bool) int {
	// TODO: use the all flag!
	s := hx.CallString("", "Scheduler.stackDump", 0)
	if len(s) > len(buf) {
		return 0
	}
	for i := 0; i < len(s); i++ {
		buf[i] = s[i]
	}
	return len(s)
}
