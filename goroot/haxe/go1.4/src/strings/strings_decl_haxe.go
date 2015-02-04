// Copyright 2014 Elliott Stoneham and The tardisgo Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Function signature:
// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build haxe

// Package strings contains a runtime function for the Go "strings" standard library package when used by TARDIS Go
package strings

import "github.com/tardisgo/tardisgo/haxe/hx"

// IndexByte returns the index of the first instance of c in s, or -1 if c is not present in s.
func IndexByte(s string, c byte) int {
	if s == "" {
		return -1
	}
	return hx.CodeInt("", "cast(_a.itemAddr(0).load().val,String).indexOf(_a.itemAddr(1).load().val);",
		s, string(rune(c)))
	/*
		sb := []byte(s)
		for i := 0; i < len(sb); i++ {
			if sb[i] == c {
				return i
			}
		}
		return -1
	*/
} // ../runtime/asm_$GOARCH.s
