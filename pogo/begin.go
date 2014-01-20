// Copyright 2014 Elliott Stoneham and The TARDIS Go Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package pogo

import (
	"code.google.com/p/go.tools/go/exact"
	"code.google.com/p/go.tools/go/ssa"
	"fmt"
	"go/token"
	"os"
	"strconv"

	// TARDIS Go included runtime libraries, so that they get installed and the SSA code can go and load their binary form
	_ "github.com/tardisgo/tardisgo/golibruntime"
	_ "github.com/tardisgo/tardisgo/haxegoruntime"
	_ "github.com/tardisgo/tardisgo/tardisgolib"
)

// global variables to save having to pass them about
var rootProgram *ssa.Program // pointer to the root datastructure TODO make this state non-global
var mainPackage *ssa.Package // pointer to the "main" package TODO make this state non-global

// The entry point for the pogo package, called from ssadump.
// It finishes with an os.Exit() so 0 for success.
func EntryPoint(mainPkg *ssa.Package) {
	mainPackage = mainPkg
	rootProgram = mainPkg.Prog
	setupPosHash()
	emitFileStart()
	emitFunctions()
	emitGoClass(mainPackage)
	emitTypeInfo()
	emitFileEnd()
	if hadErrors && stopOnError {
		LogError("", "pogo", fmt.Errorf("No output files generated"))
		os.Exit(2) // should signal failure to shell
	} else {
		writeFiles()
		os.Exit(0) // should signal success to shell
	}
}

// The main Go class contains those elements that don't fit in functions
func emitGoClass(mainPkg *ssa.Package) {
	emitGoClassStart()
	emitNamedConstants()
	emitGlobals()
	emitGoClassEnd(mainPkg)
}

// special constant name used in TARDIS Go to put text in the header of files
const pogoHeader = "tardisgoHeader"

// special constant name used in TARDIS Go to say where the runtime for the go standard libraries is located
const pogoLibRuntimePath = "tardisgoLibRuntimePath"

var LibRuntimePath string = "github.com/tardisgo/tardisgo/golibruntime" // This value is required to stop the init function in runtime replacement functions being generated.
// NOTE default value can be overwritten if required using the name in pogoLibRuntimePath see above in the code

// emit the standard file header for target language
func emitFileStart() {
	/* TODO - add some sort of dated preamble, perhaps including something like:
	for _, pkg := range rootProgram.PackagesByPath {
		// Print out the package info.
		pkg.DumpTo(os.Stdout)
		fmt.Println()
	}
	*/
	hxPkg := ""
	l := TargetLang
	ph := LanguageList[l].HeaderConstVarName
	targetPackage := LanguageList[l].PackageConstVarName
	header := ""
	allPack := rootProgram.AllPackages()
	for pkgIdx := range allPack {
		pkg := allPack[pkgIdx]
		for mName, mem := range pkg.Members {
			if mem.Token() == token.CONST {
				switch mName {
				case ph, pogoHeader: // either the language-specific constant, or the standard one
					lit := mem.(*ssa.NamedConst).Value
					switch lit.Value.Kind() {
					case exact.String:
						h, err := strconv.Unquote(lit.Value.String())
						if err != nil {
							LogError(CodePosition(lit.Pos())+"Special pogo header constant "+ph+" or "+pogoHeader,
								"pogo", err)
						} else {
							header += h + "\n"
						}
					}
				case targetPackage:
					lit := mem.(*ssa.NamedConst).Value
					switch lit.Value.Kind() {
					case exact.String:
						hp, err := strconv.Unquote(lit.Value.String())
						if err != nil {
							LogError(CodePosition(lit.Pos())+"Special targetPackage constant ", "pogo", err)
						}
						hxPkg = hp
					default:
						LogError(CodePosition(lit.Pos()), "pogo",
							fmt.Errorf("Special targetPackage constant not a string"))
					}
				case pogoLibRuntimePath:
					lit := mem.(*ssa.NamedConst).Value
					switch lit.Value.Kind() {
					case exact.String:
						lrp, err := strconv.Unquote(lit.Value.String())
						if err != nil {
							LogError(CodePosition(lit.Pos())+"Special LibRuntimePath constant ", "pogo", err)
						}
						LibRuntimePath = lrp
					default:
						LogError(CodePosition(lit.Pos()), "pogo",
							fmt.Errorf("Special targetPackage constant not a string"))
					}
				}
			}
		}
	}
	fmt.Fprintln(&LanguageList[l].buffer, LanguageList[l].FileStart(hxPkg, header))
}

// emit the tail of the required language file
func emitFileEnd() {
	l := TargetLang
	fmt.Fprintln(&LanguageList[l].buffer, LanguageList[l].FileEnd())
	for w := range warnings {
		emitComment(warnings[w])
	}
	emitComment("Package List:")
	allPack := rootProgram.AllPackages()
	for pkgIdx := range allPack {
		emitComment(" " + allPack[pkgIdx].String())
	}
}

// emit the start of the top level type definition for each language
func emitGoClassStart() {
	l := TargetLang
	fmt.Fprintln(&LanguageList[l].buffer, LanguageList[l].GoClassStart())
}

// emit the end of the top level type definition for each language file
func emitGoClassEnd(pak *ssa.Package) {
	l := TargetLang
	fmt.Fprintln(&LanguageList[l].buffer, LanguageList[l].GoClassEnd(pak))
}
