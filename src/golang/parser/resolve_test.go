package parser

import (
	"golang/ast"
	"path/filepath"
	"strings"
	"testing"
)

func getTypeDecl(s ast.Scope, name string) *ast.TypeDecl {
	sym := s.Find(name)
	if sym == nil {
		return nil
	}
	t, _ := sym.(*ast.TypeDecl)
	return t
}

func TestResolveTypeUniverse(t *testing.T) {
	const TESTDIR = "_test/typedecl/src/correct"
	up, err := ParsePackage(TESTDIR, []string{"a.go", "b.go"})
	if err != nil {
		t.Error(err)
	}
	p, err := ResolvePackage(up, nil)
	if err != nil {
		t.Error(err)
	}

	cases := []struct{ decl, ref, file string }{
		{"a", "bool", "a.go"},
		{"b", "byte", "a.go"},
		{"c", "uint8", "a.go"},
		{"d", "uint16", "a.go"},
		{"e", "uint32", "a.go"},
		{"f", "uint64", "a.go"},
		{"g", "int8", "a.go"},
		{"h", "int16", "a.go"},
		{"i", "rune", "a.go"},
		{"j", "int32", "a.go"},
		{"k", "int64", "a.go"},
		{"l", "float32", "b.go"},
		{"m", "float64", "b.go"},
		{"n", "complex64", "b.go"},
		{"o", "complex128", "b.go"},
		{"p", "uint", "b.go"},
		{"q", "int", "b.go"},
		{"r", "uintptr", "b.go"},
		{"s", "string", "b.go"},
	}
	for _, cs := range cases {
		// Test type is declared at package scope.
		sym := p.Decls[cs.decl]
		if sym == nil {
			t.Fatalf("name `%s` not found at package scope\n", cs.decl)
		}
		dcl, ok := sym.(*ast.TypeDecl)
		if !ok {
			t.Fatalf("`%s` must be a TypeDecl\n", cs.decl)
		}
		// Test type is registered under its name.
		name, _, file := sym.DeclaredAt()
		if name != cs.decl {
			t.Errorf("type `%s` declared as `%s`\n", cs.decl, name)
		}
		// Test type is declared in the respective file.
		if file == nil || filepath.Base(file.Name) != cs.file {
			t.Errorf("type '%s' must be declared in file `%s`", cs.decl, cs.file)
		}
		// Test type declaration refers to the respective typename at universe
		// scope.
		dcl, ok = dcl.Type.(*ast.TypeDecl)
		if !ok {
			t.Fatalf("declaration of `%s` does not refer to a typename\n", cs.decl)
		}
		if dcl != ast.UniverseScope.Find(cs.ref) {
			if dcl == nil {
				t.Errorf("`%s` refers to an invalid Typename\n", cs.decl)
			} else {
				t.Errorf(
					"`%s` expected to refer to `%s`, refers to `%s` instead\n",
					cs.decl, cs.ref, dcl.Name,
				)
			}
		}
	}
}

func TestResolveTypeDuplicatelAtPackageScope(t *testing.T) {
	const TESTDIR = "_test/typedecl/src/errors/dup_decl"
	up, err := ParsePackage(TESTDIR, []string{"a.go"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = ResolvePackage(up, nil)
	if err == nil || !strings.Contains(err.Error(), "redeclared") {
		if err != nil {
			t.Log(err)
		}
		t.Error("expected duplicated declaration error")
	}

	up, err = ParsePackage(TESTDIR, []string{"b.go", "c.go"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = ResolvePackage(up, nil)
	if err == nil || !strings.Contains(err.Error(), "redeclared") {
		if err != nil {
			t.Log(err)
		}
		t.Error("expected duplicated declaration error")
	}
}

func TestResolveTypePackageScope(t *testing.T) {
	const TESTDIR = "_test/typedecl/src/correct"
	up, err := ParsePackage(TESTDIR, []string{"a.go", "b.go"})
	if err != nil {
		t.Fatal(err)
	}
	p, err := ResolvePackage(up, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Test the type is declared at package scope, the declaration is a
	// TypeDecl and that it refers a Typename, which in turn refers to the
	// right TypeDecl. The various test cases explore different combinations
	// of source files for the type declaration and the referred to type.
	cases := []struct{ decl, ref string }{
		{"aa", "a"},
		{"aaa", "aa"},
		{"ss", "s"},
		{"sss", "ss"},
		{"rr", "r"},
		{"rrr", "rr"},
		{"kk", "k"},
		{"kkk", "kk"},
	}
	for _, cs := range cases {
		d := p.Find(cs.decl)
		if d == nil {
			t.Fatalf("name `%s` not found at package scope\n", cs.decl)
		}
		td, ok := d.(*ast.TypeDecl)
		if !ok {
			t.Fatalf("symbol `%s` is not a TypeDecl\n", cs.decl)
		}
		td, ok = td.Type.(*ast.TypeDecl)
		if !ok {
			t.Fatalf("declaration of `%s` does not refer to a typename", cs.decl)
		}
		if td != p.Find(cs.ref).(*ast.TypeDecl) {
			t.Errorf(
				"declaration of `%s` does nor refer to the declaration of `%s`\n",
				cs.decl, cs.ref,
			)
		}
	}

	// Test that blank identifier is not declared.
	if p.Find("_") != nil {
		t.Error("type with name `_` must not be declared")
	}
}

func TestResolveTypeBlockScope(t *testing.T) {
	const TESTDIR = "_test/typedecl/src/correct"
	up, err := ParsePackage(TESTDIR, []string{"a.go", "b.go"})
	if err != nil {
		t.Error(err)
	}
	p, err := ResolvePackage(up, nil)
	if err != nil {
		t.Error(err)
	}
	fn, ok := p.Find("F").(*ast.FuncDecl)
	if !ok {
		t.Fatal("function declaration `F` not found at package scope")
	}

	blk := fn.Func.Blk
	sym := blk.Find("A")
	tdA, _ := sym.(*ast.TypeDecl)
	if sym == nil || tdA == nil {
		t.Fatal("type declaration `A` not found at `F`s scope")
	}
	dcl, ok := tdA.Type.(*ast.TypeDecl)
	if !ok {
		t.Fatal("declaration of `A` does not refer to a typename")
	}
	if dcl != ast.UniverseScope.Find("int") {
		t.Error("declaration of `A` does not refer to predeclared `int`")
	}

	sym = fn.Func.Blk.Find("AA")
	tdAA, _ := sym.(*ast.TypeDecl)
	if sym == nil || tdAA == nil {
		t.Fatal("type declaration `AA` not found at `F`s scope")
	}
	dcl, ok = tdAA.Type.(*ast.TypeDecl)
	if !ok {
		t.Fatal("declaration of `AA` does not refer to a typename")
	}
	if dcl != tdA {
		t.Error("declaration of `AA` does not refer to `A`")
	}

	blk = blk.Body[0].(*ast.Block)
	tdB := getTypeDecl(blk, "B")
	if tdB == nil {
		t.Fatal("type declaration `B` not found in a nested block in `F`")
	}
	dcl, ok = tdB.Type.(*ast.TypeDecl)
	if !ok {
		t.Fatal("declaration of `B` does not refer to a typename")
	}
	if dcl != ast.UniverseScope.Find("int") {
		t.Error("declaration of `B` does not refer to predeclared `int`")
	}

	tdBB := getTypeDecl(blk, "BB")
	if tdB == nil {
		t.Fatal("type declaration `BB` not found in a nested block in `F`")
	}
	dcl, ok = tdBB.Type.(*ast.TypeDecl)
	if !ok {
		t.Fatal("declaration of `BB` does not refer to a typename")
	}
	if dcl != tdAA {
		t.Error("declaration of `BB` does not refer to the declaration if `AA`")
	}

	tdBBB := getTypeDecl(blk, "BBB")
	if tdB == nil {
		t.Fatal("type declaration `BBB` not found in a nested block in `F`")
	}
	dcl, ok = tdBBB.Type.(*ast.TypeDecl)
	if !ok {
		t.Fatal("declaration of `BBB` does not refer to a typename")
	}
	if dcl != tdBB {
		t.Error("declaration of `BBB` does not refer to the declaration if `BB`")
	}
}

func TestResolveTypeSelfReference(t *testing.T) {
	const TESTDIR = "_test/typedecl/src/correct/"
	up, err := ParsePackage(TESTDIR, []string{"self.go"})
	if err != nil {
		t.Error(err)
	}
	p, err := ResolvePackage(up, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Check that in type declarations the type names are inserted into the
	// symbol table and available during name resolution of the respective
	// TypeSpec.  Note: Invalid recursive types is an error, detected by the
	// type check phase.
	tA := p.Find("A")
	tdA := tA.(*ast.TypeDecl)
	if tA == nil || tdA == nil {
		t.Fatal("type declaration `A` not found at package scope")
	}
	if tdA.Type != tdA {
		t.Error("type declaration `A` does not refer to itself")
	}

	tB := p.Find("B")
	tdB := tB.(*ast.TypeDecl)
	if tB == nil || tdB == nil {
		t.Fatal("type declaration `B` not found at package scope")
	}
	if tdB.Type != tdB {
		t.Error("type declaration `B` does not refer to itself")
	}

	fn, ok := p.Find("F").(*ast.FuncDecl)
	if !ok {
		t.Fatal("function declaration `F` not found at package scope")
	}

	tC := fn.Func.Blk.Find("C")
	tdC := tC.(*ast.TypeDecl)
	if tC == nil || tdC == nil {
		t.Fatal("type declaration `C` not found at `F`s scope")
	}
	if tdC.Type != tdC {
		t.Error("type declaration `C` does not refer to itself")
	}

	tD := fn.Func.Blk.Find("D")
	tdD := tD.(*ast.TypeDecl)
	if tD == nil || tdD == nil {
		t.Fatal("type declaration `D` not found at `F`s scope")
	}
	if tdD.Type != tdD {
		t.Error("type declaration `D` does not refer to itself")
	}
}

func TestResolveConstructedType(t *testing.T) {
	const TESTDIR = "_test/typedecl/src/correct"
	up, err := ParsePackage(TESTDIR, []string{"a.go", "b.go"})
	if err != nil {
		t.Error(err)
	}
	p, err := ResolvePackage(up, nil)
	if err != nil {
		t.Error(err)
	}

	a := getTypeDecl(p, "a")
	u := getTypeDecl(p, "u")
	if u.Type.(*ast.ArrayType).Elt != a {
		t.Error("elements of `u` are not of type `a`")
	}

	v := getTypeDecl(p, "v")
	if v.Type.(*ast.SliceType).Elt != u {
		t.Error("elements of `u` are not of type `v`")
	}

	w := getTypeDecl(p, "w")
	if w.Type.(*ast.PtrType).Base != v {
		t.Error("base of `w` is not of type `v`")
	}

	x := getTypeDecl(p, "x")
	if x.Type.(*ast.MapType).Key != a {
		t.Error("keys of `x` are not of type `a`")
	}
	if x.Type.(*ast.MapType).Elt != w {
		t.Error("elements of `x` are not of type `w`")
	}

	y := getTypeDecl(p, "y")
	if y.Type.(*ast.ChanType).Elt != x {
		t.Error("elements of `y` are not of type `x`")
	}

	z := getTypeDecl(p, "z").Type.(*ast.StructType)
	if z.Fields[0].Type != u {
		t.Error("field `z.X` is not of type `u`")
	}
	if z.Fields[1].Type != v {
		t.Error("field `z.Y` is not of type `v`")
	}

	Fn := getTypeDecl(p, "Fn").Type.(*ast.FuncType)
	if Fn.Params[0].Type != u {
		t.Error("parameter #0 of `Fn` is not of type `u`")
	}
	if Fn.Params[1].Type != v {
		t.Error("parameter #1 of `Fn` is not of type `v`")
	}
	if Fn.Returns[0].Type != w {
		t.Error("return #0 of `Fn` is not of type `w`")
	}
	if Fn.Returns[1].Type != x {
		t.Error("return #1 of `Fn` is not of type `x`")
	}

	IfA := getTypeDecl(p, "IfA")
	mF := IfA.Type.(*ast.InterfaceType).Methods[0].Type.(*ast.FuncType)
	if mF.Params[0].Type != u {
		t.Error("parameterof method `IfA.F` is not of type `u`")
	}
	if mF.Returns[0].Type != v {
		t.Error("return of method `IfA.F` is not of type `v`")
	}

	IfB := getTypeDecl(p, "IfB").Type.(*ast.InterfaceType)
	if IfB.Methods[0].Type != IfA {
		t.Error("embedded interface of `IfB` is not `IfA`")
	}
	mG := IfB.Methods[1].Type.(*ast.FuncType)
	if mG.Params[0].Type != x {
		t.Error("parameterof method `IfB.G` is not of type `x`")
	}
	if mG.Returns[0].Type != y {
		t.Error("return of method `IfB.G` is not of type `y`")
	}
}

func TestResolveFuncTypeError(t *testing.T) {
	for _, src := range []string{"func-1.go", "func-2.go"} {
		up, err := ParsePackage("_test/typedecl/src/errors", []string{src})
		if err != nil {
			t.Fatal(err)
		}
		_, err = ResolvePackage(up, nil)
		if err == nil || !strings.Contains(err.Error(), "all be present") {
			t.Error("expecting `all be present or all be absent` error")
			if err != nil {
				t.Error(err)
			}
		}
	}
}

func TestResolveTypeTypenameNotFound(t *testing.T) {
	for _, src := range []string{
		"typename.go", "array.go", "slice.go", "ptr.go", "map-1.go", "map-2.go",
		"chan.go", "struct.go", "func-1.go", "func-2.go", "iface-1.go", "iface-2.go",
	} {
		up, err := ParsePackage(
			"_test/typedecl/src/errors/typename_not_found", []string{src},
		)
		if err != nil {
			t.Error(err)
		}
		_, err = ResolvePackage(up, nil)
		if err == nil || !strings.Contains(err.Error(), "X not declared") {
			t.Error("expecting `X not declared` error")
		}
	}
}

func TestResolveTypeNotATypename(t *testing.T) {
	for _, src := range []string{
		"typename.go", "array.go", "slice.go", "ptr.go", "map-1.go", "map-2.go",
		"chan.go", "struct.go", "func-1.go", "func-2.go", "iface-1.go", "iface-2.go",
		"blank.go",
	} {
		up, err := ParsePackage("_test/typedecl/src/errors/not_typename", []string{src})
		if err != nil {
			t.Error(err)
		}
		_, err = ResolvePackage(up, nil)
		if err == nil ||
			!(strings.Contains(err.Error(), "X is not a typename") ||
				strings.Contains(err.Error(), "is not a valid type name")) {
			t.Error("expecting `X is not a typename` error")
			if err != nil {
				t.Error(err)
			}
		}
	}
}

func TestResolveTypeDuplicateField(t *testing.T) {
	for _, src := range []string{"dup-1.go", "dup-2.go", "dup-3.go"} {
		up, err := ParsePackage("_test/typedecl/src/errors/dup_field", []string{src})
		if err != nil {
			t.Error(err)
		}
		_, err = ResolvePackage(up, nil)
		if err == nil || !strings.Contains(err.Error(), "X is duplicated") {
			t.Error("expecting `X is duplicated typename` error")
		}
	}

	up, err := ParsePackage("_test/typedecl/src/correct/", []string{"c.go"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = ResolvePackage(up, nil)
	if err != nil {
		t.Error(err)
	}
}

func TestResolveExprLiteral(t *testing.T) {
	up, err := ParsePackage("_test/expr/src/ok", []string{"lit.go"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = ResolvePackage(up, nil)
	if err != nil {
		t.Error(err)
	}
}

func TestResolveExprComposite(t *testing.T) {
	up, err := ParsePackage("_test/expr/src/ok", []string{"comp.go"})
	if err != nil {
		t.Fatal(err)
	}
	p, err := ResolvePackage(up, nil)
	if err != nil {
		t.Fatal(err)
	}

	A := p.Find("A").(*ast.Var)
	B := p.Find("B").(*ast.Var)

	C := p.Find("C").(*ast.Var)
	x := C.Init.RHS[0].(*ast.CompLiteral)
	elt := x.Type.(*ast.SliceType).Elt.(*ast.TypeDecl)
	if elt != ast.UniverseScope.Find("int") {
		t.Error("`C`s initializer type must be `[]int`")
	}
	if x.Elts[0].Value.(*ast.Var) != A {
		t.Error("first element in the slice literal must be `A`")
	}
	if x.Elts[1].Value.(*ast.Var) != B {
		t.Error("first element in the slice literal must be `B`")
	}

	D := p.Find("D").(*ast.Var)
	x = D.Init.RHS[0].(*ast.CompLiteral)
	if x.Elts[0].Key.(*ast.Var) != A {
		t.Error("key in the map literal must be `A`")
	}
	if x.Elts[0].Value.(*ast.Var) != B {
		t.Error("value in the map literal must be `B`")
	}

	E := p.Find("E").(*ast.Var)
	x = E.Init.RHS[0].(*ast.CompLiteral)
	if id, ok := x.Elts[0].Key.(*ast.QualifiedId); !ok {
		t.Error("field names in struct composite literal must not be resolved")
	} else if id.Id != "A" {
		t.Errorf("unexpected fied name `%s`, must be `A`\n", id.Id)
	}
	if x.Elts[0].Value.(*ast.Var) != A {
		t.Error("initializer of fieeld `A` must be variable `A`")
	}
	if id, ok := x.Elts[1].Key.(*ast.QualifiedId); !ok {
		t.Error("field names in struct composite literal must not be resolved")
	} else if id.Id != "B" {
		t.Errorf("unexpected fied name `%s`, must be `B`\n", id.Id)
	}
	if x.Elts[1].Value.(*ast.Var) != B {
		t.Error("initializer of fieeld `B` must be variable `B`")
	}
}

func TestResolveExprConversion(t *testing.T) {
	up, err := ParsePackage("_test/expr/src/ok", []string{"conv.go"})
	if err != nil {
		t.Fatal(err)
	}
	p, err := ResolvePackage(up, nil)
	if err != nil {
		t.Fatal(err)
	}

	T := p.Find("T").(*ast.TypeDecl)
	A := p.Find("A").(*ast.Var)
	B := p.Find("B").(*ast.Var)
	x := B.Init.RHS[0].(*ast.Conversion)
	if x.Type.(*ast.SliceType).Elt != T {
		t.Error("type in the conversion must be a `[]T`")
	}
	if x.X != A {
		t.Error("the converted value must be `A`")
	}

	C := p.Find("C").(*ast.Var)
	D := p.Find("D").(*ast.Var)
	x, ok := D.Init.RHS[0].(*ast.Conversion)
	if !ok {
		t.Error("initializer of `D` must be Conversion")
	}
	if x.Type != T {
		t.Error("in initializer of `D` the conversion type must be `T`")
	}
	if x.X != C {
		t.Error("in initializer of `D` the converted value must be `C`")
	}
}

func TestResolveExprCall(t *testing.T) {
	up, err := ParsePackage("_test/expr/src/ok", []string{"call.go"})
	if err != nil {
		t.Fatal(err)
	}
	p, err := ResolvePackage(up, nil)
	if err != nil {
		t.Fatal(err)
	}

	F := p.Find("F").(*ast.FuncDecl)
	A := p.Find("A").(*ast.Var)
	if A.Init == nil {
		t.Error("`A` must have an initializer statement")
	}
	x := A.Init.RHS[0].(*ast.Call)
	if x.Func != F {
		t.Error("in the initializer of `A` the called function must be `F`")
	}
	X := p.Find("X").(*ast.Var)
	Y := p.Find("Y").(*ast.Var)
	if x.Xs[0] != X {
		t.Error("first argument to the call of `F` must be `X`")
	}
	if x.Xs[1] != Y {
		t.Error("second argument to the call of `F` must be `Y`")
	}
}

func TestResolveExprParens(t *testing.T) {
	up, err := ParsePackage("_test/expr/src/ok", []string{"parens.go"})
	if err != nil {
		t.Fatal(err)
	}
	p, err := ResolvePackage(up, nil)
	if err != nil {
		t.Fatal(err)
	}

	X := p.Find("X").(*ast.Var)
	Y := p.Find("Y").(*ast.Var)
	x := Y.Init.RHS[0]
	if _, ok := x.(*ast.ParensExpr); ok {
		t.Error("ParensExpr must be removed by the resolver")
	}
	if x != X {
		t.Error("initializer or `Y` must be `X`")
	}
}

func TestResolveExprFuncLiteral(t *testing.T) {
	up, err := ParsePackage("_test/expr/src/ok", []string{"func.go"})
	if err != nil {
		t.Fatal(err)
	}
	p, err := ResolvePackage(up, nil)
	if err != nil {
		t.Fatal(err)
	}

	I := p.Find("I").(*ast.TypeDecl)
	R := p.Find("R").(*ast.TypeDecl)
	F := p.Find("F").(*ast.Var)
	fn := F.Init.RHS[0].(*ast.Func)
	if fn.Sig.Params[0].Type != R || fn.Sig.Params[1].Type != R {
		t.Error("parameters in the function literal must have type `R`")
	}
	if fn.Sig.Returns[0].Type != R {
		t.Error("first return in the function literal must have type `R`")
	}
	if fn.Sig.Returns[1].Type != I {
		t.Error("second return in the function literal must have type `I`")
	}
}

func TestResolveTypeAssertion(t *testing.T) {
	up, err := ParsePackage("_test/expr/src/ok", []string{"assert.go"})
	if err != nil {
		t.Fatal(err)
	}
	p, err := ResolvePackage(up, nil)
	if err != nil {
		t.Fatal(err)
	}

	T := p.Find("T").(*ast.TypeDecl)
	A := p.Find("A").(*ast.Var)
	B := p.Find("B").(*ast.Var)
	x := B.Init.RHS[0].(*ast.TypeAssertion)
	if x.X != A {
		t.Error("type assertion expression must be `A`")
	}
	if x.Type != T {
		t.Error("type in type assertion expression must be `T`")
	}
}

func TestResolveExprSelector(t *testing.T) {
	up, err := ParsePackage("_test/expr/src/ok", []string{"selector.go"})
	if err != nil {
		t.Fatal(err)
	}
	p, err := ResolvePackage(up, nil)
	if err != nil {
		t.Fatal(err)
	}

	A := p.Find("A").(*ast.Var)
	B := p.Find("B").(*ast.Var)
	x := B.Init.RHS[0].(*ast.Selector)
	if x.X != A {
		t.Error("expression in selector must be `A`")
	}
	if x.Id != "X" {
		t.Error("identifier in selector must be `X`")
	}

	C := p.Find("C").(*ast.Var)
	D := p.Find("D").(*ast.Var)
	x = D.Init.RHS[0].(*ast.Selector)
	y := x.X.(*ast.Selector)
	if y.X != C {
		t.Error("in initializer of `D` expression in the first selector must be `C`")
	}
	if y.Id != "A" {
		t.Error("in initializer of `D` identifier in the first selector must be `A`")
	}

	S := p.Find("S").(*ast.TypeDecl)
	E := p.Find("E").(*ast.Var)
	z, ok := E.Init.RHS[0].(*ast.MethodExpr)
	if !ok {
		t.Error("initializer of `E` must be a MethodExpr")
	}
	ptr, ok := z.Type.(*ast.PtrType)
	if !ok || ptr.Base != S {
		t.Error("in initializer of `E`: the type in the method expression must be `*S`")
	}
	if z.Id != "Y" {
		t.Error(
			"in initializer of `E`: the identifier in the method expression must be `Y`",
		)
	}
}

func TestResolveExprIndex(t *testing.T) {
	up, err := ParsePackage("_test/expr/src/ok", []string{"index.go"})
	if err != nil {
		t.Fatal(err)
	}
	p, err := ResolvePackage(up, nil)
	if err != nil {
		t.Fatal(err)
	}

	A := p.Find("A").(*ast.Var)
	B := p.Find("B").(*ast.Const)
	C := p.Find("C").(*ast.Var)
	x := C.Init.RHS[0].(*ast.IndexExpr)
	if x.X != A {
		t.Error("indexed obejct must be `A`")
	}
	if x.I != B {
		t.Error("index expression must be `A`")
	}
}

func TestResolveExprSlice(t *testing.T) {
	up, err := ParsePackage("_test/expr/src/ok", []string{"slice.go"})
	if err != nil {
		t.Fatal(err)
	}
	p, err := ResolvePackage(up, nil)
	if err != nil {
		t.Fatal(err)
	}

	A := p.Find("A").(*ast.Var)
	B := p.Find("B").(*ast.Var)
	L := p.Find("L").(*ast.Var)
	H := p.Find("H").(*ast.Var)
	C := p.Find("C").(*ast.Var)
	x := B.Init.RHS[0].(*ast.SliceExpr)
	if x.X != A {
		t.Error("sliced object must be `A`")
	}
	if x.Lo != L {
		t.Error("slice expression low index must be `L`")
	}
	if x.Hi != H {
		t.Error("slice expression high index must be `H`")
	}
	if x.Cap != C {
		t.Error("slice expression capacity be `C`")
	}
}

func TestResolveExprUnary(t *testing.T) {
	up, err := ParsePackage("_test/expr/src/ok", []string{"unary.go"})
	if err != nil {
		t.Fatal(err)
	}
	p, err := ResolvePackage(up, nil)
	if err != nil {
		t.Fatal(err)
	}

	A := p.Find("A").(*ast.Var)
	B := p.Find("B").(*ast.Var)
	x := B.Init.RHS[0].(*ast.UnaryExpr)
	if x.X != A {
		t.Error("unary expression operand must be `A`")
	}
}

func TestResolveExprBinary(t *testing.T) {
	up, err := ParsePackage("_test/expr/src/ok", []string{"binary.go"})
	if err != nil {
		t.Fatal(err)
	}
	p, err := ResolvePackage(up, nil)
	if err != nil {
		t.Fatal(err)
	}

	A := p.Find("A").(*ast.Var)
	B := p.Find("B").(*ast.Var)
	C := p.Find("C").(*ast.Var)
	x := C.Init.RHS[0].(*ast.BinaryExpr)
	if x.X != A {
		t.Error("left operand must be `A`")
	}
	if x.Y != B {
		t.Error("right operand must be `B`")
	}
}

func TestResolveExprNotTypenameError(t *testing.T) {
	srcs := []string{
		"comp.go", "conv.go", "func.go", "assert-1.go", "assert-2.go",
		"blank-1.go", "blank-2.go",
	}
	for _, src := range srcs {
		up, err := ParsePackage("_test/expr/src/err/not_typename", []string{src})
		if err != nil {
			t.Fatal(err)
		}
		_, err = ResolvePackage(up, nil)
		if err == nil ||
			!(strings.Contains(err.Error(), "is not a typename") ||
				strings.Contains(err.Error(), "is not a valid")) {
			t.Error(src, ": expected 'not a typename' error")
			if err != nil {
				t.Error(src, ": actual error:", err)
			}
		}
	}
}

func TestResolveExprNotDeclaredError(t *testing.T) {
	srcs := []string{
		"comp-1.go", "comp-2.go", "comp-3.go", "call-1.go", "call-2.go", "call-3.go",
		"conv-1.go", "conv-2.go", "conv-3.go", "sel-1.go", "sel-2.go", "index-1.go",
		"index-2.go", "slice-1.go", "slice-2.go", "slice-3.go", "slice-4.go", "unary.go",
		"binary-1.go", "binary-2.go"}
	for _, src := range srcs {
		up, err := ParsePackage("_test/expr/src/err/not_declared", []string{src})
		if err != nil {
			t.Fatal(err)
		}
		_, err = ResolvePackage(up, nil)
		if err == nil || !strings.Contains(err.Error(), "not declared") {
			t.Error(src, ": expected 'not declared' error")
			if err != nil {
				t.Error(src, ": actual error:", err)
			}
		}
	}
}

func TestResolveExprBlank(t *testing.T) {
	up, err := ParsePackage("_test/expr/src/err", []string{"blank.go"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = ResolvePackage(up, nil)
	if err == nil || !strings.Contains(err.Error(), "`_` is not a valid operand") {
		t.Error("expected 'not declared' error")
		if err != nil {
			t.Error("actual error:", err)
		}
	}
}

func TestResolveExprInvOperand(t *testing.T) {
	up, err := ParsePackage("_test/expr/src/err", []string{"inv-op.go"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = ResolvePackage(up, nil)
	if err == nil || !strings.Contains(err.Error(), "invalid operand") {
		t.Error("expected 'invalid operand' error")
		if err != nil {
			t.Error("actual error:", err)
		}
	}
}

func TestResolveExprInvConversion(t *testing.T) {
	srcs := []string{"inv-conv-1.go", "inv-conv-2.go", "inv-conv-3.go"}
	for _, src := range srcs {
		up, err := ParsePackage("_test/expr/src/err/inv_conv", []string{src})
		if err != nil {
			t.Fatal(err)
		}
		_, err = ResolvePackage(up, nil)

		if err == nil || !strings.Contains(err.Error(), "invalid conversion") {
			t.Error(src, ": expected 'invalid conversion argument' error")
			if err != nil {
				t.Error(src, ": actual error:", err)
			}
		}
	}
}

func TestResolveScope(t *testing.T) {
	up, err := ParsePackage("_test/scope/src/ok", []string{"scope.go"})
	if err != nil {
		t.Fatal(err)
	}
	p, err := ResolvePackage(up, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Very scope chain.
	if p.Parent() != ast.UniverseScope {
		t.Error("package scope must point to the universal scope")
	}
	file := p.Files[0]
	if file.Parent() != p {
		t.Error("file scope must point to package scope")
	}
	F := p.Find("F").(*ast.FuncDecl)
	fblk := F.Func.Blk
	if fblk.Parent() != file {
		t.Error("function block scope must point to file scope")
	}
	blk := fblk.Body[0].(*ast.Block)
	if blk.Parent() != fblk {
		t.Error("nested block scope must point to enclosing block scope")
	}
	blkA, blkB := blk.Body[0].(*ast.Block), blk.Body[1].(*ast.Block)
	if blkA.Parent() != blk || blkA.Parent() != blkB.Parent() {
		t.Error("nested block scope must point to enclosing block scope")
	}
	A := blkA.Find("A")
	_, _, f := A.DeclaredAt()
	if f != file {
		t.Error("incorrect source file for variable `A`")
	}
	B := blkB.Find("B")
	_, _, f = B.DeclaredAt()
	if f != file {
		t.Error("incorrect source file for variable `B`")
	}
}

func checkVarDeclared(t *testing.T, s ast.Scope, ns []string) (map[string]*ast.Var, bool) {
	m := make(map[string]*ast.Var)
	for _, n := range ns {
		d := s.Find(n)
		if d == nil {
			t.Errorf("variable `%s` not declared\n", n)
			return nil, false
		}
		v, ok := d.(*ast.Var)
		if !ok {
			t.Errorf("`%s` is not a variable\n", n)
		}
		m[n] = v
	}
	return m, true
}

func testVarDecl(t *testing.T, Fn *ast.FuncDecl, v map[string]*ast.Var) {
	// Check A and B have no initializer
	if v["A"].Init != nil {
		t.Error("`A` must be zero initialized")
	}
	if v["B"].Init != nil {
		t.Error("`B` must be zero initialized")
	}

	// Check initializer C and D have the same non-nil initializer.
	x := v["C"].Init
	if x == nil {
		t.Fatal("`C` must have an initialization statement")
	}
	if x != v["D"].Init {
		t.Error("`C` and `D` must have the same initialization statement")
	}
	// Check initialization statement is an assignment
	if x.Op != '=' || len(x.LHS) != 2 || len(x.RHS) != 2 {
		t.Error("unexpected initialization statement for `C, D`")
	}
	// Check C and D are initialized with A and B, respectively
	if x.LHS[0] != v["C"] || x.LHS[1] != v["D"] {
		t.Error("LHS of the initialization assignment must be `C, D`")
	}
	if x.RHS[0] != v["A"] || x.RHS[1] != v["B"] {
		t.Error("RHS of the initialization assignment must be `A, B`")
	}

	// Check initializer E and F have the same non-nil initializer.
	x = v["E"].Init
	if x == nil {
		t.Fatal("`E` must have an initialization statement")
	}
	if x != v["F"].Init {
		t.Error("`E` and `F` must have the same initialization statement")
	}
	// Check initialization statement is an assignment
	x = v["E"].Init
	if x.Op != '=' || len(x.LHS) != 2 || len(x.RHS) != 1 {
		t.Error("unexpected initialization statement for `E, F`")
	}
	// Check E and F are initialized with a call to Fn.
	if x.LHS[0] != v["E"] || x.LHS[1] != v["F"] {
		t.Error("LHS of the initialization assignment must be E, F`")
	}
	call, ok := x.RHS[0].(*ast.Call)
	if !ok || call.Func != Fn {
		t.Error("RHS of the initialization assignment must be a call to `Fn`")
	}
	// Check `G` has initialization statement.
	x = v["G"].Init
	if x == nil {
		t.Fatal("`G` must have an initialization statement")
	}
	// Check initialization statement is a call to Fn.
	call, ok = x.RHS[0].(*ast.Call)
	if !ok || call.Func != Fn {
		t.Error("RHS of the initialization assignment must be a call to `Fn`")
	}
}

func TestResolveVarPkgDecl(t *testing.T) {
	// Test package-level variable declarations
	up, err := ParsePackage("_test/vardecl/src/ok", []string{"decl.go"})
	if err != nil {
		t.Fatal(err)
	}
	p, err := ResolvePackage(up, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Check declarations are present and refer to variables.
	fn := p.Find("Fn").(*ast.FuncDecl)
	if v, ok := checkVarDeclared(t, p, []string{"A", "B", "C", "D", "E", "F", "G"}); ok {
		testVarDecl(t, fn, v)
	}
	// Check there is no declaration of the blank identifier (_).
	if p.Find("_") != nil {
		t.Error("variable with name `_` must not be declared")
	}

	// Test block-level varuiable declarations.
	up, err = ParsePackage("_test/vardecl/src/ok", []string{"blk-decl.go"})
	if err != nil {
		t.Fatal(err)
	}
	p, err = ResolvePackage(up, nil)
	if err != nil {
		t.Fatal(err)
	}
	s := p.Find("G").(*ast.FuncDecl).Func.Blk
	Fn := p.Find("Fn").(*ast.FuncDecl)
	// Check declarations are present and refer to variables.
	if v, ok := checkVarDeclared(t, s, []string{"A", "B", "C", "D", "E", "F", "G"}); ok {
		testVarDecl(t, Fn, v)
	}
	// Check there is no declaration of the blank identifier (_).
	if s.Find("_") != nil {
		t.Error("variable with name `_` must not be declared")
	}
}

func TestResolveVarDupDeclError(t *testing.T) {
	srcs := [][]string{
		{"dup-1.go"},
		{"dup-2.go"},
		{"dup-3.go"},
		{"blk-dup-1.go"},
		{"blk-dup-2.go"},
		{"blk-dup-3.go"},
	}
	for _, src := range srcs {
		up, err := ParsePackage("_test/vardecl/src/err", src)
		if err != nil {
			t.Fatal(err)
		}
		_, err = ResolvePackage(up, nil)
		if err == nil || !strings.Contains(err.Error(), "redeclared") {
			t.Error(src, ": expected 'redeclared' error")
			if err != nil {
				t.Error(src, ": actual error:", err)
			}
		}
	}
}

func TestResolveVarPkgNotDeclError(t *testing.T) {
	srcs := []string{
		"not-decl-1.go", "not-decl-2.go", "not-decl-3.go",
		"blk-not-decl-1.go", "blk-not-decl-2.go", "blk-not-decl-3.go",
	}
	for _, src := range srcs {
		up, err := ParsePackage("_test/vardecl/src/err", []string{src})
		if err != nil {
			t.Fatal(err)
		}
		_, err = ResolvePackage(up, nil)
		if err == nil || !strings.Contains(err.Error(), "not declared") {
			t.Error(src, ": expected 'not declared' error")
			if err != nil {
				t.Error(src, ": actual error:", err)
			}
		}
	}
}

func checkConstDeclared(
	t *testing.T, s ast.Scope, ns []string) (map[string]*ast.Const, bool) {

	m := make(map[string]*ast.Const)
	for _, n := range ns {
		d := s.Find(n)
		if d == nil {
			t.Errorf("constant `%s` not declared\n", n)
			return nil, false
		}
		c, ok := d.(*ast.Const)
		if !ok {
			t.Errorf("`%s` is not a constant\n", n)
		}
		m[n] = c
	}
	return m, true
}

func testConstDecl(t *testing.T, c map[string]*ast.Const) {
	x := c["A"].Init
	if x == nil {
		t.Error("`A` missing initializer expression")
	}
	if _, ok := x.(*ast.Literal); !ok {
		t.Error("`A` initializer not a literal")
	}
	if c["A"].Type != nil {
		t.Error("`A` must not have a type yet")
	}

	x = c["B"].Init
	if x == nil {
		t.Error("`B` missing initializer expression")
	}
	if x != c["A"] {
		t.Error("`B` initializer is not `A`")
	}
	typ := c["B"].Type
	if typ == nil {
		t.Error("`B` must have a type")
	}
	if td, ok := typ.(*ast.TypeDecl); !ok || td.Type != ast.BuiltinInt {
		t.Error("`B` must have type `int`")
	}

	x = c["C"].Init
	if x == nil {
		t.Error("`C` missing initializer expression")
	}
	if x != c["A"] {
		t.Error("`C` initializer is not `A`")
	}
	typ = c["C"].Type
	if typ != nil {
		t.Error("`C` must not have a type yet")
	}

	x = c["D"].Init
	if x == nil {
		t.Error("`D` missing initializer expression")
	}
	if x != c["B"] {
		t.Error("`D` initializer is not `B`")
	}
	typ = c["D"].Type
	if typ != nil {
		t.Error("`D` must not have a type yet")
	}

	x = c["E"].Init
	if x == nil {
		t.Error("`E` missing initializer expression")
	}
	if _, ok := x.(*ast.Literal); !ok {
		t.Error("`E` initializer not a literal")
	}
	typ = c["E"].Type
	if typ == nil {
		t.Error("`E` must have a type")
	}
	if td, ok := typ.(*ast.TypeDecl); !ok || td.Type != ast.BuiltinString {
		t.Error("`E` must have type `string`")
	}

	x = c["F"].Init
	if x == nil {
		t.Error("`F` missing initializer expression")
	}
	if _, ok := x.(*ast.Literal); !ok {
		t.Error("`F` initializer not a literal")
	}
	typ = c["F"].Type
	if typ == nil {
		t.Error("`F` must have a type")
	}
	if td, ok := typ.(*ast.TypeDecl); !ok || td.Type != ast.BuiltinString {
		t.Error("`F` must have type `string`")
	}
}

func TestResolveConstDecl(t *testing.T) {
	up, err := ParsePackage("_test/constdecl/src/ok", []string{"decl.go"})
	if err != nil {
		t.Fatal(err)
	}
	p, err := ResolvePackage(up, nil)
	if err != nil {
		t.Fatal(err)
	}

	if c, ok := checkConstDeclared(t, p, []string{"A", "B", "C", "D", "E", "F"}); ok {
		testConstDecl(t, c)
	}
	if p.Find("_") != nil {
		t.Error("const with name `_` must not be declared")
	}

	fn := p.Find("Fn").(*ast.FuncDecl)
	s := fn.Func.Blk
	if c, ok := checkConstDeclared(t, s, []string{"A", "B", "C", "D", "E", "F"}); ok {
		testConstDecl(t, c)
	}
	if s.Find("_") != nil {
		t.Error("const with name `_` must not be declared")
	}
}

func testConstDeclGroup(t *testing.T, c map[string]*ast.Const) {
	x := c["A"].Init
	if x == nil {
		t.Error("`A` missing initializer expression")
	}
	if _, ok := x.(*ast.Literal); !ok {
		t.Error("`A` initializer not a literal")
	}
	if c["A"].Type != nil {
		t.Error("`A` must not have a type yet")
	}
	if c["A"].Iota != 0 {
		t.Error("`A` must have iota 0")
	}

	x = c["B"].Init
	if x == nil {
		t.Error("`B` missing initializer expression")
	}
	if x != c["A"].Init {
		t.Error("`B` must have the same initializer as `A`")
	}
	typ := c["B"].Type
	if typ != nil {
		t.Error("`B` must not have a type")
	}
	if c["B"].Iota != 1 {
		t.Error("`B` must have iota 1")
	}

	x = c["C"].Init
	if x == nil {
		t.Error("`C` missing initializer expression")
	}
	if x != c["A"] {
		t.Error("`C` initializer is not `A`")
	}
	typ = c["C"].Type
	if typ == nil {
		t.Error("`C` must have a type")
	}
	if td, ok := typ.(*ast.TypeDecl); !ok || td.Type != ast.BuiltinInt {
		t.Error("`C` must have type `int`")
	}
	if c["C"].Iota != 2 {
		t.Error("`C` must have iota 2")
	}

	x = c["D"].Init
	if x == nil {
		t.Error("`D` missing initializer expression")
	}
	if x != c["B"] {
		t.Error("`D` initializer is not `B`")
	}
	typ = c["D"].Type
	if typ != c["C"].Type {
		t.Error("`D` must have the same type as `C`")
	}
	if c["D"].Iota != c["C"].Iota {
		t.Error("`D` must have iota the same iota as `C`")
	}

	x = c["E"].Init
	if x == nil {
		t.Error("`E` missing initializer expression")
	}
	if x != c["C"].Init {
		t.Error("`E` must have the same initializer as `C`")
	}
	if c["E"].Type != c["C"].Type {
		t.Error("`E` must have the same type as `C`")
	}
	if c["E"].Iota != 3 {
		t.Error("`E` must have iota 3")
	}

	x = c["F"].Init
	if x == nil {
		t.Error("`E` missing initializer expression")
	}
	if x != c["D"].Init {
		t.Error("`F` must have the same initializer as `D`")
	}
	if c["F"].Type != c["D"].Type {
		t.Error("`F` must have the same type as `D`")
	}
	if c["F"].Iota != c["E"].Iota {
		t.Error("`F` must have iota the same iota as `E`")
	}
}

func TestResolveConstDeclGroup(t *testing.T) {
	up, err := ParsePackage("_test/constdecl/src/ok", []string{"decl-group.go"})
	if err != nil {
		t.Fatal(err)
	}
	p, err := ResolvePackage(up, nil)
	if err != nil {
		t.Fatal(err)
	}

	if c, ok := checkConstDeclared(t, p, []string{"A", "B", "C", "D", "E", "F"}); ok {
		testConstDeclGroup(t, c)
	}

	fn := p.Find("Fn").(*ast.FuncDecl)
	s := fn.Func.Blk
	if c, ok := checkConstDeclared(t, s, []string{"A", "B", "C", "D", "E", "F"}); ok {
		testConstDeclGroup(t, c)
	}
}

func TestResolveConstDupDeclError(t *testing.T) {
	srcs := []string{
		"dup-decl-1.go", "dup-decl-2.go", "dup-decl-3.go", "dup-decl-4.go",
		"dup-decl-5.go", "dup-decl-6.go", "dup-decl-7.go",
	}
	for _, src := range srcs {
		up, err := ParsePackage("_test/constdecl/src/err", []string{src})
		if err != nil {
			t.Fatal(err)
		}
		_, err = ResolvePackage(up, nil)
		if err == nil || !strings.Contains(err.Error(), "redeclared") {
			t.Error(src, ": expected 'redeclared' error")
			if err != nil {
				t.Error(src, ": actual error:", err)
			}
		}
	}
}

func TestResolveConstNotDeclError(t *testing.T) {
	srcs := []string{
		"not-decl-1.go", "not-decl-2.go", "not-decl-3.go", "not-decl-4.go",
		"not-decl-5.go", "not-decl-6.go", "not-decl-7.go", "not-decl-8.go",
	}
	for _, src := range srcs {
		up, err := ParsePackage("_test/constdecl/src/err", []string{src})
		if err != nil {
			t.Fatal(err)
		}
		_, err = ResolvePackage(up, nil)
		if err == nil || !strings.Contains(err.Error(), "not declared") {
			t.Error(src, ": expected 'not declared' error")
			if err != nil {
				t.Error(src, ": actual error:", err)
			}
		}
	}
}

func TestResolveConstCountMismatch(t *testing.T) {
	srcs := []string{
		"not-eq-1.go", "not-eq-2.go", "not-eq-3.go", "not-eq-4.go", "not-eq-5.go",
	}
	for _, src := range srcs {
		up, err := ParsePackage("_test/constdecl/src/err", []string{src})
		if err != nil {
			t.Fatal(err)
		}
		_, err = ResolvePackage(up, nil)
		if err == nil || !strings.Contains(err.Error(), "must be equal") {
			t.Error(src, ": expected 'must be equal' error")
			if err != nil {
				t.Error(src, ": actual error:", err)
			}
		}
	}
}

func TestResolveFuncDecl(t *testing.T) {
	up, err := ParsePackage("_test/funcdecl/src/ok", []string{"decl.go"})
	if err != nil {
		t.Fatal(err)
	}
	p, err := ResolvePackage(up, nil)
	if err != nil {
		t.Fatal(err)
	}

	F := p.Find("F").(*ast.FuncDecl)
	if F == nil {
		t.Fatal("`F` not declared")
	}

	// Check parameters are declared.
	v, ok := checkVarDeclared(t, F.Func.Blk, []string{"a", "b", "c", "d", "e", "f", "g"})
	if !ok {
		return
	}

	A := p.Find("A").(*ast.TypeDecl)
	B := p.Find("B").(*ast.TypeDecl)
	C := p.Find("C").(*ast.TypeDecl)

	func(t *testing.T, ts []*ast.TypeDecl, vs []*ast.Var) {
		for i := range ts {
			if ts[i] != vs[i].Type {
				t.Errorf("type of `%s` must be `%s`\n", vs[i].Name, ts[i].Name)
			}
		}
	}(t,
		[]*ast.TypeDecl{A, B, C, C, A, B, B},
		[]*ast.Var{v["a"], v["b"], v["c"], v["d"], v["e"], v["f"], v["g"]},
	)

	// Check blank function name is not declared.
	if p.Find("_") != nil {
		t.Error("function name `_` must not be declared")
	}

	// Check blank parameter or return value is not declared.
	if F.Func.Blk.Find("_") != nil {
		t.Error("blank parameter or return value must not be declared")
	}
}

func TestResolveFuncDeclDupParam(t *testing.T) {
	srcs := []string{"dup-1.go", "dup-2.go", "dup-3.go", "dup-4.go"}
	for _, src := range srcs {
		up, err := ParsePackage("_test/funcdecl/src/err", []string{src})
		if err != nil {
			t.Fatal(err)
		}
		_, err = ResolvePackage(up, nil)
		if err == nil || !strings.Contains(err.Error(), "redeclared") {
			t.Error(src, ": expected 'redeclared' error")
			if err != nil {
				t.Error(src, ": actual error:", err)
			}
		}
	}
}
