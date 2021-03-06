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
	if ax.Type != ast.BuiltinInt8 {
		t.Error("the type of `ax` must be `int8`")
	}

	bx := p.Find("bx").(*ast.Var)
	d := bx.Type.(*ast.TypeDecl)
	if d.Name != "I16" || underlyingType(d.Type) != ast.BuiltinInt16 {
		t.Error("the type of `bx` must be `I16`")
	}

	cx := p.Find("cx").(*ast.Var)
	d = cx.Type.(*ast.TypeDecl)
	if d.Name != "I32" || underlyingType(d.Type) != ast.BuiltinInt32 {
		t.Error("the type of `cx` must be `I32`")
	}

	dx := p.Find("dx").(*ast.Var)
	if dx.Type != ast.BuiltinUint8 {
		t.Error("the type of `dx` must be `byte`")
	}

	ex := p.Find("ex").(*ast.Var)
	d = ex.Type.(*ast.TypeDecl)
	if d.Name != "F32" || underlyingType(d.Type) != ast.BuiltinFloat32 {
		t.Error("the type of `ex` must be `F32`")
	}

	fx := p.Find("fx").(*ast.Var)
	d = fx.Type.(*ast.TypeDecl)
	if d.Name != "F32" || underlyingType(d.Type) != ast.BuiltinFloat32 {
		t.Error("the type of `fx` must be `F32`")
	}
}

func TestIndexExprErrr(t *testing.T) {
	expectError(t, "_test/src/index", []string{"index-err-1.go"},
		"type `int` does not support indexing")
	expectError(t, "_test/src/index", []string{"index-err-2.go"},
		"type `*map[int]int` does not support indexing")
	expectError(t, "_test/src/index", []string{"index-err-3.go"},
		"type `chan int` does not support indexing")
	expectError(t, "_test/src/index", []string{"index-err-4.go"},
		"index must be of integer type")
	expectError(t, "_test/src/index", []string{"index-err-5.go"},
		"index must be of integer type")
	expectError(t, "_test/src/index", []string{"index-err-6.go"},
		"index `-1` out of bounds")
	expectError(t, "_test/src/index", []string{"index-err-7.go"},
		"1.1 (`untyped float`) cannot be converted to `int`")
	expectError(t, "_test/src/index", []string{"index-err-8.go"},
		"index `2` out of bounds")
	expectError(t, "_test/src/index", []string{"index-err-9.go"},
		"index `5` out of bounds")
	expectError(t, "_test/src/index", []string{"index-err-10.go"},
		"1.1 (`untyped float`) cannot be converted to `I16`")
	expectError(t, "_test/src/index", []string{"index-err-11.go"},
		"`float64` is not assignable to `I16`")
	expectError(t, "_test/src/index", []string{"index-err-12.go"},
		"1.1 (`untyped float`) cannot be converted to `int`")
}
