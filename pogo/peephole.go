// Copyright 2014 Elliott Stoneham and The TARDIS Go Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package pogo

import (
	"fmt"
	"go/token"
	"strings"

	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/types"
)

func (comp *Compilation) InlineMap(key string) (val string, ok bool) {
	val, ok = comp.inlineMap[key]
	count, seen := comp.keysSeen[key]
	if seen {
		comp.keysSeen[key] = count + 1
	} else {
		comp.keysSeen[key] = 1
	}
	return
}

func (comp *Compilation) newInlineMap() {
	comp.inlineMap = make(map[string]string)
	comp.keysSeen = make(map[string]int)
}

// peephole optimizes and emits short sequences of instructions that do not contain control flow
func (comp *Compilation) peephole(instrs []ssa.Instruction) {
	comp.newInlineMap()
	for i := 0; i < len(instrs); i++ {
		//var v ssa.Value
		var isV, inline bool
		if len(instrs[i:]) >= 1 {
			for j := len(instrs); j > (i /* +1 */); j-- {
				opt, reg := comp.peepholeFindOpt(instrs[i:j])
				if opt != "" {
					//fmt.Println("DEBUG PEEPHOLE", opt, reg)
					fmt.Fprintln(&LanguageList[comp.TargetLang].buffer,
						LanguageList[comp.TargetLang].PeepholeOpt(opt,
							reg, instrs[i:j], "[ PEEPHOLE ]"))
					i = j - 1
					goto instrsEmitted
				}
			}
		}
		inline = false
		_, isV = instrs[i].(ssa.Value)
		if isV {
			if LanguageList[comp.TargetLang].CanInline(instrs[i]) {
				inline = true
			}
		}
		if inline {
			postWrite := ""
			preBuffLen := LanguageList[comp.TargetLang].buffer.Len()
			comp.emitInstruction(instrs[i], instrs[i].Operands(make([]*ssa.Value, 0)))
			raw := strings.TrimSpace(string(LanguageList[comp.TargetLang].buffer.Bytes()[preBuffLen:]))
			for _, ignorePrefix := range LanguageList[comp.TargetLang].IgnorePrefixes {
				if strings.HasPrefix(raw, ignorePrefix) {
					sph := strings.SplitAfterN(raw, LanguageList[comp.TargetLang].StatementTerminator, 2)
					if len(sph) != 2 {
						panic("code to ignore not as expected: " + raw)
					}
					postWrite = sph[0]
					raw = strings.TrimSpace(sph[1])
					break
				}
			}
			bits := strings.SplitAfter(raw, LanguageList[comp.TargetLang].LineCommentMark) //comment marker must not be in strings
			code := strings.TrimSuffix(
				strings.TrimSpace(strings.TrimSuffix(bits[0],
					LanguageList[comp.TargetLang].LineCommentMark)), LanguageList[comp.TargetLang].StatementTerminator) // usually a semi-colon
			parts := strings.SplitAfterN(code, "=", 2)
			if len(parts) != 2 {
				panic("no = after register name in: " + code)
			}
			parts[0] = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(parts[0]), "="))
			//println("DEBUG inlineMap[" + parts[0] + "]=" + parts[1])
			found := 0
			for _, v := range comp.inlineMap {
				if v == parts[1] {
					found++
				}
			}
			comp.inlineMap[parts[0]] = parts[1]
			LanguageList[comp.TargetLang].buffer.Truncate(preBuffLen)
			fmt.Fprintln(&LanguageList[comp.TargetLang].buffer, "//[ PEEPHOLE INLINE "+parts[0]+" ] "+instrs[i].String())
			if postWrite != "" {
				fmt.Fprintln(&LanguageList[comp.TargetLang].buffer, postWrite)
			}
			if found > 0 {
				fmt.Fprintf(&LanguageList[comp.TargetLang].buffer,
					"// DEBUG %d duplicate(s) found for %s\n", found, parts[1]) // this optimisation TODO
			}
		} else {
			comp.emitInstruction(instrs[i], instrs[i].Operands(make([]*ssa.Value, 0)))
		}
	instrsEmitted:
	}
	//println("DEBUG new inlineMap")
	comp.newInlineMap() // needed here too to stop these temp values bleeding to elsewhere
}

func (comp *Compilation) peepholeFindOpt(instrs []ssa.Instruction) (optName, regName string) {
	switch instrs[0].(type) {
	case *ssa.IndexAddr, *ssa.FieldAddr:
		ptrChainSize := 1
		if len(instrs) < 2 {
			return // fail
		}
		//fmt.Println("DEBUG looking for ptrChain num refs=", len(*(instrs[0].(ssa.Value).Referrers())))
		if len(*(instrs[0].(ssa.Value).Referrers())) == 0 || !addrInstrUsesPointer(instrs[0]) {
			goto nextOpts
		}
		//fmt.Println("DEBUG instr 0: ", instrs[0].String())
		for ; ptrChainSize < len(instrs); ptrChainSize++ {
			//fmt.Println("DEBUG instr ", ptrChainSize, instrs[ptrChainSize].String())
			switch instrs[ptrChainSize].(type) {
			case *ssa.IndexAddr, *ssa.FieldAddr:
				if !addrInstrUsesPointer(instrs[ptrChainSize]) {
					goto nextOpts
				}
				/*
					fmt.Println("DEBUG i, refs,  prev, this, instr(prev), instr(this)=",
						ptrChainSize,
						len(*instrs[ptrChainSize].(ssa.Value).Referrers()),
						RegisterName(instrs[ptrChainSize-1].(ssa.Value)),
						"_"+(*instrs[ptrChainSize].Operands(nil)[0]).Name(),
						RegisterName(instrs[ptrChainSize-1].(ssa.Value))+"="+instrs[ptrChainSize-1].String(),
						RegisterName(instrs[ptrChainSize].(ssa.Value))+"="+instrs[ptrChainSize].String())
				*/
				if len(*instrs[ptrChainSize-1].(ssa.Value).Referrers()) != 1 ||
					"_"+(*instrs[ptrChainSize].Operands(nil)[0]).Name() != comp.RegisterName(instrs[ptrChainSize-1].(ssa.Value)) {
					goto nextOpts
				}
			default:
				goto nextOpts
			}
		}
		if ptrChainSize > 1 {
			/*
				fmt.Println("DEBUG pointer chain found")
				for i := 0; i < ptrChainSize; i++ {
					fmt.Println("DEBUG pointer chain ", i, len(*instrs[i].(ssa.Value).Referrers()),
						RegisterName(instrs[i].(ssa.Value))+"="+instrs[i].String())
				}
			*/
			return "pointerChain", comp.RegisterName(instrs[ptrChainSize-1].(ssa.Value))
		}
	nextOpts:
		return // fail

	case *ssa.UnOp:
		if len(instrs) < 2 {
			return // fail
		}
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
						regName = comp.RegisterName(instrs[1].(ssa.Value))
						return // success
					}
					return // fail
				}
			}
		}

	case *ssa.Phi:
		if len(instrs) == 0 {
			return // fail
		}
		for _, instr := range instrs {
			_ /*phi*/, ok := instr.(*ssa.Phi)
			if !ok {
				return // fail
			}
			//if len(*phi.Referrers()) == 0 {
			//	return // fail
			//}
		}
		optName = "phiList"
		//regName is unused
		return //success

	}
	return // fail
}

func addrInstrUsesPointer(i ssa.Instruction) bool {
	switch i.(type) {
	case *ssa.IndexAddr:
		_, ok := i.(*ssa.IndexAddr).X.Type().Underlying().(*types.Pointer)
		return ok
	case *ssa.FieldAddr:
		_, ok := i.(*ssa.FieldAddr).X.Type().Underlying().(*types.Pointer)
		return ok
	default:
		return false
	}
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
