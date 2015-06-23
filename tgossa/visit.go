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

func vprintln(args ...interface{}) {
	//fmt.Println(args...)
}

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
	//fmt.Printf("DEBUG VisitedFunctions.seen %v\n", visit.seen)
	return visit.seen, visit.usesGR
}

type visitor struct {
	prog   *ssa.Program
	packs  []*ssa.Package // new
	seen   map[*ssa.Function]bool
	usesGR map[*ssa.Function]bool // new
}

func (visit *visitor) program(isOvl isOverloaded) {
	//fmt.vprintln("DEBUG base packages:", visit.packs)
	for p := range visit.packs {
		if visit.packs[p] != nil {
			if visit.packs[p].Members != nil {
				for _, mem := range visit.packs[p].Members { //was pkg.Members {
					if fn, ok := mem.(*ssa.Function); ok {
						//fmt.vprintln("DEBUG base function:", fn.String())
						visit.function(fn, isOvl)
						if visit.usesGR[fn] {
							visit.refsUseGR(fn.Referrers(), make(map[*ssa.Function]bool))
						}
					}
				}
			}
		}
	}
	for _, T := range visit.prog.RuntimeTypes() {
		mset := visit.prog.MethodSets.MethodSet(T)
		for i, n := 0, mset.Len(); i < n; i++ {
			mf := visit.prog.Method(mset.At(i))
			visit.function(mf, isOvl)
			// ??? conservatively mark every method as requiring goroutines, in order to simplify method calls?
			// visit.usesGR[mf] = true
			// TODO use Oracle techniques to discover which of these methods could actually be called
			if visit.usesGR[mf] {
				visit.refsUseGR(mf.Referrers(), make(map[*ssa.Function]bool))
			}
		}
	}
}

func (visit *visitor) refsUseGR(refs *[]ssa.Instruction, refed map[*ssa.Function]bool) {
	if refs != nil {
		for r := range *refs {
			fn := (*refs)[r].Parent()
			if !refed[fn] {
				visit.usesGR[fn] = true
				refed[fn] = true
				visit.refsUseGR(fn.Referrers(), refed)
			}
		}
	}
}

func (visit *visitor) function(fn *ssa.Function, isOvl isOverloaded) {
	if !visit.seen[fn] { // been, exists := visit.seen[fn]; !been || !exists {
		vprintln("DEBUG 1st visit to: ", fn.String())
		visit.seen[fn] = true
		visit.usesGR[fn] = false
		if isOvl(fn) {
			vprintln("DEBUG overloaded: ", fn.String())
			return
		}
		if len(fn.Blocks) == 0 { // exclude functions that reference C/assembler code
			// NOTE: not marked as seen, because we don't want to include in output
			// if used, the symbol will be included in the golibruntime replacement packages
			// TODO review
			vprintln("DEBUG no code for: ", fn.String())
			return // external functions cannot use goroutines
		}
		var buf [10]*ssa.Value // avoid alloc in common case
		for _, b := range fn.Blocks {
			for _, instr := range b.Instrs {
				for _, op := range instr.Operands(buf[:0]) {
					areRecursing := false
					afn, isFn := (*op).(*ssa.Function)
					if isFn {
						if afn == fn {
							areRecursing = true
						}
						visit.function(afn, isOvl)
						if visit.usesGR[afn] {
							vprintln("marked as using GR because referenced func uses GR")
							visit.usesGR[fn] = true
						}
						vprintln(fn.Name(), " calls ", afn.Name())
					}
					// TODO, review if this code should be included
					if !visit.usesGR[fn] {
						if _, ok := (*op).(ssa.Value); ok {
							typ := (*op).Type()
							typ = DeRefUl(typ)
							switch typ.(type) {
							// TODO use oracle techniques to determine which interfaces or functions may require GR
							case *types.Chan, *types.Interface:
								visit.usesGR[fn] = true // may be too conservative
								vprintln("marked as using GR because uses Chan/Interface")
							case *types.Signature:
								if !areRecursing {
									if !isFn {
										visit.usesGR[fn] = true
										vprintln("marked as using GR because uses Signature")
									}
								}
							}
						}
					}
				}
				if _, ok := instr.(*ssa.Call); ok {
					switch instr.(*ssa.Call).Call.Value.(type) {
					case *ssa.Builtin:
						//NoOp
					default:
						cc := instr.(*ssa.Call).Common()
						if cc != nil {
							afn := cc.StaticCallee()
							if afn != nil {
								visit.function(afn, isOvl)
								if visit.usesGR[afn] {
									visit.usesGR[fn] = true
									vprintln("marked as using GR because call target uses GR")
								}
								vprintln(fn.Name(), " calls ", afn.Name())
							}
						}
					}
				}
				if !visit.usesGR[fn] {
					switch instr.(type) {
					case *ssa.Go, *ssa.MakeChan, *ssa.Defer, *ssa.Panic,
						*ssa.Send, *ssa.Select:
						vprintln("usesGR because uses Go...", fn.Name())
						visit.usesGR[fn] = true
					case *ssa.UnOp:
						if instr.(*ssa.UnOp).Op.String() == "<-" {
							vprintln("usesGR because uses <-", fn.Name())
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
