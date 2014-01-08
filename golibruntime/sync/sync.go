// Copyright 2014 Elliott Stoneham and The tardisgo Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Built using code:
// Copyright 2012 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sync

// THIS PACKAGE ONLY PARTLY USABLE

//import "unsafe"

// defined in package runtime

// Semacquire waits until *s > 0 and then atomically decrements it.
// It is intended as a simple sleep primitive for use by the synchronization
// library and should not be used directly.
func runtime_Semacquire(s *uint32) {
	for *s == 0 {
		runtime_Syncsemcheck(0) // effectivly gosched, as a no-op
	}
	*s--
}

// Semrelease atomically increments *s and notifies a waiting goroutine
// if one is blocked in Semacquire.
// It is intended as a simple wakeup primitive for use by the synchronization
// library and should not be used directly.
func runtime_Semrelease(s *uint32) {
	*s++
}

// Opaque representation of SyncSema in runtime/sema.goc.
type syncSema [3]uintptr

// Syncsemacquire waits for a pairing Syncsemrelease on the same semaphore s.
func runtime_Syncsemacquire(s *syncSema) {
	panic("runtime_Syncsemacquire not yet implemented")
}

// Syncsemrelease waits for n pairing Syncsemacquire on the same semaphore s.
func runtime_Syncsemrelease(s *syncSema, n uint32) {
	panic("runtime_Syncsemrelease not yet implemented")
}

// Ensure that sync and runtime agree on size of syncSema.
func runtime_Syncsemcheck(size uintptr) {
	/*NoOp*/
}

// INIT FUNCTIONS IN THESE RUNTIME LIBRARIES ARE NOT USED
//func init() {
//	var s syncSema
//	runtime_Syncsemcheck(unsafe.Sizeof(s))
//}
