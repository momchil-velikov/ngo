package typecheck

import (
	"golang/ast"
	"testing"
)

func TestArrayLiteral(t *testing.T) {
	p, err := compilePackage("_test/src/comp", []string{"array.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	A := p.Find("A").(*ast.Var)
	typ := unnamedType(A.Type).(*ast.ArrayType)
	c, ok := typ.Dim.(*ast.ConstValue)
	if !ok {
		t.Fatal("the dimension of the `A` is not a constant")
	}
	if c.Typ != ast.BuiltinInt {
		t.Error("the dimension of the `A` should have type `int`")
	}
	if i, ok := c.Value.(ast.Int); !ok || i != 3 {
		t.Error("the dimension of the `A` should be 3")
	}

	B := p.Find("B").(*ast.Var)
	typ = B.Type.(*ast.ArrayType)
	c, ok = typ.Dim.(*ast.ConstValue)
	if !ok {
		t.Fatal("the dimension of the `B` is not a constant")
	}
	if c.Typ != ast.BuiltinInt {
		t.Error("the dimension of the `B` should have type `int`")
	}
	if i, ok := c.Value.(ast.Int); !ok || i != 2 {
		t.Error("the dimension of the `B` should be 2")
	}
	e0 := B.Init.RHS[0].(*ast.CompLiteral).Elts[0].Elt.(*ast.CompLiteral)
	typ = unnamedType(e0.Typ).(*ast.ArrayType)
	c, ok = typ.Dim.(*ast.ConstValue)
	if c.Typ != ast.BuiltinInt {
		t.Error("the dimension of the `B[0]` should have type `int`")
	}
	if i, ok := c.Value.(ast.Int); !ok || i != 3 {
		t.Error("the dimension of the `B[0]` should be 3")
	}
	e1 := B.Init.RHS[0].(*ast.CompLiteral).Elts[1].Elt.(*ast.CompLiteral)
	typ = unnamedType(e1.Typ).(*ast.ArrayType)
	c, ok = typ.Dim.(*ast.ConstValue)
	if c.Typ != ast.BuiltinInt {
		t.Error("the dimension of the `B[1]` should have type `int`")
	}
	if i, ok := c.Value.(ast.Int); !ok || i != 3 {
		t.Error("the dimension of the `B[1]` should be 3")
	}

	C := p.Find("C").(*ast.Var)
	typ = C.Type.(*ast.ArrayType)
	c, ok = typ.Dim.(*ast.ConstValue)
	if !ok {
		t.Fatal("the dimension of the `C` is not a constant")
	}
	if c.Typ != ast.BuiltinInt {
		t.Error("the dimension of the `C` should have type `int`")
	}
	if i, ok := c.Value.(ast.Int); !ok || i != 3 {
		t.Error("the dimension of the `C` should be 3")
	}

	D := p.Find("D").(*ast.Var)
	typ = D.Type.(*ast.ArrayType)
	c, ok = typ.Dim.(*ast.ConstValue)
	if !ok {
		t.Fatal("the dimension of the `D` is not a constant")
	}
	if c.Typ != ast.BuiltinInt {
		t.Error("the dimension of the `D` should have type `int`")
	}
	if i, ok := c.Value.(ast.Int); !ok || i != 2 {
		t.Error("the dimension of the `D` should be 2")
	}

	e0 = D.Init.RHS[0].(*ast.CompLiteral).Elts[0].Elt.(*ast.CompLiteral)
	typ = e0.Typ.(*ast.ArrayType)
	c, ok = typ.Dim.(*ast.ConstValue)
	if c.Typ != ast.BuiltinInt {
		t.Error("the dimension of the `D[0]` should have type `int`")
	}
	if i, ok := c.Value.(ast.Int); !ok || i != 4 {
		t.Logf("%#v\n", i)
		t.Error("the dimension of the `D[0]` should be 4")
	}
	e1 = D.Init.RHS[0].(*ast.CompLiteral).Elts[1].Elt.(*ast.CompLiteral)
	typ = e1.Typ.(*ast.ArrayType)
	c, ok = typ.Dim.(*ast.ConstValue)
	if c.Typ != ast.BuiltinInt {
		t.Error("the dimension of the `D[1]` should have type `int`")
	}
	if i, ok := c.Value.(ast.Int); !ok || i != 4 {
		t.Error("the dimension of the `D[1]` should be 4")
	}

	E := p.Find("E").(*ast.Var)
	typ = E.Type.(*ast.ArrayType)
	c, ok = typ.Dim.(*ast.ConstValue)
	if !ok {
		t.Fatal("the dimension of the `E` is not a constant")
	}
	if c.Typ != ast.BuiltinInt {
		t.Error("the dimension of the `E` should have type `int`")
	}
	if i, ok := c.Value.(ast.Int); !ok || i != 5 {
		t.Error("the dimension of the `E` should be 5")
	}

	F := p.Find("F").(*ast.Var)
	typ = F.Type.(*ast.ArrayType)
	c, ok = typ.Dim.(*ast.ConstValue)
	if !ok {
		t.Fatal("the dimension of the `F` is not a constant")
	}
	if c.Typ != ast.BuiltinInt {
		t.Error("the dimension of the `F` should have type `int`")
	}
	if i, ok := c.Value.(ast.Int); !ok || i != 5 {
		t.Error("the dimension of the `F` should be 5")
	}
}

func TestArrayLiteralErr(t *testing.T) {
	expectError(t, "_test/src/comp", []string{"array-err-1.go"},
		"unspecified array length not allowed")
	expectError(t, "_test/src/comp", []string{"array-err-2.go"},
		"2.100000 (untyped) cannot be converted to int")
	expectError(t, "_test/src/comp", []string{"array-err-3.go"},
		"array index must be a non-negative `int` constant")
	expectError(t, "_test/src/comp", []string{"array-err-4.go"},
		"1.200000 (untyped) cannot be converted to int")
	expectError(t, "_test/src/comp", []string{"array-err-5.go"},
		"index out of bounds")
	expectError(t, "_test/src/comp", []string{"array-err-6.go"},
		"index out of bounds")
	expectError(t, "_test/src/comp", []string{"array-err-7.go"},
		"index out of bounds")
	expectError(t, "_test/src/comp", []string{"array-err-8.go"},
		"duplicate index in array/slice literal: 1")
	expectError(t, "_test/src/comp", []string{"array-err-9.go"},
		"duplicate index in array/slice literal: 3")
	expectError(t, "_test/src/comp", []string{"array-err-10.go"},
		"unspecified array length not allowed")
	expectError(t, "_test/src/comp", []string{"array-err-11.go"},
		"unspecified array length not allowed")
}

func TestSliceLiteral(t *testing.T) {
	_, err := compilePackage("_test/src/comp", []string{"slice.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMapLiteral(t *testing.T) {
	_, err := compilePackage("_test/src/comp", []string{"map.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMapLiteralErr(t *testing.T) {
	expectError(t, "_test/src/comp", []string{"map-err-1.go"},
		"all elements in a map composite literal must have a key")
	expectError(t, "_test/src/comp", []string{"map-err-2.go"},
		"duplicate key in map literal: true")
	expectError(t, "_test/src/comp", []string{"map-err-3.go"},
		"duplicate key in map literal: -1")
	expectError(t, "_test/src/comp", []string{"map-err-4.go"},
		"duplicate key in map literal: 1")
	expectError(t, "_test/src/comp", []string{"map-err-5.go"},
		"duplicate key in map literal: 5.8")
	expectError(t, "_test/src/comp", []string{"map-err-6.go"},
		"duplicate key in map literal: foo")
}

func TestMapStructLiteral(t *testing.T) {
	_, err := compilePackage("_test/src/comp", []string{"struct.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestStructLiteralErr(t *testing.T) {
	expectError(t, "_test/src/comp", []string{"struct-err-1.go"},
		"key is not a field name")
	expectError(t, "_test/src/comp", []string{"struct-err-2.go"},
		"type does not have a field named Z")
	expectError(t, "_test/src/comp", []string{"struct-err-3.go"},
		"type does not have a field named X")
	expectError(t, "_test/src/comp", []string{"struct-err-4.go"},
		"struct literal mixes field:value and value initializers")
	expectError(t, "_test/src/comp", []string{"struct-err-5.go"},
		"the literal must contain exactly one element for each struct field")
	expectError(t, "_test/src/comp", []string{"struct-err-6.go"},
		"duplicate field name in struct literal: X")
	expectError(t, "_test/src/comp", []string{"struct-err-7.go"},
		"missing type for composite literal")
}

func TestLiteralErr(t *testing.T) {
	expectError(t, "_test/src/comp", []string{"lit-err.go"},
		"invalid type for composite literal")
}
