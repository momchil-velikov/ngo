package typecheck

import (
	"golang/ast"
	"testing"
)

func TestShift(t *testing.T) {
	p, err := compilePackage("_test/src/binary", []string{"shift.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Test type inference.
	V := p.Find("A").(*ast.Var)
	if V.Init.RHS[0].Type() != ast.BuiltinInt {
		t.Error("initializer of `A` must have type `int`")
	}

	V = p.Find("B").(*ast.Var)
	if V.Init.RHS[0].Type() != ast.BuiltinUint {
		t.Error("initializer of `B` must have type `uint`")
	}

	V = p.Find("C").(*ast.Var)
	if V.Init.RHS[0].Type() != ast.BuiltinUint8 {
		t.Error("initializer of `C` must have type `uint8`")
	}

	V = p.Find("D").(*ast.Var)
	if V.Init.RHS[0].Type() != ast.BuiltinInt16 {
		t.Error("initializer of `D` must have type `int16`")
	}

	V = p.Find("E").(*ast.Var)
	if V.Init.RHS[0].Type() != ast.BuiltinInt8 {
		t.Error("initializer of `E` must have type `int8`")
	}

	V = p.Find("F").(*ast.Var)
	if V.Init.RHS[0].Type() != ast.BuiltinUint16 {
		t.Error("initializer of `F` must have type `uint16`")
	}

	V = p.Find("G").(*ast.Var)
	if V.Init.RHS[0].Type() != ast.BuiltinInt {
		t.Error("initializer of `G` must have type `uint16`")
	}

	V = p.Find("H").(*ast.Var)
	x := V.Init.RHS[0].(*ast.CompLiteral)
	if x.Elts[0].Elt.Type() != ast.BuiltinInt64 {
		t.Error("in initializer of `H`: element must have type `int64`")
	}

	V = p.Find("I").(*ast.Var)
	x = V.Init.RHS[0].(*ast.CompLiteral)
	if x.Elts[0].Key.Type() != ast.BuiltinInt8 {
		t.Error("in initializer of `I`: key must have type `int8`")
	}
	if x.Elts[0].Elt.Type() != ast.BuiltinUint16 {
		t.Error("in initializer of `I`: key must have type `uint16`")
	}

	V = p.Find("J").(*ast.Var)
	x = V.Init.RHS[0].(*ast.CompLiteral)
	if x.Elts[0].Elt.Type() != ast.BuiltinUint {
		t.Error("in initializer of `J.X`: element must have type `uint`")
	}
	if x.Elts[1].Elt.Type() != ast.BuiltinInt8 {
		t.Error("in initializer of `J.Y`: key must have type `int8`")
	}

	V = p.Find("K").(*ast.Var)
	x = V.Init.RHS[0].(*ast.CompLiteral)
	if x.Elts[1].Elt.Type() != ast.BuiltinUint {
		t.Error("in initializer of `K.X`: element must have type `uint`")
	}
	if x.Elts[0].Elt.Type() != ast.BuiltinInt8 {
		t.Error("in initializer of `K.Y`: key must have type `int8`")
	}

	V = p.Find("L").(*ast.Var)
	ix := V.Init.RHS[0].(*ast.IndexExpr).I
	if ix.Type() != ast.BuiltinInt8 {
		t.Error("in initializer of `L`: key must have type `int8`")
	}

	V = p.Find("M").(*ast.Var)
	ix = V.Init.RHS[0].(*ast.IndexExpr).I
	if ix.Type() != ast.BuiltinInt {
		t.Error("in initializer of `M`: index must have type `int`")
	}

	V = p.Find("N").(*ast.Var)
	ix = V.Init.RHS[0].(*ast.SliceExpr).Lo
	if ix.Type() != ast.BuiltinInt {
		t.Error("in initializer of `N`: low must have type `int`")
	}
	ix = V.Init.RHS[0].(*ast.SliceExpr).Hi
	if ix.Type() != ast.BuiltinInt {
		t.Error("in initializer of `N`: hi must have type `int`")
	}
	ix = V.Init.RHS[0].(*ast.SliceExpr).Cap
	if ix.Type() != ast.BuiltinInt {
		t.Error("in initializer of `N`: cap must have type `int`")
	}
}
