// Copyright 2014 Elliott Stoneham and The TARDIS Go Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package haxe

import "github.com/tardisgo/tardisgo/pogo"

func init() {
	var langVar langType
	var langEntry pogo.LanguageEntry
	langEntry.Language = langVar

	il := 1024 // 1024 is an internal Haxe C# limit (`lvregs_len < 1024`)

	langEntry.InstructionLimit = il      /* size before we make subfns */
	langEntry.SubFnInstructionLimit = il /* 256 required for php */
	langEntry.PackageConstVarName = "tardisgoHaxePackage"
	langEntry.HeaderConstVarName = "tardisgoHaxeHeader"
	langEntry.Goruntime = "haxegoruntime" // a string containing the location of the core language runtime functions delivered in Go
	langEntry.PseudoPkgPaths = []string{"github.com/tardisgo/tardisgo/haxe/hx"}
	langEntry.LineCommentMark = "//"
	langEntry.StatementTerminator = ";"
	langEntry.IgnorePrefixes = []string{"this.setPH("}
	langEntry.GOROOT = "/src/github.com/tardisgo/tardisgo/goroot/haxe/go1.4"
	langEntry.TgtDir = "tardis" // TODO move to the correct directory based on a command line argument

	pogo.LanguageList = append(pogo.LanguageList, langEntry)
}
