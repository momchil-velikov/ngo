package typecheck

import (
	"golang/ast"
	"testing"
)

func TestSliceExpr(t *testing.T) {
	p, err := compilePackage("_test/src/slice", []string{"slice.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	b := p.Find("b").(*ast.Var)
	if b.Type != ast.BuiltinString {
		t.Error("`b` must have type `string`")
	}
	e := p.Find("e").(*ast.Var)
	if e.Type != ast.BuiltinString {
		t.Error("`e` must have type `string`")
	}

	f := p.Find("f").(*ast.Var)
	typ := f.Type.(*ast.SliceType)
	d := typ.Elt.(*ast.TypeDecl)
	if d.Name != "Int" {
		t.Error("`f` must have type `[]Int`")
	}
	g := p.Find("g").(*ast.Var)
	typ = g.Type.(*ast.SliceType)
	d = typ.Elt.(*ast.TypeDecl)
	if d.Name != "Int" {
		t.Error("`g` must have type `[]Int`")
	}

	ff := p.Find("ff").(*ast.Var)
	d = ff.Type.(*ast.TypeDecl)
	if d.Name != "S" {
		t.Error("`ff` must have type `S`")
	}
}

func TestSliceExprErr(t *testing.T) {
	expectError(t, "_test/src/slice", []string{"slice-err-1.go"},
		"does not support 3-index slicing")
	expectError(t, "_test/src/slice", []string{"slice-err-2.go"},
		"does not support 3-index slicing")
	expectError(t, "_test/src/slice", []string{"slice-err-3.go"},
		"index `-1` out of bounds")
	expectError(t, "_test/src/slice", []string{"slice-err-4.go"},
		"index `1` out of bounds")
	expectError(t, "_test/src/slice", []string{"slice-err-5.go"},
		"index `6` out of bounds")
	expectError(t, "_test/src/slice", []string{"slice-err-6.go"},
		"index `6` out of bounds")
	expectError(t, "_test/src/slice", []string{"slice-err-7.go"},
		"index must be of integer type")
	expectError(t, "_test/src/slice", []string{"slice-err-8.go"},
		"index must be of integer type")
	expectError(t, "_test/src/slice", []string{"slice-err-9.go"},
		"type `*int` does not support indexing or slicing")
	expectError(t, "_test/src/slice", []string{"slice-err-10.go"},
		"type `*[]int` does not support indexing or slicing")
	expectError(t, "_test/src/slice", []string{"slice-err-11.go"},
		"type `chan int` does not support indexing or slicing")
	expectError(t, "_test/src/slice", []string{"slice-err-12.go"},
		"index `6` out of bounds")
	expectError(t, "_test/src/slice", []string{"slice-err-13.go"},
		"index `6` out of bounds")
	expectError(t, "_test/src/slice", []string{"slice-err-14.go"},
		"index `6` out of bounds")
	expectError(t, "_test/src/slice", []string{"slice-err-15.go"},
		"index `2` out of bounds")
	expectError(t, "_test/src/slice", []string{"slice-err-16.go"},
		"index `2` out of bounds")
	expectError(t, "_test/src/slice", []string{"slice-err-17.go"},
		"index `2` out of bounds")
	expectError(t, "_test/src/slice", []string{"slice-err-18.go"},
		"index `-1` out of bounds")
	expectError(t, "_test/src/slice", []string{"slice-err-19.go"},
		"index `-1` out of bounds")
	expectError(t, "_test/src/slice", []string{"slice-err-20.go"},
		"index `-1` out of bounds")
	expectError(t, "_test/src/slice", []string{"slice-err-21.go"},
		"index `1` out of bounds")
	expectError(t, "_test/src/slice", []string{"slice-err-22.go"},
		"index `1` out of bounds")
	expectError(t, "_test/src/slice", []string{"slice-err-23.go"},
		"index `1` out of bounds")
	expectError(t, "_test/src/slice", []string{"slice-err-24.go"},
		"type `int` does not support indexing or slicing")
	expectError(t, "_test/src/slice", []string{"slice-err-25.go"},
		"index `-1` out of bounds")
	expectError(t, "_test/src/slice", []string{"slice-err-26.go"},
		"index `-1` out of bounds")
	expectError(t, "_test/src/slice", []string{"slice-err-27.go"},
		"index `-1` out of bounds")
}
