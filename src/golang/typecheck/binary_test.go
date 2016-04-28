package typecheck

import (
	"fmt"
	"golang/ast"
	"math/big"
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

	V = p.Find("O").(*ast.Var)
	xx := V.Init.RHS[0].(*ast.Conversion).X.(*ast.BinaryExpr)
	if xx.Type() != ast.BuiltinUint16 || xx.X.Type() != ast.BuiltinUint16 ||
		xx.Y.Type() != ast.BuiltinUint16 {
		t.Error("in initializer of `O`: subexpressions must have type `uint16`")
	}
	xx = xx.X.(*ast.BinaryExpr)
	if xx.Type() != ast.BuiltinUint16 ||
		xx.X.Type() != ast.BuiltinUint16 ||
		xx.Y.Type() != ast.BuiltinUint16 {
		t.Error("in initializer of `O`: subexpressions must have type `uint16`")
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

	// FIXME: temporarily disabled until assignability checks eveywhere
	// V = p.Find("l0").(*ast.Var)
	// lit := V.Init.RHS[0].(*ast.CompLiteral)
	// c = lit.Elts[0].Elt.(*ast.ConstValue)
	// if c.Typ != ast.BuiltinUint8 {
	// 	t.Error("`l0` element initializer should should have type `uint8`")
	// }

	// V = p.Find("l1").(*ast.Var)
	// lit = V.Init.RHS[0].(*ast.CompLiteral)
	// c = lit.Elts[0].Key.(*ast.ConstValue)
	// if c.Typ != ast.BuiltinUint16 {
	// 	t.Error("`l1` key initializer should have type `uint16`")
	// }
	// c = lit.Elts[0].Elt.(*ast.ConstValue)
	// if c.Typ != ast.BuiltinUint8 {
	// 	t.Error("`l1` element initializer should have type `uint8`")
	// }

}

func TestShiftErr(t *testing.T) {
	expectError(t, "_test/src/binary", []string{"shift-err-1.go"},
		"shift count must be unsigned integer (`int` given)")
	expectError(t, "_test/src/binary", []string{"shift-err-2.go"},
		"1.1 (`untyped float`) cannot be converted to `uint64`")
	expectError(t, "_test/src/binary", []string{"shift-err-3.go"},
		"shift count must be unsigned integer (`float32` given)")
	expectError(t, "_test/src/binary", []string{"shift-err-4.go"},
		"shift count too big: 512")
	expectError(t, "_test/src/binary", []string{"shift-err-5.go"},
		"invalid operand to `<<`: `1.1`")
	expectError(t, "_test/src/binary", []string{"shift-err-6.go"},
		"invalid operand to `>>`: `1.1i`")
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
		if v, ok := c.Value.(ast.Bool); !ok || c.Typ != ast.BuiltinUntypedBool ||
			bool(v) != cs.exp {
			t.Errorf("`%s` should be untyped boolean %v", cs.name, cs.exp)
		}
	}

	for i := 0; i <= 11; i++ {
		a := p.Find(fmt.Sprintf("x%d", i)).(*ast.Var)
		if a.Type != ast.BuiltinBool {
			t.Errorf("`x%d` should have type `bool`", i)
		}
	}
}

func TestCompareErr(t *testing.T) {
	expectError(t, "_test/src/binary", []string{"cmp-err-01.go"},
		"operation `==` not supported for `nil`")
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
	expectError(t, "_test/src/binary", []string{"cmp-err-07.go"},
		"neither `int` nor `int32` is assignable to the other")
	expectError(t, "_test/src/binary", []string{"cmp-err-08.go"},
		"neither `int` nor `nil` is assignable to the other")
	expectError(t, "_test/src/binary", []string{"cmp-err-09.go"},
		"values of type `S` cannot be compared for equality")
	expectError(t, "_test/src/binary", []string{"cmp-err-10.go"},
		"values of type `S` cannot be compared for equality")
	expectError(t, "_test/src/binary", []string{"cmp-err-11.go"},
		"neither `S` nor `I` is assignable to the other")
	expectError(t, "_test/src/binary", []string{"cmp-err-12.go"},
		"neither `I` nor `J` is assignable to the other")
	expectError(t, "_test/src/binary", []string{"cmp-err-13.go"},
		"values of type `S` cannot be compared for equality")
	expectError(t, "_test/src/binary", []string{"cmp-err-14.go"},
		"values of type `bool` are not ordered")
	expectError(t, "_test/src/binary", []string{"cmp-err-15.go"},
		"values of type `Complex` are not ordered")
	expectError(t, "_test/src/binary", []string{"cmp-err-16.go"},
		"values of type `*int` are not ordered")
	expectError(t, "_test/src/binary", []string{"cmp-err-17.go"},
		"values of type `chan int` are not ordered")
	expectError(t, "_test/src/binary", []string{"cmp-err-18.go"},
		"values of type `interface{...}` are not ordered")
	expectError(t, "_test/src/binary", []string{"cmp-err-19.go"},
		"values of type `S` are not ordered")
	expectError(t, "_test/src/binary", []string{"cmp-err-20.go"},
		"values of type `Vec` are not ordered")
	expectError(t, "_test/src/binary", []string{"cmp-err-21.go"},
		"neither `Float` nor `float32` is assignable to the other")
	expectError(t, "_test/src/binary", []string{"cmp-err-22.go"},
		"invalid operation `<=`: mismatched types `untyped float` and `untyped string`")
}

func TestBinaryErr(t *testing.T) {
	expectError(t, "_test/src/binary", []string{"bin-err-01.go"},
		"invalid operation `+`: mismatched types `int` and `int64`")
	expectError(t, "_test/src/binary", []string{"bin-err-02.go"},
		"invalid operation `+`: mismatched types `int` and `Int`")
	expectError(t, "_test/src/binary", []string{"bin-err-03.go"},
		"invalid operation `+`: mismatched types `Int` and `Int32`")
}

func TestAdd(t *testing.T) {
	p, err := compilePackage("_test/src/binary", []string{"add.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	Int := p.Find("Int").(*ast.TypeDecl)

	for _, cs := range []struct {
		name string
		typ  ast.Type
		val  ast.Value
	}{
		{"C", ast.BuiltinInt8, ast.Int(3)},
		{"E", ast.BuiltinUint16, ast.Int(4)},
		{"H", ast.BuiltinFloat64, ast.Float(3.375)},
		{"J", ast.BuiltinFloat32, ast.Float(float32(2.0000005))},
		{"L", ast.BuiltinFloat64, ast.Float(2.00000051)},
		{"SS", ast.BuiltinString, ast.String("abcαβγ")},
		{"V", ast.BuiltinComplex64, ast.Complex(complex(1.0, 0.125))},
		{"Y", ast.BuiltinComplex128, ast.Complex(complex(1.25, 0.125))},
		{"Z1", Int, ast.Int(7)},
		{"Z2", Int, ast.Int(7)},
		{"Z3", Int, ast.Int(10)},
	} {
		c := p.Find(cs.name).(*ast.Const)
		if c.Type != cs.typ {
			t.Errorf("`%s` must have type `%s`\n", c.Name, cs.typ)
		}
		if v := c.Init.(*ast.ConstValue).Value; v != cs.val {
			t.Errorf("`%s` must have value `%v`\n", c.Name, cs.val)
		}
	}

	// int, int
	for _, cs := range []struct {
		name  string
		value int64
	}{
		{"N", 2},
		{"D12", 2},
	} {
		c := p.Find(cs.name).(*ast.Const)
		if c.Type != ast.BuiltinUntypedInt {
			t.Errorf("`%s` must have type `%s`\n", c.Name, ast.BuiltinUntypedInt)
		}
		v := c.Init.(*ast.ConstValue).Value
		if v, ok := v.(ast.UntypedInt); !ok || v.Int64() != cs.value {
			t.Errorf("`%s` must have value `%v`\n", c.Name, cs.value)
		}
	}

	// int, rune
	// rune, rune
	for _, cs := range []struct {
		name  string
		value int64
	}{
		{"D0", 'b'},
		{"D1", 'b'},
		{"D13", 'a' + 'a'},
	} {
		c := p.Find(cs.name).(*ast.Const)
		if c.Type != ast.BuiltinUntypedRune {
			t.Errorf("`%s` must have type `%s`\n", c.Name, ast.BuiltinUntypedRune)
		}
		v := c.Init.(*ast.ConstValue).Value
		if v, ok := v.(ast.Rune); !ok || v.Int64() != cs.value {
			t.Errorf("`%s` must have value `%v`\n", c.Name, cs.value)
		}
	}

	// int, float
	// rune, float
	// float, float
	for _, cs := range []struct {
		name  string
		value float64
	}{
		{"D2", 2.125},
		{"D3", 2.125},
		{"D6", 98.125},
		{"D7", 98.125},
		{"D14", 2.25},
	} {
		c := p.Find(cs.name).(*ast.Const)
		if c.Type != ast.BuiltinUntypedFloat {
			t.Errorf("`%s` must have type `%s`\n", c.Name, ast.BuiltinUntypedFloat)
		}
		v := c.Init.(*ast.ConstValue).Value
		u, acc := v.(ast.UntypedFloat).Float.Float64()
		if acc != big.Exact || u != cs.value {
			t.Errorf("`%s` must have value `%v`\n", c.Name, cs.value)
		}
	}

	// int, complex
	// rune, complex
	// float, complex
	// complex, complex
	for _, cs := range []struct {
		name string
		re   float64
		im   float64
	}{
		{"D4", 1.0, 1.25},
		{"D5", 1.0, 1.25},
		{"D8", 97.0, 1.25},
		{"D9", 97.0, 1.25},
		{"D10", 1.125, 1.25},
		{"D11", 1.125, 1.25},
		{"D15", 0.0, 2.5},
	} {
		c := p.Find(cs.name).(*ast.Const)
		if c.Type != ast.BuiltinUntypedComplex {
			t.Errorf("`%s` must have type `%s`\n", c.Name, ast.BuiltinUntypedComplex)
		}
		v := c.Init.(*ast.ConstValue).Value
		u, acc := v.(ast.UntypedComplex).Re.Float64()
		if acc != big.Exact || u != cs.re {
			t.Errorf("`real(%s)` must have value `%v`\n", cs.name, cs.re)
		}
		u, acc = v.(ast.UntypedComplex).Im.Float64()
		if acc != big.Exact || u != cs.im {
			t.Errorf("`imag(%s)` must have value `%v`\n", cs.name, cs.im)
		}
	}

	D := p.Find("D16").(*ast.Const)
	if D.Type != ast.BuiltinUntypedString {
		t.Error("`D16` must be `untyped string`")
	}
	v := D.Init.(*ast.ConstValue).Value
	if v := v.(ast.String); v != ast.String("aβγxyζ") {
		t.Error("`D16` must have value \"aβγxyζ\"")
	}

	for _, name := range []string{"a", "b", "c", "d"} {
		v := p.Find(name).(*ast.Var)
		if v.Type != Int {
			t.Errorf("`%s` should have type `Int`", name)
		}
	}

	ss := p.Find("ss").(*ast.Var)
	if ss.Type != ast.BuiltinString {
		t.Error("`ss` should have type `string`")
	}
}

func TestAddErr(t *testing.T) {
	expectError(t, "_test/src/binary", []string{"add-err-01.go"},
		"invalid operation `+`: mismatched types `int8` and `uint`")
	expectError(t, "_test/src/binary", []string{"add-err-02.go"},
		"operation `+` not supported for `bool`")
	expectError(t, "_test/src/binary", []string{"add-err-03.go"},
		"operation `+` not supported for `untyped bool`")
	expectError(t, "_test/src/binary", []string{"add-err-04.go"},
		"invalid operation `+`: mismatched types `untyped string` and `untyped int`")
	expectError(t, "_test/src/binary", []string{"add-err-05.go"},
		"invalid operation `+`: mismatched types `untyped int` and `untyped string`")
	expectError(t, "_test/src/binary", []string{"add-err-06.go"},
		"operation `+` not supported for `*int`")
	expectError(t, "_test/src/binary", []string{"add-err-07.go"},
		"operation `+` not supported for `nil`")
	expectError(t, "_test/src/binary", []string{"add-err-08.go"},
		"operation `+` not supported for `bool`")
	expectError(t, "_test/src/binary", []string{"add-err-09.go"},
		"invalid operation `+`: mismatched types `int` and `int64`")
	expectError(t, "_test/src/binary", []string{"add-err-10.go"},
		"invalid operation `+`: mismatched types `*int` and `[]int`")
}

func TestSub(t *testing.T) {
	p, err := compilePackage("_test/src/binary", []string{"sub.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	Int := p.Find("Int").(*ast.TypeDecl)

	for _, cs := range []struct {
		name string
		typ  ast.Type
		val  ast.Value
	}{
		{"C", ast.BuiltinInt8, ast.Int(0xffffffffffffffff)},
		{"E", ast.BuiltinUint16, ast.Int(0xfffffffffffffffe)},
		{"H", ast.BuiltinFloat64, ast.Float(-1.125)},
		{"J", ast.BuiltinFloat32, ast.Float(float32(1.0000001))},
		{"L", ast.BuiltinFloat64, ast.Float(1.0000000900000001)},
		{"V", ast.BuiltinComplex64, ast.Complex(complex(-1.0, 0.125))},
		{"Y", ast.BuiltinComplex128, ast.Complex(complex(-1.25, 0.125))},
		{"Z1", Int, ast.Int(0xffffffffffffffff)},
		{"Z2", Int, ast.Int(1)},
		{"Z3", Int, ast.Int(4)},
	} {
		c := p.Find(cs.name).(*ast.Const)
		if c.Type != cs.typ {
			t.Errorf("`%s` must have type `%s`\n", c.Name, cs.typ)
		}
		if v := c.Init.(*ast.ConstValue).Value; v != cs.val {
			t.Errorf("`%s` must have value `%v`\n", c.Name, cs.val)
		}
	}

	// int, int
	for _, cs := range []struct {
		name  string
		value int64
	}{
		{"N", 0},
		{"D12", 0},
	} {
		c := p.Find(cs.name).(*ast.Const)
		if c.Type != ast.BuiltinUntypedInt {
			t.Errorf("`%s` must have type `%s`\n", c.Name, ast.BuiltinUntypedInt)
		}
		v := c.Init.(*ast.ConstValue).Value
		if v, ok := v.(ast.UntypedInt); !ok || v.Int64() != cs.value {
			t.Errorf("`%s` must have value `%v`\n", c.Name, cs.value)
		}
	}

	// int, rune
	// rune, rune
	for _, cs := range []struct {
		name  string
		value int64
	}{
		{"D0", -96},
		{"D1", 96},
		{"D13", 0},
	} {
		c := p.Find(cs.name).(*ast.Const)
		if c.Type != ast.BuiltinUntypedRune {
			t.Errorf("`%s` must have type `%s`\n", c.Name, ast.BuiltinUntypedRune)
		}
		v := c.Init.(*ast.ConstValue).Value
		if v, ok := v.(ast.Rune); !ok || v.Int64() != cs.value {
			t.Errorf("`%s` must have value `%v`\n", c.Name, cs.value)
		}
	}

	// int, float
	// rune, float
	// float, float
	for _, cs := range []struct {
		name  string
		value float64
	}{
		{"D2", -0.125},
		{"D3", 0.125},
		{"D6", 95.875},
		{"D7", -95.875},
		{"D14", 0},
	} {
		c := p.Find(cs.name).(*ast.Const)
		if c.Type != ast.BuiltinUntypedFloat {
			t.Errorf("`%s` must have type `%s`\n", c.Name, ast.BuiltinUntypedFloat)
		}
		v := c.Init.(*ast.ConstValue).Value
		u, acc := v.(ast.UntypedFloat).Float.Float64()
		if acc != big.Exact || u != cs.value {
			t.Errorf("`%s` must have value `%v`\n", c.Name, cs.value)
		}
	}

	// int, complex
	// rune, complex
	// float, complex
	// complex, complex
	for _, cs := range []struct {
		name string
		re   float64
		im   float64
	}{
		{"D4", 1.0, -1.25},
		{"D5", -1.0, 1.25},
		{"D8", 97.0, -1.25},
		{"D9", -97.0, 1.25},
		{"D10", 1.125, -1.25},
		{"D11", -1.125, 1.25},
		{"D15", 0.0, 0.0},
	} {
		c := p.Find(cs.name).(*ast.Const)
		if c.Type != ast.BuiltinUntypedComplex {
			t.Errorf("`%s` must have type `%s`\n", c.Name, ast.BuiltinUntypedComplex)
		}
		v := c.Init.(*ast.ConstValue).Value
		u, acc := v.(ast.UntypedComplex).Re.Float64()
		if acc != big.Exact || u != cs.re {
			t.Errorf("`real(%s)` must have value `%v`\n", cs.name, cs.re)
		}
		u, acc = v.(ast.UntypedComplex).Im.Float64()
		if acc != big.Exact || u != cs.im {
			t.Errorf("`imag(%s)` must have value `%v`\n", cs.name, cs.im)
		}
	}

	for _, name := range []string{"a", "b", "c", "d"} {
		v := p.Find(name).(*ast.Var)
		if v.Type != Int {
			t.Errorf("`%s` should have type `Int`", name)
		}
	}
}

func TestMul(t *testing.T) {
	p, err := compilePackage("_test/src/binary", []string{"mul.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	Int := p.Find("Int").(*ast.TypeDecl)

	for _, cs := range []struct {
		name string
		typ  ast.Type
		val  ast.Value
	}{
		{"C", ast.BuiltinInt8, ast.Int(2)},
		{"E", ast.BuiltinUint16, ast.Int(3)},
		{"H", ast.BuiltinFloat64, ast.Float(2.53125)},
		{"J", ast.BuiltinFloat32, ast.Float(float32(0.50000006))},
		{"L", ast.BuiltinFloat64, ast.Float(0.50000005)},
		{"V", ast.BuiltinComplex64, ast.Complex(complex(0.0, 0.15625))},
		{"Y", ast.BuiltinComplex128, ast.Complex(complex(0.0, 0.140625))},
		{"Z1", Int, ast.Int(12)},
		{"Z2", Int, ast.Int(12)},
		{"Z3", Int, ast.Int(36)},
	} {
		c := p.Find(cs.name).(*ast.Const)
		if c.Type != cs.typ {
			t.Errorf("`%s` must have type `%s`\n", c.Name, cs.typ)
		}
		if v := c.Init.(*ast.ConstValue).Value; v != cs.val {
			t.Errorf("`%s` must have value `%v`\n", c.Name, cs.val)
		}
	}

	// int, int
	for _, cs := range []struct {
		name  string
		value int64
	}{
		{"N", 1},
		{"D12", 1},
	} {
		c := p.Find(cs.name).(*ast.Const)
		if c.Type != ast.BuiltinUntypedInt {
			t.Errorf("`%s` must have type `%s`\n", c.Name, ast.BuiltinUntypedInt)
		}
		v := c.Init.(*ast.ConstValue).Value
		if v, ok := v.(ast.UntypedInt); !ok || v.Int64() != cs.value {
			t.Errorf("`%s` must have value `%v`\n", c.Name, cs.value)
		}
	}

	// int, rune
	// rune, rune
	for _, cs := range []struct {
		name  string
		value int64
	}{
		{"D0", 97},
		{"D1", 97},
		{"D13", 9409},
	} {
		c := p.Find(cs.name).(*ast.Const)
		if c.Type != ast.BuiltinUntypedRune {
			t.Errorf("`%s` must have type `%s`\n", c.Name, ast.BuiltinUntypedRune)
		}
		v := c.Init.(*ast.ConstValue).Value
		if v, ok := v.(ast.Rune); !ok || v.Int64() != cs.value {
			t.Errorf("`%s` must have value `%v`\n", c.Name, cs.value)
		}
	}

	// int, float
	// rune, float
	// float, float
	for _, cs := range []struct {
		name  string
		value float64
	}{
		{"D2", 1.125},
		{"D3", 1.125},
		{"D6", 109.125},
		{"D7", 109.125},
		{"D14", 1.265625},
	} {
		c := p.Find(cs.name).(*ast.Const)
		if c.Type != ast.BuiltinUntypedFloat {
			t.Errorf("`%s` must have type `%s`\n", c.Name, ast.BuiltinUntypedFloat)
		}
		v := c.Init.(*ast.ConstValue).Value
		u, acc := v.(ast.UntypedFloat).Float.Float64()
		if acc != big.Exact || u != cs.value {
			t.Errorf("`%s` must have value `%v`\n", c.Name, cs.value)
		}
	}

	// int, complex
	// rune, complex
	// float, complex
	// complex, complex
	for _, cs := range []struct {
		name string
		re   float64
		im   float64
	}{
		{"D4", 0.0, 1.25},
		{"D5", 0.0, 1.25},
		{"D8", 0.0, 121.25},
		{"D9", 0.0, 121.25},
		{"D10", 0.0, 1.40625},
		{"D11", 0.0, 1.40625},
		{"D15", -1.5625, 0.0},
	} {
		c := p.Find(cs.name).(*ast.Const)
		if c.Type != ast.BuiltinUntypedComplex {
			t.Errorf("`%s` must have type `%s`\n", c.Name, ast.BuiltinUntypedComplex)
		}
		v := c.Init.(*ast.ConstValue).Value
		u, acc := v.(ast.UntypedComplex).Re.Float64()
		if acc != big.Exact || u != cs.re {
			t.Errorf("`real(%s)` must have value `%v`\n", cs.name, cs.re)
		}
		u, acc = v.(ast.UntypedComplex).Im.Float64()
		if acc != big.Exact || u != cs.im {
			t.Errorf("`imag(%s)` must have value `%v`\n", cs.name, cs.im)
		}
	}

	for _, name := range []string{"a", "b", "c", "d"} {
		v := p.Find(name).(*ast.Var)
		if v.Type != Int {
			t.Errorf("`%s` should have type `Int`", name)
		}
	}
}

func TestSubErr(t *testing.T) {
	expectError(t, "_test/src/binary", []string{"sub-err-01.go"},
		"invalid operation `-`: mismatched types `int8` and `uint`")
	expectError(t, "_test/src/binary", []string{"sub-err-02.go"},
		"operation `-` not supported for `bool`")
	expectError(t, "_test/src/binary", []string{"sub-err-03.go"},
		"operation `-` not supported for `untyped bool`")
	expectError(t, "_test/src/binary", []string{"sub-err-04.go"},
		"operation `-` not supported for `untyped string`")
	expectError(t, "_test/src/binary", []string{"sub-err-05.go"},
		"invalid operation `-`: mismatched types `untyped int` and `untyped string`")
	expectError(t, "_test/src/binary", []string{"sub-err-06.go"},
		"operation `-` not supported for `*int`")
	expectError(t, "_test/src/binary", []string{"sub-err-07.go"},
		"operation `-` not supported for `nil`")
	expectError(t, "_test/src/binary", []string{"sub-err-08.go"},
		"operation `-` not supported for `bool`")
	expectError(t, "_test/src/binary", []string{"sub-err-09.go"},
		"invalid operation `-`: mismatched types `int` and `int64`")
	expectError(t, "_test/src/binary", []string{"sub-err-10.go"},
		"invalid operation `-`: mismatched types `*int` and `[]int`")
	expectError(t, "_test/src/binary", []string{"sub-err-11.go"},
		"operation `-` not supported for `string`")
	expectError(t, "_test/src/binary", []string{"sub-err-12.go"},
		"operation `-` not supported for `untyped string`")
	expectError(t, "_test/src/binary", []string{"sub-err-13.go"},
		"operation `-` not supported for `string`")
}

func TestMulErr(t *testing.T) {
	expectError(t, "_test/src/binary", []string{"mul-err-01.go"},
		"invalid operation `*`: mismatched types `int8` and `uint`")
	expectError(t, "_test/src/binary", []string{"mul-err-02.go"},
		"operation `*` not supported for `bool`")
	expectError(t, "_test/src/binary", []string{"mul-err-03.go"},
		"operation `*` not supported for `untyped bool`")
	expectError(t, "_test/src/binary", []string{"mul-err-04.go"},
		"operation `*` not supported for `untyped string`")
	expectError(t, "_test/src/binary", []string{"mul-err-05.go"},
		"invalid operation `*`: mismatched types `untyped int` and `untyped string`")
	expectError(t, "_test/src/binary", []string{"mul-err-06.go"},
		"operation `*` not supported for `*int`")
	expectError(t, "_test/src/binary", []string{"mul-err-07.go"},
		"operation `*` not supported for `nil`")
	expectError(t, "_test/src/binary", []string{"mul-err-08.go"},
		"operation `*` not supported for `bool`")
	expectError(t, "_test/src/binary", []string{"mul-err-09.go"},
		"invalid operation `*`: mismatched types `int` and `int64`")
	expectError(t, "_test/src/binary", []string{"mul-err-10.go"},
		"invalid operation `*`: mismatched types `*int` and `[]int`")
	expectError(t, "_test/src/binary", []string{"mul-err-11.go"},
		"operation `*` not supported for `string`")
	expectError(t, "_test/src/binary", []string{"mul-err-12.go"},
		"operation `*` not supported for `untyped string`")
	expectError(t, "_test/src/binary", []string{"mul-err-13.go"},
		"operation `*` not supported for `string`")
}

func TestDiv(t *testing.T) {
	p, err := compilePackage("_test/src/binary", []string{"div.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	Int := p.Find("Int").(*ast.TypeDecl)

	for _, cs := range []struct {
		name string
		typ  ast.Type
		val  ast.Value
	}{
		{"C", ast.BuiltinInt8, ast.Int(5)},
		{"E", ast.BuiltinUint16, ast.Int(3)},
		{"H", ast.BuiltinFloat64, ast.Float(0.5)},
		{"J", ast.BuiltinFloat32, ast.Float(float32(2.0000002))},
		{"L", ast.BuiltinFloat64, ast.Float(2.0000002)},
		{"V", ast.BuiltinComplex64, ast.Complex(complex(0.0, 0.5))},
		{"Y", ast.BuiltinComplex128, ast.Complex(complex(0.0, 0.5))},
		{"Z1", Int, ast.Int(3)},
		{"Z2", Int, ast.Int(0)},
		{"Z3", Int, ast.Int(4)},
	} {
		c := p.Find(cs.name).(*ast.Const)
		if c.Type != cs.typ {
			t.Errorf("`%s` must have type `%s`\n", c.Name, cs.typ)
		}
		if v := c.Init.(*ast.ConstValue).Value; v != cs.val {
			t.Errorf("`%s` must have value `%v`\n", c.Name, cs.val)
		}
	}

	// int, int
	for _, cs := range []struct {
		name  string
		value int64
	}{
		{"N", 5},
		{"D12", 1},
	} {
		c := p.Find(cs.name).(*ast.Const)
		if c.Type != ast.BuiltinUntypedInt {
			t.Errorf("`%s` must have type `%s`\n", c.Name, ast.BuiltinUntypedInt)
		}
		v := c.Init.(*ast.ConstValue).Value
		if v, ok := v.(ast.UntypedInt); !ok || v.Int64() != cs.value {
			t.Errorf("`%s` must have value `%v`\n", c.Name, cs.value)
		}
	}

	// int, rune
	// rune, rune
	for _, cs := range []struct {
		name  string
		value int64
	}{
		{"D0", 0},
		{"D1", 9},
		{"D13", 1},
	} {
		c := p.Find(cs.name).(*ast.Const)
		if c.Type != ast.BuiltinUntypedRune {
			t.Errorf("`%s` must have type `%s`\n", c.Name, ast.BuiltinUntypedRune)
		}
		v := c.Init.(*ast.ConstValue).Value
		if v, ok := v.(ast.Rune); !ok || v.Int64() != cs.value {
			t.Errorf("`%s` must have value `%v`\n", c.Name, cs.value)
		}
	}

	// int, float
	// rune, float
	// float, float
	for _, cs := range []struct {
		name  string
		value float64
	}{
		{"D2", 8.0},
		{"D3", 0.125},
		{"D6", 77.6},
		{"D7", 0.01288659793814433},
		{"D14", 1.0},
	} {
		c := p.Find(cs.name).(*ast.Const)
		if c.Type != ast.BuiltinUntypedFloat {
			t.Errorf("`%s` must have type `%s`\n", c.Name, ast.BuiltinUntypedFloat)
		}
		v := c.Init.(*ast.ConstValue).Value
		u, _ := v.(ast.UntypedFloat).Float.Float64()
		if u != cs.value {
			t.Errorf("`%s` must have value `%v`\n", c.Name, cs.value)
		}
	}

	// int, complex
	// rune, complex
	// float, complex
	// complex, complex
	for _, cs := range []struct {
		name string
		re   float64
		im   float64
	}{
		{"D4", 0.0, -8.0},
		{"D5", 0.0, 0.125},
		{"D8", 0.0, -77.6},
		{"D9", 0.0, 0.01288659793814433},
		{"D10", 0.0, -1.0},
		{"D11", 0.0, 1.0},
		{"D15", 1.0, 0.0},
	} {
		c := p.Find(cs.name).(*ast.Const)
		if c.Type != ast.BuiltinUntypedComplex {
			t.Errorf("`%s` must have type `%s`\n", c.Name, ast.BuiltinUntypedComplex)
		}
		v := c.Init.(*ast.ConstValue).Value
		u, _ := v.(ast.UntypedComplex).Re.Float64()
		if u != cs.re {
			t.Errorf("`real(%s)` must have value `%v`\n", cs.name, cs.re)
		}
		u, _ = v.(ast.UntypedComplex).Im.Float64()
		if u != cs.im {
			t.Errorf("`imag(%s)` must have value `%v`\n", cs.name, cs.im)
		}
	}

	for _, name := range []string{"a", "b", "c", "d"} {
		v := p.Find(name).(*ast.Var)
		if v.Type != Int {
			t.Errorf("`%s` should have type `Int`", name)
		}
	}
}

func TestDivErr(t *testing.T) {
	expectError(t, "_test/src/binary", []string{"div-err-01.go"},
		"invalid operation `/`: mismatched types `int8` and `uint`")
	expectError(t, "_test/src/binary", []string{"div-err-02.go"},
		"operation `/` not supported for `bool`")
	expectError(t, "_test/src/binary", []string{"div-err-03.go"},
		"operation `/` not supported for `untyped bool`")
	expectError(t, "_test/src/binary", []string{"div-err-04.go"},
		"operation `/` not supported for `untyped string`")
	expectError(t, "_test/src/binary", []string{"div-err-05.go"},
		"invalid operation `/`: mismatched types `untyped int` and `untyped string`")
	expectError(t, "_test/src/binary", []string{"div-err-06.go"},
		"operation `/` not supported for `*int`")
	expectError(t, "_test/src/binary", []string{"div-err-07.go"},
		"operation `/` not supported for `nil`")
	expectError(t, "_test/src/binary", []string{"div-err-08.go"},
		"operation `/` not supported for `bool`")
	expectError(t, "_test/src/binary", []string{"div-err-09.go"},
		"invalid operation `/`: mismatched types `int` and `int64`")
	expectError(t, "_test/src/binary", []string{"div-err-10.go"},
		"invalid operation `/`: mismatched types `*int` and `[]int`")
	expectError(t, "_test/src/binary", []string{"div-err-11.go"},
		"operation `/` not supported for `string`")
	expectError(t, "_test/src/binary", []string{"div-err-12.go"},
		"operation `/` not supported for `untyped string`")
	expectError(t, "_test/src/binary", []string{"div-err-13.go"},
		"operation `/` not supported for `string`")
	expectError(t, "_test/src/binary", []string{"div-err-14.go"},
		"division by zero")
	expectError(t, "_test/src/binary", []string{"div-err-15.go"},
		"division by zero")
	expectError(t, "_test/src/binary", []string{"div-err-16.go"},
		"division by zero")
	expectError(t, "_test/src/binary", []string{"div-err-17.go"},
		"division by zero")
	expectError(t, "_test/src/binary", []string{"div-err-18.go"},
		"division by zero")
	expectError(t, "_test/src/binary", []string{"div-err-19.go"},
		"division by zero")
	expectError(t, "_test/src/binary", []string{"div-err-20.go"},
		"division by zero")
	expectError(t, "_test/src/binary", []string{"div-err-21.go"},
		"division by zero")
	expectError(t, "_test/src/binary", []string{"div-err-22.go"},
		"division by zero")
}
