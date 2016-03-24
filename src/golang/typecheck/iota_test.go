package typecheck

import (
	"golang/ast"
	"testing"
)

func TestIota(t *testing.T) {
	p, err := compilePackage("_test/src/iota", []string{"iota.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	A := p.Find("A").(*ast.Const)
	c := A.Init.(*ast.ConstValue)
	if v, ok := c.Value.(ast.UntypedInt); !ok || v.Int64() != 0 {
		t.Error("`A` should have value 0")
	}
	B := p.Find("B").(*ast.Const)
	c = B.Init.(*ast.ConstValue)
	if v, ok := c.Value.(ast.UntypedInt); !ok || v.Int64() != 1 {
		t.Error("`B` should have value 1")
	}

	D := p.Find("D").(*ast.Const)
	c = D.Init.(*ast.ConstValue)
	if c.Value.(ast.Int) != 0 {
		t.Error("`D` should have value 0")
	}
	E := p.Find("E").(*ast.Const)
	c = E.Init.(*ast.ConstValue)
	if c.Value.(ast.Int) != 1 {
		t.Error("`E` should have value 1")
	}

	C := p.Find("C").(*ast.Const)
	c = C.Init.(*ast.ConstValue)
	if c.Value.(ast.Int) != 2 {
		t.Error("`C` should have value 2")
	}
	F := p.Find("F").(*ast.Const)
	c = F.Init.(*ast.ConstValue)
	if c.Value.(ast.Int) != 2 {
		t.Error("`F` should have value 2")
	}
	G := p.Find("G").(*ast.Const)
	c = G.Init.(*ast.ConstValue)
	if c.Value.(ast.Int) != 0xffef {
		t.Error("`G` should have value 0xffef")
	}

	U := p.Find("U").(*ast.Const)
	c = U.Init.(*ast.ConstValue)
	if c.Value.(ast.Int) != 0xfff7 {
		t.Error("`U` should have value 0xfff7")
	}
	V := p.Find("V").(*ast.Const)
	c = V.Init.(*ast.ConstValue)
	if c.Value.(ast.Int) != 0xffef {
		t.Error("`V` should have value 0xffef")
	}

	N := p.Find("N").(*ast.Const)
	c = N.Init.(*ast.ConstValue)
	if N.Type != nil || c.Typ != nil {
		t.Error("`N` should be untyped")
	}
	if v, ok := c.Value.(ast.UntypedInt); !ok || v.Int64() != 0 {
		t.Error("`N` should have value 0")
	}
}

func TestIotaErr(t *testing.T) {
	expectError(t, "_test/src/iota", []string{"iota-err-01.go"},
		"`iota` used outside const declaration")
}
