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
)

// IsValidInPogo exists to screen out any types that the system does not handle correctly.
// Currently it should say everything is valid.
// TODO review if still required in this form.
func (comp *Compilation) IsValidInPogo(et types.Type, posStr string) bool {
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
			comp.LogError(posStr, "pogo", fmt.Errorf("basic type %s is not supported", et.(*types.Basic).String()))
		}
	case *types.Interface, *types.Slice, *types.Struct, *types.Tuple, *types.Map, *types.Pointer, *types.Array,
		*types.Named, *types.Signature, *types.Chan:
		return true
	default:
		rTyp := reflect.TypeOf(et).String()
		if rTyp == "*ssa.opaqueType" { // the type of map itterators!
			return true
		}
		comp.LogError(posStr, "pogo", fmt.Errorf("type %s is not supported", rTyp))
	}
	return false
}

func (comp *Compilation) initTypes() {
	comp.catchReferencedTypesSeen = make(map[string]bool)
	comp.NextTypeID = 1 // entry zero is invalid
}

// LogTypeUse : As the code generator encounters new types it logs them here, returning a string of the ID for insertion into the code.
func (comp *Compilation) LogTypeUse(t types.Type) string {
	r := comp.TypesEncountered.At(t)
	if r != nil {
		return fmt.Sprintf("%d", r)
	}
	comp.TypesEncountered.Set(t, comp.NextTypeID)
	r = comp.NextTypeID
	comp.NextTypeID++
	return fmt.Sprintf("%d", r)
}

// TypesWithMethodSets in a utility function to only return seen types
func (comp *Compilation) TypesWithMethodSets() (sets []types.Type) {
	typs := comp.rootProgram.RuntimeTypes()
	for _, t := range typs {
		if comp.TypesEncountered.At(t) != nil {
			sets = append(sets, t)
		}
	}
	return sets
}

// MethodSetFor is a conveniance function
func (comp *Compilation) MethodSetFor(T types.Type) *types.MethodSet {
	return comp.rootProgram.MethodSets.MethodSet(T)
}

// RootProgram is a conveniance function
func (comp *Compilation) RootProgram() *ssa.Program {
	return comp.rootProgram
}

func (comp *Compilation) catchReferencedTypes(et types.Type) {
	id := comp.LogTypeUse(et)
	_, seen := comp.catchReferencedTypesSeen[id]
	if seen {
		return
	}
	comp.catchReferencedTypesSeen[id] = true

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
		comp.catchReferencedTypes(et.Underlying())
		for m := 0; m < et.(*types.Named).NumMethods(); m++ {
			comp.catchReferencedTypes(et.(*types.Named).Method(m).Type())
		}
	case *types.Array:
		comp.catchReferencedTypes(et.(*types.Array).Elem())
		//catchReferencedTypes(types.NewSlice(et.(*types.Array).Elem()))
	case *types.Pointer:
		comp.catchReferencedTypes(et.(*types.Pointer).Elem())
	case *types.Slice:
		comp.catchReferencedTypes(et.(*types.Slice).Elem())
	case *types.Struct:
		for f := 0; f < et.(*types.Struct).NumFields(); f++ {
			if et.(*types.Struct).Field(f).IsField() {
				comp.catchReferencedTypes(et.(*types.Struct).Field(f).Type())
			}
		}
	case *types.Map:
		comp.catchReferencedTypes(et.(*types.Map).Key())
		comp.catchReferencedTypes(et.(*types.Map).Elem())
	case *types.Signature:
		for i := 0; i < et.(*types.Signature).Params().Len(); i++ {
			comp.catchReferencedTypes(et.(*types.Signature).Params().At(i).Type())
		}
		for o := 0; o < et.(*types.Signature).Results().Len(); o++ {
			comp.catchReferencedTypes(et.(*types.Signature).Results().At(o).Type())
		}
	case *types.Chan:
		comp.catchReferencedTypes(et.(*types.Chan).Elem())
	}
}

func (comp *Compilation) visitAllTypes() {
	// add the supplied method, required to make sure no synthetic types or types referenced via interfaces have been missed
	rt := comp.rootProgram.RuntimeTypes()
	sort.Sort(TypeSorter(rt))
	for _, T := range rt {
		comp.LogTypeUse(T)
	}
	// ...so just get the full info on the types we've seen
	for t := 1; t < comp.NextTypeID; t++ { // make sure we do this in a consistent order
		for _, k := range comp.TypesEncountered.Keys() {
			if comp.TypesEncountered.At(k).(int) == t {
				comp.catchReferencedTypes(k)
			}
		}
	}
}

// Wrapper for target language emitTypeInfo()
func (comp *Compilation) emitTypeInfo() {
	comp.visitAllTypes()
	l := comp.TargetLang

	if len(comp.LibListNoDCE) > 0 { // output target lang type to access named object
		for t := 1; t < comp.NextTypeID; t++ { // make sure we do this in a consistent order
			for _, k := range comp.TypesEncountered.Keys() {
				if comp.TypesEncountered.At(k).(int) == t {
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
