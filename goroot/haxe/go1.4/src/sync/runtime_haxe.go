// Copyright 2012 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build haxe

// runtime functions rewritten for single-threaded Haxe

package sync

import (
	"runtime"
	"unsafe"
)

// defined in package runtime

// Semacquire waits until *s > 0 and then atomically decrements it.
// It is intended as a simple sleep primitive for use by the synchronization
// library and should not be used directly.
func runtime_Semacquire(s *uint32) {
	for *s < 1 {
		runtime.Gosched()
	}
	*s -= 1
}

// Semrelease atomically increments *s and notifies a waiting goroutine
// if one is blocked in Semacquire.
// It is intended as a simple wakeup primitive for use by the synchronization
// library and should not be used directly.
func runtime_Semrelease(s *uint32) {
	*s += 1
	runtime.Gosched()
}

// Approximation of syncSema in runtime/sema.go.
type syncSema struct {
	lock uintptr
	head unsafe.Pointer
	tail unsafe.Pointer
}

// Syncsemacquire waits for a pairing Syncsemrelease on the same semaphore s.
func runtime_Syncsemacquire(s *syncSema) {
	panic("TODO:sync.runtime_Syncsemaquire")
}

// Syncsemrelease waits for n pairing Syncsemacquire on the same semaphore s.
func runtime_Syncsemrelease(s *syncSema, n uint32) {
	panic("TODO:sync.runtime_Syncsemrelease")
}

// Ensure that sync and runtime agree on size of syncSema.
func runtime_Syncsemcheck(size uintptr) {}

func init() {
	var s syncSema
	runtime_Syncsemcheck(unsafe.Sizeof(s))
}
