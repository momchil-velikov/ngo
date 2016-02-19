package typecheck

import (
	"golang/ast"
	"testing"
)

func testUnambiguousSelector(
	t *testing.T, typ ast.Type, name string, found fieldOrMethod) {

	u, v, _ := findSelector(typ, name)
	if u.F == nil && u.M == nil {
		t.Fatalf("selector `%s` not found", name)
	}
	if v.F != nil || v.M != nil {
		t.Errorf("selector `%s` is ambiguous", name)
	}
	if u != found {
		t.Error("unexpected field or member found")
	}
}

func testSelectorNotFound(t *testing.T, typ ast.Type, name string) {
	u, _, _ := findSelector(typ, name)
	if u.F != nil || u.M != nil {
		t.Fatalf("selector `%s` found", name)
	}
}

func testAmbiguousSelector(t *testing.T, typ ast.Type, name string) {
	_, v, _ := findSelector(typ, name)
	if v.F == nil && v.M == nil {
		t.Errorf("selector `%s` is not ambiguous", name)
	}
}

func TestSelector(t *testing.T) {
	p, err := compilePackage("_test/src/sel", []string{"selector-1.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	A := p.Find("A").(*ast.TypeDecl)
	AX := fieldOrMethod{F: &A.Type.(*ast.StructType).Fields[0]}
	MX := fieldOrMethod{M: A.Methods[0]}
	testUnambiguousSelector(t, A, "X", AX)
	testUnambiguousSelector(t, A.Type, "X", AX)
	testUnambiguousSelector(t, &ast.PtrType{Base: A}, "X", AX)
	testUnambiguousSelector(t, &ast.PtrType{Base: A.Type}, "X", AX)
	testUnambiguousSelector(t, A, "MX", MX)
	testSelectorNotFound(t, A.Type, "MX")
	testUnambiguousSelector(t, &ast.PtrType{Base: A}, "MX", MX)
	testSelectorNotFound(t, &ast.PtrType{Base: A.Type}, "MX")

	AA := p.Find("AA").(*ast.TypeDecl)
	testUnambiguousSelector(t, AA, "X", AX)
	testUnambiguousSelector(t, AA.Type, "X", AX)
	testUnambiguousSelector(t, &ast.PtrType{Base: AA}, "X", AX)
	testUnambiguousSelector(t, &ast.PtrType{Base: AA.Type}, "X", AX)
	testSelectorNotFound(t, AA, "MX")
	testSelectorNotFound(t, &ast.PtrType{Base: AA}, "MX")

	B := p.Find("B").(*ast.TypeDecl)
	BY := fieldOrMethod{F: &B.Type.(*ast.StructType).Fields[1]}
	AMY := fieldOrMethod{M: A.Methods[1]}
	BMY := fieldOrMethod{M: B.PMethods[0]}
	testUnambiguousSelector(t, B, "X", AX)
	testUnambiguousSelector(t, B.Type, "X", AX)
	testUnambiguousSelector(t, &ast.PtrType{Base: B}, "X", AX)
	testUnambiguousSelector(t, &ast.PtrType{Base: B.Type}, "X", AX)
	testUnambiguousSelector(t, B, "MX", MX)
	testUnambiguousSelector(t, B.Type, "MX", MX)
	testUnambiguousSelector(t, &ast.PtrType{Base: B}, "MX", MX)
	testUnambiguousSelector(t, &ast.PtrType{Base: B.Type}, "MX", MX)

	testUnambiguousSelector(t, B, "Y", BY)
	testUnambiguousSelector(t, B.Type, "Y", BY)
	testUnambiguousSelector(t, &ast.PtrType{Base: B}, "Y", BY)
	testUnambiguousSelector(t, &ast.PtrType{Base: B.Type}, "Y", BY)
	testUnambiguousSelector(t, B, "MY", BMY)
	testUnambiguousSelector(t, B.Type, "MY", AMY)
	testUnambiguousSelector(t, &ast.PtrType{Base: B}, "MY", BMY)
	testUnambiguousSelector(t, &ast.PtrType{Base: B.Type}, "MY", AMY)

	C := p.Find("C").(*ast.TypeDecl)
	CY := fieldOrMethod{F: &C.Type.(*ast.StructType).Fields[1]}
	CMY := fieldOrMethod{M: C.Methods[0]}
	testUnambiguousSelector(t, C, "X", AX)
	testUnambiguousSelector(t, C.Type, "X", AX)
	testUnambiguousSelector(t, &ast.PtrType{Base: C}, "X", AX)
	testUnambiguousSelector(t, &ast.PtrType{Base: C.Type}, "X", AX)
	testUnambiguousSelector(t, C, "MX", MX)
	testUnambiguousSelector(t, C.Type, "MX", MX)
	testUnambiguousSelector(t, &ast.PtrType{Base: C}, "MX", MX)
	testUnambiguousSelector(t, &ast.PtrType{Base: C.Type}, "MX", MX)

	testUnambiguousSelector(t, C, "Y", CY)
	testUnambiguousSelector(t, C.Type, "Y", CY)
	testUnambiguousSelector(t, &ast.PtrType{Base: C}, "Y", CY)
	testUnambiguousSelector(t, &ast.PtrType{Base: C.Type}, "Y", CY)
	testUnambiguousSelector(t, C, "MY", CMY)
	testUnambiguousSelector(t, C.Type, "MY", AMY)
	testUnambiguousSelector(t, &ast.PtrType{Base: C}, "MY", CMY)
	testUnambiguousSelector(t, &ast.PtrType{Base: C.Type}, "MY", AMY)

	I := p.Find("I").(*ast.TypeDecl)
	F := fieldOrMethod{M: I.Type.(*ast.InterfaceType).Methods[0]}
	testUnambiguousSelector(t, I, "F", F)
	testUnambiguousSelector(t, I.Type, "F", F)
	testSelectorNotFound(t, &ast.PtrType{Base: I}, "F")

	II := p.Find("II").(*ast.TypeDecl)
	testUnambiguousSelector(t, II, "F", F)
	testUnambiguousSelector(t, II.Type, "F", F)
	testSelectorNotFound(t, &ast.PtrType{Base: II}, "F")

	J := p.Find("J").(*ast.TypeDecl)
	G := fieldOrMethod{M: J.Type.(*ast.InterfaceType).Methods[0]}
	testUnambiguousSelector(t, J, "F", F)
	testUnambiguousSelector(t, J, "G", G)
	testSelectorNotFound(t, &ast.PtrType{Base: J}, "F")
	testSelectorNotFound(t, &ast.PtrType{Base: J}, "G")

	testUnambiguousSelector(t, A, "F", F)
	testUnambiguousSelector(t, A.Type, "F", F)
	testUnambiguousSelector(t, &ast.PtrType{Base: A}, "F", F)
	testUnambiguousSelector(t, A, "G", G)
	testUnambiguousSelector(t, A.Type, "G", G)
	testUnambiguousSelector(t, &ast.PtrType{Base: A}, "G", G)

	testUnambiguousSelector(t, B, "F", F)
	testUnambiguousSelector(t, B.Type, "F", F)
	testUnambiguousSelector(t, &ast.PtrType{Base: B}, "F", F)
	testUnambiguousSelector(t, B, "G", G)
	testUnambiguousSelector(t, B.Type, "G", G)
	testUnambiguousSelector(t, &ast.PtrType{Base: B}, "G", G)

	testUnambiguousSelector(t, C, "F", F)
	testUnambiguousSelector(t, C.Type, "F", F)
	testUnambiguousSelector(t, &ast.PtrType{Base: C}, "F", F)
	testUnambiguousSelector(t, C, "G", G)
	testUnambiguousSelector(t, C.Type, "G", G)
	testUnambiguousSelector(t, &ast.PtrType{Base: C}, "G", G)

	P := p.Find("P").(*ast.TypeDecl)
	testUnambiguousSelector(t, P, "X", AX)
	testUnambiguousSelector(t, P.Type, "X", AX)
	testSelectorNotFound(t, &ast.PtrType{Base: P}, "X")
	testSelectorNotFound(t, &ast.PtrType{Base: P.Type}, "X")
	testSelectorNotFound(t, P, "MX")
	testUnambiguousSelector(t, P.Type, "MX", MX)
	testSelectorNotFound(t, P, "MY")

	testUnambiguousSelector(t, P, "Y", CY)
	testUnambiguousSelector(t, P.Type, "Y", CY)
	testSelectorNotFound(t, &ast.PtrType{Base: P}, "Y")
	testSelectorNotFound(t, &ast.PtrType{Base: P.Type}, "Y")

	testSelectorNotFound(t, P, "F")
	testUnambiguousSelector(t, P.Type, "F", F)
	testSelectorNotFound(t, &ast.PtrType{Base: P}, "F")
	testSelectorNotFound(t, P, "G")
	testUnambiguousSelector(t, P.Type, "G", G)
	testSelectorNotFound(t, &ast.PtrType{Base: P}, "G")
}

func TestSelectorErr(t *testing.T) {
	p, err := compilePackage("_test/src/sel", []string{"selector-2.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	C := p.Find("C").(*ast.TypeDecl)
	testAmbiguousSelector(t, C, "X")
	testAmbiguousSelector(t, C.Type, "X")
	testAmbiguousSelector(t, &ast.PtrType{Base: C}, "X")
	testAmbiguousSelector(t, &ast.PtrType{Base: C.Type}, "X")

	testAmbiguousSelector(t, C, "F")
	testAmbiguousSelector(t, C.Type, "F")
	testAmbiguousSelector(t, &ast.PtrType{Base: C}, "F")
	testAmbiguousSelector(t, &ast.PtrType{Base: C.Type}, "F")

	P := p.Find("P").(*ast.TypeDecl)
	testAmbiguousSelector(t, P, "X")
	testAmbiguousSelector(t, P.Type, "X")
	testSelectorNotFound(t, &ast.PtrType{Base: P}, "X")
	testSelectorNotFound(t, &ast.PtrType{Base: P.Type}, "X")

	testAmbiguousSelector(t, P, "F")
	testAmbiguousSelector(t, P.Type, "F")
	testSelectorNotFound(t, &ast.PtrType{Base: P}, "F")
	testSelectorNotFound(t, &ast.PtrType{Base: P.Type}, "F")

	E := p.Find("E").(*ast.TypeDecl)
	testAmbiguousSelector(t, E, "X")
	testAmbiguousSelector(t, E.Type, "X")
	testAmbiguousSelector(t, &ast.PtrType{Base: E}, "X")
	testAmbiguousSelector(t, &ast.PtrType{Base: E.Type}, "X")

	D := p.Find("D").(*ast.TypeDecl)
	testSelectorNotFound(t, D, "Y")
}
