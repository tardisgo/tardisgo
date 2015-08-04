package pogo

import (
	"sort"

	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/types"
)

type PackageSorter []*ssa.Package

func (a PackageSorter) Len() int           { return len(a) }
func (a PackageSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a PackageSorter) Less(i, j int) bool { return a[i].String() < a[j].String() }

func MemberNamesSorted(pkg *ssa.Package) []string {
	allMem := []string{}
	for mName := range pkg.Members {
		allMem = append(allMem, mName)
	}
	sort.Strings(allMem)
	return allMem
}

type fnMapSorter []*ssa.Function

func (a fnMapSorter) Len() int           { return len(a) }
func (a fnMapSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a fnMapSorter) Less(i, j int) bool { return a[i].String() < a[j].String() }

func (comp *Compilation) fnMapSorted() []*ssa.Function {
	var fms = fnMapSorter([]*ssa.Function{})
	for f := range comp.fnMap {
		fms = append(fms, f)
	}
	sort.Sort(fms)
	return []*ssa.Function(fms)
}

type TypeSorter []types.Type

func (a TypeSorter) Len() int           { return len(a) }
func (a TypeSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a TypeSorter) Less(i, j int) bool { return a[i].String() < a[j].String() }
