// Copyright 2014 Elliott Stoneham and The tardisgo Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Some code below:
// Copyright 2010 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Runtime functions for the Go "bytes" standard library package when used by TARDIS Go
package bytes

//****go:noescape

// IndexByte returns the index of the first instance of c in s, or -1 if c is not present in s.
func IndexByte(s []byte, c byte) int {
	for i := range s {
		if s[i] == c {
			return i
		}
	}
	return -1
} // asm_$GOARCH.s

//****go:noescape

// Equal returns a boolean reporting whether a == b.
// A nil argument is equivalent to an empty slice.
func Equal(a, b []byte) bool {
	if a == nil {
		if b == nil {
			return true
		} else {
			return len(b) == 0
		}
	}
	if b == nil {
		return len(a) == 0
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
} // asm_arm.s or ../runtime/asm_{386,amd64}.s

//****go:noescape

// Compare returns an integer comparing two byte slices lexicographically.
// The result will be 0 if a==b, -1 if a < b, and +1 if a > b.
// A nil argument is equivalent to an empty slice.
func Compare(a, b []byte) int {
	if a == nil {
		if b == nil {
			return 0
		} else {
			if len(b) == 0 {
				return 0
			} else {
				return 1
			}
		}
	}
	if b == nil {
		if len(a) == 0 {
			return 0
		} else {
			return 1
		}
	}
	i := 0
	for (i < len(a)) && (i < len(b)) {
		if a[i] < b[i] {
			return -1
		}
		if a[i] > b[i] {
			return +1
		}
		i++
	}
	if len(a) == len(b) {
		return 0
	}
	if len(a) < len(b) {
		return -1
	}
	return +1
} // ../runtime/noasm_arm.goc or ../runtime/asm_{386,amd64}.s
