// Copyright 2014 Elliott Stoneham and The TARDIS Go Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package haxe

import (
	"fmt"
	"sort"

	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/types"
)

type phiEntry struct{ reg, val string }

// PeepholeOpt implements the optimisations spotted by pogo.peephole
func (l langType) PeepholeOpt(opt, register string, code []ssa.Instruction, errorInfo string) string {
	ret := ""
	switch opt {
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
		opts := make(map[int][]phiEntry)
		for _, cod := range code {
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
				opts[phiEntries[o]] = append(opts[phiEntries[o]], phiEntry{thisReg, valEntries[o]})
			}
		}

		ret += "switch(_Phi) { \n"
		idxs := make([]int, 0, len(opts))
		for phi := range opts {
			idxs = append(idxs, phi)
		}
		sort.Ints(idxs)
		for _, phi := range idxs {
			opt := opts[phi]
			ret += fmt.Sprintf("\tcase %d:\n", phi)
			for _, ent := range opt {
				ret += fmt.Sprintf("\t\tvar tmp_%s=%s;\n", ent.reg, ent.val) // need temp vars for a,b = b,a
			}
			for _, ent := range opt {
				rn := "_" + ent.reg
				if useRegisterArray {
					rn = rn[:2] + "[" + rn[2:] + "]"
				}
				ret += fmt.Sprintf("\t\t%s=tmp_%s;\n", rn, ent.reg)
			}
		}
		ret += "}\n"
	}
	return ret
}
