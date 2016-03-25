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
		"shift count must be unsigned integer (`int` given)")
	expectError(t, "_test/src/binary", []string{"shift-err-2.go"},
		"1.100000 (`untyped float`) cannot be converted to `uint64`")
	expectError(t, "_test/src/binary", []string{"shift-err-3.go"},
		"shift count must be unsigned integer (`float32` given)")
	expectError(t, "_test/src/binary", []string{"shift-err-4.go"},
		"shift count too big: 512")
	expectError(t, "_test/src/binary", []string{"shift-err-5.go"},
		"invalid operand to `<<`: `1.100000`")
	expectError(t, "_test/src/binary", []string{"shift-err-6.go"},
		"invalid operand to `>>`: `(0 + 1.1i)`")
	expectError(t, "_test/src/binary", []string{"shift-err-7.go"},
		"invalid operand to `<<`: operand must have integer type (`untyped bool` given)")
	expectError(t, "_test/src/binary", []string{"shift-err-8.go"},
		"invalid operand")
	expectError(t, "_test/src/binary", []string{"shift-err-9.go"},
		"invalid operand to `<<`: operand must have integer type (`untyped string` given")
	expectError(t, "_test/src/binary", []string{"shift-err-10.go"},
		"invalid operand to `<<`")
	expectError(t, "_test/src/binary", []string{"shift-err-11.go"},
		"invalid operand to `<<`")
	expectError(t, "_test/src/binary", []string{"shift-err-12.go"},
		"shift count must be unsigned and integer")
	expectError(t, "_test/src/binary", []string{"shift-err-13.go"},
		"shift count must be unsigned and integer")
}

func TestCompareExpr(t *testing.T) {
	p, err := compilePackage("_test/src/binary", []string{"cmp.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, cs := range []struct {
		name string
		exp  bool
	}{
		{"ab", false}, {"bb", false},
		{"aub", false}, {"bub", false},

		{"ci8", true}, {"di8", true}, {"ei8", false}, {"fi8", true}, {"gi8", false},
		{"hi8", true},

		{"cu8", false}, {"du8", true}, {"eu8", false}, {"fu8", true}, {"gu8", false},
		{"hu8", true},

		{"ci64", true}, {"di64", true}, {"ei64", false}, {"fi64", true}, {"gi64", false},
		{"hi64", true},

		{"af", false}, {"bf", true}, {"cf", false}, {"df", true}, {"ef", false},
		{"ff", true},

		{"as", false}, {"bs", true}, {"cs", false}, {"ds", true}, {"es", false},
		{"fs", true},

		{"aus", false}, {"bus", true}, {"cus", false}, {"dus", true}, {"eus", false},
		{"fus", true},

		{"cui", true}, {"dui", true}, {"eui", false}, {"fui", true}, {"gui", false},
		{"hui", true},

		{"cur", true}, {"dur", true}, {"eur", false}, {"fur", true}, {"gur", false},
		{"hur", true},

		{"cuf", true}, {"duf", true}, {"euf", false}, {"fuf", true}, {"guf", false},
		{"huf", true},

		{"cir", true}, {"dir", true}, {"eir", false}, {"fir", true}, {"gir", true},
		{"hir", true},

		{"cif", true}, {"dif", true}, {"eif", false}, {"fif", true}, {"gif", false},
		{"hif", true},

		{"cif", true}, {"dif", true}, {"eif", false}, {"fif", true}, {"gif", false},
		{"hif", true},
	} {
		a := p.Find(cs.name).(*ast.Const)
		c := a.Init.(*ast.ConstValue)
		if v, ok := c.Value.(ast.Bool); !ok || c.Typ != nil || bool(v) != cs.exp {
			t.Errorf("`%s` should be untyped boolean %v", cs.name, cs.exp)
		}
	}
}

func TestCompareErr(t *testing.T) {
	// FIXME
	// expectError(t, "_test/src/binary", []string{"cmp-err-01.go"},
	// 	"operation `==` not supported for `nil`")
	expectError(t, "_test/src/binary", []string{"cmp-err-02.go"},
		"mismatched types `int` and `uint`")
	expectError(t, "_test/src/binary", []string{"cmp-err-03.go"},
		"operation `>` not supported for `bool`")
	expectError(t, "_test/src/binary", []string{"cmp-err-04.go"},
		"mismatched types `untyped bool` and `untyped int`")
	expectError(t, "_test/src/binary", []string{"cmp-err-05.go"},
		"operation `>` not supported for `untyped bool`")
	expectError(t, "_test/src/binary", []string{"cmp-err-06.go"},
		"mismatched types `untyped string` and `untyped float`")
}
