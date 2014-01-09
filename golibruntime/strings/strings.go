// Copyright 2014 Elliott Stoneham and The tardisgo Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Function signature:
// Copyright 2013 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Runtime function for the Go "strings" standard library package when used by TARDIS Go
package strings

// IndexByte returns the index of the first instance of c in s, or -1 if c is not present in s.
func IndexByte(s string, c byte) int {
	sb := []byte(s)
	for i := 0; i < len(sb); i++ {
		if sb[i] == c {
			return i
		}
	}
	return -1
} // ../runtime/asm_$GOARCH.s
