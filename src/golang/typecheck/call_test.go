package typecheck

import (
	"golang/ast"
	"golang/parser"
	"testing"
)

func TestSideEffects(t *testing.T) {
	p, err := parser.ParsePackage("_test/src/call", []string{"side-effects.go"})
	if err != nil {
		t.Fatal(err)
	}
	err = parser.ResolvePackage(p, nil)
	if err != nil {
		t.Fatal(err)
	}

	ti := newInferer(p)
	for _, cs := range []struct {
		name string
		eff  bool
	}{
		{"A", true}, {"B", false}, {"C", true}, {"D", false}, {"E", true},
		{"F", false}, {"G", true}, {"H", false}, {"I", true}, {"J", false},
		{"K", true}, {"L", true}, {"M", true}, {"N", false}, {"O", false},
		{"P", false}, {"Q", false}, {"R", false}, {"S", false}, {"T", true},
		{"U", true}, {"V", true}, {"W", true}, {"X", true}, {"Y", false},
		{"Z", true},

		{"Cr", true}, {"Gr", true}, {"Ir", true}, {"Kr", true}, {"Lr", true},
		{"Mr", true}, {"Tr", true}, {"Ur", true}, {"Vr", true}, {"Wr", true},

		{"AB", true}, {"AC", true}, {"AD", false}, {"AE", true}, {"AF", true},
		{"AG", true}, {"AH", true},

		{"L0", false}, {"L1", true}, {"L2", true},
		{"C0", false}, {"C1", true}, {"C2", true},
	} {
		v := p.Find(cs.name).(*ast.Var)
		if b := ti.hasSideEffects(v.Init.RHS[0]); b != cs.eff {
			t.Errorf("hasSideEffects(`%s`): unexpected result `%v`", cs.name, b)
		}
	}
}

func TestConstExpr(t *testing.T) {
	p, err := parser.ParsePackage("_test/src/call", []string{"const-expr.go"})
	if err != nil {
		t.Fatal(err)
	}
	err = parser.ResolvePackage(p, nil)
	if err != nil {
		t.Fatal(err)
	}

	ti := newInferer(p)
	for _, cs := range []struct {
		name string
		eff  bool
	}{
		{"A", true}, {"B", true}, {"C", false}, {"D", false}, {"E", false},
		{"F", false}, {"G", false}, {"H", true}, {"I", false}, {"J", false},
		{"K", false}, {"L", false}, {"M", false}, {"N", true}, {"O", false},
		{"P", false}, {"Q", false}, {"R", true}, {"S", true}, {"T", true},
		{"U", false}, {"V", false}, {"W", true}, {"X", false}, {"Y", false},
		{"AA", false}, {"AB", true}, {"AD", true}, {"AE", false}, {"AF", false},
		{"AG", false}, {"AH", true}, {"AI", false}, {"AJ", false},

		{"C0", true}, {"C1", false}, {"C2", false},

		{"IM0", false}, {"IM1", true}, {"IM2", false}, {"RE0", false}, {"RE1", true},
		{"RE2", false},

		{"S0", true}, {"S1", false}, {"S2", false}, {"S3", false}, {"S4", false},
		{"S5", true}, {"S6", false},

		{"F0", false}, {"F1", false},
	} {
		v := p.Find(cs.name).(*ast.Var)
		if b := ti.isConst(v.Init.RHS[0]); b != cs.eff {
			t.Errorf("isConst(`%s`): unexpected result `%v`", cs.name, b)
		}
	}
}

func TestCall(t *testing.T) {
	_, err := compilePackage("_test/src/call", []string{"call.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCallErr(t *testing.T) {
	expectError(t, "_test/src/call", []string{"call-err-01.go"},
		"argument count mismatch")
	expectError(t, "_test/src/call", []string{"call-err-02.go"},
		"argument count mismatch")
	expectError(t, "_test/src/call", []string{"call-err-03.go"},
		"argument count mismatch")
	expectError(t, "_test/src/call", []string{"call-err-04.go"},
		"argument count mismatch")
	expectError(t, "_test/src/call", []string{"call-err-05.go"},
		"argument count mismatch")
	expectError(t, "_test/src/call", []string{"call-err-06.go"},
		"argument count mismatch")
	expectError(t, "_test/src/call", []string{"call-err-07.go"},
		"argument count mismatch")
	expectError(t, "_test/src/call", []string{"call-err-08.go"},
		"argument count mismatch")
	expectError(t, "_test/src/call", []string{"call-err-09.go"},
		"invalid use of `...` to call a non-variadic function")
	expectError(t, "_test/src/call", []string{"call-err-10.go"},
		"argument count mismatch")
	expectError(t, "_test/src/call", []string{"call-err-11.go"},
		"argument count mismatch")
	expectError(t, "_test/src/call", []string{"call-err-12.go"},
		"1.2 (`untyped float`) cannot be converted to `int`")
	expectError(t, "_test/src/call", []string{"call-err-13.go"},
		"`float32` is not assignable to `int`")
	expectError(t, "_test/src/call", []string{"call-err-14.go"},
		"128 (`untyped int`) cannot be converted to `int8`")
	expectError(t, "_test/src/call", []string{"call-err-15.go"},
		"`float32` is not assignable to `float64`")
	expectError(t, "_test/src/call", []string{"call-err-16.go"},
		"`float32` is not assignable to `float64`")
	expectError(t, "_test/src/call", []string{"call-err-17.go"},
		"`float32` is not assignable to `float64`")
	expectError(t, "_test/src/call", []string{"call-err-18.go"},
		"`float32` is not assignable to `float64`")
	expectError(t, "_test/src/call", []string{"call-err-19.go"},
		"`[]float32` is not assignable to `[]float64`")
	expectError(t, "_test/src/call", []string{"call-err-20.go"},
		"invalid use of `...` with a non-slice argument")
	expectError(t, "_test/src/call", []string{"call-err-21.go"},
		"called object is not a function")
	expectError(t, "_test/src/call", []string{"call-err-22.go"},
		"type argument not allowed")
	expectError(t, "_test/src/call", []string{"call-err-23.go"},
		"`void` is not assignable to `int`")
	expectError(t, "_test/src/call", []string{"call-err-24.go"},
		"multiple value expression in single-value context")
	expectError(t, "_test/src/call", []string{"call-err-25.go"},
		"argument count mismatch")
	expectError(t, "_test/src/call", []string{"call-err-26.go"},
		"`float32` is not assignable to `J`")
	expectError(t, "_test/src/call", []string{"call-err-27.go"},
		"1.1i (`untyped complex`) cannot be converted to `J`")
}

func TestBuiltinMake(t *testing.T) {
	p, err := compilePackage("_test/src/call", []string{"make.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"a0", "a1", "a2"} {
		v := p.Find(name).(*ast.Var)
		if s, ok := v.Type.(*ast.SliceType); !ok || s.Elt != ast.BuiltinInt {
			t.Errorf("`%s` should have type `[]int`", name)
		}
	}

	SInt := p.Find("SInt").(*ast.TypeDecl)
	x := p.Find("x").(*ast.Var)
	if x.Type != SInt {
		t.Error("`x` should have type, `SInt`")
	}

	for _, name := range []string{"b0", "b1"} {
		v := p.Find(name).(*ast.Var)
		if s, ok := v.Type.(*ast.MapType); !ok || s.Key != ast.BuiltinString ||
			s.Elt != ast.BuiltinInt {
			t.Errorf("`%s` should have type `map[string]int`", name)
		}
	}

	for _, name := range []string{"c0", "c1"} {
		v := p.Find(name).(*ast.Var)
		if s, ok := v.Type.(*ast.ChanType); !ok || s.Elt != ast.BuiltinInt {
			t.Errorf("`%s` should have type `chan int`", name)
		}
	}
}

func TestBuiltinMakeErr(t *testing.T) {
	expectError(t, "_test/src/call", []string{"make-err-01.go"},
		"the first argument to `make` must be a slice, map, or chan type")
	expectError(t, "_test/src/call", []string{"make-err-02.go"},
		"the first argument to `make` must be a slice, map, or chan type")
	expectError(t, "_test/src/call", []string{"make-err-03.go"},
		"1.1 (`untyped float`) cannot be converted to `int`")
	expectError(t, "_test/src/call", []string{"make-err-04.go"},
		"`make` argument must be of integer type (given `float32`)")
	expectError(t, "_test/src/call", []string{"make-err-05.go"},
		"`make` argument must be of integer type (given `float32`)")
	expectError(t, "_test/src/call", []string{"make-err-06.go"},
		"`make` argument must be of integer type (given `float32`)")
	expectError(t, "_test/src/call", []string{"make-err-07.go"},
		"argument count mismatch")
	expectError(t, "_test/src/call", []string{"make-err-08.go"},
		"length exceeds capacity in a call to `make`")
	expectError(t, "_test/src/call", []string{"make-err-09.go"},
		"argument count mismatch")
	expectError(t, "_test/src/call", []string{"make-err-10.go"},
		"argument count mismatch")
	expectError(t, "_test/src/call", []string{"make-err-11.go"},
		"1.2 (`untyped float`) cannot be converted to `int`")
	expectError(t, "_test/src/call", []string{"make-err-12.go"},
		"`make` argument must be non-negative")
}

func TestBuiltinAppend(t *testing.T) {
	p, err := compilePackage("_test/src/call", []string{"append.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"a", "b"} {
		v := p.Find(name).(*ast.Var)
		if s, ok := v.Type.(*ast.SliceType); !ok || s.Elt != ast.BuiltinInt {
			t.Errorf("`%s` should have type `[]int`", name)
		}
	}

	SInt := p.Find("SInt").(*ast.TypeDecl)
	v := p.Find("c").(*ast.Var)
	if v.Type != SInt {
		t.Error("`c` should have type `SInt`")
	}

	for _, name := range []string{"d", "e", "f"} {
		v := p.Find(name).(*ast.Var)
		if s, ok := v.Type.(*ast.SliceType); !ok || s.Elt != ast.BuiltinUint8 {
			t.Errorf("`%s` should have type `[]byte`", name)
		}
	}
}

func TestBuiltinAppendErr(t *testing.T) {
	expectError(t, "_test/src/call", []string{"append-err-01.go"},
		"argument count mismatch")
	expectError(t, "_test/src/call", []string{"append-err-02.go"},
		"the first argument to `append` must be of a slice type")
	expectError(t, "_test/src/call", []string{"append-err-03.go"},
		"the first argument to `append` must be of a slice type")
	expectError(t, "_test/src/call", []string{"append-err-04.go"},
		"the first argument to `append` must be of a slice type")
	expectError(t, "_test/src/call", []string{"append-err-05.go"},
		"1.1 (`untyped float`) cannot be converted to `int`")
	expectError(t, "_test/src/call", []string{"append-err-06.go"},
		"argument count mismatch")
	expectError(t, "_test/src/call", []string{"append-err-07.go"},
		"argument count mismatch")
	expectError(t, "_test/src/call", []string{"append-err-08.go"},
		"1.2 (`untyped float`) cannot be converted to `int`")
	expectError(t, "_test/src/call", []string{"append-err-09.go"},
		"`[]uint` is not assignable to `[]int`")
	expectError(t, "_test/src/call", []string{"append-err-10.go"},
		"type argument not allowed")
	expectError(t, "_test/src/call", []string{"append-err-11.go"},
		"`string` is not assignable to `[]string`")
}

func TestBuiltinCap(t *testing.T) {
	p, err := compilePackage("_test/src/call", []string{"cap.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"c0", "c1", "c4", "c5"} {
		v := p.Find(name).(*ast.Var)
		if v.Type != ast.BuiltinInt {
			t.Errorf("`%s` should have type `int`", name)
		}
		c, ok := v.Init.RHS[0].(*ast.ConstValue)
		if !ok {
			t.Fatalf("initializer of `%s` should be a constant expression", name)
		}
		if i, ok := c.Value.(ast.Int); !ok || i != 3 {
			t.Errorf("`%s` should ghave value `3`", name)
		}
	}

	for _, name := range []string{"c2", "c3", "c6", "c7"} {
		v := p.Find(name).(*ast.Var)
		if v.Type != ast.BuiltinInt {
			t.Errorf("`%s` should have type `int`", name)
		}
		_, ok := v.Init.RHS[0].(*ast.ConstValue)
		if ok {
			t.Fatalf("initializer of `%s` should not be a constant expression", name)
		}
	}
}

func TestBuiltinCapErr(t *testing.T) {
	expectError(t, "_test/src/call", []string{"cap-err-01.go"},
		" type argument not allowed")
	expectError(t, "_test/src/call", []string{"cap-err-02.go"},
		"argument count mismatch")
	expectError(t, "_test/src/call", []string{"cap-err-03.go"},
		"`int` is invalid parameter type to the builtin `cap` function")
	expectError(t, "_test/src/call", []string{"cap-err-04.go"},
		"`*int` is invalid parameter type to the builtin `cap` function")
}
