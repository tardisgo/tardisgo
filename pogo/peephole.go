// Copyright 2014 Elliott Stoneham and The TARDIS Go Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package pogo

import (
	"fmt"
	"go/token"

	"golang.org/x/tools/go/ssa"
)

// peephole optimizes and emits short sequences of instructions that do not contain control flow
func peephole(instrs []ssa.Instruction) {

	for i := 0; i < len(instrs); i++ {
		if len(instrs[i:]) >= 2 {
			for j := len(instrs); j > (i + 1); j-- {
				opt, reg := peepholeFindOpt(instrs[i:j])
				if opt != "" {
					//fmt.Println("DEBUG PEEPHOLE", opt, reg)
					fmt.Fprintln(&LanguageList[TargetLang].buffer,
						LanguageList[TargetLang].PeepholeOpt(opt,
							reg, instrs[i:j], "[ PEEPHOLE ]"))
					i = j - 1
					goto instrsEmitted
				}
			}
		}
		emitInstruction(instrs[i], instrs[i].Operands(make([]*ssa.Value, 0)))
	instrsEmitted:
	}
}

// TODO WIP...
func peepholeFindOpt(instrs []ssa.Instruction) (optName, regName string) {
	if len(instrs) < 2 {
		return // fail
	}
	switch instrs[0].(type) {
	case *ssa.UnOp:
		if instrs[0].(*ssa.UnOp).Op == token.MUL &&
			len(*instrs[0].(*ssa.UnOp).Referrers()) == 1 {
			switch instrs[len(instrs)-1].(type) {
			case *ssa.Index, *ssa.Field:
				// candidate to remove load_object
				if len(instrs) == 2 {
					// we are at the first two in the load_object(UnOp*)+Index/Field sequence
					if instrs[0].(*ssa.UnOp).Name() == indexOrFieldXName(instrs[1]) &&
						indexOrFieldRefCount(instrs[1]) > 0 {
						optName = "loadObject"
						regName = RegisterName(instrs[1].(ssa.Value))
						return // success
					}
					return // fail
				}
				// we are in some sequence of Index/Field ops, one after another
				// so first check that the earlier parts of the sequene are OK
				on, rn := peepholeFindOpt(instrs[0 : len(instrs)-1])
				if on == "loadObject" {
					if rn == "_"+indexOrFieldXName(instrs[len(instrs)-1]) && // end one links to one before
						indexOrFieldRefCount(instrs[len(instrs)-2]) == 1 && // one before only used by this one
						indexOrFieldRefCount(instrs[len(instrs)-1]) > 0 { // end result is used
						optName = on
						regName = RegisterName(instrs[len(instrs)-1].(ssa.Value))
						return // success
					}
				}
			}
		}
	case *ssa.Phi:
		for _, instr := range instrs {
			phi, ok := instr.(*ssa.Phi)
			if !ok {
				return // fail
			}
			if len(*phi.Referrers()) == 0 {
				return // fail
			}
		}
		optName = "phiList"
		//regName is unused
		return //success

	}
	return // fail
}

func indexOrFieldXName(i ssa.Instruction) string {
	switch i.(type) {
	case *ssa.Index:
		return i.(*ssa.Index).X.Name()
	case *ssa.Field:
		return i.(*ssa.Field).X.Name()
	default:
		return ""
	}
}

func indexOrFieldRefCount(i ssa.Instruction) int {
	switch i.(type) {
	case *ssa.Index:
		return len(*i.(*ssa.Index).Referrers())
	case *ssa.Field:
		return len(*i.(*ssa.Field).Referrers())
	default:
		return 0
	}
}
