// Copyright 2014 Elliott Stoneham and The TARDIS Go Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package pogo

import (
	"fmt"
	"go/token"
	"sort"
	"strconv"

	"golang.org/x/tools/go/exact"
	"golang.org/x/tools/go/ssa"
)

// emit the constant declarations
func (comp *Compilation) emitNamedConstants() {
	allPack := comp.rootProgram.AllPackages()
	sort.Sort(PackageSorter(allPack))
	for pkgIdx := range allPack {
		pkg := allPack[pkgIdx]
		allMem := MemberNamesSorted(pkg)
		for _, mName := range allMem {
			mem := pkg.Members[mName]
			if mem.Token() == token.CONST {
				lit := mem.(*ssa.NamedConst).Value
				posStr := comp.CodePosition(lit.Pos())
				pName := mem.(*ssa.NamedConst).Object().Pkg().Path() // was .Name()
				switch lit.Value.Kind() {                            // non language specific validation
				case exact.Bool, exact.String, exact.Float, exact.Int, exact.Complex: //OK
					isPublic := mem.Object().Exported()
					if isPublic { // constants will be inserted inline, these declarations of public constants are for exteral use in target language
						l := comp.TargetLang
						fmt.Fprintln(&LanguageList[l].buffer, LanguageList[l].NamedConst(pName, mName, *lit, posStr))
					}
				default:
					comp.LogError(posStr, "pogo", fmt.Errorf("%s.%s : emitConstants() internal error, unrecognised constant type: %v",
						pName, mName, lit.Value.Kind()))
				}
			}
		}
	}
}

// FloatVal is a utility function returns a string constant value from an exact.Value.
func (comp *Compilation) FloatVal(eVal exact.Value, bits int, posStr string) string {
	fVal, isExact := exact.Float64Val(eVal)
	if !isExact {
		comp.LogWarning(posStr, "inexact", fmt.Errorf("constant value %g cannot be accurately represented in float64", fVal))
	}
	ret := strconv.FormatFloat(fVal, byte('g'), -1, bits)
	if fVal < 0.0 {
		return fmt.Sprintf("(%s)", ret)
	}
	return ret
}

// IntVal is a utility function returns an int64 constant value from an exact.Value, split into high and low int32.
func (comp *Compilation) IntVal(eVal exact.Value, posStr string) (high, low int32) {
	iVal, isExact := exact.Int64Val(eVal)
	if !isExact {
		comp.LogWarning(posStr, "inexact", fmt.Errorf("constant value %d cannot be accurately represented in int64", iVal))
	}
	return int32(iVal >> 32), int32(iVal & 0xFFFFFFFF)
}
