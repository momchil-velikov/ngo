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
