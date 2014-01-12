// Copyright 2014 Elliott Stoneham and The TARDIS Go Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Runtime functions for the Go standard libraries
package golibruntime

import (
	_ "github.com/tardisgo/tardisgo/golibruntime/bytes"
	_ "github.com/tardisgo/tardisgo/golibruntime/math"
	_ "github.com/tardisgo/tardisgo/golibruntime/runtime"
	_ "github.com/tardisgo/tardisgo/golibruntime/strings"
	_ "github.com/tardisgo/tardisgo/golibruntime/sync"
	_ "github.com/tardisgo/tardisgo/golibruntime/sync/atomic"
	_ "runtime" // TODO currently fails with a MStats vs MemStatsType size mis-match on 32-bit Ubuntu, works on OSX
)
