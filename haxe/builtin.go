// Copyright 2014 Elliott Stoneham and The TARDIS Go Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package haxe

import "golang.org/x/tools/go/ssa"

func (l langType) append(args []ssa.Value, errorInfo string) string {
	source := l.IndirectValue(args[1], errorInfo)
	if l.LangType(args[1].Type().Underlying(), false, errorInfo) == "String" {
		source = "Force.toUTF8slice(this._goroutine," + source + ")" // if we have a string, we must convert it to a slice
	}
	target := l.IndirectValue(args[0], errorInfo)
	ret := target + ".append(" + source + ")"
	if target == "null" {
		ret = source
	}
	//fmt.Printf("APPEND DEBUG: %s - %+v - %s\n", ulSize, args, ret)

	return ret
}

func (l langType) copy(register string, args []ssa.Value, errorInfo string) string {
	ret := ""
	if register != "" {
		ret += register
	}
	source := l.IndirectValue(args[1], errorInfo)
	if l.LangType(args[1].Type().Underlying(), false, errorInfo) == "String" {
		source = "Force.toUTF8slice(this._goroutine," + source + ")" // if we have a string, we must convert it to a slice
	}
	code := l.IndirectValue(args[0], errorInfo) + ".copy(" + source + ")"
	// TODO consider makting this a runtime function
	return ret + code
}
