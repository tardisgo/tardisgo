// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This file implements export filtering of an AST.

package doc

import (
	"go/ast"
	"go/token"
)

// filterIdentList removes unexported names from list in place
// and returns the resulting list. If blankOk is set, blank
// identifiers are considered exported names.
//
func filterIdentList(list []*ast.Ident, blankOk bool) []*ast.Ident {
	j := 0
	for _, x := range list {
		if ast.IsExported(x.Name) || (blankOk && x.Name == "_") {
			list[j] = x
			j++
		}
	}
	return list[0:j]
}

// removeErrorField removes anonymous fields named "error" from an interface.
// This is called when "error" has been determined to be a local name,
// not the predeclared type.
//
func removeErrorField(ityp *ast.InterfaceType) {
	list := ityp.Methods.List // we know that ityp.Methods != nil
	j := 0
	for _, field := range list {
		keepField := true
		if n := len(field.Names); n == 0 {
			// anonymous field
			if fname, _ := baseTypeName(field.Type); fname == "error" {
				keepField = false
			}
		}
		if keepField {
			list[j] = field
			j++
		}
	}
	if j < len(list) {
		ityp.Incomplete = true
	}
	ityp.Methods.List = list[0:j]
}

// filterFieldList removes unexported fields (field names) from the field list
// in place and returns true if fields were removed. Anonymous fields are
// recorded with the parent type. filterType is called with the types of
// all remaining fields.
//
func (r *reader) filterFieldList(parent *namedType, fields *ast.FieldList, ityp *ast.InterfaceType) (removedFields bool) {
	if fields == nil {
		return
	}
	list := fields.List
	j := 0
	for _, field := range list {
		keepField := false
		if n := len(field.Names); n == 0 {
			// anonymous field
			fname := r.recordAnonymousField(parent, field.Type)
			if ast.IsExported(fname) {
				keepField = true
			} else if ityp != nil && fname == "error" {
				// possibly the predeclared error interface; keep
				// it for now but remember this interface so that
				// it can be fixed if error is also defined locally
				keepField = true
				r.remember(ityp)
			}
		} else {
			field.Names = filterIdentList(field.Names, false)
			if len(field.Names) < n {
				removedFields = true
			}
			if len(field.Names) > 0 {
				keepField = true
			}
		}
		if keepField {
			r.filterType(nil, field.Type)
			list[j] = field
			j++
		}
	}
	if j < len(list) {
		removedFields = true
	}
	fields.List = list[0:j]
	return
}

// filterParamList applies filterType to each parameter type in fields.
//
func (r *reader) filterParamList(fields *ast.FieldList) {
	if fields != nil {
		for _, f := range fields.List {
			r.filterType(nil, f.Type)
		}
	}
}

// filterType strips any unexported struct fields or method types from typ
// in place. If fields (or methods) have been removed, the corresponding
// struct or interface type has the Incomplete field set to true.
//
func (r *reader) filterType(parent *namedType, typ ast.Expr) {
	switch t := typ.(type) {
	case *ast.Ident:
		// nothing to do
	case *ast.ParenExpr:
		r.filterType(nil, t.X)
	case *ast.ArrayType:
		r.filterType(nil, t.Elt)
	case *ast.StructType:
		if r.filterFieldList(parent, t.Fields, nil) {
			t.Incomplete = true
		}
	case *ast.FuncType:
		r.filterParamList(t.Params)
		r.filterParamList(t.Results)
	case *ast.InterfaceType:
		if r.filterFieldList(parent, t.Methods, t) {
			t.Incomplete = true
		}
	case *ast.MapType:
		r.filterType(nil, t.Key)
		r.filterType(nil, t.Value)
	case *ast.ChanType:
		r.filterType(nil, t.Value)
	}
}

func (r *reader) filterSpec(spec ast.Spec, tok token.Token) bool {
	switch s := spec.(type) {
	case *ast.ImportSpec:
		// always keep imports so we can collect them
		return true
	case *ast.ValueSpec:
		// special case: consider blank constants as exported
		// (work-around for issue 5397)
		s.Names = filterIdentList(s.Names, tok == token.CONST)
		if len(s.Names) > 0 {
			r.filterType(nil, s.Type)
			return true
		}
	case *ast.TypeSpec:
		if name := s.Name.Name; ast.IsExported(name) {
			r.filterType(r.lookupType(s.Name.Name), s.Type)
			return true
		} else if name == "error" {
			// special case: remember that error is declared locally
			r.errorDecl = true
		}
	}
	return false
}

func (r *reader) filterSpecList(list []ast.Spec, tok token.Token) []ast.Spec {
	j := 0
	for _, s := range list {
		if r.filterSpec(s, tok) {
			list[j] = s
			j++
		}
	}
	return list[0:j]
}

func (r *reader) filterDecl(decl ast.Decl) bool {
	switch d := decl.(type) {
	case *ast.GenDecl:
		d.Specs = r.filterSpecList(d.Specs, d.Tok)
		return len(d.Specs) > 0
	case *ast.FuncDecl:
		// ok to filter these methods early because any
		// conflicting method will be filtered here, too -
		// thus, removing these methods early will not lead
		// to the false removal of possible conflicts
		return ast.IsExported(d.Name.Name)
	}
	return false
}

// fileExports removes unexported declarations from src in place.
//
func (r *reader) fileExports(src *ast.File) {
	j := 0
	for _, d := range src.Decls {
		if r.filterDecl(d) {
			src.Decls[j] = d
			j++
		}
	}
	src.Decls = src.Decls[0:j]
}
