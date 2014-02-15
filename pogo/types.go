// Copyright 2014 Elliott Stoneham and The TARDIS Go Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package pogo

import (
	"code.google.com/p/go.tools/go/types"
	"code.google.com/p/go.tools/go/types/typemap"
	"fmt"
	"reflect"
)

// IsValidInPogo exists to screen out any types that the system does not handle correctly.
// Currently it should say everything is valid. TODO review if still required in this form.
func IsValidInPogo(et types.Type, posStr string) bool {
	switch et.(type) {
	case *types.Basic:
		switch et.(*types.Basic).Kind() {
		case types.Bool, types.UntypedBool, types.String, types.UntypedString, types.Float64, types.Float32, types.UntypedFloat,
			types.Int, types.Int8, types.Int16, types.Int32, types.UntypedInt, types.UntypedRune,
			types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Int64, types.Uint64,
			types.Complex64, types.Complex128, types.UntypedComplex,
			types.Uintptr, types.UnsafePointer: // these last two return a poisoned value
			return true
		default:
			if et.(*types.Basic).String() == "invalid type" { // the type of unused map value itterators!
				return true
			}
			LogError(posStr, "pogo", fmt.Errorf("basic type %s is not supported", et.(*types.Basic).String()))
		}
	case *types.Interface, *types.Slice, *types.Struct, *types.Tuple, *types.Map, *types.Pointer, *types.Array,
		*types.Named, *types.Signature, *types.Chan:
		return true
	default:
		rTyp := reflect.TypeOf(et).String()
		if rTyp == "*ssa.opaqueType" { // the type of map itterators!
			return true
		}
		LogError(posStr, "pogo", fmt.Errorf("type %s is not supported", rTyp))
	}
	return false
}

// TypesEncountered keeps track of the types we encounter using the excellent go.tools/go/types/typesmap package.
var TypesEncountered typemap.M

// NextTypeID is used to give each type we come across its own ID.
var NextTypeID = 0

// LogTypeUse : As the code generator encounters new types it logs them here, returning a string of the ID for insertion into the code.
func LogTypeUse(t types.Type) string {
	r := TypesEncountered.At(t)
	if r != nil {
		return fmt.Sprintf("%d", r)
	}
	TypesEncountered.Set(t, NextTypeID)
	r = NextTypeID
	NextTypeID++
	return fmt.Sprintf("%d", r)
}

// TypesWithMethodSets ia a utility function to avoid exposing rootProgram
func TypesWithMethodSets() []types.Type {
	return rootProgram.TypesWithMethodSets()
}

// Wrapper for target language emitTypeInfo()
func emitTypeInfo() {
	l := TargetLang
	fmt.Fprintln(&LanguageList[l].buffer, LanguageList[l].EmitTypeInfo())
}
