package typecheck

import (
	"golang/ast"
	"testing"
)

func TestIndexExpr(t *testing.T) {
	p, err := compilePackage("_test/src/index", []string{"index.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	ax := p.Find("ax").(*ast.Var)
	typ := ax.Type
	if unnamedType(typ) != ast.BuiltinInt8 {
		t.Error("the type of `ax` must be `int8`")
	}

	bx := p.Find("bx").(*ast.Var)
	typ = bx.Type
	if unnamedType(typ) != ast.BuiltinInt16 {
		t.Error("the type of `bx` must be `int16`")
	}

	cx := p.Find("cx").(*ast.Var)
	typ = cx.Type
	if unnamedType(typ) != ast.BuiltinInt32 {
		t.Error("the type of `bx` must be `int32`")
	}

	dx := p.Find("dx").(*ast.Var)
	typ = dx.Type
	if unnamedType(typ) != ast.BuiltinUint8 {
		t.Error("the type of `dx` must be `uint8`")
	}

	ex := p.Find("ex").(*ast.Var)
	typ = ex.Type
	if unnamedType(typ) != ast.BuiltinFloat32 {
		t.Error("the type of `dx` must be `float32`")
	}
}

func TestIndexExprErrr(t *testing.T) {
	expectError(t, "_test/src/index", []string{"index-err-1.go"},
		"type does not support indexing")
	expectError(t, "_test/src/index", []string{"index-err-2.go"},
		"type does not support indexing")
	expectError(t, "_test/src/index", []string{"index-err-3.go"},
		"type does not support indexing")
	expectError(t, "_test/src/index", []string{"index-err-4.go"},
		"index must be of integer type")
	expectError(t, "_test/src/index", []string{"index-err-5.go"},
		"index must be of integer type")
	expectError(t, "_test/src/index", []string{"index-err-6.go"},
		"index out of bounds")
	expectError(t, "_test/src/index", []string{"index-err-7.go"},
		"1.100000 (untyped) cannot be converted to int")
	expectError(t, "_test/src/index", []string{"index-err-8.go"},
		"index out of bounds")
	expectError(t, "_test/src/index", []string{"index-err-9.go"},
		"index out of bounds")
}
