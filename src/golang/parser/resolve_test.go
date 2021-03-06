package parser

import (
	"bufio"
	"errors"
	"golang/ast"
	"golang/pdb"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func compilePackage(
	dir string, srcs []string, loc ast.PackageLocator) (*ast.Package, error) {

	p, err := ParsePackage(dir, srcs)
	if err != nil {
		return nil, err
	}
	if err := ResolvePackage(p, loc); err != nil {
		return nil, err
	}
	return p, err
}

func getTypeDecl(s ast.Scope, name string) *ast.TypeDecl {
	sym := s.Find(name)
	if sym == nil {
		return nil
	}
	t, _ := sym.(*ast.TypeDecl)
	return t
}

func expectErrorWithLoc(
	t *testing.T, pkg string, srcs []string, loc ast.PackageLocator, msg string) {

	_, err := compilePackage(pkg, srcs, loc)
	if err == nil || !strings.Contains(err.Error(), msg) {
		t.Errorf("%s:%v: expected `%s` error", pkg, srcs, msg)
		if err == nil {
			t.Log("actual: no error")
		} else {
			t.Logf("actual: %s", err.Error())
		}
	}
}

func expectError(t *testing.T, pkg string, srcs []string, msg string) {
	expectErrorWithLoc(t, pkg, srcs, nil, msg)
}

func TestResolveTypeUniverse(t *testing.T) {
	const TESTDIR = "_test/typedecl/src/correct"
	p, err := compilePackage(TESTDIR, []string{"a.go", "b.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		decl, ref, file string
		typ             ast.Type
	}{
		{"a", "bool", "a.go", ast.BuiltinBool},
		{"b", "byte", "a.go", ast.BuiltinUint8},
		{"c", "uint8", "a.go", ast.BuiltinUint8},
		{"d", "uint16", "a.go", ast.BuiltinUint16},
		{"e", "uint32", "a.go", ast.BuiltinUint32},
		{"f", "uint64", "a.go", ast.BuiltinUint64},
		{"g", "int8", "a.go", ast.BuiltinInt8},
		{"h", "int16", "a.go", ast.BuiltinInt16},
		{"i", "rune", "a.go", ast.BuiltinInt32},
		{"j", "int32", "a.go", ast.BuiltinInt32},
		{"k", "int64", "a.go", ast.BuiltinInt64},
		{"l", "float32", "b.go", ast.BuiltinFloat32},
		{"m", "float64", "b.go", ast.BuiltinFloat64},
		{"n", "complex64", "b.go", ast.BuiltinComplex64},
		{"o", "complex128", "b.go", ast.BuiltinComplex128},
		{"p", "uint", "b.go", ast.BuiltinUint},
		{"q", "int", "b.go", ast.BuiltinInt},
		{"r", "uintptr", "b.go", ast.BuiltinUintptr},
		{"s", "string", "b.go", ast.BuiltinString},
	}
	for _, cs := range cases {
		// Test type is declared at package scope.
		sym := p.Syms[cs.decl]
		if sym == nil {
			t.Fatalf("name `%s` not found at package scope\n", cs.decl)
		}
		dcl, ok := sym.(*ast.TypeDecl)
		if !ok {
			t.Fatalf("`%s` must be a TypeDecl\n", cs.decl)
		}
		// Test type is registered under its name.
		name := sym.Id()
		_, file := sym.DeclaredAt()
		if name != cs.decl {
			t.Errorf("type `%s` declared as `%s`\n", cs.decl, name)
		}
		// Test type is declared in the respective file.
		if file == nil || filepath.Base(file.Name) != cs.file {
			t.Errorf("type '%s' must be declared in file `%s`", cs.decl, cs.file)
		}
		// Test type declaration refers to the respective builtin type literals.
		typ, ok := dcl.Type.(*ast.BuiltinType)
		if !ok {
			t.Fatalf("declaration of `%s` does not refer to a builtin type\n", cs.decl)
		}
		if typ != cs.typ {
			t.Errorf("`%s` refers to an invalid builtin type: %d\n", cs.decl, typ.Kind)
		}
	}
}

func TestResolveTypeDuplicatelAtPackageScope(t *testing.T) {
	expectError(t, "_test/typedecl/src/errors/dup_decl", []string{"a.go"}, "redeclared")
	expectError(
		t, "_test/typedecl/src/errors/dup_decl", []string{"b.go", "c.go"}, "redeclared")
}

func TestResolveTypePackageScope(t *testing.T) {
	const TESTDIR = "_test/typedecl/src/correct"
	p, err := compilePackage(TESTDIR, []string{"a.go", "b.go"}, nil)
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
				"declaration of `%s` does not refer to the declaration of `%s`\n",
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
	p, err := compilePackage(TESTDIR, []string{"a.go", "b.go"}, nil)
	if err != nil {
		t.Fatal(err)
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
	if tdA.Type != ast.BuiltinInt {
		t.Error("declaration of `A` does not refer to predeclared `int`")
	}

	sym = fn.Func.Blk.Find("AA")
	tdAA, _ := sym.(*ast.TypeDecl)
	if sym == nil || tdAA == nil {
		t.Fatal("type declaration `AA` not found at `F`s scope")
	}
	dcl, ok := tdAA.Type.(*ast.TypeDecl)
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
	if tdB.Type != ast.BuiltinInt {
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
	p, err := compilePackage(TESTDIR, []string{"self.go"}, nil)
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
	p, err := compilePackage(TESTDIR, []string{"a.go", "b.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	D := p.Find("D").(*ast.Const)
	a := getTypeDecl(p, "a")
	u := getTypeDecl(p, "u")
	if u.Type.(*ast.ArrayType).Elt != a {
		t.Error("elements of `u` are not of type `a`")
	}
	op := u.Type.(*ast.ArrayType).Dim.(*ast.OperandName)
	if op.Decl != D {
		t.Error("the dimension of `u` is not `D`")
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
	if z.Fields[2].Type != v {
		t.Error("field `z.Z` is not of type `v`")
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
	mF := IfA.Type.(*ast.InterfaceType).Methods[0].Func.Sig
	if mF.Params[0].Type != u {
		t.Error("parameterof method `IfA.F` is not of type `u`")
	}
	if mF.Returns[0].Type != v {
		t.Error("return of method `IfA.F` is not of type `v`")
	}

	IfB := getTypeDecl(p, "IfB").Type.(*ast.InterfaceType)
	if IfB.Embedded[0] != IfA {
		t.Error("embedded interface of `IfB` is not `IfA`")
	}
	mG := IfB.Methods[0].Func.Sig
	if mG.Params[0].Type != x {
		t.Error("parameterof method `IfB.G` is not of type `x`")
	}
	if mG.Returns[0].Type != y {
		t.Error("return of method `IfB.G` is not of type `y`")
	}

	expectError(t, "_test/typedecl/src/errors", []string{"iface.go"}, "blank method")
}

func TestResolveFuncTypeError(t *testing.T) {
	for _, src := range []string{"func-1.go", "func-2.go"} {
		expectError(t, "_test/typedecl/src/errors", []string{src}, "all be present")
	}
}

func TestResolveTypeTypenameNotFound(t *testing.T) {
	for _, src := range []string{
		"typename.go", "array.go", "slice.go", "ptr.go", "map-1.go", "map-2.go",
		"chan.go", "struct.go", "func-1.go", "func-2.go", "iface-1.go", "iface-2.go",
	} {
		expectError(t, "_test/typedecl/src/errors/typename_not_found", []string{src},
			"X not declared")
	}
}

func TestResolveTypeNotATypename(t *testing.T) {
	for _, src := range []string{
		"typename.go", "array.go", "slice.go", "ptr.go", "map-1.go", "map-2.go",
		"chan.go", "struct.go", "func-1.go", "func-2.go", "iface-1.go", "iface-2.go",
		"blank.go",
	} {
		expectError(t, "_test/typedecl/src/errors/not_typename", []string{src},
			"is not a typename")
	}
}

func TestResolveExprLiteral(t *testing.T) {
	_, err := compilePackage("_test/expr/src/ok", []string{"lit.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestResolveExprComposite(t *testing.T) {
	p, err := compilePackage("_test/expr/src/ok", []string{"comp.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	A := p.Find("A").(*ast.Var)
	B := p.Find("B").(*ast.Var)

	C := p.Find("C").(*ast.Var)
	x := C.Init.RHS[0].(*ast.CompLiteral)
	elt := x.Typ.(*ast.SliceType).Elt
	if elt != ast.BuiltinInt {
		t.Error("`C`s initializer type must be `[]int`")
	}
	if op, ok := x.Elts[0].Elt.(*ast.OperandName); !ok || op.Decl != A {
		t.Error("first element in the slice literal must be `A`")
	}
	if op, ok := x.Elts[1].Elt.(*ast.OperandName); !ok || op.Decl != B {
		t.Error("first element in the slice literal must be `B`")
	}

	D := p.Find("D").(*ast.Var)
	x = D.Init.RHS[0].(*ast.CompLiteral)
	if op, ok := x.Elts[0].Key.(*ast.OperandName); !ok || op.Decl != A {
		t.Error("key in the map literal must be `A`")
	}
	if op, ok := x.Elts[0].Elt.(*ast.OperandName); !ok || op.Decl != B {
		t.Error("value in the map literal must be `B`")
	}

	E := p.Find("E").(*ast.Var)
	x = E.Init.RHS[0].(*ast.CompLiteral)
	if id, ok := x.Elts[0].Key.(*ast.QualifiedId); !ok {
		t.Error("field names in struct composite literal must not be resolved")
	} else if id.Id != "A" {
		t.Errorf("unexpected fied name `%s`, must be `A`\n", id.Id)
	}
	if op, ok := x.Elts[0].Elt.(*ast.OperandName); !ok || op.Decl != A {
		t.Error("initializer of fieeld `A` must be variable `A`")
	}
	if id, ok := x.Elts[1].Key.(*ast.QualifiedId); !ok {
		t.Error("field names in struct composite literal must not be resolved")
	} else if id.Id != "B" {
		t.Errorf("unexpected fied name `%s`, must be `B`\n", id.Id)
	}
	if op, ok := x.Elts[1].Elt.(*ast.OperandName); !ok || op.Decl != B {
		t.Error("initializer of fieeld `B` must be variable `B`")
	}
}

func TestResolveExprConversion(t *testing.T) {
	p, err := compilePackage("_test/expr/src/ok", []string{"conv.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	T := p.Find("T").(*ast.TypeDecl)
	A := p.Find("A").(*ast.Var)
	B := p.Find("B").(*ast.Var)
	x := B.Init.RHS[0].(*ast.Conversion)
	if x.Typ.(*ast.SliceType).Elt != T {
		t.Error("type in the conversion must be a `[]T`")
	}
	if op, ok := x.X.(*ast.OperandName); !ok || op.Decl != A {
		t.Error("the converted value must be `A`")
	}

	C := p.Find("C").(*ast.Var)
	D := p.Find("D").(*ast.Var)
	x, ok := D.Init.RHS[0].(*ast.Conversion)
	if !ok {
		t.Error("initializer of `D` must be Conversion")
	}
	if x.Typ != T {
		t.Error("in initializer of `D` the conversion type must be `T`")
	}
	if op, ok := x.X.(*ast.OperandName); !ok || op.Decl != C {
		t.Error("in initializer of `D` the converted value must be `C`")
	}
}

func TestResolveExprCall(t *testing.T) {
	p, err := compilePackage("_test/expr/src/ok", []string{"call.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	F := p.Find("F").(*ast.FuncDecl)
	A := p.Find("A").(*ast.Var)
	if A.Init == nil {
		t.Error("`A` must have an initializer statement")
	}
	x := A.Init.RHS[0].(*ast.Call)
	if x.Func.(*ast.OperandName).Decl != F {
		t.Error("in the initializer of `A` the called function must be `F`")
	}
	X := p.Find("X").(*ast.Var)
	Y := p.Find("Y").(*ast.Var)
	if op, ok := x.Xs[0].(*ast.OperandName); !ok || op.Decl != X {
		t.Error("first argument to the call of `F` must be `X`")
	}
	if op, ok := x.Xs[1].(*ast.OperandName); !ok || op.Decl != Y {
		t.Error("second argument to the call of `F` must be `Y`")
	}

	B := p.Find("B").(*ast.Var)
	x = B.Init.RHS[0].(*ast.Call)
	if x.ATyp != ast.BuiltinInt {
		t.Error("in initializer call of `B`: first argment must be type `int`")
	}
	if len(x.Xs) != 1 {
		t.Error("in initializer of `B`; call should have one expression argument")
	}

	T := p.Find("T").(*ast.TypeDecl)
	C := p.Find("C").(*ast.Var)
	x = C.Init.RHS[0].(*ast.Call)
	if x.ATyp != T {
		t.Error("in initializer `C`: first argment must be type `int`")
	}
	if len(x.Xs) != 0 {
		t.Error("in initializer of `C`: call should have zero expression arguments")
	}

	D := p.Find("D").(*ast.Var)
	x = D.Init.RHS[0].(*ast.Call)
	if s, ok := x.ATyp.(*ast.SliceType); !ok || s.Elt != ast.BuiltinInt {
		t.Error("in initializer `D`: first argment must be type `[]int`")
	}
	if len(x.Xs) != 2 {
		t.Error("in initializer of `D`: call should have two expression arguments")
	}
}

func TestResolveExprParens(t *testing.T) {
	p, err := compilePackage("_test/expr/src/ok", []string{"parens.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	X := p.Find("X").(*ast.Var)
	Y := p.Find("Y").(*ast.Var)
	x := Y.Init.RHS[0]
	if _, ok := x.(*ast.ParensExpr); ok {
		t.Error("ParensExpr must be removed by the resolver")
	}
	if op, ok := x.(*ast.OperandName); !ok || op.Decl != X {
		t.Error("initializer or `Y` must be `X`")
	}
}

func TestResolveExprFuncLiteral(t *testing.T) {
	p, err := compilePackage("_test/expr/src/ok", []string{"func.go"}, nil)
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

	x := fn.Blk.Find("x")
	if x == nil {
		t.Error("parameter `x` must be declared at function block scope")
	}
	x = fn.Blk.Find("y")
	if x == nil {
		t.Error("parameter `y` must be declared at function block scope")
	}
	x = fn.Blk.Find("u")
	if x == nil {
		t.Error("return `u` must be declared at function block scope")
	}
	x = fn.Blk.Find("v")
	if x == nil {
		t.Error("return `v` must be declared at function block scope")
	}
}

func TestResolveTypeAssertion(t *testing.T) {
	p, err := compilePackage("_test/expr/src/ok", []string{"assert.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	T := p.Find("T").(*ast.TypeDecl)
	A := p.Find("A").(*ast.Var)
	B := p.Find("B").(*ast.Var)
	x := B.Init.RHS[0].(*ast.TypeAssertion)
	if op, ok := x.X.(*ast.OperandName); !ok || op.Decl != A {
		t.Error("type assertion expression must be `A`")
	}
	if x.ATyp != T {
		t.Error("type in type assertion expression must be `T`")
	}
}

func TestResolveExprSelector(t *testing.T) {
	p, err := compilePackage("_test/expr/src/ok", []string{"selector.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	A := p.Find("A").(*ast.Var)
	B := p.Find("B").(*ast.Var)
	x := B.Init.RHS[0].(*ast.Selector)
	if op, ok := x.X.(*ast.OperandName); !ok || op.Decl != A {
		t.Error("expression in selector must be `A`")
	}
	if x.Id != "X" {
		t.Error("identifier in selector must be `X`")
	}

	C := p.Find("C").(*ast.Var)
	D := p.Find("D").(*ast.Var)
	x = D.Init.RHS[0].(*ast.Selector)
	y := x.X.(*ast.Selector)
	if op, ok := y.X.(*ast.OperandName); !ok || op.Decl != C {
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
	ptr, ok := z.RTyp.(*ast.PtrType)
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
	p, err := compilePackage("_test/expr/src/ok", []string{"index.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	A := p.Find("A").(*ast.Var)
	B := p.Find("B").(*ast.Const)
	C := p.Find("C").(*ast.Var)
	x := C.Init.RHS[0].(*ast.IndexExpr)
	if op, ok := x.X.(*ast.OperandName); !ok || op.Decl != A {
		t.Error("indexed obejct must be `A`")
	}
	if op, ok := x.I.(*ast.OperandName); !ok || op.Decl != B {
		t.Error("index expression must be `A`")
	}
}

func TestResolveExprSlice(t *testing.T) {
	p, err := compilePackage("_test/expr/src/ok", []string{"slice.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	A := p.Find("A").(*ast.Var)
	B := p.Find("B").(*ast.Var)
	L := p.Find("L").(*ast.Var)
	H := p.Find("H").(*ast.Var)
	C := p.Find("C").(*ast.Var)
	x := B.Init.RHS[0].(*ast.SliceExpr)
	if op, ok := x.X.(*ast.OperandName); !ok || op.Decl != A {
		t.Error("sliced object must be `A`")
	}
	if op, ok := x.Lo.(*ast.OperandName); !ok || op.Decl != L {
		t.Error("slice expression low index must be `L`")
	}
	if op, ok := x.Hi.(*ast.OperandName); !ok || op.Decl != H {
		t.Error("slice expression high index must be `H`")
	}
	if op, ok := x.Cap.(*ast.OperandName); !ok || op.Decl != C {
		t.Error("slice expression capacity be `C`")
	}
}

func TestResolveExprUnary(t *testing.T) {
	p, err := compilePackage("_test/expr/src/ok", []string{"unary.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	A := p.Find("A").(*ast.Var)
	B := p.Find("B").(*ast.Var)
	x := B.Init.RHS[0].(*ast.UnaryExpr)
	if op, ok := x.X.(*ast.OperandName); !ok || op.Decl != A {
		t.Error("unary expression operand must be `A`")
	}
}

func TestResolveExprBinary(t *testing.T) {
	p, err := compilePackage("_test/expr/src/ok", []string{"binary.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	A := p.Find("A").(*ast.Var)
	B := p.Find("B").(*ast.Var)
	C := p.Find("C").(*ast.Var)
	x := C.Init.RHS[0].(*ast.BinaryExpr)
	if op, ok := x.X.(*ast.OperandName); !ok || op.Decl != A {
		t.Error("left operand must be `A`")
	}
	if op, ok := x.Y.(*ast.OperandName); !ok || op.Decl != B {
		t.Error("right operand must be `B`")
	}
}

func TestResolveExprNotTypenameError(t *testing.T) {
	srcs := []string{
		"comp.go",
		"conv.go",
		"func.go", "assert-1.go", "assert-2.go",
	}
	for _, src := range srcs {
		expectError(t, "_test/expr/src/err/not_typename", []string{src},
			"is not a typename")
	}
	srcs = []string{
		"blank-1.go", "blank-2.go",
	}
	for _, src := range srcs {
		expectError(t, "_test/expr/src/err/not_typename", []string{src},
			"is not a valid operand or a typename")
	}
}

func TestResolveExprNotDeclaredError(t *testing.T) {
	srcs := []string{
		"comp-1.go", "comp-2.go", "comp-3.go", "call-1.go", "call-2.go", "call-3.go",
		"conv-1.go", "conv-2.go", "conv-3.go", "sel-1.go", "sel-2.go", "index-1.go",
		"index-2.go", "slice-1.go", "slice-2.go", "slice-3.go", "slice-4.go", "unary.go",
		"binary-1.go", "binary-2.go"}
	for _, src := range srcs {
		expectError(t, "_test/expr/src/err/not_declared", []string{src}, "not declared")
	}
}

func TestResolveExprBlank(t *testing.T) {
	expectError(t, "_test/expr/src/err", []string{"blank.go"},
		"`_` is not a valid operand")
}

func TestResolveExprInvOperand(t *testing.T) {
	expectError(t, "_test/expr/src/err", []string{"inv-op.go"}, "invalid operand")
}

func TestResolveExprInvConversion(t *testing.T) {
	srcs := []string{"inv-conv-1.go", "inv-conv-2.go", "inv-conv-3.go"}
	for _, src := range srcs {
		expectError(t, "_test/expr/src/err/inv_conv", []string{src},
			"invalid conversion argument")
	}
}

func TestResolveScope(t *testing.T) {
	p, err := compilePackage("_test/scope/src/ok", []string{"scope.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Verify scope chain.
	if p.Parent() != ast.UniverseScope {
		t.Error("package scope must point to the universal scope")
	}
	file := p.Files[0]
	if file.Parent() != p {
		t.Error("file scope must point to package scope")
	}
	F := p.Find("F").(*ast.FuncDecl)
	if F.Func.Parent() != file {
		t.Error("function scope must point to file scope")
	}
	fblk := F.Func.Blk
	if fblk.Parent() != &F.Func {
		t.Error("function block scope must point to function scope")
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
	_, f := A.DeclaredAt()
	if f != file {
		t.Error("incorrect source file for variable `A`")
	}
	B := blkB.Find("B")
	_, f = B.DeclaredAt()
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
	// Check A and B have an initializer with an empty RHS.
	if v["A"].Init != nil && len(v["A"].Init.RHS) > 0 {
		t.Error("`A` must be zero initialized")
	}
	if v["B"].Init != nil && len(v["B"].Init.RHS) > 0 {
		t.Error("`B` must be zero initialized")
	}

	// Check A and B have type `int`.
	int := ast.BuiltinInt
	if v["A"].Type != int {
		t.Error("`A` type is not the predeclared `int`")
	}
	if v["B"].Type != int {
		t.Error("`B` type is not the predeclared `int`")
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
	if x.Op != ast.NOP || len(x.LHS) != 2 || len(x.RHS) != 2 {
		t.Error("unexpected initialization statement for `C, D`")
	}
	// Check C and D are initialized with A and B, respectively
	op0, ok0 := x.LHS[0].(*ast.OperandName)
	op1, ok1 := x.LHS[1].(*ast.OperandName)
	if !ok0 || !ok1 || op0.Decl != v["C"] || op1.Decl != v["D"] {
		t.Error("LHS of the initialization assignment must be `C, D`")
	}
	op0, ok0 = x.RHS[0].(*ast.OperandName)
	op1, ok1 = x.RHS[1].(*ast.OperandName)
	if !ok0 || !ok1 || op0.Decl != v["A"] || op1.Decl != v["B"] {
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
	if x.Op != ast.NOP || len(x.LHS) != 2 || len(x.RHS) != 1 {
		t.Error("unexpected initialization statement for `E, F`")
	}
	// Check E and F are initialized with a call to Fn.
	if x.LHS[0].(*ast.OperandName).Decl != v["E"] ||
		x.LHS[1].(*ast.OperandName).Decl != v["F"] {
		t.Error("LHS of the initialization assignment must be E, F`")
	}
	call, ok := x.RHS[0].(*ast.Call)
	if !ok || call.Func.(*ast.OperandName).Decl != Fn {
		t.Error("RHS of the initialization assignment must be a call to `Fn`")
	}
	// Check `G` has initialization statement.
	x = v["G"].Init
	if x == nil {
		t.Fatal("`G` must have an initialization statement")
	}
	// Check initialization statement is a call to Fn.
	call, ok = x.RHS[0].(*ast.Call)
	if !ok || call.Func.(*ast.OperandName).Decl != Fn {
		t.Error("RHS of the initialization assignment must be a call to `Fn`")
	}
}

func TestResolveVarPkgDecl(t *testing.T) {
	// Test package-level variable declarations
	p, err := compilePackage("_test/vardecl/src/ok", []string{"decl.go"}, nil)
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

	// Test block-level variable declarations.
	p, err = compilePackage("_test/vardecl/src/ok", []string{"blk-decl.go"}, nil)
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
		expectError(t, "_test/vardecl/src/err", src, "redeclared")
	}
}

func TestResolveVarPkgNotDeclError(t *testing.T) {
	srcs := []string{
		"not-decl-1.go", "not-decl-2.go", "not-decl-3.go",
		"blk-not-decl-1.go", "blk-not-decl-2.go", "blk-not-decl-3.go",
	}
	for _, src := range srcs {
		expectError(t, "_test/vardecl/src/err", []string{src}, "not declared")
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
	if _, ok := x.(*ast.ConstValue); !ok {
		t.Error("`A` initializer not a literal")
	}
	if c["A"].Type != nil {
		t.Error("`A` must not have a type yet")
	}

	x = c["B"].Init
	if x == nil {
		t.Error("`B` missing initializer expression")
	}
	if x.(*ast.OperandName).Decl != c["A"] {
		t.Error("`B` initializer is not `A`")
	}
	typ := c["B"].Type
	if typ == nil {
		t.Error("`B` must have a type")
	}
	if typ != ast.BuiltinInt {
		t.Error("`B` must have type `int`")
	}

	x = c["C"].Init
	if x == nil {
		t.Error("`C` missing initializer expression")
	}
	if x.(*ast.OperandName).Decl != c["A"] {
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
	if x.(*ast.OperandName).Decl != c["B"] {
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
	if _, ok := x.(*ast.ConstValue); !ok {
		t.Error("`E` initializer not a literal")
	}
	typ = c["E"].Type
	if typ == nil {
		t.Error("`E` must have a type")
	}
	if typ != ast.BuiltinString {
		t.Error("`E` must have type `string`")
	}

	x = c["F"].Init
	if x == nil {
		t.Error("`F` missing initializer expression")
	}
	if _, ok := x.(*ast.ConstValue); !ok {
		t.Error("`F` initializer not a literal")
	}
	typ = c["F"].Type
	if typ == nil {
		t.Error("`F` must have a type")
	}
	if typ != ast.BuiltinString {
		t.Error("`F` must have type `string`")
	}
}

func TestResolveConstDecl(t *testing.T) {
	p, err := compilePackage("_test/constdecl/src/ok", []string{"decl.go"}, nil)
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
	if _, ok := x.(*ast.ConstValue); !ok {
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
	if x.(*ast.OperandName).Decl != c["A"] {
		t.Error("`C` initializer is not `A`")
	}
	typ = c["C"].Type
	if typ == nil {
		t.Error("`C` must have a type")
	}
	if typ != ast.BuiltinInt {
		t.Error("`C` must have type `int`")
	}
	if c["C"].Iota != 2 {
		t.Error("`C` must have iota 2")
	}

	x = c["D"].Init
	if x == nil {
		t.Error("`D` missing initializer expression")
	}
	if x.(*ast.OperandName).Decl != c["B"] {
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
	p, err := compilePackage("_test/constdecl/src/ok", []string{"decl-group.go"}, nil)
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
		expectError(t, "_test/constdecl/src/err", []string{src}, "redeclared")
	}
}

func TestResolveConstNotDeclError(t *testing.T) {
	srcs := []string{
		"not-decl-1.go", "not-decl-2.go", "not-decl-3.go", "not-decl-4.go",
		"not-decl-5.go", "not-decl-6.go", "not-decl-7.go", "not-decl-8.go",
	}
	for _, src := range srcs {
		expectError(t, "_test/constdecl/src/err", []string{src}, "not declared")
	}
}

func TestResolveConstCountMismatch(t *testing.T) {
	srcs := []string{
		"not-eq-1.go", "not-eq-2.go", "not-eq-3.go", "not-eq-4.go", "not-eq-5.go",
	}
	for _, src := range srcs {
		expectError(t, "_test/constdecl/src/err", []string{src}, "must be equal")
	}
}

func TestResolveFuncDecl(t *testing.T) {
	p, err := compilePackage("_test/funcdecl/src/ok", []string{"decl.go"}, nil)
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
		expectError(t, "_test/funcdecl/src/err", []string{src}, "redeclared")
	}
}

func TestResolveStmtEmpty(t *testing.T) {
	p, err := compilePackage("_test/stmt/src/ok", []string{"empty.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	F := p.Find("F").(*ast.FuncDecl)
	if len(F.Func.Blk.Body) != 0 {
		t.Error("empty statement must be discarded entirely")
	}
}

func TestResolveStmtGo(t *testing.T) {
	p, err := compilePackage("_test/stmt/src/ok", []string{"go.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	F := p.Find("F").(*ast.FuncDecl)
	G := p.Find("G").(*ast.FuncDecl)
	g := G.Func.Blk.Body[0].(*ast.GoStmt)
	if x, ok := g.X.(*ast.Call); !ok || x.Func.(*ast.OperandName).Decl != F {
		t.Error("the go statememt must be a call to `F`")
	}
}

func TestResolveStmtReturn(t *testing.T) {
	p, err := compilePackage("_test/stmt/src/ok", []string{"return.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	F := p.Find("F").(*ast.FuncDecl)
	r := F.Func.Blk.Body[0].(*ast.ReturnStmt)
	if len(r.Xs) != 0 {
		t.Error("the return stmt in `F` must be without expressions")
	}

	G := p.Find("G").(*ast.FuncDecl)
	r = G.Func.Blk.Body[0].(*ast.ReturnStmt)
	if len(r.Xs) != 2 {
		t.Error("the return stmt in `G` must return 2 expressions")
	}
	A := p.Find("A").(*ast.Var)
	B := p.Find("B").(*ast.Var)
	if r.Xs[0].(*ast.OperandName).Decl != A {
		t.Error("the first return expression must be `A`")
	}
	if r.Xs[1].(*ast.OperandName).Decl != B {
		t.Error("the second return expression must be `B`")
	}
}

func TestResolveStmtSend(t *testing.T) {
	p, err := compilePackage("_test/stmt/src/ok", []string{"send.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	F := p.Find("F").(*ast.FuncDecl)
	s := F.Func.Blk.Body[0].(*ast.SendStmt)
	A := p.Find("A").(*ast.Var)
	B := p.Find("B").(*ast.Var)
	if s.Ch.(*ast.OperandName).Decl != A {
		t.Error("the channel in the send stmt must be `A`")
	}
	if s.X.(*ast.OperandName).Decl != B {
		t.Error("the value in the send stmt must be `B`")
	}
}

func TestResolveStmtIncDec(t *testing.T) {
	p, err := compilePackage("_test/stmt/src/ok", []string{"incdec.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	A := p.Find("A").(*ast.Var)
	B := p.Find("B").(*ast.Var)
	F := p.Find("F").(*ast.FuncDecl)
	i := F.Func.Blk.Body[0].(*ast.IncStmt)
	if i.X.(*ast.OperandName).Decl != A {
		t.Error("the incremented location must be `A`")
	}
	d := F.Func.Blk.Body[1].(*ast.DecStmt)
	if d.X.(*ast.OperandName).Decl != B {
		t.Error("the decremented location must be `B`")
	}
}

func TestResolveStmtAssign(t *testing.T) {
	p, err := compilePackage("_test/stmt/src/ok", []string{"assign.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	A := p.Find("A").(*ast.Var)
	B := p.Find("B").(*ast.Var)
	F := p.Find("F").(*ast.FuncDecl)

	s := F.Func.Blk.Body[0].(*ast.AssignStmt)
	if s.LHS[0].(*ast.OperandName).Decl != A {
		t.Error("st#0: LHS#0 must resolve to `A`")
	}

	s = F.Func.Blk.Body[1].(*ast.AssignStmt)
	if s.LHS[0].(*ast.OperandName).Decl != A {
		t.Error("st#1: LHS#0 must resolve to `A`")
	}
	if s.LHS[1].(*ast.OperandName).Decl != B {
		t.Error("st#1: LHS#1 must resolve to `B`")
	}
	if s.RHS[0].(*ast.OperandName).Decl != B {
		t.Error("st#1: RHS#0 must resolve to `B`")
	}
	if s.RHS[1].(*ast.OperandName).Decl != A {
		t.Error("st#1: RHS#1 must resolve to `A`")
	}
}

func TestResolveStmtShortVarDecl(t *testing.T) {
	p, err := compilePackage("_test/stmt/src/ok", []string{"short-var-decl.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	A := p.Find("A").(*ast.Var)
	B := p.Find("B").(*ast.Var)
	F := p.Find("F").(*ast.FuncDecl)

	a := F.Func.Blk.Find("A").(*ast.Var)
	if a == nil {
		t.Error("declaration of local `A` not found")
	}
	b := F.Func.Blk.Find("B").(*ast.Var)
	if b == nil {
		t.Error("declaration of local `B` not found")
	}
	c := F.Func.Blk.Find("C").(*ast.Var)
	if b == nil {
		t.Error("declaration of local `C` not found")
	}

	if len(F.Func.Blk.Body) != 3 {
		t.Error("body of `F` must consist of 3 assignments")
	}

	s := F.Func.Blk.Body[0].(*ast.AssignStmt)
	if s.LHS[0].(*ast.OperandName).Decl != a {
		t.Error("st#0: LHS #0 must resolve to local `A`")
	}
	if s.RHS[0].(*ast.OperandName).Decl != A {
		t.Error("st#0: RHS #0 must resolve to global `A`")
	}

	s = F.Func.Blk.Body[1].(*ast.AssignStmt)
	if s.LHS[0].(*ast.OperandName).Decl != a {
		t.Error("st#1: LHS #0 must resolve to local `A`")
	}
	if s.LHS[1].(*ast.OperandName).Decl != b {
		t.Error("st#1: LHS #1 must resolve to local `B`")
	}
	if s.RHS[0].(*ast.OperandName).Decl != B {
		t.Error("st#1: RHS #0 must resolve to global `B`")
	}
	if s.RHS[1].(*ast.OperandName).Decl != a {
		t.Error("st#1: RHS #1 must resolve to local `A`")
	}

	s = F.Func.Blk.Body[2].(*ast.AssignStmt)
	if F.Func.Blk.Find("_") != nil {
		t.Error("blank identifier must not be declared")
	}
	if s.LHS[0].(*ast.OperandName).Decl != c {
		t.Error("st#2: LHS #0 must resolve to local `C`")
	}
	if s.LHS[1] != ast.Blank {
		t.Error("st#2: LHS #1 must resolve to singleton `ast.Blank`")
	}
}

func TestResolveShortVarDeclError(t *testing.T) {
	expectError(t, "_test/stmt/src/err", []string{"lhs-dup.go"}, "duplicate ident")
	expectError(t, "_test/stmt/src/err", []string{"recv-dup.go"}, "duplicate ident")
	expectError(t, "_test/stmt/src/err", []string{"lhs-no-new-vars.go"}, "no new var")
	expectError(t, "_test/stmt/src/err", []string{"lhs-non-name.go"},
		"non-name on the left")
	expectError(t, "_test/stmt/src/err", []string{"lhs-non-name-for.go"},
		"non-name on the left")
}

func TestResolveStmtExpr(t *testing.T) {
	_, err := compilePackage("_test/stmt/src/ok", []string{"expr.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestResolveStmtIf(t *testing.T) {
	p, err := compilePackage("_test/stmt/src/ok", []string{"if.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	G := p.Find("G").(*ast.FuncDecl)
	x := G.Func.Blk.Find("x").(*ast.Var)
	y := G.Func.Blk.Find("y").(*ast.Var)

	// If #1
	s := G.Func.Blk.Body[1].(*ast.IfStmt)
	c := s.Cond.(*ast.BinaryExpr)
	if c.X.(*ast.OperandName).Decl != x || c.Y.(*ast.OperandName).Decl != y {
		t.Error("st#1: condition operands must refer to function local `x` and `y`")
	}

	// If #2
	s = G.Func.Blk.Body[2].(*ast.IfStmt)
	i := s.Init.(*ast.AssignStmt)
	if i.LHS[0].(*ast.OperandName).Decl != x || i.LHS[1].(*ast.OperandName).Decl != y {
		t.Error("st#2: LHS of the init stmt must refer to function local `x` and `y`")
	}
	c = s.Cond.(*ast.BinaryExpr)
	if c.X.(*ast.OperandName).Decl != x || c.Y.(*ast.OperandName).Decl != y {
		t.Error("st#2: condition operands must refer to function local `x` and `y`")
	}

	// If #3
	s = G.Func.Blk.Body[3].(*ast.IfStmt)
	u := s.Find("u").(*ast.Var)
	v := s.Find("v").(*ast.Var)
	if u == nil || v == nil {
		t.Error("st#3: missing declarations of `u` and `v`")
	}
	i = s.Init.(*ast.AssignStmt)
	if i.LHS[0].(*ast.OperandName).Decl != u || i.LHS[1].(*ast.OperandName).Decl != v {
		t.Error("st#3: LHS of the init stmt must refer to block local `u` and `v`")
	}
	c = s.Cond.(*ast.BinaryExpr)
	if c.X.(*ast.OperandName).Decl != u || c.Y.(*ast.OperandName).Decl != v {
		t.Error("st#3: condition operands must refer to block local `u` and `v`")
	}

	// If #4
	s = G.Func.Blk.Body[4].(*ast.IfStmt)
	x = s.Find("x").(*ast.Var)
	v = s.Find("v").(*ast.Var)
	if x == nil || v == nil {
		t.Error("st#4: missing declarations of `x` and `v`")
	}
	i = s.Init.(*ast.AssignStmt)
	if i.LHS[0].(*ast.OperandName).Decl != x || i.LHS[1].(*ast.OperandName).Decl != v {
		t.Error("st#4: LHS of the init stmt must refer to block local `x` and `v`")
	}
	c = s.Cond.(*ast.BinaryExpr)
	if c.X.(*ast.OperandName).Decl != x || c.Y.(*ast.OperandName).Decl != v {
		t.Error("st#4: condition operands must refer to block local `x` and `v`")
	}

	// If #5
	s = G.Func.Blk.Body[5].(*ast.IfStmt)
	x = s.Find("x").(*ast.Var)
	y = s.Find("y").(*ast.Var)
	if x == nil || y == nil {
		t.Error("st#5: missing declarations of `x` and `y`")
	}
	i = s.Init.(*ast.AssignStmt)
	if i.LHS[0].(*ast.OperandName).Decl != x || i.LHS[1].(*ast.OperandName).Decl != y {
		t.Error("st#5: LHS of the init stmt must refer to block local `x` and `y`")
	}
	c = s.Cond.(*ast.BinaryExpr)
	if c.X.(*ast.OperandName).Decl != x || c.Y.(*ast.OperandName).Decl != y {
		t.Error("st#5: condition operands must refer to block local `x` and `y`")
	}
	s = s.Else.(*ast.IfStmt)
	c = s.Cond.(*ast.BinaryExpr)
	if c.X.(*ast.OperandName).Decl != x || c.Y.(*ast.OperandName).Decl != y {
		t.Error("st#5: second condition operands must refer to block local `x` and `y`")
	}

	// If #6
	s = G.Func.Blk.Body[6].(*ast.IfStmt)
	x = s.Find("x").(*ast.Var)
	y = s.Find("y").(*ast.Var)
	if x == nil || y == nil {
		t.Error("st#6: missing declarations of `x` and `y`")
	}
	i = s.Init.(*ast.AssignStmt)
	if i.LHS[0].(*ast.OperandName).Decl != x || i.LHS[1].(*ast.OperandName).Decl != y {
		t.Error("st#6: LHS of the init stmt must refer to block local `x` and `y`")
	}
	c = s.Cond.(*ast.BinaryExpr)
	if c.X.(*ast.OperandName).Decl != x || c.Y.(*ast.OperandName).Decl != y {
		t.Error("st#6: condition operands must refer to block local `x` and `y`")
	}
	s = s.Else.(*ast.IfStmt)
	i = s.Init.(*ast.AssignStmt)
	if i.LHS[0].(*ast.OperandName).Decl != x || i.LHS[1].(*ast.OperandName).Decl != y {
		t.Error("st#6: LHS of the init stmt must refer to block local `x` and `y`")
	}
	c = s.Cond.(*ast.BinaryExpr)
	if c.X.(*ast.OperandName).Decl != x || c.Y.(*ast.OperandName).Decl != y {
		t.Error("st#6: second condition operands must refer to block local `x` and `y`")
	}

	// If #7
	s = G.Func.Blk.Body[7].(*ast.IfStmt)
	x = s.Find("x").(*ast.Var)
	y = s.Find("y").(*ast.Var)
	if x == nil || y == nil {
		t.Error("st#7/1: missing declarations of `x` and `y`")
	}
	i = s.Init.(*ast.AssignStmt)
	if i.LHS[0].(*ast.OperandName).Decl != x || i.LHS[1].(*ast.OperandName).Decl != y {
		t.Error("st#7/1: LHS of the init stmt must refer to block local `x` and `y`")
	}
	c = s.Cond.(*ast.BinaryExpr)
	if c.X.(*ast.OperandName).Decl != x || c.Y.(*ast.OperandName).Decl != y {
		t.Error("st#7/1: condition operands must refer to block local `x` and `y`")
	}
	s = s.Else.(*ast.IfStmt)
	x = s.Find("x").(*ast.Var)
	y = s.Find("y").(*ast.Var)
	if x == nil || y == nil {
		t.Error("st#7/2: missing declarations of `x` and `y`")
	}
	i = s.Init.(*ast.AssignStmt)
	if i.LHS[0].(*ast.OperandName).Decl != x || i.LHS[1].(*ast.OperandName).Decl != y {
		t.Error("st#7/2: LHS of the init stmt must refer to block local `x` and `y`")
	}
	c = s.Cond.(*ast.BinaryExpr)
	if c.X.(*ast.OperandName).Decl != x || c.Y.(*ast.OperandName).Decl != y {
		t.Error("st#7/2: second condition operands must refer to block local `x` and `y`")
	}
	if _, ok := s.Else.(*ast.Block); !ok {
		t.Error("`else` statement is not a block")
	}
}

func TestResolveStmtFor(t *testing.T) {
	p, err := compilePackage("_test/stmt/src/ok", []string{"for.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	G := p.Find("G").(*ast.FuncDecl)
	x := G.Func.Blk.Find("x").(*ast.Var)
	y := G.Func.Blk.Find("y").(*ast.Var)

	// For #1
	s := G.Func.Blk.Body[1].(*ast.ForStmt)
	c := s.Cond.(*ast.BinaryExpr)
	if c.X.(*ast.OperandName).Decl != x || c.Y.(*ast.OperandName).Decl != y {
		t.Error("st#1: condition operands must refer to function local `x` and `y`")
	}

	// For #2
	s = G.Func.Blk.Body[2].(*ast.ForStmt)
	i := s.Init.(*ast.AssignStmt)
	if i.LHS[0].(*ast.OperandName).Decl != x || i.LHS[1].(*ast.OperandName).Decl != y {
		t.Error("st#2: LHS of the init stmt must refer to function local `x` and `y`")
	}
	c = s.Cond.(*ast.BinaryExpr)
	if c.X.(*ast.OperandName).Decl != x || c.Y.(*ast.OperandName).Decl != y {
		t.Error("st#2: condition operands must refer to function local `x` and `y`")
	}

	// For #3
	s = G.Func.Blk.Body[3].(*ast.ForStmt)
	u := s.Find("u").(*ast.Var)
	v := s.Find("v").(*ast.Var)
	if u == nil || v == nil {
		t.Error("st#3: missing declarations of `u` and `v`")
	}
	i = s.Init.(*ast.AssignStmt)
	if i.LHS[0].(*ast.OperandName).Decl != u || i.LHS[1].(*ast.OperandName).Decl != v {
		t.Error("st#3: LHS of the init stmt must refer to block local `u` and `v`")
	}
	c = s.Cond.(*ast.BinaryExpr)
	if c.X.(*ast.OperandName).Decl != u || c.Y.(*ast.OperandName).Decl != v {
		t.Error("st#3: condition operands must refer to block local `u` and `v`")
	}

	// For #4
	s = G.Func.Blk.Body[4].(*ast.ForStmt)
	x = s.Find("x").(*ast.Var)
	v = s.Find("v").(*ast.Var)
	if x == nil || v == nil {
		t.Error("st#4: missing declarations of `x` and `v`")
	}
	i = s.Init.(*ast.AssignStmt)
	if i.LHS[0].(*ast.OperandName).Decl != x || i.LHS[1].(*ast.OperandName).Decl != v {
		t.Error("st#4: LHS of the init stmt must refer to block local `x` and `v`")
	}
	c = s.Cond.(*ast.BinaryExpr)
	if c.X.(*ast.OperandName).Decl != x || c.Y.(*ast.OperandName).Decl != v {
		t.Error("st#4: condition operands must refer to block local `x` and `v`")
	}
}

func TestResolveStmtForRange(t *testing.T) {
	p, err := compilePackage("_test/stmt/src/ok", []string{"for-range.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	G := p.Find("G").(*ast.FuncDecl)
	x := G.Func.Blk.Find("x").(*ast.Var)
	y := G.Func.Blk.Find("y").(*ast.Var)

	// Range For #1
	s := G.Func.Blk.Body[1].(*ast.ForRangeStmt)
	if s.LHS[0].(*ast.OperandName).Decl != x || s.LHS[1].(*ast.OperandName).Decl != y {
		t.Error("st#1: LHS in range-for must refer to function local `x` and `y`")
	}

	// Range For #2
	s = G.Func.Blk.Body[2].(*ast.ForRangeStmt)
	u := s.Find("u").(*ast.Var)
	v := s.Find("v").(*ast.Var)
	if u == nil || v == nil {
		t.Error("st#2: missing declarations of `u` and `v`")
	}
	if s.LHS[0].(*ast.OperandName).Decl != u || s.LHS[1].(*ast.OperandName).Decl != v {
		t.Error("st#2: LHS in range-for must refer to block local `u` and `v`")
	}

	// Range For #3
	s = G.Func.Blk.Body[3].(*ast.ForRangeStmt)
	x = s.Find("x").(*ast.Var)
	v = s.Find("v").(*ast.Var)
	if x == nil || v == nil {
		t.Error("st#3: missing declarations of `x` and `v`")
	}
	if s.LHS[0].(*ast.OperandName).Decl != x || s.LHS[1].(*ast.OperandName).Decl != v {
		t.Error("st#2: LHS in range-for must refer to block local `x` and `v`")
	}
}

func TestResolveStmtDefer(t *testing.T) {
	p, err := compilePackage("_test/stmt/src/ok", []string{"defer.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	F := p.Find("F").(*ast.FuncDecl)
	G := p.Find("G").(*ast.FuncDecl)
	s := G.Func.Blk.Body[0].(*ast.DeferStmt)
	x := s.X.(*ast.Call)
	if x.Func.(*ast.OperandName).Decl != F {
		t.Error("the expression in defer stmt must be a call to `F`")
	}
}

func TestResolveStmtExprSwitch(t *testing.T) {
	p, err := compilePackage("_test/stmt/src/ok", []string{"expr-switch.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	G := p.Find("G").(*ast.FuncDecl)
	x := G.Func.Blk.Find("x").(*ast.Var)
	y := G.Func.Blk.Find("y").(*ast.Var)

	// Expr switch #1
	s := G.Func.Blk.Body[1].(*ast.ExprSwitchStmt)
	c := s.Cases[0].Xs[0].(*ast.BinaryExpr)
	if c.X.(*ast.OperandName).Decl != x || c.Y.(*ast.OperandName).Decl != y {
		t.Error("st#1: expr in case clause must refer to function local `x` and `y`")
	}

	// Expr switch #2
	s = G.Func.Blk.Body[2].(*ast.ExprSwitchStmt)
	c = s.Cases[0].Xs[0].(*ast.BinaryExpr)
	if c.X.(*ast.OperandName).Decl != x || c.Y.(*ast.OperandName).Decl != y {
		t.Error("st#2: expr in case clause must refer to function local `x` and `y`")
	}

	// Expr switch #3
	s = G.Func.Blk.Body[3].(*ast.ExprSwitchStmt)
	u := s.Find("u").(*ast.Var)
	v := s.Find("v").(*ast.Var)
	if u == nil || v == nil {
		t.Error("st#3: missing declarations of `u` and `v`")
	}
	i := s.Init.(*ast.AssignStmt)
	if i.LHS[0].(*ast.OperandName).Decl != u || i.LHS[1].(*ast.OperandName).Decl != v {
		t.Error("st#3: LHS of the init stmt must refer to block local `u` and `v`")
	}
	c = s.Cases[0].Xs[0].(*ast.BinaryExpr)
	if c.X.(*ast.OperandName).Decl != u || c.Y.(*ast.OperandName).Decl != v {
		t.Error("st#2: expr in case clause must refer to block local `u` and `v`")
	}

	// Expr switch #4
	s = G.Func.Blk.Body[4].(*ast.ExprSwitchStmt)
	x = s.Find("x").(*ast.Var)
	v = s.Find("v").(*ast.Var)
	if x == nil || v == nil {
		t.Error("st#4: missing declarations of `x` and `v`")
	}
	i = s.Init.(*ast.AssignStmt)
	if i.LHS[0].(*ast.OperandName).Decl != x || i.LHS[1].(*ast.OperandName).Decl != v {
		t.Error("st#4: LHS of the init stmt must refer to block local `x` and `v`")
	}
	c = s.Cases[0].Xs[0].(*ast.BinaryExpr)
	if c.X.(*ast.OperandName).Decl != x || c.Y.(*ast.OperandName).Decl != v {
		t.Error("st#4: expr in case clause must refer to block local `x` and `v`")
	}
}

func TestResolveStmtTypeSwitch(t *testing.T) {
	p, err := compilePackage("_test/stmt/src/ok", []string{"type-switch.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	G := p.Find("G").(*ast.FuncDecl)
	x := G.Func.Blk.Find("x").(*ast.Var)
	y := G.Func.Blk.Find("y").(*ast.Var)

	// Type switch #1
	s := G.Func.Blk.Body[1].(*ast.TypeSwitchStmt)
	if s.X.(*ast.OperandName).Decl != x {
		t.Error("st#1: expression in type switch refer to function local `x`")
	}
	typ := s.Cases[0].Types[0]
	if typ != ast.BuiltinInt {
		t.Error("st#1: type in case clause must resolve to builtin `int`")
	}
	a := s.Cases[0].Blk.Body[0].(*ast.ExprStmt).X.(*ast.BinaryExpr)
	if a.X.(*ast.OperandName).Decl != x || a.Y.(*ast.OperandName).Decl != y {
		t.Error("st#2: expr in case clause must refer to function local `x` and `y`")
	}

	// Type switch #2
	s = G.Func.Blk.Body[2].(*ast.TypeSwitchStmt)
	i := s.Init.(*ast.AssignStmt)
	if i.LHS[0].(*ast.OperandName).Decl != x || i.LHS[1].(*ast.OperandName).Decl != y {
		t.Error("st#2: LHS of the init stmt must refer to function local `x` and `y`")
	}
	a = s.Cases[0].Blk.Body[0].(*ast.ExprStmt).X.(*ast.BinaryExpr)
	if a.X.(*ast.OperandName).Decl != x || a.Y.(*ast.OperandName).Decl != y {
		t.Error("st#2: expr in case body must refer to function local `x` and `y`")
	}

	// Type switch #3
	s = G.Func.Blk.Body[3].(*ast.TypeSwitchStmt)
	u := s.Find("u").(*ast.Var)
	v := s.Find("v").(*ast.Var)
	if u == nil || v == nil {
		t.Error("st#3: missing declarations of `u` and `v`")
	}
	i = s.Init.(*ast.AssignStmt)
	if i.LHS[0].(*ast.OperandName).Decl != u || i.LHS[1].(*ast.OperandName).Decl != v {
		t.Error("st#3: LHS of the init stmt must refer to block local `u` and `v`")
	}
	a = s.Cases[0].Blk.Body[0].(*ast.ExprStmt).X.(*ast.BinaryExpr)
	if a.X.(*ast.OperandName).Decl != u || a.Y.(*ast.OperandName).Decl != v {
		t.Error("st#3: expr in case body must refer to block local `u` and `v`")
	}

	// Type switch #4
	s = G.Func.Blk.Body[4].(*ast.TypeSwitchStmt)
	x = s.Find("x").(*ast.Var)
	v = s.Find("v").(*ast.Var)
	if x == nil || v == nil {
		t.Error("st#4: missing declarations of `x` and `v`")
	}
	i = s.Init.(*ast.AssignStmt)
	if i.LHS[0].(*ast.OperandName).Decl != x || i.LHS[1].(*ast.OperandName).Decl != v {
		t.Error("st#4: LHS of the init stmt must refer to block local `x` and `v`")
	}
	a = s.Cases[0].Blk.Body[0].(*ast.ExprStmt).X.(*ast.BinaryExpr)
	if a.X.(*ast.OperandName).Decl != x || a.Y.(*ast.OperandName).Decl != v {
		t.Error("st#4: expr in case body must refer to block local `x` and `v`")
	}

	// Type switch #5
	s = G.Func.Blk.Body[5].(*ast.TypeSwitchStmt)
	if G.Func.Blk.Find(s.Id) != nil || s.Find(s.Id) != nil {
		t.Error("st#5: name from type switch guard must be declared only in case blocks")
	}
	bb := s.Cases[0].Blk
	if bb.Find(s.Id) == nil {
		t.Error("st#5: name from type switch guard not declared in case block")
	}
	bb = s.Cases[1].Blk
	if bb.Find(s.Id) == nil {
		t.Error("#st#5: name from type switch guard not declared in case block")
	}
	bb = s.Cases[2].Blk
	if bb.Find(s.Id) != nil {
		t.Error(
			"#st#5: name from type switch should not be declared in the default block")
	}
}

func TestResolveStmtSelect(t *testing.T) {
	p, err := compilePackage("_test/stmt/src/ok", []string{"select.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	G := p.Find("G").(*ast.FuncDecl)
	x := G.Func.Blk.Find("x").(*ast.Var)
	y := G.Func.Blk.Find("y").(*ast.Var)

	// Select case #0
	s := G.Func.Blk.Body[1].(*ast.SelectStmt)
	snd := s.Comms[0].Comm.(*ast.SendStmt)
	if snd.Ch.(*ast.OperandName).Decl != x || snd.X.(*ast.OperandName).Decl != y {
		t.Error("comm#0: send stmt must refer to function local `x` and `y`")
	}

	// Select case #1
	rcv := s.Comms[1].Comm.(*ast.RecvStmt)
	if rcv.X.(*ast.OperandName).Decl != y {
		t.Error("comm#1: recv stmt must receive into function local `y`")
	}
	if rcv.Rcv.(*ast.UnaryExpr).X.(*ast.OperandName).Decl != x {
		t.Error("comm#1: recv stmt must receive from function local `x`")
	}

	// Select case #2
	b := s.Comms[2].Blk
	xx := b.Find("x").(*ast.Var)
	if xx == nil {
		t.Error("new `x` must be declared in comm clause block")
	}
	rcv = s.Comms[2].Comm.(*ast.RecvStmt)
	if rcv.X.(*ast.OperandName).Decl != xx {
		t.Error("comm#2: recv stmt must receive into block local `x`")
	}
	if rcv.Y != ast.Blank {
		t.Error("comm#2: recv stmt must receive into singletone 'ast.Blank'")
	}
	if rcv.Rcv.(*ast.UnaryExpr).X.(*ast.OperandName).Decl != x {
		t.Error("comm#2: recv stmt must receive from function local `x`")
	}
	if b.Body[0].(*ast.ExprStmt).X.(*ast.OperandName).Decl != xx {
		t.Error("comm#2: expr in case block must refer to block local `x`")
	}
}

func TestResolveStmtErrorNotDecl(t *testing.T) {
	for _, src := range []string{
		"not-decl-1.go", "not-decl-2.go", "not-decl-3.go", "not-decl-4.go",
		"not-decl-5.go", "not-decl-6.go", "not-decl-7.go", "not-decl-8.go",
		"not-decl-9.go", "not-decl-10.go", "not-decl-11.go", "not-decl-12.go",
		"not-decl-13.go", "not-decl-14.go", "not-decl-15.go", "not-decl-16.go",
		"not-decl-17.go", "not-decl-18.go", "not-decl-19.go", "not-decl-20.go",
		"not-decl-21.go", "not-decl-22.go", "not-decl-23.go", "not-decl-24.go",
		"not-decl-25.go", "not-decl-26.go", "not-decl-27.go", "not-decl-28.go",
		"not-decl-29.go", "not-decl-30.go", "not-decl-31.go", "not-decl-32.go",
		"not-decl-33.go", "not-decl-34.go",
	} {
		expectError(t, "_test/stmt/src/err", []string{src}, "not declared")
	}
}

func TestResolveStmtErrorDupDecl(t *testing.T) {
	for _, src := range []string{"dup-decl-1.go", "dup-decl-2.go", "dup-decl-3.go"} {
		expectError(t, "_test/stmt/src/err", []string{src}, "redeclared")
	}
}

func TestResolveStmtForPostError(t *testing.T) {
	expectError(t, "_test/stmt/src/err", []string{"for-post.go"},
		"cannot declare in for post")
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

func TestResolveImport(t *testing.T) {
	loc := &MockPackageLocator{pkgs: make(map[string]*ast.Package)}
	pkg, err := compilePackage("_test/pkg/src/a", []string{"a.go"}, loc)
	if err != nil {
		t.Fatal(err)
	}
	pkgA, err := reloadPackage(pkg, loc)
	if err != nil {
		t.Fatal(err)
	}
	loc.pkgs["a"] = pkgA

	A := pkgA.Find("A")
	if A == nil {
		t.Fatal("name `A` not found in package")
	}
	_, ok := A.(*ast.TypeDecl)
	if !ok {
		t.Fatal("`A` is not a typename")
	}

	// Test regular import
	pkg, err = compilePackage("_test/pkg/src/b", []string{"b1.go"}, loc)
	if err != nil {
		t.Fatal(err)
	}

	a, ok := pkg.Files[0].Find("a").(*ast.ImportDecl)
	if !ok {
		t.Fatalf("import `%s` not found in source `%s`\n", a.Name, pkg.Files[0].Name)
	}
	if a.Pkg != pkgA {
		t.Fatal("import refers to unexpected package")
	}

	// Test import alias
	pkg, err = compilePackage("_test/pkg/src/b", []string{"b2.go"}, loc)
	if err != nil {
		t.Fatal(err)
	}

	x, ok := pkg.Files[0].Find("x").(*ast.ImportDecl)
	if !ok {
		t.Fatalf("import name `%s` not found in source `%s`\n", x.Name, pkg.Files[0].Name)
	}

	// Test blank import
	pkg, err = compilePackage("_test/pkg/src/b", []string{"b3.go"}, loc)
	if err != nil {
		t.Fatal(err)
	}

	y := pkg.Files[0].Find("a")
	if y != nil {
		t.Fatalf("import name `a` should not be found in file `%s`\n", pkg.Files[0].Name)
	}

	// Test "transparent "import
	pkg, err = compilePackage("_test/pkg/src/b", []string{"b4.go"}, loc)
	if err != nil {
		t.Fatal(err)
	}

	y = pkg.Files[0].Find("a")
	if y != nil {
		t.Fatalf("import name `a` should not be found in source `%s`\n", pkg.Files[0].Name)
	}

	AA, ok := pkg.Files[0].Find("A").(*ast.TypeDecl)
	if !ok {
		t.Fatalf("type name `A` not found in the scope of file `%s`\n", pkg.Files[0].Name)
	}
	if A != AA {
		t.Error("name `A` in packages `a` and `b` refer to different declarations")
	}
}

func TestResolveImportError(t *testing.T) {
	loc := &MockPackageLocator{pkgs: make(map[string]*ast.Package)}
	pkg, err := compilePackage("_test/pkg/src/a", []string{"a.go"}, loc)
	if err != nil {
		t.Fatal(err)
	}
	pkgA, err := reloadPackage(pkg, loc)
	if err != nil {
		t.Fatal(err)
	}

	// Test import name not found
	expectErrorWithLoc(t, "_test/pkg/src/c", []string{"err1.go"}, loc, "not found")

	loc.pkgs["a"] = pkgA
	pkg, err = compilePackage("_test/pkg/src/b", []string{"b1.go"}, loc)
	if err != nil {
		t.Fatal(err)
	}
	pkgB, err := reloadPackage(pkg, loc)
	if err != nil {
		t.Fatal(err)
	}
	loc.pkgs["b"] = pkgB

	// Test duplicated import names
	expectErrorWithLoc(t, "_test/pkg/src/c", []string{"err2.go"}, loc, "redeclared")

	// Test "transparent" import conflicts
	expectErrorWithLoc(t, "_test/pkg/src/c", []string{"err3.go"}, loc, "redeclared")

	// Test conflict between import name and a "transparent" import
	expectErrorWithLoc(t, "_test/pkg/src/c", []string{"err4.go"}, loc, "redeclared")
	expectErrorWithLoc(t, "_test/pkg/src/c", []string{"err5.go"}, loc, "redeclared")
}

func TestResolveTypeMultiPackage(t *testing.T) {
	loc := &MockPackageLocator{pkgs: make(map[string]*ast.Package)}
	pkgA, err := compilePackage("_test/pkg/src/a", []string{"a.go"}, loc)
	if err != nil {
		t.Fatal(err)
	}
	loc.pkgs["a"] = pkgA
	A := pkgA.Find("A").(*ast.TypeDecl)
	B := pkgA.Find("B").(*ast.Var)

	pkg, err := compilePackage("_test/pkg/src/d", []string{"d1.go"}, loc)
	if err != nil {
		t.Fatal(err)
	}

	D := pkg.Find("D")
	if D == nil {
		t.Error("name `D` not found")
	}
	typD, ok := D.(*ast.TypeDecl)
	if !ok {
		t.Error("`D` is not a typename")
	}

	if typD.Type != A {
		t.Error("typename `D` does not refer to `a.A`")
	}

	// Test package-level initialization statements are resolved in the scope
	// of the file, which contains the variable declaration.
	d := pkg.Find("d").(*ast.Var)
	if d.Init == nil || d.Init.RHS == nil {
		t.Error("`d` missing an initialization statement")
	}
	if d.Init.RHS[0].(*ast.OperandName).Decl != B {
		t.Error("`d` must be initialized by `a.B`")
	}

	// Test unknown package name
	expectErrorWithLoc(t, "_test/pkg/src/d", []string{"err1.go"}, loc, "not declared")

	// Test package part of a QualifiedId does not refer to imported package
	expectErrorWithLoc(t, "_test/pkg/src/d", []string{"err2.go"}, loc,
		"does not refer to package name")

	// Test name not declared in imported package.
	expectErrorWithLoc(t, "_test/pkg/src/d", []string{"err3.go"}, loc, "not declared")

	// Test predeclared names are not found via a qualified id.
	expectErrorWithLoc(t, "_test/pkg/src/d", []string{"err6.go"}, loc, "not exported")

	// Test name declared, but not exported.
	expectErrorWithLoc(t, "_test/pkg/src/d", []string{"err4.go"}, loc, "not exported")

	// After loading the PDB, the above error becomes `not declared`, as the
	// identifier is not present at all in the PDB.
	pkgA, err = reloadPackage(pkgA, loc)
	if err != nil {
		t.Fatal(err)
	}
	loc.pkgs["a"] = pkgA
	expectErrorWithLoc(t, "_test/pkg/src/d", []string{"err4.go"}, loc, "not declared")

	// Test imported indentifier is not a typename.
	expectErrorWithLoc(t, "_test/pkg/src/d", []string{"err5.go"}, loc, "not a typename")
}

func TestResolveExprMultiPackage(t *testing.T) {
	loc := &MockPackageLocator{pkgs: make(map[string]*ast.Package)}
	pkgA, err := compilePackage("_test/pkg/src/a", []string{"a.go"}, loc)
	if err != nil {
		t.Fatal(err)
	}
	loc.pkgs["a"] = pkgA
	A := pkgA.Find("A").(*ast.TypeDecl)
	B := pkgA.Find("B").(*ast.Var)

	pkg, err := compilePackage("_test/pkg/src/e", []string{"e1.go"}, loc)
	if err != nil {
		t.Fatal(err)
	}

	F := pkg.Find("F").(*ast.FuncDecl)
	s := F.Func.Blk.Body[0].(*ast.ReturnStmt)
	if s.Xs[0].(*ast.OperandName).Decl != B {
		t.Error("the return expression is not `a.B`")
	}

	// Test a Call changed to Conversion.
	G := pkg.Find("G").(*ast.FuncDecl)
	s = G.Func.Blk.Body[0].(*ast.ReturnStmt)
	x, ok := s.Xs[0].(*ast.Conversion)
	if !ok {
		t.Error("the return expression is not a Conversion")
	}
	if x.Typ != A {
		t.Error("the conversion type is not `a.A`")
	}

	// Test QualifiedId resolved to MethodExpr
	H := pkg.Find("H").(*ast.FuncDecl)
	ss := H.Func.Blk.Body[0].(*ast.ExprStmt)
	y, ok := ss.X.(*ast.MethodExpr)
	if !ok {
		t.Error("the return expression is not a Conversion")
	}
	if y.RTyp != A {
		t.Error("the type in the method expression is not `a.A`")
	}

	// Test package part of a QualifiedId does not refer to imported package
	expectErrorWithLoc(t, "_test/pkg/src/e", []string{"err1.go"}, loc, "aa not declared")

	// // Test name not declared in imported package.
	expectErrorWithLoc(t, "_test/pkg/src/e", []string{"err2.go"}, loc, "notB not declared")

	// Test name declared, but not exported.
	expectErrorWithLoc(t, "_test/pkg/src/e", []string{"err3.go"}, loc, "not exported")

	// Test name cannot be an operand.
	expectErrorWithLoc(t, "_test/pkg/src/e", []string{"err4.go"}, loc, "invalid operand")

	// Test package part of a QualifiedId does not refer to imported package
	expectErrorWithLoc(t, "_test/pkg/src/e", []string{"conv-err1.go"}, loc,
		"aa not declared")

	// Test name not declared in imported package.
	expectErrorWithLoc(t, "_test/pkg/src/e", []string{"conv-err2.go"}, loc,
		"notB not declared")

	// Test name declared, but not exported.
	expectErrorWithLoc(t, "_test/pkg/src/e", []string{"conv-err3.go"}, loc,
		"not exported")
}

func TestResolveReload(t *testing.T) {
	loc := &MockPackageLocator{pkgs: make(map[string]*ast.Package)}
	pkg, err := compilePackage("_test/pkg/src/a", []string{"a.go"}, loc)
	if err != nil {
		t.Fatal(err)
	}
	pkgA, err := reloadPackage(pkg, loc)
	if err != nil {
		t.Fatal(err)
	}
	loc.pkgs["a"] = pkgA

	pkg, err = compilePackage("_test/pkg/src/f", []string{"f.go"}, loc)
	if err != nil {
		t.Fatal(err)
	}
	pkg, err = reloadPackage(pkg, loc)
	if err != nil {
		t.Fatal(err)
	}
}

func TestResolveLabel(t *testing.T) {
	p, err := compilePackage("_test/funcdecl/src/ok", []string{"labels.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	F := p.Find("F").(*ast.FuncDecl)
	vL := F.Func.Blk.Lookup("L")
	L := F.Func.FindLabel("L")
	L1 := F.Func.FindLabel("L1")
	if L == vL {
		t.Error("label `L` must not be found by ordinary lookup")
	}
	if L.Blk != F.Func.Blk {
		t.Error("label `L` must point to the containing block")
	}
	if L.Stmt != L1.Stmt {
		t.Error("`L` and `L1` must refer to the same statement")
	}
	s := F.Func.Blk.Body[1].(*ast.AssignStmt)
	fn := s.RHS[0].(*ast.Func)
	vLL := fn.Blk.Lookup("L")
	LL := fn.FindLabel("L")
	LL1 := fn.FindLabel("L1")
	if LL == vLL {
		t.Error("label `L` must not be found by ordinary lookup")
	}
	if LL.Blk != fn.Blk {
		t.Error("label `L` must point to the containing block")
	}
	if vLL != vL {
		t.Error(
			"`L` in function literal must resolve to `L` variable in the outer function")
	}
	if LL.Stmt != LL1.Stmt {
		t.Error("`L` and `L1` must refer to the same statement")
	}
	if LL.Stmt != nil {
		t.Error("`L` must refer to empty statement")
	}
	expectError(t, "_test/funcdecl/src/err", []string{"dup-label.go"}, "L redeclared")
}

func TestResolveBreak(t *testing.T) {
	p, err := compilePackage("_test/labels/src/ok", []string{"break.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	F := p.Find("F").(*ast.FuncDecl)
	L := []string{"L0", "L1", "L2", "L3", "L3", "L4", "L5", "L6", "L6", "L7", "L7"}
	B := []string{"B0", "B1", "B2", "B3", "B4", "B5", "B6", "B7", "B8", "B9", "B10"}
	for i := range L {
		l := F.Func.FindLabel(L[i])
		b := F.Func.FindLabel(B[i]).Stmt.(*ast.BreakStmt)
		if b.Dst != l.Stmt {
			t.Errorf("break at `%s` must jump to label `%s`\n", B[i], L[i])
		}
	}

	expectError(t, "_test/labels/src/err", []string{"break-1.go"},
		"break label does not refer to a `for`")
	expectError(t, "_test/labels/src/err", []string{"break-2.go"},
		"break label does not refer to a `for`")
	expectError(t, "_test/labels/src/err", []string{"break-3.go"},
		"break label does not refer to a `for`")
	expectError(t, "_test/labels/src/err", []string{"break-4.go"},
		"break label does not refer to an enclosing")
	expectError(t, "_test/labels/src/err", []string{"break-5.go"}, "unknown label")
}

func TestResolveContinue(t *testing.T) {
	p, err := compilePackage("_test/labels/src/ok", []string{"cont.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	F := p.Find("F").(*ast.FuncDecl)
	L := []string{"L0", "L1", "L2", "L3"}
	C := []string{"C0", "C1", "C2", "C3"}
	for i := range L {
		l := F.Func.FindLabel(L[i])
		c := F.Func.FindLabel(C[i]).Stmt.(*ast.ContinueStmt)
		if c.Dst != l.Stmt {
			t.Errorf("continue at `%s` must jump to label `%s`\n", C[i], L[i])
		}
	}

	expectError(t, "_test/labels/src/err", []string{"cont-1.go"},
		"continue label does not refer to a `for`")
	expectError(t, "_test/labels/src/err", []string{"cont-2.go"},
		"continue label does not refer to a `for`")
	expectError(t, "_test/labels/src/err", []string{"cont-3.go"},
		"continue label does not refer to a `for`")
	expectError(t, "_test/labels/src/err", []string{"cont-4.go"},
		"continue label does not refer to an enclosing")
	expectError(t, "_test/labels/src/err", []string{"cont-5.go"}, "unknown label")
}

func TestResolveGoto(t *testing.T) {
	p, err := compilePackage("_test/labels/src/ok", []string{"goto.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	F := p.Find("F").(*ast.FuncDecl)
	L := []string{"E0", "E1", "E0", "E0", "B0", "B1", "B0", "B0"}
	G := []string{"G0", "G1", "G2", "G3", "G4", "G5", "G6", "G7"}
	for i := range L {
		l := F.Func.FindLabel(L[i])
		g := F.Func.FindLabel(G[i]).Stmt.(*ast.GotoStmt)
		if g.Dst != l.Stmt {
			t.Errorf("goto at `%s` must jump to label `%s`\n", G[i], L[i])
		}
	}

	expectError(t, "_test/labels/src/err", []string{"goto-1.go"}, "unknown label")
	expectError(t, "_test/labels/src/err", []string{"goto-2.go"}, "cross block")
	expectError(t, "_test/labels/src/err", []string{"goto-3.go"}, "cross block")
	expectError(t, "_test/labels/src/err", []string{"goto-4.go"}, "cross block")
	expectError(t, "_test/labels/src/err", []string{"goto-5.go"}, "cross block")
	expectError(t, "_test/labels/src/err", []string{"goto-6.go"},
		"goto jumps into the scope")
	expectError(t, "_test/labels/src/err", []string{"goto-7.go"},
		"goto jumps into the scope")
	expectError(t, "_test/labels/src/err", []string{"goto-8.go"},
		"goto jumps into the scope")
}

func TestResolveFallthrough(t *testing.T) {
	p, err := compilePackage("_test/labels/src/ok", []string{"fallthrough.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	F := p.Find("F").(*ast.FuncDecl)
	f := F.Func.FindLabel("f").Stmt.(*ast.FallthroughStmt)
	s := F.Func.Blk.Body[0].(*ast.ExprSwitchStmt)
	if f.Dst != s.Cases[1].Blk {
		t.Error("the destination of the `fallthrough` must be the second case")
	}

	expectError(t, "_test/labels/src/err", []string{"fall-1.go"}, "misplaced fallthrough")
	expectError(t, "_test/labels/src/err", []string{"fall-2.go"}, "misplaced fallthrough")
	expectError(t, "_test/labels/src/err", []string{"fall-3.go"}, "misplaced fallthrough")
	expectError(t, "_test/labels/src/err", []string{"fall-4.go"},
		"cannot fallthrough the last case")
}

func TestResolveIndirectImport(t *testing.T) {
	loc := &MockPackageLocator{pkgs: make(map[string]*ast.Package)}
	pkg, err := compilePackage("_test/indirect-import/src/a", []string{"a.go"}, loc)
	if err != nil {
		t.Fatal(err)
	}
	pkgA, err := reloadPackage(pkg, loc)
	if err != nil {
		t.Fatal(err)
	}
	loc.pkgs["a"] = pkgA

	pkg, err = compilePackage("_test/indirect-import/src/b", []string{"b.go"}, loc)
	if err != nil {
		t.Fatal(err)
	}
	pkgB, err := reloadPackage(pkg, loc)
	if err != nil {
		t.Fatal(err)
	}
	loc.pkgs["b"] = pkgB

	pkg, err = compilePackage("_test/indirect-import/src/c", []string{"c.go"}, loc)
	if err != nil {
		t.Fatal(err)
	}
	pkgC, err := reloadPackage(pkg, loc)
	if err != nil {
		t.Fatal(err)
	}
	loc.pkgs["c"] = pkgC
}

func TestResolveMethodDecl(t *testing.T) {
	p, err := compilePackage("_test/methods/src/ok", []string{"method.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	A := p.Find("A").(*ast.TypeDecl)
	if len(A.Methods) != 1 || A.Methods[0].Name != "F" {
		t.Error("`F` must be in the method set of `A`")
	}
	if len(A.PMethods) != 1 || A.PMethods[0].Name != "G" {
		t.Error("`G` must be in the method set of `*A`")
	}

	F := A.Methods[0]
	if F.Func.Blk.Find("a") == nil {
		t.Error("receiver `a` must be declared at function block scope")
	}

	_, err = compilePackage("_test/methods/src/ok", []string{"uniq-1.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = compilePackage("_test/methods/src/ok", []string{"uniq-2.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	expectError(t, "_test/methods/src/err", []string{"bad-type-1.go"},
		"invalid receiver type (is an unnamed type)")
	expectError(t, "_test/methods/src/err", []string{"bad-type-2.go"},
		"invalid receiver type (is an unnamed type)")

	pkgA, err := compilePackage("_test/methods/src/err/a", []string{"bad-type-3.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	loc := &MockPackageLocator{pkgs: make(map[string]*ast.Package)}
	loc.pkgs["a"] = pkgA
	expectErrorWithLoc(t, "_test/methods/src/err/b", []string{"bad-type-4.go"}, loc,
		"type and method must be declared in the same package")

	expectError(t, "_test/methods/src/err", []string{"bad-type-5.go"}, "not declared")
	expectError(t, "_test/methods/src/err", []string{"bad-type-6.go"},
		"cannot be a pointer")
	expectError(t, "_test/methods/src/err", []string{"bad-type-7.go"},
		"cannot be an interface")
	expectError(t, "_test/methods/src/err", []string{"dup-recv.go"},
		"redeclared")
}
