// Copyright 2014 Elliott Stoneham and The TARDIS Go Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package pogo

import (
	"fmt"
	"go/token"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/tools/go/exact"
	"golang.org/x/tools/go/ssa"
)

// Recycle the Compilation resources
func (comp *Compilation) Recycle() { LanguageList[comp.TargetLang] = LanguageEntry{} }

// EntryPoint provides the entry point for the pogo package, called from ssadump_copy.
func EntryPoint(mainPkg *ssa.Package, debug, trace bool, langName string) (*Compilation, error) {
	comp := &Compilation{
		mainPackage: mainPkg,
		rootProgram: mainPkg.Prog,
		DebugFlag:   debug,
		TraceFlag:   trace,
	}

	k, e := FindTargetLang(langName)
	if e != nil {
		return nil, e
	}
	// make a new language list entry for this compilation
	newLang := LanguageList[k]
	languageListAppendMutex.Lock()
	LanguageList = append(LanguageList, newLang)
	comp.TargetLang = len(LanguageList) - 1
	languageListAppendMutex.Unlock()
	LanguageList[comp.TargetLang].Language =
		LanguageList[comp.TargetLang].Language.InitLang(comp.TargetLang,comp)
	//fmt.Printf("DEBUG created TargetLang[%d]=%#v\n",
	//	comp.TargetLang, LanguageList[comp.TargetLang])

	comp.initErrors()
	comp.initTypes()
	comp.setupPosHash()
	comp.loadSpecialConsts()
	comp.emitFileStart()
	comp.emitFunctions()
	comp.emitGoClass(comp.mainPackage)
	comp.emitTypeInfo()
	comp.emitFileEnd()
	if comp.hadErrors && comp.stopOnError {
		err := fmt.Errorf("no output files generated")
		comp.LogError("", "pogo", err)
		return nil, err
	}
	comp.writeFiles()
	return comp, nil
}

// The main Go class contains those elements that don't fit in functions
func (comp *Compilation) emitGoClass(mainPkg *ssa.Package) {
	comp.emitGoClassStart()
	comp.emitNamedConstants()
	comp.emitGlobals()
	comp.emitGoClassEnd(mainPkg)
	comp.WriteAsClass("Go", "")
}

// special constant name used in TARDIS Go to put text in the header of files
const pogoHeader = "tardisgoHeader"
const pogoLibList = "tardisgoLibList"

func (comp *Compilation) loadSpecialConsts() {
	hxPkg := ""
	l := comp.TargetLang
	ph := LanguageList[l].HeaderConstVarName
	targetPackage := LanguageList[l].PackageConstVarName
	header := ""
	allPack := comp.rootProgram.AllPackages()
	sort.Sort(PackageSorter(allPack))
	for _, pkg := range allPack {
		allMem := MemberNamesSorted(pkg)
		for _, mName := range allMem {
			mem := pkg.Members[mName]
			if mem.Token() == token.CONST {
				switch mName {
				case ph, pogoHeader: // either the language-specific constant, or the standard one
					lit := mem.(*ssa.NamedConst).Value
					switch lit.Value.Kind() {
					case exact.String:
						h, err := strconv.Unquote(lit.Value.String())
						if err != nil {
							comp.LogError(comp.CodePosition(lit.Pos())+"Special pogo header constant "+ph+" or "+pogoHeader,
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
							comp.LogError(comp.CodePosition(lit.Pos())+"Special targetPackage constant ", "pogo", err)
						}
						hxPkg = hp
					default:
						comp.LogError(comp.CodePosition(lit.Pos()), "pogo",
							fmt.Errorf("special targetPackage constant not a string"))
					}
				case pogoLibList:
					lit := mem.(*ssa.NamedConst).Value
					switch lit.Value.Kind() {
					case exact.String:
						lrp, err := strconv.Unquote(lit.Value.String())
						if err != nil {
							comp.LogError(comp.CodePosition(lit.Pos())+"Special "+pogoLibList+" constant ", "pogo", err)
						}
						comp.LibListNoDCE = strings.Split(lrp, ",")
						for lib := range comp.LibListNoDCE {
							comp.LibListNoDCE[lib] = strings.TrimSpace(comp.LibListNoDCE[lib])
						}
					default:
						comp.LogError(comp.CodePosition(lit.Pos()), "pogo",
							fmt.Errorf("special targetPackage constant not a string"))
					}
				}
			}
		}
	}
	comp.hxPkgName = hxPkg
	comp.headerText = header
}

// emit the standard file header for target language
func (c *Compilation) emitFileStart() {
	l := c.TargetLang
	fmt.Fprintln(&LanguageList[l].buffer, LanguageList[l].FileStart(c.hxPkgName, c.headerText))
}

// emit the tail of the required language file
func (c *Compilation) emitFileEnd() {
	l := c.TargetLang
	fmt.Fprintln(&LanguageList[l].buffer, LanguageList[l].FileEnd())
	for w := range c.warnings {
		c.emitComment(c.warnings[w])
	}
	c.emitComment("Package List:")
	allPack := c.rootProgram.AllPackages()
	sort.Sort(PackageSorter(allPack))
	for pkgIdx := range allPack {
		c.emitComment(" " + allPack[pkgIdx].String())
	}
}

// emit the start of the top level type definition for each language
func (c *Compilation) emitGoClassStart() {
	l := c.TargetLang
	fmt.Fprintln(&LanguageList[l].buffer, LanguageList[l].GoClassStart())
}

// emit the end of the top level type definition for each language file
func (c *Compilation) emitGoClassEnd(pak *ssa.Package) {
	l := c.TargetLang
	fmt.Fprintln(&LanguageList[l].buffer, LanguageList[l].GoClassEnd(pak))
}

func (c *Compilation) UsingPackage(pkgName string) bool {
	//println("DEBUG UsingPackage() looking for: ", pkgName)
	pkgName = "package " + pkgName
	pkgs := c.rootProgram.AllPackages()
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
