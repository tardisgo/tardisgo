// Copyright 2014 Elliott Stoneham and The TARDIS Go Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package asmgo

import (
	"fmt"

	"go/types"
)

var haxeStdSizes = types.StdSizes{
	WordSize: 4, // word size in bytes - must be >= 4 (32bits)
	MaxAlign: 8, // maximum alignment in bytes - must be >= 1
}

func fieldOffset(str *types.Struct, fldNum int) int64 {
	fieldList := make([]*types.Var, str.NumFields())
	for f := 0; f < str.NumFields(); f++ {
		fieldList[f] = str.Field(f)
	}
	return haxeStdSizes.Offsetsof(fieldList)[fldNum]
}

func arrayOffsetCalc(ele types.Type) string {
	ent := types.NewVar(0, nil, "___temp", ele)
	fieldList := []*types.Var{ent, ent}
	off := haxeStdSizes.Offsetsof(fieldList)[1] // to allow for word alignment
	//off := haxeStdSizes.Sizeof(ele) // ?? or should it be the code above ?
	if off == 1 {
		return ""
	}
	for ls := uint(1); ls < 31; ls++ {
		target := int64(1 << ls)
		if off == target {
			return fmt.Sprintf("<<%d", ls)
		}
		if off < target {
			break // no point in looking any further
		}
	}
	return fmt.Sprintf("*%d", off)
}
