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
}
