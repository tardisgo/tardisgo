// Copyright 2014 Elliott Stoneham and The tardisgo Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

//Original:
// Copyright 2011 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package atomic contains runtime functions for the Go "sync/atomic" standard library package when used by TARDIS Go
//
package atomic

import (
	"unsafe"
)

// Non-parallel basic replacement for the sync/atomic package
//***********************************************************

// *********** ignore: +build !race

// Package atomic provides low-level atomic memory primitives
// useful for implementing synchronization algorithms.
//
// These functions require great care to be used correctly.
// Except for special, low-level applications, synchronization is better
// done with channels or the facilities of the sync package.
// Share memory by communicating;
// don't communicate by sharing memory.
//
// The compare-and-swap operation, implemented by the CompareAndSwapT
// functions, is the atomic equivalent of:
//
//	if *addr == old {
//		*addr = new
//		return true
//	}
//	return false
//
// The add operation, implemented by the AddT functions, is the atomic
// equivalent of:
//
//	*addr += delta
//	return *addr
//
// The load and store operations, implemented by the LoadT and StoreT
// functions, are the atomic equivalents of "return *addr" and
// "*addr = val".
//

// CompareAndSwapInt32 executes the compare-and-swap operation for an int32 value.
func CompareAndSwapInt32(addr *int32, old, new int32) (swapped bool) {
	if *addr == old {
		*addr = new
		return true
	}
	return false
}

// CompareAndSwapInt64 executes the compare-and-swap operation for an int64 value.
func CompareAndSwapInt64(addr *int64, old, new int64) (swapped bool) {
	if *addr == old {
		*addr = new
		return true
	}
	return false
}

// CompareAndSwapUint32 executes the compare-and-swap operation for a uint32 value.
func CompareAndSwapUint32(addr *uint32, old, new uint32) (swapped bool) {
	if *addr == old {
		*addr = new
		return true
	}
	return false
}

// CompareAndSwapUint64 executes the compare-and-swap operation for a uint64 value.
func CompareAndSwapUint64(addr *uint64, old, new uint64) (swapped bool) {
	if *addr == old {
		*addr = new
		return true
	}
	return false
}

// CompareAndSwapUintptr executes the compare-and-swap operation for a uintptr value.
func CompareAndSwapUintptr(addr *uintptr, old, new uintptr) (swapped bool) {
	if *addr == old {
		*addr = new
		return true
	}
	return false
}

// CompareAndSwapPointer executes the compare-and-swap operation for a unsafe.Pointer value.
func CompareAndSwapPointer(addr *unsafe.Pointer, old, new unsafe.Pointer) (swapped bool) {
	if *addr == old {
		*addr = new
		return true
	}
	return false
}

// AddInt32 atomically adds delta to *addr and returns the new value.
func AddInt32(addr *int32, delta int32) (new int32) {
	*addr += delta
	return *addr
}

// AddUint32 atomically adds delta to *addr and returns the new value.
func AddUint32(addr *uint32, delta uint32) (new uint32) {
	*addr += delta
	return *addr
}

// AddInt64 atomically adds delta to *addr and returns the new value.
func AddInt64(addr *int64, delta int64) (new int64) {
	*addr += delta
	return *addr
}

// AddUint64 atomically adds delta to *addr and returns the new value.
func AddUint64(addr *uint64, delta uint64) (new uint64) {
	*addr += delta
	return *addr
}

// AddUintptr atomically adds delta to *addr and returns the new value.
func AddUintptr(addr *uintptr, delta uintptr) (new uintptr) {
	*addr += delta
	return *addr
}

// LoadInt32 atomically loads *addr.
func LoadInt32(addr *int32) (val int32) { return *addr }

// LoadInt64 atomically loads *addr.
func LoadInt64(addr *int64) (val int64) { return *addr }

// LoadUint32 atomically loads *addr.
func LoadUint32(addr *uint32) (val uint32) { return *addr }

// LoadUint64 atomically loads *addr.
func LoadUint64(addr *uint64) (val uint64) { return *addr }

// LoadUintptr atomically loads *addr.
func LoadUintptr(addr *uintptr) (val uintptr) { return *addr }

// LoadPointer atomically loads *addr.
func LoadPointer(addr *unsafe.Pointer) (val unsafe.Pointer) { return *addr }

// StoreInt32 atomically stores val into *addr.
func StoreInt32(addr *int32, val int32) { *addr = val }

// StoreInt64 atomically stores val into *addr.
func StoreInt64(addr *int64, val int64) { *addr = val }

// StoreUint32 atomically stores val into *addr.
func StoreUint32(addr *uint32, val uint32) { *addr = val }

// StoreUint64 atomically stores val into *addr.
func StoreUint64(addr *uint64, val uint64) { *addr = val }

// StoreUintptr atomically stores val into *addr.
func StoreUintptr(addr *uintptr, val uintptr) { *addr = val }

// StorePointer atomically stores val into *addr.
func StorePointer(addr *unsafe.Pointer, val unsafe.Pointer) { *addr = val }

// this only for the SSA compiler, will not be code generated
func init() {
	/*
		if false {
			CompareAndSwapInt32(nil, 0, 0)
			CompareAndSwapInt64(nil, 0, 0)
			CompareAndSwapUint32(nil, 0, 0)
			CompareAndSwapUint64(nil, 0, 0)
			CompareAndSwapUintptr(nil, 0, 0)
			CompareAndSwapPointer(nil, nil, nil)

			AddInt32(nil, 0)
			AddUint32(nil, 0)
			AddInt64(nil, 0)
			AddUint64(nil, 0)
			AddUintptr(nil, 0)

			LoadInt32(nil)
			LoadInt64(nil)
			LoadUint32(nil)
			LoadUint64(nil)
			LoadUintptr(nil)
			LoadPointer(nil)

			StoreInt32(nil, 0)
			StoreInt64(nil, 0)
			StoreUint32(nil, 0)
			StoreUint64(nil, 0)
			StoreUintptr(nil, 0)
			StorePointer(nil, nil)
		}
	*/
}
