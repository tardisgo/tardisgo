// Copyright 2014 Elliott Stoneham and The TARDIS Go Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// Package golibruntime provides runtime functions for the Go standard libraries.
// This overall package is incomplete and only compiles on OSX.
// Some individual sub-packages do work in some circumstances, see the example code.
package golibruntime

import _ "github.com/tardisgo/tardisgo/golibruntime/tgoaddmath"

/*
import (
	_ "github.com/tardisgo/tardisgo/golibruntime/bytes" // blank imports are used here because it allows the haxe name-spaces to overlap, TODO find a better method long-term
	_ "github.com/tardisgo/tardisgo/golibruntime/os"
	_ "github.com/tardisgo/tardisgo/golibruntime/runtime"
	_ "github.com/tardisgo/tardisgo/golibruntime/strings"
	_ "github.com/tardisgo/tardisgo/golibruntime/syscall"
	_ "github.com/tardisgo/tardisgo/golibruntime/tgoaddmath"
	//_ "github.com/tardisgo/tardisgo/golibruntime/tgofmt"
	_ "github.com/tardisgo/tardisgo/golibruntime/tgoreflect"
	_ "github.com/tardisgo/tardisgo/golibruntime/tgosync"
	_ "github.com/tardisgo/tardisgo/golibruntime/tgosync/atomic"
	_ "github.com/tardisgo/tardisgo/golibruntime/tgotesting"
	_ "github.com/tardisgo/tardisgo/golibruntime/tgotime"
)
*/
