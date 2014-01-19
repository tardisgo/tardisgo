// Copyright 2014 Elliott Stoneham and The TARDIS Go Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package haxe

import (
	"code.google.com/p/go.tools/go/ssa"
	"code.google.com/p/go.tools/go/types"
)

//TODO rewrite as very inefficiant, espcially for slices, and possibly not side-effect-free
func (l langType) append(args []ssa.Value, errorInfo string) string {
	typeElem := l.LangType(args[0].Type().Underlying().(*types.Slice).Elem().Underlying(), false, errorInfo)
	initElem := l.LangType(args[0].Type().Underlying().(*types.Slice).Elem().Underlying(), true, errorInfo)
	source := l.IndirectValue(args[1], errorInfo)
	if l.LangType(args[1].Type().Underlying(), false, errorInfo) == "String" {
		source = "Force.toUTF8slice(this._goroutine," + source + /* "," + ph +*/ ")" // if we have a string, we must convert it to a slice
	}
	lengthOverall := l.IndirectValue(args[0], errorInfo) +
		".len()+" + source + ".len()"

	ret := "{var _v:" + l.LangType(args[0].Type().Underlying(), false, errorInfo) + ";"
	ret += "if(" + l.IndirectValue(args[0], errorInfo) +
		"==null) _v=" + source + ";"
	ret += "else if(" + l.IndirectValue(args[0], errorInfo) + ".len()==0) _v=" +
		source + ";"
	ret += "else if(" + source + "==null) _v=" +
		l.IndirectValue(args[0], errorInfo) + ";"
	ret += "else if(" + source + ".len()==0) _v=" +
		l.IndirectValue(args[0], errorInfo) + ";"
	ret += "else {var l0:Int=" + l.IndirectValue(args[0], errorInfo) + ".len();"
	ret += "_v=" + newSliceCode(typeElem, initElem, lengthOverall, lengthOverall, errorInfo) + ";" //len==cap
	ret += "for(_i in 0...l0) _v.setAt(_i,Deep.copy(" + l.IndirectValue(args[0], errorInfo) +
		".getAt(_i)));"
	ret += "for(_i in 0..." + source + ".len()) _v.setAt(_i+l0,Deep.copy(" +
		source + ".getAt(_i)));"
	return ret + "};_v;}"
}
