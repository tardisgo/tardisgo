// Copyright 2014 Elliott Stoneham and The TARDIS Go Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package pogo

import (
	"fmt"
	"go/token"
	"strings"
	"unicode"

	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/types"
)

/* THIS SECTION ONLY REQUIRED IF GLOBALS ARE ADDRESSABLE USING OFFSETS RATTHER THAN PSEUDO-POINTERS
// TODO review adding this method of addressing global values as an option

// everything we need to remember about globals
type GlobalInfo struct {
	Typ             types.Type
	TypDesc         string
	Package, Member string
	Location, Size  uint
}

var GlobalsMap = make(map[*ssa.Global]GlobalInfo)
var GoGlobalsSize uint = 0

// Get the location and size of all of the globals
func scanGlobals() {
	var address uint = 0
	for pName, pack := range rootProgram.PackagesByPath {
		for mName, member := range pack.Members {
			switch member.(type) {
			case *ssa.Global:
				var g GlobalInfo
				glob := member.(*ssa.Global)
				g.Package = pName
				g.Member = mName
				g.Typ = glob.Type().(*types.Pointer).Elem() // all globals are pointers
				g.Size = uint(types.DefaultSizeof(g.Typ))
				// do not error on g.Size == 0 as it is in some library code, for example:
				// "var zero [0]byte" in /src/pkg/io/pipe.go
				g.TypDesc = g.Typ.String()
				// make sure the next address is on a correct byte boundary for the type
				boundary := uint(types.DefaultAlignof(g.Typ))
				for address%boundary != 0 {
					address++ // increase the address until we are on the correct byte boundary
				}
				g.Location = address
				address += g.Size
				GlobalsMap[glob] = g
			default:
				// do nothing if it is not a global declaration
			}
		}
	}
	GoGlobalsSize = address
}

 END ADDRESSABLE GLOBALS SECTION */

// Emit the Global declarations, run inside the Go class declaration output.
func emitGlobals() {
	allPack := rootProgram.AllPackages()
	for pkgIdx := range allPack {
		pkg := allPack[pkgIdx]
		for mName, mem := range pkg.Members {
			if mem.Token() == token.VAR {
				glob := mem.(*ssa.Global)
				pName := glob.Pkg.Object.Name()
				//println("DEBUG processing global:", pName, mName)
				posStr := CodePosition(glob.Pos())
				MakePosHash(glob.Pos()) // mark that we are dealing with this global
				if IsValidInPogo(
					glob.Type().(*types.Pointer).Elem(), // globals are always pointers to a global
					"Global:"+pName+"."+mName+":"+posStr) {
					if !hadErrors { // no point emitting code if we have already encounderd an error
						isPublic := unicode.IsUpper(rune(mName[0])) // Object value sometimes not available
						l := TargetLang
						_, _, isOverloaded := LanguageList[l].PackageOverloaded(pName)
						if !isOverloaded &&
							!(mName == "init$guard" && strings.HasPrefix(glob.RelString(nil), LibRuntimePath) && isDupPkg(pName)) {
							fmt.Fprintln(&LanguageList[l].buffer, LanguageList[l].Global(pName, mName, *glob, posStr, isPublic))
						}
					}
				}
			}
		}
	}
}
