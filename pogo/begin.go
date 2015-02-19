// Copyright 2014 Elliott Stoneham and The TARDIS Go Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package pogo

import (
	"fmt"
	"go/token"
	"strconv"

	"golang.org/x/tools/go/exact"
	"golang.org/x/tools/go/ssa"
)

// global variables to save having to pass them about (TODO these, and other, status vars should be in a structure)
var rootProgram *ssa.Program // pointer to the root datastructure TODO make this state non-global
var mainPackage *ssa.Package // pointer to the "main" package TODO make this state non-global

// DebugFlag is used to signal if we are emitting debug information
var DebugFlag bool

// TraceFlag is used to signal if we are emitting trace information (big)
var TraceFlag bool

// EntryPoint provides the entry point for the pogo package, called from ssadump_copy.
func EntryPoint(mainPkg *ssa.Package) error {
	mainPackage = mainPkg
	rootProgram = mainPkg.Prog

	for k, v := range LanguageList {
		//fmt.Println("DEBUG Language List[", k, "]=", v.LanguageName())
		if v.LanguageName() == "haxe" {
			TargetLang = k
			//fmt.Println("DEBUG Target Language =", k)
			break
		}
	}

	setupPosHash()
	emitFileStart()
	emitFunctions()
	emitGoClass(mainPackage)
	emitTypeInfo()
	emitFileEnd()
	if hadErrors && stopOnError {
		err := fmt.Errorf("no output files generated")
		LogError("", "pogo", err)
		return err
	}
	writeFiles()
	return nil
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
//const pogoLibRuntimePath = "tardisgoLibRuntimePath"

// LibRuntimePath is required to stop the init function in runtime replacement functions being generated.
// NOTE default value can be overwritten if required using the name in pogoLibRuntimePath see above in the code
//var LibRuntimePath = "github.com/tardisgo/tardisgo/golibruntime"

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
							fmt.Errorf("special targetPackage constant not a string"))
					}
					//case pogoLibRuntimePath:
					//	lit := mem.(*ssa.NamedConst).Value
					//	switch lit.Value.Kind() {
					//	case exact.String:
					//		lrp, err := strconv.Unquote(lit.Value.String())
					//		if err != nil {
					//			LogError(CodePosition(lit.Pos())+"Special LibRuntimePath constant ", "pogo", err)
					//		}
					//		LibRuntimePath = lrp
					//	default:
					//		LogError(CodePosition(lit.Pos()), "pogo",
					//			fmt.Errorf("special targetPackage constant not a string"))
					//	}
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

func UsingPackage(pkgName string) bool {
	//println("DEBUG UsingPackage() looking for: ", pkgName)
	pkgName = "package " + pkgName
	pkgs := rootProgram.AllPackages()
	for p := range pkgs {
		//println("DEBUG UsingPackage() considering pkg: ", pkgs[p].String())
		if pkgs[p].String() == pkgName {
			//println("DEBUG UsingPackage()  ", pkgName, " = true")
			return true
		}
	}
	//println("DEBUG UsingPackage()  ", pkgName, " =false")
	return false
}
