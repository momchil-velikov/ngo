package typecheck

import (
	"bufio"
	"errors"
	"golang/ast"
	"golang/pdb"
	"io/ioutil"
	"os"
	"testing"
)

func testUnambiguousSelector(
	t *testing.T, pkg *ast.Package, typ ast.Type, name string, found fieldOrMethod) {

	u, v, _ := findSelector(pkg, typ, name)
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

func testSelectorNotFound(t *testing.T, pkg *ast.Package, typ ast.Type, name string) {
	u, _, _ := findSelector(pkg, typ, name)
	if u.F != nil || u.M != nil {
		t.Fatalf("selector `%s` found", name)
	}
}

func testAmbiguousSelector(t *testing.T, pkg *ast.Package, typ ast.Type, name string) {
	_, v, _ := findSelector(pkg, typ, name)
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
	testUnambiguousSelector(t, p, A, "X", AX)
	testUnambiguousSelector(t, p, A.Type, "X", AX)
	testUnambiguousSelector(t, p, &ast.PtrType{Base: A}, "X", AX)
	testUnambiguousSelector(t, p, &ast.PtrType{Base: A.Type}, "X", AX)
	testUnambiguousSelector(t, p, A, "MX", MX)
	testSelectorNotFound(t, p, A.Type, "MX")
	testUnambiguousSelector(t, p, &ast.PtrType{Base: A}, "MX", MX)
	testSelectorNotFound(t, p, &ast.PtrType{Base: A.Type}, "MX")

	AA := p.Find("AA").(*ast.TypeDecl)
	testUnambiguousSelector(t, p, AA, "X", AX)
	testUnambiguousSelector(t, p, AA.Type, "X", AX)
	testUnambiguousSelector(t, p, &ast.PtrType{Base: AA}, "X", AX)
	testUnambiguousSelector(t, p, &ast.PtrType{Base: AA.Type}, "X", AX)
	testSelectorNotFound(t, p, AA, "MX")
	testSelectorNotFound(t, p, &ast.PtrType{Base: AA}, "MX")

	B := p.Find("B").(*ast.TypeDecl)
	BY := fieldOrMethod{F: &B.Type.(*ast.StructType).Fields[1]}
	AMY := fieldOrMethod{M: A.Methods[1]}
	BMY := fieldOrMethod{M: B.PMethods[0]}
	testUnambiguousSelector(t, p, B, "X", AX)
	testUnambiguousSelector(t, p, B.Type, "X", AX)
	testUnambiguousSelector(t, p, &ast.PtrType{Base: B}, "X", AX)
	testUnambiguousSelector(t, p, &ast.PtrType{Base: B.Type}, "X", AX)
	testUnambiguousSelector(t, p, B, "MX", MX)
	testUnambiguousSelector(t, p, B.Type, "MX", MX)
	testUnambiguousSelector(t, p, &ast.PtrType{Base: B}, "MX", MX)
	testUnambiguousSelector(t, p, &ast.PtrType{Base: B.Type}, "MX", MX)

	testUnambiguousSelector(t, p, B, "Y", BY)
	testUnambiguousSelector(t, p, B.Type, "Y", BY)
	testUnambiguousSelector(t, p, &ast.PtrType{Base: B}, "Y", BY)
	testUnambiguousSelector(t, p, &ast.PtrType{Base: B.Type}, "Y", BY)
	testUnambiguousSelector(t, p, B, "MY", BMY)
	testUnambiguousSelector(t, p, B.Type, "MY", AMY)
	testUnambiguousSelector(t, p, &ast.PtrType{Base: B}, "MY", BMY)
	testUnambiguousSelector(t, p, &ast.PtrType{Base: B.Type}, "MY", AMY)

	C := p.Find("C").(*ast.TypeDecl)
	CY := fieldOrMethod{F: &C.Type.(*ast.StructType).Fields[1]}
	CMY := fieldOrMethod{M: C.Methods[0]}
	testUnambiguousSelector(t, p, C, "X", AX)
	testUnambiguousSelector(t, p, C.Type, "X", AX)
	testUnambiguousSelector(t, p, &ast.PtrType{Base: C}, "X", AX)
	testUnambiguousSelector(t, p, &ast.PtrType{Base: C.Type}, "X", AX)
	testUnambiguousSelector(t, p, C, "MX", MX)
	testUnambiguousSelector(t, p, C.Type, "MX", MX)
	testUnambiguousSelector(t, p, &ast.PtrType{Base: C}, "MX", MX)
	testUnambiguousSelector(t, p, &ast.PtrType{Base: C.Type}, "MX", MX)

	testUnambiguousSelector(t, p, C, "Y", CY)
	testUnambiguousSelector(t, p, C.Type, "Y", CY)
	testUnambiguousSelector(t, p, &ast.PtrType{Base: C}, "Y", CY)
	testUnambiguousSelector(t, p, &ast.PtrType{Base: C.Type}, "Y", CY)
	testUnambiguousSelector(t, p, C, "MY", CMY)
	testUnambiguousSelector(t, p, C.Type, "MY", AMY)
	testUnambiguousSelector(t, p, &ast.PtrType{Base: C}, "MY", CMY)
	testUnambiguousSelector(t, p, &ast.PtrType{Base: C.Type}, "MY", AMY)

	I := p.Find("I").(*ast.TypeDecl)
	F := fieldOrMethod{M: I.Type.(*ast.InterfaceType).Methods[0]}
	testUnambiguousSelector(t, p, I, "F", F)
	testUnambiguousSelector(t, p, I.Type, "F", F)
	testSelectorNotFound(t, p, &ast.PtrType{Base: I}, "F")

	II := p.Find("II").(*ast.TypeDecl)
	testUnambiguousSelector(t, p, II, "F", F)
	testUnambiguousSelector(t, p, II.Type, "F", F)
	testSelectorNotFound(t, p, &ast.PtrType{Base: II}, "F")

	J := p.Find("J").(*ast.TypeDecl)
	G := fieldOrMethod{M: J.Type.(*ast.InterfaceType).Methods[0]}
	testUnambiguousSelector(t, p, J, "F", F)
	testUnambiguousSelector(t, p, J, "G", G)
	testSelectorNotFound(t, p, &ast.PtrType{Base: J}, "F")
	testSelectorNotFound(t, p, &ast.PtrType{Base: J}, "G")

	testUnambiguousSelector(t, p, A, "F", F)
	testUnambiguousSelector(t, p, A.Type, "F", F)
	testUnambiguousSelector(t, p, &ast.PtrType{Base: A}, "F", F)
	testUnambiguousSelector(t, p, A, "G", G)
	testUnambiguousSelector(t, p, A.Type, "G", G)
	testUnambiguousSelector(t, p, &ast.PtrType{Base: A}, "G", G)

	testUnambiguousSelector(t, p, B, "F", F)
	testUnambiguousSelector(t, p, B.Type, "F", F)
	testUnambiguousSelector(t, p, &ast.PtrType{Base: B}, "F", F)
	testUnambiguousSelector(t, p, B, "G", G)
	testUnambiguousSelector(t, p, B.Type, "G", G)
	testUnambiguousSelector(t, p, &ast.PtrType{Base: B}, "G", G)

	testUnambiguousSelector(t, p, C, "F", F)
	testUnambiguousSelector(t, p, C.Type, "F", F)
	testUnambiguousSelector(t, p, &ast.PtrType{Base: C}, "F", F)
	testUnambiguousSelector(t, p, C, "G", G)
	testUnambiguousSelector(t, p, C.Type, "G", G)
	testUnambiguousSelector(t, p, &ast.PtrType{Base: C}, "G", G)

	P := p.Find("P").(*ast.TypeDecl)
	testUnambiguousSelector(t, p, P, "X", AX)
	testUnambiguousSelector(t, p, P.Type, "X", AX)
	testSelectorNotFound(t, p, &ast.PtrType{Base: P}, "X")
	testSelectorNotFound(t, p, &ast.PtrType{Base: P.Type}, "X")
	testSelectorNotFound(t, p, P, "MX")
	testUnambiguousSelector(t, p, P.Type, "MX", MX)
	testSelectorNotFound(t, p, P, "MY")

	testUnambiguousSelector(t, p, P, "Y", CY)
	testUnambiguousSelector(t, p, P.Type, "Y", CY)
	testSelectorNotFound(t, p, &ast.PtrType{Base: P}, "Y")
	testSelectorNotFound(t, p, &ast.PtrType{Base: P.Type}, "Y")

	testSelectorNotFound(t, p, P, "F")
	testUnambiguousSelector(t, p, P.Type, "F", F)
	testSelectorNotFound(t, p, &ast.PtrType{Base: P}, "F")
	testSelectorNotFound(t, p, P, "G")
	testUnambiguousSelector(t, p, P.Type, "G", G)
	testSelectorNotFound(t, p, &ast.PtrType{Base: P}, "G")
}

func TestSelectorErr(t *testing.T) {
	p, err := compilePackage("_test/src/sel", []string{"selector-2.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	C := p.Find("C").(*ast.TypeDecl)
	testAmbiguousSelector(t, p, C, "X")
	testAmbiguousSelector(t, p, C.Type, "X")
	testAmbiguousSelector(t, p, &ast.PtrType{Base: C}, "X")
	testAmbiguousSelector(t, p, &ast.PtrType{Base: C.Type}, "X")

	testAmbiguousSelector(t, p, C, "F")
	testAmbiguousSelector(t, p, C.Type, "F")
	testAmbiguousSelector(t, p, &ast.PtrType{Base: C}, "F")
	testAmbiguousSelector(t, p, &ast.PtrType{Base: C.Type}, "F")

	P := p.Find("P").(*ast.TypeDecl)
	testAmbiguousSelector(t, p, P, "X")
	testAmbiguousSelector(t, p, P.Type, "X")
	testSelectorNotFound(t, p, &ast.PtrType{Base: P}, "X")
	testSelectorNotFound(t, p, &ast.PtrType{Base: P.Type}, "X")

	testAmbiguousSelector(t, p, P, "F")
	testAmbiguousSelector(t, p, P.Type, "F")
	testSelectorNotFound(t, p, &ast.PtrType{Base: P}, "F")
	testSelectorNotFound(t, p, &ast.PtrType{Base: P.Type}, "F")

	E := p.Find("E").(*ast.TypeDecl)
	testAmbiguousSelector(t, p, E, "X")
	testAmbiguousSelector(t, p, E.Type, "X")
	testAmbiguousSelector(t, p, &ast.PtrType{Base: E}, "X")
	testAmbiguousSelector(t, p, &ast.PtrType{Base: E.Type}, "X")

	D := p.Find("D").(*ast.TypeDecl)
	testSelectorNotFound(t, p, D, "Y")
}

type MockPackageLocator struct {
	pkgs map[string]*ast.Package
}

func (loc *MockPackageLocator) FindPackage(path string) (*ast.Package, error) {
	pkg, ok := loc.pkgs[path]
	if !ok {
		return nil, errors.New("import `" + path + "` not found")
	}
	return pkg, nil
}

func reloadPackage(pkg *ast.Package, loc ast.PackageLocator) (*ast.Package, error) {
	f, err := ioutil.TempFile("", "resolve-test-pkg")
	if err != nil {
		return nil, err
	}
	os.Remove(f.Name())
	defer f.Close()
	w := bufio.NewWriter(f)
	if err = pdb.Write(w, pkg); err != nil {
		return nil, err
	}
	if err = w.Flush(); err != nil {
		return nil, err
	}
	if _, err = f.Seek(0, os.SEEK_SET); err != nil {
		return nil, err
	}
	return pdb.Read(bufio.NewReader(f), loc)
}

func TestSelectorNonExported(t *testing.T) {
	loc := &MockPackageLocator{pkgs: make(map[string]*ast.Package)}
	pa, err := compilePackage("_test/src/sel/a", []string{"a.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	pa, err = reloadPackage(pa, loc)
	if err != nil {
		t.Fatal(err)
	}
	loc.pkgs["sel/a"] = pa

	expectErrorWithLoc(t, "_test/src/sel/b", []string{"b-err-1.go"}, loc,
		"ambiguous selector X")

	pb, err := compilePackage("_test/src/sel/b", []string{"b.go"}, loc)
	if err != nil {
		t.Fatal(err)
	}
	pb, err = reloadPackage(pb, loc)
	if err != nil {
		t.Fatal(err)
	}
	loc.pkgs["sel/b"] = pb

	_, err = compilePackage("_test/src/sel", []string{"selector-3.go"}, loc)
	if err != nil {
		t.Fatal(err)
	}

	expectErrorWithLoc(t, "_test/src/sel", []string{"selector-4.go"}, loc,
		"type does not have a field or method named y")
	expectErrorWithLoc(t, "_test/src/sel", []string{"selector-5.go"}, loc,
		"type does not have a field or method named y")
	expectErrorWithLoc(t, "_test/src/sel", []string{"selector-6.go"}, loc,
		"ambiguous selector X")
	expectErrorWithLoc(t, "_test/src/sel", []string{"selector-7.go"}, loc,
		"type does not have a field or method named y")
}

func TestMethodExpr(t *testing.T) {
	p, err := compilePackage("_test/src/sel", []string{"method-expr.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	I := p.Find("I").(*ast.TypeDecl)
	A := p.Find("A").(*ast.Var)
	typ, ok := A.Type.(*ast.FuncType)
	if !ok || len(typ.Params) != 2 {
		t.Fatal("`A` should be a function with two parameters")
	}
	if typ.Params[0].Type != I {
		t.Error("first parameter of `A` must have type `I`")
	}
	if unnamedType(typ.Params[1].Type) != ast.BuiltinInt {
		t.Error("second parameter of `A` must have type `int`")
	}

	S := p.Find("S").(*ast.TypeDecl)
	B := p.Find("B").(*ast.Var)
	typ, ok = B.Type.(*ast.FuncType)
	if !ok || len(typ.Params) != 2 {
		t.Fatal("`B` should be a function with two parameters")
	}
	if typ.Params[0].Type != S {
		t.Error("first parameter of `B` must have type `S`")
	}
	if unnamedType(typ.Params[1].Type) != ast.BuiltinInt {
		t.Error("second parameter of `B` must have type `int`")
	}

	C := p.Find("C").(*ast.Var)
	typ, ok = C.Type.(*ast.FuncType)
	if !ok || len(typ.Params) != 2 {
		t.Fatal("`C` should be a function with two parameters")
	}
	if ptr, ok := typ.Params[0].Type.(*ast.PtrType); !ok || ptr.Base != S {
		t.Error("first parameter of `C` must have type `S`")
	}
	if unnamedType(typ.Params[1].Type) != ast.BuiltinInt {
		t.Error("second parameter of `C` must have type `int`")
	}

	D := p.Find("D").(*ast.Var)
	typ, ok = D.Type.(*ast.FuncType)
	if !ok || len(typ.Params) != 2 {
		t.Fatal("`D` should be a function with two parameters")
	}
	if ptr, ok := typ.Params[0].Type.(*ast.PtrType); !ok || ptr.Base != S {
		t.Error("first parameter of `D` must have type `*S`")
	}
	if unnamedType(typ.Params[1].Type) != ast.BuiltinFloat64 {
		t.Error("second parameter of `D` must have type `float64`")
	}

	J := p.Find("J").(*ast.TypeDecl)
	A0 := p.Find("A0").(*ast.Var)
	typ, ok = A0.Type.(*ast.FuncType)
	if !ok || len(typ.Params) != 2 {
		t.Fatal("`A0` should be a function with two parameters")
	}
	if typ.Params[0].Type != J {
		t.Error("first parameter of `A0` must have type `J`")
	}
	if unnamedType(typ.Params[1].Type) != ast.BuiltinInt {
		t.Error("second parameter of `A0` must have type `int`")
	}

	A1 := p.Find("A1").(*ast.Var)
	typ, ok = A1.Type.(*ast.FuncType)
	if !ok || len(typ.Params) != 2 {
		t.Fatal("`A1` should be a function with two parameters")
	}
	if ptr, ok := typ.Params[0].Type.(*ast.PtrType); !ok || ptr.Base != J {
		t.Error("first parameter of `A1` must have type `*J`")
	}
	if unnamedType(typ.Params[1].Type) != ast.BuiltinInt {
		t.Error("second parameter of `A1` must have type `int`")
	}

	T0 := p.Find("T0").(*ast.TypeDecl)
	B0 := p.Find("B0").(*ast.Var)
	typ, ok = B0.Type.(*ast.FuncType)
	if !ok || len(typ.Params) != 2 {
		t.Fatal("`B0` should be a function with two parameters")
	}
	if typ.Params[0].Type != T0 {
		t.Error("first parameter of `B0` must have type `T0`")
	}
	if unnamedType(typ.Params[1].Type) != ast.BuiltinInt {
		t.Error("second parameter of `B0` must have type `int`")
	}

	C0 := p.Find("C0").(*ast.Var)
	typ, ok = C0.Type.(*ast.FuncType)
	if !ok || len(typ.Params) != 2 {
		t.Fatal("`C0` should be a function with two parameters")
	}
	if ptr, ok := typ.Params[0].Type.(*ast.PtrType); !ok || ptr.Base != T0 {
		t.Error("first parameter of `C0` must have type `*T0`")
	}
	if unnamedType(typ.Params[1].Type) != ast.BuiltinInt {
		t.Error("second parameter of `C0` must have type `int`")
	}

	D0 := p.Find("D0").(*ast.Var)
	typ, ok = D0.Type.(*ast.FuncType)
	if !ok || len(typ.Params) != 2 {
		t.Fatal("`D0` should be a function with two parameters")
	}
	if ptr, ok := typ.Params[0].Type.(*ast.PtrType); !ok || ptr.Base != T0 {
		t.Error("first parameter of `D0` must have type `*T0`")
	}
	if unnamedType(typ.Params[1].Type) != ast.BuiltinFloat64 {
		t.Error("second parameter of `D0` must have type `float64`")
	}

	T1 := p.Find("T1").(*ast.TypeDecl)
	B1 := p.Find("B1").(*ast.Var)
	typ, ok = B1.Type.(*ast.FuncType)
	if !ok || len(typ.Params) != 2 {
		t.Fatal("`B1` should be a function with two parameters")
	}
	if typ.Params[0].Type != T1 {
		t.Error("first parameter of `B0` must have type `T1`")
	}
	if unnamedType(typ.Params[1].Type) != ast.BuiltinInt {
		t.Error("second parameter of `B1` must have type `int`")
	}

	C1 := p.Find("C1").(*ast.Var)
	typ, ok = C1.Type.(*ast.FuncType)
	if !ok || len(typ.Params) != 2 {
		t.Fatal("`C1` should be a function with two parameters")
	}
	if ptr, ok := typ.Params[0].Type.(*ast.PtrType); !ok || ptr.Base != T1 {
		t.Error("first parameter of `C1` must have type `*T1`")
	}
	if unnamedType(typ.Params[1].Type) != ast.BuiltinInt {
		t.Error("second parameter of `C1` must have type `int`")
	}

	D10 := p.Find("D10").(*ast.Var)
	typ, ok = D10.Type.(*ast.FuncType)
	if !ok || len(typ.Params) != 2 {
		t.Fatal("`D10` should be a function with two parameters")
	}
	if typ.Params[0].Type != T1 {
		t.Error("first parameter of `D10` must have type `T1`")
	}
	if unnamedType(typ.Params[1].Type) != ast.BuiltinFloat64 {
		t.Error("second parameter of `D10` must have type `float64`")
	}

	D11 := p.Find("D11").(*ast.Var)
	typ, ok = D11.Type.(*ast.FuncType)
	if !ok || len(typ.Params) != 2 {
		t.Fatal("`D11` should be a function with two parameters")
	}
	if ptr, ok := typ.Params[0].Type.(*ast.PtrType); !ok || ptr.Base != T1 {
		t.Error("first parameter of `D11` must have type `*T1`")
	}
	if unnamedType(typ.Params[1].Type) != ast.BuiltinFloat64 {
		t.Error("second parameter of `D11` must have type `float64`")
	}
}

func TestMethodExprErr(t *testing.T) {
	expectError(t, "_test/src/sel", []string{"method-expr-err-1.go"},
		"must have a pointer receiver")
	expectError(t, "_test/src/sel", []string{"method-expr-err-2.go"},
		"must have a pointer receiver")

	loc := &MockPackageLocator{pkgs: make(map[string]*ast.Package)}
	pa, err := compilePackage("_test/src/sel/a", []string{"a.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	pa, err = reloadPackage(pa, loc)
	if err != nil {
		t.Fatal(err)
	}
	loc.pkgs["sel/a"] = pa

	expectErrorWithLoc(t, "_test/src/sel/b", []string{"b-err-2.go"}, loc,
		"ambiguous selector F")

	pb, err := compilePackage("_test/src/sel/b", []string{"b.go"}, loc)
	if err != nil {
		t.Fatal(err)
	}
	pb, err = reloadPackage(pb, loc)
	if err != nil {
		t.Fatal(err)
	}
	loc.pkgs["sel/b"] = pb

	expectErrorWithLoc(t, "_test/src/sel", []string{"method-expr-err-3.go"}, loc,
		"type does not have a method named g")
	expectErrorWithLoc(t, "_test/src/sel", []string{"method-expr-err-4.go"}, loc,
		"ambiguous selector F")
	expectErrorWithLoc(t, "_test/src/sel", []string{"method-expr-err-5.go"}, loc,
		"type does not have a method named g")
}
