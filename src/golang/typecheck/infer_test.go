package typecheck

import (
	"golang/ast"
	"testing"
)

func TestInfer01(t *testing.T) {
	p, err := compilePackage("_test/src/infer", []string{"infer-01.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	P := p.Find("P").(*ast.TypeDecl)
	Q := p.Find("Q").(*ast.TypeDecl)

	a := p.Find("a").(*ast.Var)
	if a.Type != P {
		t.Error("the type of `a` must be inferred to `P`")
	}
	b := p.Find("b").(*ast.Var)
	if b.Type != Q {
		t.Error("the type of `b` must be inferred to `Q`")
	}

	S := p.Find("S").(*ast.TypeDecl).Type.(*ast.StructType)
	tX := S.Fields[0].Type.(*ast.ArrayType)
	c, ok := tX.Dim.(*ast.ConstValue)
	if !ok {
		t.Fatal("the dimension of the `S.X`s type is not a constant")
	}
	if c.Typ != ast.BuiltinInt {
		t.Error("the dimension of the `S.X`s type should have type `int`")
	}
	if i, ok := c.Value.(ast.Int); !ok || i != 1 {
		t.Error("the dimension of the `S.X`s type should be 1")
	}
}

func TestInfer02(t *testing.T) {
	p, err := compilePackage("_test/src/infer", []string{"infer-02.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	S := p.Find("S").(*ast.TypeDecl).Type.(*ast.StructType)
	tX := S.Fields[0].Type.(*ast.ArrayType)
	tY := S.Fields[1].Type.(*ast.ArrayType)

	c, ok := tX.Dim.(*ast.ConstValue)
	if !ok {
		t.Fatal("the dimension of the `S.X`s type is not a constant")
	}
	if c.Typ != ast.BuiltinInt {
		t.Error("the dimension of the `S.X`s type should have type `int`")
	}
	if v, ok := c.Value.(ast.Int); !ok || v != 1 {
		t.Error("the dimension of the `S.X`s type should be 1")
	}

	c, ok = tY.Dim.(*ast.ConstValue)
	if !ok {
		t.Fatal("the dimension of the `S.Y`s type is not a constant")
	}
	if c.Typ != ast.BuiltinInt {
		t.Error("the dimension of the `S.Y`s type should have type `int`")
	}
	if v, ok := c.Value.(ast.Int); !ok || v != 1 {
		t.Error("the dimension of the `S.Y`s type should be 1")
	}

	u := p.Find("u").(*ast.Var)
	if u.Type != tX {
		t.Error("the type of `u` must be inferred to the type of `S.X`")
	}
	v := p.Find("v").(*ast.Var)
	if v.Type != tY {
		t.Error("the type of `b` must be inferred to the type of `S.Y`")
	}
}

func TestInfer03(t *testing.T) {
	p, err := compilePackage("_test/src/infer", []string{"infer-03.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	u := p.Find("u").(*ast.Var)
	typ := u.Type.(*ast.ArrayType)
	c, ok := typ.Dim.(*ast.ConstValue)
	if !ok {
		t.Fatal("the dimension of `u`s type is not a constant")
	}
	if c.Typ != ast.BuiltinInt {
		t.Error("the dimension of the `u`s type should have type `int`")
	}
	if v, ok := c.Value.(ast.Int); !ok || v != 1 {
		t.Error("the dimension of the `u`s type should be 1")
	}

	v := p.Find("v").(*ast.Var)
	typ = v.Type.(*ast.ArrayType)
	c, ok = typ.Dim.(*ast.ConstValue)
	if !ok {
		t.Fatal("the dimension of `v`s type is not a constant")
	}
	if c.Typ != ast.BuiltinInt {
		t.Error("the dimension of the `v`s type should have type `int`")
	}
	if v, ok := c.Value.(ast.Int); !ok || v != 1 {
		t.Error("the dimension of the `v`s type should be 1")
	}
}

func TestInfer04(t *testing.T) {
	p, err := compilePackage("_test/src/infer", []string{"infer-04.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	B := p.Find("B").(*ast.TypeDecl)
	X := &B.Type.(*ast.StructType).Fields[0]
	tX := X.Type.(*ast.ArrayType)
	c, ok := tX.Dim.(*ast.ConstValue)
	if !ok {
		t.Fatal("the dimension of `B.X`s type is not a constant")
	}
	if c.Typ != ast.BuiltinInt {
		t.Error("the dimension of the `B.X`s type should have type `int`")
	}
	if v, ok := c.Value.(ast.Int); !ok || v != 2 {
		t.Error("the dimension of the `B.X`s type should be 2")
	}
}

func TestInfer05(t *testing.T) {
	p, err := compilePackage("_test/src/infer", []string{"infer-05.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	P := p.Find("P").(*ast.TypeDecl)
	Q := p.Find("Q").(*ast.TypeDecl)

	a := p.Find("a").(*ast.Var)
	if a.Type != P {
		t.Error("the type of `a` must be inferred to `P`")
		t.Log(a.Type)
	}
	b := p.Find("b").(*ast.Var)
	if b.Type != Q {
		t.Error("the type of `b` must be inferred to `Q`")
		t.Log(b.Type)
	}

	S := p.Find("S").(*ast.TypeDecl).Type.(*ast.StructType)
	tX := S.Fields[0].Type.(*ast.ArrayType)
	c, ok := tX.Dim.(*ast.ConstValue)
	if !ok {
		t.Fatal("the dimension of the `S.X`s type is not a constant")
	}
	if c.Typ != ast.BuiltinInt {
		t.Error("the dimension of the `S.X`s type should have type `int`")
	}
	if i, ok := c.Value.(ast.Int); !ok || i != 1 {
		t.Error("the dimension of the `S.X`s type should be 1")
	}
}

func TestInfer06(t *testing.T) {
	p, err := compilePackage("_test/src/infer", []string{"infer-06.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	tX := p.Find("s").(*ast.Var).Type.(*ast.StructType).Fields[1].Type
	c, ok := tX.(*ast.ArrayType).Dim.(*ast.ConstValue)
	if !ok {
		t.Fatal("the dimension of the `s.X`s type is not a constant")
	}
	if c.Typ != ast.BuiltinInt {
		t.Error("the dimension of the `S.X`s type should have type `int`")
	}
	if i, ok := c.Value.(ast.Int); !ok || i != 1 {
		t.Error("the dimension of the `s.X`s type should be 1")
	}
}

func TestInfer07(t *testing.T) {
	expectError(t, "_test/src/infer", []string{"infer-07.go"}, "evaluation loop")
}

func TestInfer08(t *testing.T) {
	expectError(t, "_test/src/infer", []string{"infer-08.go"}, "evaluation loop")
}

func TestInfer09(t *testing.T) {
	p, err := compilePackage("_test/src/infer", []string{"infer-09.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	a := p.Find("a").(*ast.Var)
	tX := a.Type.(*ast.StructType).Fields[0].Type
	c, ok := tX.(*ast.ArrayType).Dim.(*ast.ConstValue)
	if !ok {
		t.Fatal("the dimension of the `a.X`s type is not a constant")
	}
	if c.Typ != ast.BuiltinInt {
		t.Error("the dimension of the `a.X`s type should have type `int`")
	}
	if i, ok := c.Value.(ast.Int); !ok || i != 1 {
		t.Error("the dimension of the `a.X`s type should be 1")
	}

	b := p.Find("a").(*ast.Var)
	tV := b.Type.(*ast.StructType).Fields[1].Type
	c, ok = tV.(*ast.ArrayType).Dim.(*ast.ConstValue)
	if !ok {
		t.Fatal("the dimension of the `a.X`s type is not a constant")
	}
	if c.Typ != ast.BuiltinInt {
		t.Error("the dimension of the `b.X`s type should have type `int`")
	}
	if i, ok := c.Value.(ast.Int); !ok || i != 1 {
		t.Error("the dimension of the `b.X`s type should be 1")
	}
}

func TestInfer10(t *testing.T) {
	p, err := compilePackage("_test/src/infer", []string{"infer-10.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	s := p.Find("s").(*ast.Var)
	tX := s.Type.(*ast.StructType).Fields[0].Type
	c, ok := tX.(*ast.ArrayType).Dim.(*ast.ConstValue)
	if !ok {
		t.Fatal("the dimension of the `a.X`s type is not a constant")
	}
	if c.Typ != ast.BuiltinInt {
		t.Error("the dimension of the `a.X`s type should have type `int`")
	}
	if i, ok := c.Value.(ast.Int); !ok || i != 1 {
		t.Error("the dimension of the `a.X`s type should be 1")
	}
}

func TestInfer11(t *testing.T) {
	p, err := compilePackage("_test/src/infer", []string{"infer-11.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	a := p.Find("a").(*ast.Var)
	c, ok := a.Type.(*ast.ArrayType).Dim.(*ast.ConstValue)
	if !ok {
		t.Fatal("the dimension of the `a`s type is not a constant")
	}
	if c.Typ != ast.BuiltinInt {
		t.Error("the dimension of the `a.X`s type should have type `int`")
	}
	if i, ok := c.Value.(ast.Int); !ok || i != 1 {
		t.Error("the dimension of the `a`s type should be 1")
	}

	aa := p.Find("aa").(*ast.Var)
	c, ok = aa.Type.(*ast.ArrayType).Dim.(*ast.ConstValue)
	if !ok {
		t.Fatal("the dimension of the `aa`s type is not a constant")
	}
	if c.Typ != ast.BuiltinInt {
		t.Error("the dimension of the `aa`s type should have type `int`")
	}
	if i, ok := c.Value.(ast.Int); !ok || i != 1 {
		t.Error("the dimension of the `aa`s type should be 1")
	}

	b := p.Find("b").(*ast.Var)
	c, ok = b.Type.(*ast.ArrayType).Dim.(*ast.ConstValue)
	if !ok {
		t.Fatal("the dimension of the `b`s type is not a constant")
	}
	if c.Typ != ast.BuiltinInt {
		t.Error("the dimension of the `b`s type should have type `int`")
	}
	if i, ok := c.Value.(ast.Int); !ok || i != 1 {
		t.Error("the dimension of the `b`s type should be 1")
	}

	bb := p.Find("bb").(*ast.Var)
	c, ok = bb.Type.(*ast.ArrayType).Dim.(*ast.ConstValue)
	if !ok {
		t.Fatal("the dimension of the `a.X`s type is not a constant")
	}
	if c.Typ != ast.BuiltinInt {
		t.Error("the dimension of the `a.X`s type should have type `int`")
	}
	if i, ok := c.Value.(ast.Int); !ok || i != 1 {
		t.Error("the dimension of the `a.X`s type should be 1")
	}
}

func TestInfer12(t *testing.T) {
	expectError(t, "_test/src/infer", []string{"infer-12.go"}, "inference loop")
}

func TestInfer13(t *testing.T) {
	p, err := compilePackage("_test/src/infer", []string{"infer-13.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	a := p.Find("a").(*ast.Var)
	if underlyingType(a.Type) != ast.BuiltinInt {
		t.Error("the type of `a` must be the builtin `int`")
	}
	S := p.Find("S").(*ast.TypeDecl)
	b := p.Find("b").(*ast.Var)
	if b.Type != S {
		t.Error("the type of `b` must be `S`")
	}
}

func TestInfer14(t *testing.T) {
	expectError(t, "_test/src/infer", []string{"infer-14.go"}, "inference loop")
}
