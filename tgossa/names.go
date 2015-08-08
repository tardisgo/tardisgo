package tgossa

import (
	"fmt"

	"golang.org/x/tools/go/ssa"
)

// CheckNames exists because of this comment in the SSA code documentation:
// "Many objects are nonetheless named to aid in debugging, but it is not essential that the names be either accurate or unambiguous. "
func CheckNames(f *ssa.Function) error {
	names := make(map[string]*ssa.Instruction)

	//fmt.Println("DEBUG Check Names for ", f.String())

	for blk := range f.Blocks {
		for ins := range f.Blocks[blk].Instrs {
			instrVal, hasVal := f.Blocks[blk].Instrs[ins].(ssa.Value)
			if hasVal {
				register := instrVal.Name()
				//fmt.Println("DEBUG name ", register)
				val, found := names[register]
				if found {
					if val != &f.Blocks[blk].Instrs[ins] {
						return fmt.Errorf("internal error, ssa register names not unique in function %s var name %s",
							f.String(), register)
					}
				} else {
					names[register] = &f.Blocks[blk].Instrs[ins]
				}
			}
		}
	}
	return nil
}
