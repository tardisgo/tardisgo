// Copyright 2014 Elliott Stoneham and The TARDIS Go Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package haxe

import (
	"fmt"
	"go/token"
	"sort"
	"strings"

	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/types"

	"github.com/tardisgo/tardisgo/pogo"
)

type phiEntry struct{ reg, val string }

// PeepholeOpt implements the optimisations spotted by pogo.peephole
func (l langType) PeepholeOpt(opt, register string, code []ssa.Instruction, errorInfo string) string {
	ret := ""
	switch opt {
	case "pointerChain":
		for _, cod := range code {
			switch cod.(type) {
			case *ssa.IndexAddr:
				ret += fmt.Sprintf("// %s=%s\n", cod.(*ssa.IndexAddr).Name(), cod.String())
			case *ssa.FieldAddr:
				ret += fmt.Sprintf("// %s=%s\n", cod.(*ssa.FieldAddr).Name(), cod.String())
			}
		}
		ret += register + "="
		if pogo.DebugFlag {
			ret += "Pointer.check("
		}
		switch code[0].(type) {
		case *ssa.IndexAddr:
			ret += l.IndirectValue(code[0].(*ssa.IndexAddr).X, errorInfo)
		case *ssa.FieldAddr:
			ret += l.IndirectValue(code[0].(*ssa.FieldAddr).X, errorInfo)
		default:
			panic(fmt.Errorf("unexpected type %T", code[0]))
		}
		if pogo.DebugFlag {
			ret += ")"
		}
		ret += ".addr("
		for c, cod := range code {
			if c > 0 {
				ret += "+"
			}
			switch cod.(type) {
			case *ssa.IndexAddr:
				idxString := wrapForce_toUInt(l.IndirectValue(cod.(*ssa.IndexAddr).Index, errorInfo),
					cod.(*ssa.IndexAddr).Index.(ssa.Value).Type().Underlying().(*types.Basic).Kind())
				ele := cod.(*ssa.IndexAddr).X.Type().Underlying().(*types.Pointer).Elem().Underlying().(*types.Array).Elem().Underlying()
				ret += "(" + idxString + arrayOffsetCalc(ele) + ")"
			case *ssa.FieldAddr:
				off := fieldOffset(cod.(*ssa.FieldAddr).X.Type().Underlying().(*types.Pointer).Elem().Underlying().(*types.Struct), cod.(*ssa.FieldAddr).Field)
				ret += fmt.Sprintf(`%d`, off)
			}
		}
		ret += "); // PEEPHOLE OPTIMIZATION pointerChain\n"

	case "loadObject":
		idx := ""
		ret += fmt.Sprintf("// %s=%s\n", code[0].(*ssa.UnOp).Name(), code[0].String())
		for _, cod := range code[1:] {
			switch cod.(type) {
			case *ssa.Index:
				ret += fmt.Sprintf("// %s=%s\n", cod.(*ssa.Index).Name(), cod.String())
			case *ssa.Field:
				ret += fmt.Sprintf("// %s=%s\n", cod.(*ssa.Field).Name(), cod.String())
			}
		}
		ptrString := l.IndirectValue(code[0].(*ssa.UnOp).X, errorInfo)
		ret += fmt.Sprintf("%s=%s", register, ptrString)
		for _, cod := range code[1:] {
			switch cod.(type) {
			case *ssa.Index:
				idx = wrapForce_toUInt(l.IndirectValue(cod.(*ssa.Index).Index, errorInfo),
					cod.(*ssa.Index).Index.Type().Underlying().(*types.Basic).Kind())
				//if idx != "0" {
				//	ret += fmt.Sprintf(".addr(%s%s)",
				//		idx,
				//		arrayOffsetCalc(cod.(*ssa.Index).Type().Underlying()))
				//}
			case *ssa.Field:
				fo := fieldOffset(cod.(*ssa.Field).X.Type().Underlying().(*types.Struct), cod.(*ssa.Field).Field)
				idx = fmt.Sprintf("%d", fo)
				//if idx != "0" {
				//	ret += fmt.Sprintf(".fieldAddr(%d)", fo)
				//}
			}
		}
		suffix := ""
		switch code[len(code)-1].(type) {
		case *ssa.Index:
			suffix = loadStoreSuffix(code[len(code)-1].(*ssa.Index).Type().Underlying(), true)
			//ret += fmt.Sprintf(".load%s); // PEEPHOLE OPTIMIZATION loadObject (Index)\n",
			//	loadStoreSuffix(code[len(code)-1].(*ssa.Index).Type().Underlying(), false))
		case *ssa.Field:
			suffix = loadStoreSuffix(code[len(code)-1].(*ssa.Field).Type().Underlying(), true)
			//ret += fmt.Sprintf(".load%s); // PEEPHOLE OPTIMIZATION loadObject (Field)\n",
			//	loadStoreSuffix(code[len(code)-1].(*ssa.Field).Type().Underlying(), false))
		}
		ret += fmt.Sprintf(".obj.get%s%s+%s.off); // PEEPHOLE OPTIMIZATION loadObject\n",
			suffix, idx, ptrString)

	case "phiList":
		ret += "// PEEPHOLE OPTIMIZATION phiList\n"
		//ret += l.PhiCode(true, 0, code, errorInfo)
	}
	return ret
}

func (l langType) PhiCode(allTargets bool, targetPhi int, code []ssa.Instruction, errorInfo string) string {
	ret := ""
	opts := make(map[int][]phiEntry)
	for _, cod := range code {
		_, isPhi := cod.(*ssa.Phi)
		if isPhi {
			operands := cod.(*ssa.Phi).Operands([]*ssa.Value{})
			phiEntries := make([]int, len(operands))
			valEntries := make([]string, len(operands))
			thisReg := cod.(*ssa.Phi).Name()
			ret += "// " + thisReg + "=" + cod.String() + "\n"
			for o := range operands {
				phiEntries[o] = cod.(*ssa.Phi).Block().Preds[o].Index
				if _, ok := opts[phiEntries[o]]; !ok {
					opts[phiEntries[o]] = make([]phiEntry, 0)
				}
				valEntries[o] = l.IndirectValue(*operands[o], errorInfo)
				if fnCanOptMap[thisReg] || len(*(cod.(*ssa.Phi).Referrers())) == 0 {
					thisReg = ""
				}
				opts[phiEntries[o]] = append(opts[phiEntries[o]], phiEntry{thisReg, valEntries[o]})
			}
		}
	}
	if allTargets {
		ret += "switch(_Phi) { \n"
	}
	idxs := make([]int, 0, len(opts))
	for phi := range opts {
		idxs = append(idxs, phi)
	}
	sort.Ints(idxs)
	for _, phi := range idxs {
		if allTargets || phi == targetPhi {
			opt := opts[phi]
			if allTargets {
				ret += fmt.Sprintf("\tcase %d:\n", phi)
			}
			crossover := false
			for x1, ent1 := range opt {
				for x2, ent2 := range opt {
					if x1 != x2 {
						if "_"+ent1.reg == ent2.val {
							crossover = true
							goto foundCrossover
						}
					}
				}
			}
		foundCrossover:
			if crossover {
				for _, ent := range opt {
					if ent.reg != "" {
						ret += fmt.Sprintf("\t\tvar tmp_%s=%s;\n", ent.reg, ent.val) // need temp vars for a,b = b,a
					}
				}
			}
			for _, ent := range opt {
				if ent.reg != "" {
					rn := "_" + ent.reg
					if useRegisterArray {
						rn = rn[:2] + "[" + rn[2:] + "]"
					}
					if crossover {
						ret += fmt.Sprintf("\t\t%s=tmp_%s;\n", rn, ent.reg)
					} else {
						if rn != ent.val {
							ret += fmt.Sprintf("\t\t%s=%s;\n", rn, ent.val)
						}
					}
				}
			}
		}
	}
	if allTargets {
		ret += "}\n"
	}
	return ret
}

func (l langType) CanInline(vi interface{}) bool {
	//if pogo.DebugFlag {
	//   return false
	//}
	val, isVal := vi.(ssa.Value)
	if !isVal {
		return false
	}
	switch l.LangType(val.Type(), false, "CanInline()") {
	case "Dynamic": // so a uintptr
		return false // this can yeild un-expected results & mess up the type checking
	}
	var refs *[]ssa.Instruction
	var thisBlock *ssa.BasicBlock
	switch vi.(type) {
	default:
		return false
	case *ssa.Convert:
		/* this slows things down for cpp and does not speed things up for js
		if l.LangType(val.Type(), false, "CanInline()") !=
			l.LangType(vi.(*ssa.Convert).X.Type(), false, "CanInline()") {
			return false // can't have different Haxe types
		}
		*/
		refs = vi.(*ssa.Convert).Referrers()
		thisBlock = vi.(*ssa.Convert).Block()
	case *ssa.BinOp:
		refs = vi.(*ssa.BinOp).Referrers()
		thisBlock = vi.(*ssa.BinOp).Block()
	case *ssa.UnOp:
		switch vi.(*ssa.UnOp).Op {
		case token.ARROW, token.MUL: // NOTE token.MUL does not work because of a[4],a[5]=a[5],a[4] crossover
			return false
		}
		refs = vi.(*ssa.UnOp).Referrers()
		thisBlock = vi.(*ssa.UnOp).Block()
		/*
			case *ssa.IndexAddr: // NOTE optimising this instruction means it's pointer leaks, but it does give a speed-up
				_, isSlice := vi.(*ssa.IndexAddr).X.Type().Underlying().(*types.Slice)
				if !isSlice {
					return false // only slices handled in the general in-line code, rather than pointerChain above
				}
				refs = vi.(*ssa.IndexAddr).Referrers()
				thisBlock = vi.(*ssa.IndexAddr).Block()
		*/
	}

	if thisBlock == nil {
		return false
	}
	//if len(thisBlock.Instrs) >= pogo.LanguageList[langIdx].InstructionLimit {
	//	return false
	//}
	if len(*refs) != 1 {
		return false
	}
	if (*refs)[0].Block() != thisBlock {
		return false // consumer is not in the same block
	}
	if blockContainsBreaks((*refs)[0].Block(), vi.(ssa.Instruction), (*refs)[0]) {
		return false
	}
	/*
		ia, is := vi.(*ssa.IndexAddr)
		if is {
			println("DEBUG CanInline found candidate IndexAddr:", ia.String())
		}
	*/
	return true
}

func (l langType) inlineRegisterName(vi interface{}) string {
	vp, okPtr := vi.(*ssa.Value)
	if !okPtr {
		v, ok := vi.(ssa.Value)
		if !ok {
			panic(fmt.Sprintf("inlineRegisterName not a pointer to a value, or a value; it is a %T", vi))
		}
		vp = &v
	}
	nm := strings.TrimSpace(pogo.RegisterName(*vp))
	if l.CanInline(vi) {
		code, found := pogo.InlineMap(nm)
		if !found {
			//for k, v := range pogo.InlineMap {
			//	println("DEBUG dump pogo.InlineMap[", k, "] is ", v)
			//}
			//pogo.LogError(vi.(ssa.Instruction).Parent().String(), "haxe", errors.New("internal error - cannot find "+nm+" in pogo.InlineMap"))
			return nm
		}
		return code
	}
	return nm
}

func blockContainsBreaks(b *ssa.BasicBlock, producer, consumer ssa.Instruction) bool {
	hadProducer := false
	for ii, i := range b.Instrs {
		_ = ii
		if hadProducer {
			switch i.(type) {
			case *ssa.Call, *ssa.Select, *ssa.Send, *ssa.Defer, *ssa.RunDefers, *ssa.Return:
				return true
			case *ssa.UnOp:
				if i.(*ssa.UnOp).Op == token.ARROW {
					return true
				}
			}
			if i == consumer {
				//println("DEBUG consumer is ", ii)
				return false // if there is no break before the var is consumed, then no problem
			}
		} else {
			if i == producer {
				//println("DEBUG producer is ", ii)
				hadProducer = true
			}
		}
	}
	return false
}
