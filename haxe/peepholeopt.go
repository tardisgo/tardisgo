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
		ret += fmt.Sprintf("// %s=%s\n", code[0].(*ssa.UnOp).Name(), code[0].String())
		for _, cod := range code[1:] {
			switch cod.(type) {
			case *ssa.Index:
				ret += fmt.Sprintf("// %s=%s\n", cod.(*ssa.Index).Name(), cod.String())
			case *ssa.Field:
				ret += fmt.Sprintf("// %s=%s\n", cod.(*ssa.Field).Name(), cod.String())
			}
		}
		ret += fmt.Sprintf("%s=%s", register, l.IndirectValue(code[0].(*ssa.UnOp).X, errorInfo))
		for _, cod := range code[1:] {
			switch cod.(type) {
			case *ssa.Index:
				idx := wrapForce_toUInt(l.IndirectValue(cod.(*ssa.Index).Index, errorInfo),
					cod.(*ssa.Index).Index.Type().Underlying().(*types.Basic).Kind())
				ret += fmt.Sprintf(".addr(%s%s)",
					idx,
					arrayOffsetCalc(cod.(*ssa.Index).Type().Underlying()))
			case *ssa.Field:
				ret += fmt.Sprintf(".fieldAddr(%d)",
					fieldOffset(cod.(*ssa.Field).X.Type().Underlying().(*types.Struct), cod.(*ssa.Field).Field))
			}
		}
		switch code[len(code)-1].(type) {
		case *ssa.Index:
			ret += fmt.Sprintf(".load%s); // PEEPHOLE OPTIMIZATION loadObject (Index)\n",
				loadStoreSuffix(code[len(code)-1].(*ssa.Index).Type().Underlying(), false))
		case *ssa.Field:
			ret += fmt.Sprintf(".load%s); // PEEPHOLE OPTIMIZATION loadObject (Field)\n",
				loadStoreSuffix(code[len(code)-1].(*ssa.Field).Type().Underlying(), false))
		}
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
				ret += fmt.Sprintf("\t\t_%s=tmp_%s;\n", ent.reg, ent.reg)
			}
		}
		ret += "}\n"
	}
	return ret
}
