// modified version of the file: code.google.com/p/go.tools/go/ssa/ssautil/visit.go

// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package ssaopt provides pure ssa optimization functions.
package ssaopt // was ssautil

import (
	"code.google.com/p/go.tools/go/ssa"
	"code.google.com/p/go.tools/go/types"
)

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
func VisitedFunctions(prog *ssa.Program, packs []*ssa.Package /*new*/) (seen, usesGR map[*ssa.Function]bool) {
	visit := visitor{
		prog:   prog,
		packs:  packs, // new
		seen:   make(map[*ssa.Function]bool),
		usesGR: make(map[*ssa.Function]bool),
	}
	visit.program()
	return visit.seen, visit.usesGR
}

type visitor struct {
	prog   *ssa.Program
	packs  []*ssa.Package // new
	seen   map[*ssa.Function]bool
	usesGR map[*ssa.Function]bool // new
}

func (visit *visitor) program() {
	//for _, pkg := range visit.prog.AllPackages() {
	for p := range visit.packs {
		for _, mem := range visit.packs[p].Members { //was pkg.Members {
			if fn, ok := mem.(*ssa.Function); ok {
				visit.function(fn)
			}
		}
	}
	// TODO use Oracle techniques to discover which of these methods could actually be called
	for _, T := range visit.prog.TypesWithMethodSets() {
		mset := visit.prog.MethodSets.MethodSet(T)
		for i, n := 0, mset.Len(); i < n; i++ {
			visit.function(visit.prog.Method(mset.At(i)))
		}
	}
}

func (visit *visitor) function(fn *ssa.Function) {
	if !visit.seen[fn] {
		visit.seen[fn] = true
		visit.usesGR[fn] = false
		var buf [10]*ssa.Value // avoid alloc in common case
		for _, b := range fn.Blocks {
			for _, instr := range b.Instrs {
				for _, op := range instr.Operands(buf[:0]) {
					if afn, ok := (*op).(*ssa.Function); ok {
						visit.function(afn)
						if !visit.usesGR[fn] { // if the calling function is not yet marked as using a goroutine
							visit.usesGR[fn] = visit.usesGR[afn] // use the called function's value
						}
					}
					if _, ok := (*op).(ssa.Value); ok {
						typ := (*op).Type()
						if _, ok := typ.(*types.Named); ok {
							typ = typ.Underlying()
						}
						switch typ.(type) {
						case *types.Chan, *types.Interface: // TODO use oracle techniques to determine which interfaces may require GR
							visit.usesGR[fn] = true
						case *types.Pointer:
							if _, ok := typ.(*types.Pointer).Elem().Underlying().(*types.Chan); ok {
								visit.usesGR[fn] = true
							}
						}
					}
				}
				switch instr.(type) {
				case *ssa.Go, *ssa.MakeChan:
					visit.usesGR[fn] = true
				}
			}
		}
	}
}
