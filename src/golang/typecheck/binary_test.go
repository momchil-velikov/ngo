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

	// Test constant shifts evaluation.
	V = p.Find("c0").(*ast.Var)
	c := V.Init.RHS[0].(*ast.ConstValue)
	if v, ok := c.Value.(ast.Int); !ok || v != 2 {
		t.Error("`c0 should have value 2")
	}

	V = p.Find("c1").(*ast.Var)
	c = V.Init.RHS[0].(*ast.ConstValue)
	if v, ok := c.Value.(ast.Int); !ok || int64(v) != -1 {
		t.Error("`c1` should have value -1")
	}

	V = p.Find("c2").(*ast.Var)
	c = V.Init.RHS[0].(*ast.ConstValue)
	if v, ok := c.Value.(ast.Int); !ok || v != 1 {
		t.Error("`c2` should have value 1")
	}

	V = p.Find("c3").(*ast.Var)
	c = V.Init.RHS[0].(*ast.ConstValue)
	if v, ok := c.Value.(ast.Int); !ok || v != 2 {
		t.Error("`c3` should have value 2")
	}

	V = p.Find("c4").(*ast.Var)
	c = V.Init.RHS[0].(*ast.ConstValue)
	if v, ok := c.Value.(ast.Int); !ok || int64(v) != -1 {
		t.Error("`c4` should have value -1")
	}

	V = p.Find("c5").(*ast.Var)
	c = V.Init.RHS[0].(*ast.ConstValue)
	if v, ok := c.Value.(ast.Int); !ok || v != 2 {
		t.Error("`c5` should have value 2")
	}

	V = p.Find("c6").(*ast.Var)
	c = V.Init.RHS[0].(*ast.ConstValue)
	if v, ok := c.Value.(ast.Int); !ok || v != 3 {
		t.Error("`c6` should have value 3")
	}

	V = p.Find("c7").(*ast.Var)
	c = V.Init.RHS[0].(*ast.ConstValue)
	if v, ok := c.Value.(ast.Int); !ok || v != 0 {
		t.Error("`c7` should have value 0")
	}

	V = p.Find("c8").(*ast.Var)
	c = V.Init.RHS[0].(*ast.ConstValue)
	if v, ok := c.Value.(ast.Int); !ok || v != 0 {
		t.Error("`c8` should have value 0")
	}

	V = p.Find("c9").(*ast.Var)
	c = V.Init.RHS[0].(*ast.ConstValue)
	if v, ok := c.Value.(ast.Int); !ok || v != 194 {
		t.Error("`c9` should have value 194")
	}

	V = p.Find("c10").(*ast.Var)
	c = V.Init.RHS[0].(*ast.ConstValue)
	if v, ok := c.Value.(ast.Int); !ok || v != 49 {
		t.Error("`c10` should have value 49")
	}
}

func TestShiftErr(t *testing.T) {
	expectError(t, "_test/src/binary", []string{"shift-err-1.go"},
		"shift count must be unsigned integer")
	expectError(t, "_test/src/binary", []string{"shift-err-2.go"},
		"1.100000 (untyped) cannot be converted to uint64")
	expectError(t, "_test/src/binary", []string{"shift-err-3.go"},
		"shift count must be unsigned integer")
	expectError(t, "_test/src/binary", []string{"shift-err-4.go"},
		"shift count too big")
	expectError(t, "_test/src/binary", []string{"shift-err-5.go"},
		"shifted operand must have integer type")
	expectError(t, "_test/src/binary", []string{"shift-err-6.go"},
		"shifted operand must have integer type")
	expectError(t, "_test/src/binary", []string{"shift-err-7.go"},
		"shifted operand must have integer type")
	expectError(t, "_test/src/binary", []string{"shift-err-8.go"},
		"shifted operand must have integer type")
	expectError(t, "_test/src/binary", []string{"shift-err-9.go"},
		"shifted operand must have integer type")
	expectError(t, "_test/src/binary", []string{"shift-err-10.go"},
		"invalid operand to shift")
	expectError(t, "_test/src/binary", []string{"shift-err-11.go"},
		"invalid operand to shift")
	expectError(t, "_test/src/binary", []string{"shift-err-12.go"},
		"shift count must be unsigned and integer")
	expectError(t, "_test/src/binary", []string{"shift-err-13.go"},
		"shift count must be unsigned and integer")
}
