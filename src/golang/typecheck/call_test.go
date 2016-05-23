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
		{"A", true},
		{"B", false},
		{"C", true},
		{"D", false},
		{"E", true},
		{"F", false},
		{"G", true},
		{"H", false},
		{"I", true},
		{"J", false},
		{"K", true},
		{"L", true},
		{"M", true},
		{"N", false},
		{"O", false},
		{"P", false},
		{"Q", false},
		{"R", false},
		{"S", false},
		{"T", true},
		{"U", true},
		{"V", true},
		{"W", true},
		{"X", true},
		{"Y", false},
		{"Z", true},

		// Postponed until issue #71 fixed
		// {"Cr", true},
		// {"Dr", true},
		// {"Gr", true},
		// {"Ir", true},
		// {"Kr", true},
		// {"Lr", true},
		// {"Mr", true},
		// {"Tr", true},
		// {"Ur", true},
		// {"Vr", true},
		// {"Wr", true},

		{"AB", true},
		{"AC", true},
		{"AD", false},
		{"AE", true},
		{"AF", true},
		{"AG", true},
		{"AH", true},

		{"L0", false},
		{"L1", true},
	} {
		v := p.Find(cs.name).(*ast.Var)
		if b := ti.hasSideEffects(v.Init.RHS[0]); b != cs.eff {
			t.Errorf("hasSideEffects(`%s`): unexpected result `%v`", cs.name, b)
		}
	}
}
