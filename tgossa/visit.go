// modified version of the file: golang.org/x/tools/go/ssa/ssautil/visit.go

// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package tgossa provides pure ssa optimization functions in TARDISgo.
package tgossa // was ssautil

import (
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/types"
)

type isOverloaded func(*ssa.Function) bool

// TARDISGO VERSION MODIFIED FROM
// This file defines utilities for visiting the SSA representation of
// a Program.
//
// TODO(adonovan): test coverage.

// VisitedFunctions finds only those functions visited from the main package, using the logic of:
// AllFunctions finds and returns the set of functions potentially
// needed by program prog, as determined by a simple linker-style
// reachability algorithm starting from the members and method-sets of
// each package.  The result may include anonymous functions and
// synthetic wrappers.
//
// Precondition: all packages are built.
//
func VisitedFunctions(prog *ssa.Program, packs []*ssa.Package, isOvl isOverloaded) (seen, usesGR map[*ssa.Function]bool) {
	visit := visitor{
		prog:   prog,
		packs:  packs, // new
		seen:   make(map[*ssa.Function]bool),
		usesGR: make(map[*ssa.Function]bool),
	}
	visit.program(isOvl)
	//fmt.Printf("DEBUG VisitedFunctions.usesGR %v\n", visit.usesGR)
	return visit.seen, visit.usesGR
}

type visitor struct {
	prog   *ssa.Program
	packs  []*ssa.Package // new
	seen   map[*ssa.Function]bool
	usesGR map[*ssa.Function]bool // new
}

func (visit *visitor) program(isOvl isOverloaded) {
	for p := range visit.packs {
		if visit.packs[p] != nil {
			if visit.packs[p].Members != nil {
				for _, mem := range visit.packs[p].Members { //was pkg.Members {
					if fn, ok := mem.(*ssa.Function); ok {
						visit.function(fn, isOvl)
					}
				}
			}
		}
	}
	for _, T := range visit.prog.TypesWithMethodSets() {
		mset := visit.prog.MethodSets.MethodSet(T)
		for i, n := 0, mset.Len(); i < n; i++ {
			mf := visit.prog.Method(mset.At(i))
			visit.function(mf, isOvl)
			// ??? conservatively mark every method as requiring goroutines, in order to simplify method calls?
			// visit.usesGR[mf] = true
			// TODO use Oracle techniques to discover which of these methods could actually be called
		}
	}
}

func (visit *visitor) function(fn *ssa.Function, isOvl isOverloaded) {
	if !visit.seen[fn] {
		if isOvl(fn) {
			return
		}
		if len(fn.Blocks) == 0 { // exclude functions that reference C/assembler code
			// NOTE: not marked as seen, because we don't want to include in output
			// if used, the symbol will be included in the golibruntime replacement packages
			// TODO review
			visit.usesGR[fn] = false // external functions cannot use goroutines
			return
		}
		visit.seen[fn] = true
		visit.usesGR[fn] = false
		var buf [10]*ssa.Value // avoid alloc in common case
		for _, b := range fn.Blocks {
			for _, instr := range b.Instrs {
				for _, op := range instr.Operands(buf[:0]) {
					if afn, ok := (*op).(*ssa.Function); ok {
						visit.function(afn, isOvl)
						if visit.usesGR[afn] {
							visit.usesGR[fn] = true
						}
						//println(fn.Name(), " calls ", afn.Name())
					}
					/* TODO, review if this code should be included
					if !visit.usesGR[fn] {
						if _, ok := (*op).(ssa.Value); ok {
							typ := (*op).Type()
							typ = DeRefUl(typ)
							switch typ.(type) {
							case *types.Chan , *types.Interface, *types.Signature :
								// TODO use oracle techniques to determine which interfaces or functions may require GR
								visit.usesGR[fn] = true // may be too conservative
							}
						}
					}
					*/
				}
				if !visit.usesGR[fn] {
					switch instr.(type) {
					case *ssa.Go, *ssa.MakeChan, *ssa.Defer, *ssa.Panic,
						*ssa.Send, *ssa.Select:
						//fmt.Println("usesGR", fn.Name())
						visit.usesGR[fn] = true
					case *ssa.UnOp:
						if instr.(*ssa.UnOp).Op.String() == "<-" {
							visit.usesGR[fn] = true
						}
					}
				}
			}
		}
	}
}

// DeRefUl dereferencs a type and gets to it's underlying type
func DeRefUl(T types.Type) types.Type {
deRef:
	switch T.(type) {
	case *types.Pointer:
		T = T.(*types.Pointer).Elem().Underlying()
		goto deRef
	case *types.Named:
		T = T.(*types.Named).Underlying()
		goto deRef
	default:
		return T
	}
}
