// Copyright 2014 Elliott Stoneham and The tardisgo Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package golibruntime // runtime finctions for the Go libraries in subdirectories

import (
	_ "github.com/tardisgo/tardisgo/golibruntime/bytes"
	_ "github.com/tardisgo/tardisgo/golibruntime/math"
	_ "github.com/tardisgo/tardisgo/golibruntime/runtime"
	_ "github.com/tardisgo/tardisgo/golibruntime/strings"
	_ "github.com/tardisgo/tardisgo/golibruntime/sync"
	_ "github.com/tardisgo/tardisgo/golibruntime/sync/atomic"
)
