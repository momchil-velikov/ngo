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
