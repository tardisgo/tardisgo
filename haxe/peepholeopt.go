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
		basePointer := ""
		switch code[0].(type) {
		case *ssa.IndexAddr:
			basePointer = l.IndirectValue(code[0].(*ssa.IndexAddr).X, errorInfo)
		case *ssa.FieldAddr:
			basePointer = l.IndirectValue(code[0].(*ssa.FieldAddr).X, errorInfo)
		default:
			panic(fmt.Errorf("unexpected type %T", code[0]))
		}
		chainGang := ""
		for c, cod := range code {
			if c > 0 {
				chainGang += "+"
			}
			switch cod.(type) {
			case *ssa.IndexAddr:
				idxString := wrapForceToUInt(l.IndirectValue(cod.(*ssa.IndexAddr).Index, errorInfo),
					cod.(*ssa.IndexAddr).Index.(ssa.Value).Type().Underlying().(*types.Basic).Kind())
				ele := cod.(*ssa.IndexAddr).X.Type().Underlying().(*types.Pointer).Elem().Underlying().(*types.Array).Elem().Underlying()
				chainGang += "(" + idxString + arrayOffsetCalc(ele) + ")"
			case *ssa.FieldAddr:
				off := fieldOffset(cod.(*ssa.FieldAddr).X.Type().Underlying().(*types.Pointer).Elem().Underlying().(*types.Struct), cod.(*ssa.FieldAddr).Field)
				chainGang += fmt.Sprintf(`%d`, off)
			}
		}
		if l.is1usePtr(code[len(code)-1]) {
			ret += l.set1usePtr(code[len(code)-1].(ssa.Value), oneUsePtr{obj: basePointer + ".obj", off: chainGang + "+" + basePointer + ".off"})
			ret += "// virtual oneUsePtr " + register + "=" + l.hc.map1usePtr[code[len(code)-1].(ssa.Value)].obj + ":" + l.hc.map1usePtr[code[len(code)-1].(ssa.Value)].off
			ret += " PEEPHOLE OPTIMIZATION pointerChain\n"
		} else {
			ret += register + "="
			if l.PogoComp().DebugFlag {
				ret += "Pointer.check("
			}
			ret += basePointer
			if l.PogoComp().DebugFlag {
				ret += ")"
			}
			ret += ".addr(" + chainGang
			ret += "); // PEEPHOLE OPTIMIZATION pointerChain\n"
		}

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
		for _, cod := range code[1:] {
			switch cod.(type) {
			case *ssa.Index:
				idx = wrapForceToUInt(l.IndirectValue(cod.(*ssa.Index).Index, errorInfo),
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
		if l.is1usePtr(code[0].(*ssa.UnOp).X) {
			oup, found := l.hc.map1usePtr[code[0].(*ssa.UnOp).X]
			if !found {
				panic("unable to find virtual 1usePtr")
			}
			ret += register + "=" + oup.obj + ".get" + suffix + idx + "+" + oup.off + "); // PEEPHOLE OPTIMIZATION loadObject\n"
		} else {
			ptrString := l.IndirectValue(code[0].(*ssa.UnOp).X, errorInfo)
			ret += fmt.Sprintf("%s=%s", register, ptrString)
			ret += fmt.Sprintf(".obj.get%s%s+%s.off); // PEEPHOLE OPTIMIZATION loadObject\n",
				suffix, idx, ptrString)
		}

	case "phiList":
		//ret += "// PEEPHOLE OPTIMIZATION phiList\n"
		//ret += l.PhiCode(true, 0, code, errorInfo)
	default:
		panic("Unhandled peephole optimization: " + opt)
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
				if l.hc.fnCanOptMap[thisReg] || len(*(cod.(*ssa.Phi).Referrers())) == 0 {
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
					if l.hc.useRegisterArray {
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
	nm := strings.TrimSpace(l.PogoComp().RegisterName(*vp))
	if l.CanInline(vi) {
		code, found := l.PogoComp().InlineMap(nm)
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

type oneUsePtr struct {
	obj, off, objOrig, offOrig string
	varObj, varOff             bool
}

func (l langType) reset1useMap() {
	l.hc.map1usePtr = make(map[ssa.Value]oneUsePtr)
}

func (l langType) set1usePtr(v ssa.Value, oup oneUsePtr) string {
	if l.hc.useRegisterArray {
		l.hc.map1usePtr[v] = oneUsePtr{obj: oup.obj, off: oup.off}
		return ""
	}
	nam := v.Name()
	newObj := ""
	newOff := ""
	ret := ""
	madeVarObj := false
	madeVarOff := false
	for _, eoup := range l.hc.map1usePtr { // TODO speed this up with another map or two
		if oup.obj == eoup.objOrig && eoup.varObj {
			newObj = eoup.obj
		}
		if oup.off == eoup.offOrig && eoup.varOff {
			newOff = eoup.off
		}
	}
	if newObj == "" {
		ret += "var " + nam + "obj=" + oup.obj + ";\n"
		newObj = nam + "obj"
		madeVarObj = true
		l.hc.tempVarList = append(l.hc.tempVarList, regToFree{nam + "obj", "Dynamic"})
	}
	if newOff == "" {
		ret += "var " + nam + "off=" + oup.off + ";\n"
		newOff = nam + "off"
		madeVarOff = true
	}
	l.hc.map1usePtr[v] = oneUsePtr{newObj, newOff, oup.obj, oup.off, madeVarObj, madeVarOff}
	return ret
}

func (l langType) is1usePtr(v interface{}) bool {
	var bl *ssa.BasicBlock
	switch v.(type) {
	case *ssa.FieldAddr:
		bl = v.(*ssa.FieldAddr).Block()
	case *ssa.IndexAddr:
		bl = v.(*ssa.IndexAddr).Block()
	default:
		return false
	}
	val := v.(ssa.Value)
	if l.hc.fnCanOptMap[val.Name()] {
		switch val.Type().Underlying().(type) {
		case *types.Pointer:
			refs := val.Referrers()
			for ll := range *refs {
				if (*refs)[ll].Block() == bl {
					for _, i := range l.hc.subFnInstrs {
						if i == (*refs)[ll].(ssa.Instruction) {
							goto foundInSubFn
						}
					}
					return false
				foundInSubFn:
					switch (*refs)[ll].(type) {
					case *ssa.Store:
						if (*refs)[ll].(*ssa.Store).Addr == val {
							return true
						}
					case *ssa.UnOp:
						if (*refs)[ll].(*ssa.UnOp).Op.String() == "*" {
							return true
						}
					default: // something we don't know how to handle!
						return false
					}
				}
			}
		}
	}
	return false
}
