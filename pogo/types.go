// Copyright 2014 Elliott Stoneham and The TARDIS Go Authors
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package pogo

import (
	"fmt"
	"reflect"
	"sort"

	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/types"
	"golang.org/x/tools/go/types/typeutil"
)

// IsValidInPogo exists to screen out any types that the system does not handle correctly.
// Currently it should say everything is valid.
// TODO review if still required in this form.
func IsValidInPogo(et types.Type, posStr string) bool {
	switch et.(type) {
	case *types.Basic:
		switch et.(*types.Basic).Kind() {
		case types.Bool, types.String, types.Float64, types.Float32,
			types.Int, types.Int8, types.Int16, types.Int32,
			types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Int64, types.Uint64,
			types.Complex64, types.Complex128,
			types.Uintptr, types.UnsafePointer:
			return true
		case types.UntypedInt, types.UntypedRune, types.UntypedBool,
			types.UntypedString, types.UntypedFloat, types.UntypedComplex:
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
var TypesEncountered typeutil.Map

// NextTypeID is used to give each type we come across its own ID.
var NextTypeID = 1 // entry zero is invalid

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

// TypesWithMethodSets in a utility function to only return seen types
func TypesWithMethodSets() (sets []types.Type) {
	typs := rootProgram.RuntimeTypes()
	for _, t := range typs {
		if TypesEncountered.At(t) != nil {
			sets = append(sets, t)
		}
	}
	return sets
}

func MethodSetFor(T types.Type) *types.MethodSet {
	return rootProgram.MethodSets.MethodSet(T)
}

func RootProgram() *ssa.Program {
	return rootProgram
}

var catchReferencedTypesSeen = make(map[string]bool)

func catchReferencedTypes(et types.Type) {
	id := LogTypeUse(et)
	_, seen := catchReferencedTypesSeen[id]
	if seen {
		return
	}
	catchReferencedTypesSeen[id] = true

	// check that we have all the required methods?
	/*
		for t := 1; t < NextTypeID; t++ { // make sure we do this in a consistent order
			for _, k := range TypesEncountered.Keys() {
				if TypesEncountered.At(k).(int) == t {
					switch k.(type) {
					case *types.Interface:
						if types.Implements(et,k.(*types.Interface)) {
							// TODO call missing method?
						}
					}
				}
			}
		}
	*/

	//LogTypeUse(types.NewPointer(et))
	switch et.(type) {
	case *types.Named:
		catchReferencedTypes(et.Underlying())
		for m := 0; m < et.(*types.Named).NumMethods(); m++ {
			catchReferencedTypes(et.(*types.Named).Method(m).Type())
		}
	case *types.Array:
		catchReferencedTypes(et.(*types.Array).Elem())
		//catchReferencedTypes(types.NewSlice(et.(*types.Array).Elem()))
	case *types.Pointer:
		catchReferencedTypes(et.(*types.Pointer).Elem())
	case *types.Slice:
		catchReferencedTypes(et.(*types.Slice).Elem())
	case *types.Struct:
		for f := 0; f < et.(*types.Struct).NumFields(); f++ {
			if et.(*types.Struct).Field(f).IsField() {
				catchReferencedTypes(et.(*types.Struct).Field(f).Type())
			}
		}
	case *types.Map:
		catchReferencedTypes(et.(*types.Map).Key())
		catchReferencedTypes(et.(*types.Map).Elem())
	case *types.Signature:
		for i := 0; i < et.(*types.Signature).Params().Len(); i++ {
			catchReferencedTypes(et.(*types.Signature).Params().At(i).Type())
		}
		for o := 0; o < et.(*types.Signature).Results().Len(); o++ {
			catchReferencedTypes(et.(*types.Signature).Results().At(o).Type())
		}
	case *types.Chan:
		catchReferencedTypes(et.(*types.Chan).Elem())
	}
}

func visitAllTypes() {
	// add the supplied method, required to make sure no synthetic types or types referenced via interfaces have been missed
	rt := rootProgram.RuntimeTypes()
	sort.Sort(TypeSorter(rt))
	for _, T := range rt {
		LogTypeUse(T)
	}
	// ...so just get the full info on the types we've seen
	for t := 1; t < NextTypeID; t++ { // make sure we do this in a consistent order
		for _, k := range TypesEncountered.Keys() {
			if TypesEncountered.At(k).(int) == t {
				catchReferencedTypes(k)
			}
		}
	}
}

// Wrapper for target language emitTypeInfo()
func emitTypeInfo() {
	visitAllTypes()
	l := TargetLang

	if len(LibListNoDCE) > 0 { // output target lang type to access named object
		for t := 1; t < NextTypeID; t++ { // make sure we do this in a consistent order
			for _, k := range TypesEncountered.Keys() {
				if TypesEncountered.At(k).(int) == t {
					switch k.(type) {
					case *types.Named:
						if k.(*types.Named).Obj().Exported() {
							fmt.Fprintln(&LanguageList[l].buffer,
								LanguageList[l].TypeStart(k.(*types.Named), k.String()))
							//fmt.Fprintln(&LanguageList[l].buffer,
							//	LanguageList[l].TypeEnd(k.(*types.Named), k.String()))
						}
					}
				}
			}
		}
	}

	fmt.Fprintln(&LanguageList[l].buffer, LanguageList[l].EmitTypeInfo())
}
