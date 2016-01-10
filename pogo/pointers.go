// Copyright 2014 Elliott Stoneham and The TARDIS Go Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package pogo

import (
	"fmt"
	"go/types"

	"golang.org/x/tools/go/ssa"
)

// is this value a pointer?
func valIsPointer(v interface{}) bool {
	switch v.(type) {
	case *ssa.Alloc, *ssa.FieldAddr, *ssa.Global, *ssa.IndexAddr:
		return true
	}
	return false
}

// this is not language specific, it is used to show debug information
func showIndirectValue(v interface{}) string {
	switch v.(type) {
	case *ssa.Alloc:
		// TODO check if this is what to do here...
		return fmt.Sprintf("%s", v.(*ssa.Alloc).Name())

	case *ssa.FieldAddr:
		st := v.(*ssa.FieldAddr).X.Type().Underlying().(*types.Pointer).Elem().Underlying().(*types.Struct)
		// Be robust against a bad index.
		name := "?"
		if 0 <= v.(*ssa.FieldAddr).Field && v.(*ssa.FieldAddr).Field < st.NumFields() {
			name = st.Field(v.(*ssa.FieldAddr).Field).Name()
		}
		return fmt.Sprintf("%s.%s [#%d]",
			showIndirectValue(v.(*ssa.FieldAddr).X),
			name, v.(*ssa.FieldAddr).Field)

	case *ssa.IndexAddr:
		return fmt.Sprintf("%s[%s]",
			showIndirectValue(v.(*ssa.IndexAddr).X),
			v.(*ssa.IndexAddr).Index.Name())

	case *ssa.Global:
		return fmt.Sprintf("%s", v.(*ssa.Global).Name()) //TODO qualify name

	default:
		//LogError("TODO", "pogo", fmt.Errorf("getIndirectValue() unhandled type: %v", reflect.TypeOf(v)))
		return fmt.Sprintf("%s", v.(ssa.Value).String())
	}
}
