package typecheck

import (
	"fmt"
	"golang/ast"
	"testing"
)

func TestConvNilErr(t *testing.T) {
	expectError(t, "_test/src/conv", []string{"nil-err.go"},
		"const initializer is not a constant")
}

func TestConvBool(t *testing.T) {
	p, err := compilePackage("_test/src/conv", []string{"bool.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	C := p.Find("C").(*ast.Const)
	c := C.Init.(*ast.ConstValue)

	Bool := p.Find("Bool")
	b, ok := c.Value.(ast.Bool)
	if !ok || c.Typ.(*ast.TypeDecl) != Bool || !bool(b) {
		t.Fatal("expected value `true` of type `Bool` after evaluation of `C`")
	}

	A := p.Find("A").(*ast.Const)
	c, ok = A.Init.(*ast.ConstValue)
	if !ok {
		t.Fatal("initializer of `A` is not a ConstValue")
	}
	b, ok = c.Value.(ast.Bool)
	if !ok || c.Typ != nil || !bool(b) {
		t.Fatal("expected untyped boolean value `true` after evaluation of `A`")
	}

	B := p.Find("B").(*ast.Const)
	c, ok = B.Init.(*ast.ConstValue)
	if !ok {
		t.Fatal("initializer of `B` is not a ConstValue")
	}
	b, ok = c.Value.(ast.Bool)
	if !ok || c.Typ.(*ast.TypeDecl).Type != ast.BuiltinBool || !bool(b) {
		t.Fatal("expected typed boolean value `true` after evaluation of `B`")
	}
}

func TestConvBoolErr(t *testing.T) {
	expectError(t, "_test/src/conv", []string{"bool-err-1.go"},
		"true (untyped) cannot be converted to int")
	expectError(t, "_test/src/conv", []string{"bool-err-2.go"},
		"invalid constant type")
	expectError(t, "_test/src/conv", []string{"bool-err-3.go"},
		"true (untyped) cannot be converted to string")
}

func TestConvInt(t *testing.T) {
	p, err := compilePackage("_test/src/conv", []string{"integer.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range []struct{ name, typ, val string }{
		{"u8", "uint8", "255"},
		{"u16", "uint16", "65535"},
		{"u32", "uint32", "4294967295"},
		{"u64", "uint64", "18446744073709551615"},
		{"u", "uint", "18446744073709551615"},
		{"uptr", "uintptr", "18446744073709551615"},
		{"w8", "U8", "255"},
		{"w16", "U16", "255"},
		{"w32", "U32", "255"},
		{"w64", "U64", "255"},
		{"i8", "int8", "-128"},
		{"i16", "int16", "-32768"},
		{"i32", "int32", "-2147483648"},
		{"i64", "int64", "-9223372036854775808"},
		{"i", "int", "-9223372036854775808"},
		{"v8", "I8", "-128"},
		{"v16", "I16", "-128"},
		{"v32", "I32", "-128"},
		{"v64", "I64", "-128"},
		{"s", "string", "\"й\""},
		{"x", "string", "\"\ufffd\""},
		{"y", "string", "\"\ufffd\""},
		{"z0", "string", "\"\ufffd\""},
		{"z1", "string", "\"\ufffd\""},
		{"f32", "float32", "-128.000000"},
		{"f64", "float64", "-128.000000"},
		{"g32", "float32", "128.000000"},
		{"g64", "float64", "128.000000"},
		{"h32", "float32", "-128.000000"},
		{"h64", "float64", "128.000000"},
		{"p32", "float32", "16777216.000000"},
		{"p64", "float64", "4503599627370497.000000"},
		{"q32", "float32", "2147483648.000000"},
		{"q64", "float64", "9223372036854775808.000000"},
		{"c64", "complex64", "(-128.000000+0.000000i)"},
		{"c128", "complex128", "(-128.000000+0.000000i)"},
		{"d64", "complex64", "(128.000000+0.000000i)"},
		{"d128", "complex128", "(128.000000+0.000000i)"},
		{"e64", "complex64", "(-128.000000+0.000000i)"},
		{"e128", "complex128", "(128.000000+0.000000i)"},
	} {
		n := p.Find(c.name).(*ast.Const)
		x := n.Init.(*ast.ConstValue)
		typ := x.Typ.(*ast.TypeDecl).Name

		if typ != c.typ {
			t.Errorf("unexpected type `%s` for `%s`\n", typ, n.Name)
		}
		s := valueToString(builtinType(x.Typ), x.Value)
		if s != c.val {
			t.Errorf("unexpected value `%s` for `%s`\n", s, n.Name)
		}
	}
}

func TestConvIntErr(t *testing.T) {
	const dir = "_test/src/conv"
	for i, e := range []string{
		"1 (untyped) cannot be converted to bool",
		"1 (type uint32) cannot be converted to bool",
		"-129 (untyped) cannot be converted to int8",
		"128 (untyped) cannot be converted to int8",
		"-129 (type int32) cannot be converted to int8",
		"128 (type uint32) cannot be converted to int8",
		"128 (type int32) cannot be converted to int8",
		"-1 (untyped) cannot be converted to uint8",
		"256 (untyped) cannot be converted to uint8",
		"-1 (type int32) cannot be converted to uint8",
		"256 (type uint32) cannot be converted to uint8",
		"256 (type int32) cannot be converted to uint8",
		"-32769 (untyped) cannot be converted to int16",
		"32768 (untyped) cannot be converted to int16",
		"-32769 (type int32) cannot be converted to int16",
		"32768 (type uint32) cannot be converted to int16",
		"32768 (type int32) cannot be converted to int16",
		"-1 (untyped) cannot be converted to uint16",
		"65536 (untyped) cannot be converted to uint16",
		"-1 (type int32) cannot be converted to uint16",
		"65536 (type uint32) cannot be converted to uint16",
		"65536 (type int32) cannot be converted to uint16",
		"-2147483649 (untyped) cannot be converted to int32",
		"2147483648 (untyped) cannot be converted to int32",
		"-2147483649 (type int64) cannot be converted to int32",
		"2147483648 (type uint32) cannot be converted to int32",
		"2147483648 (type int64) cannot be converted to int32",
		"-1 (untyped) cannot be converted to uint32",
		"4294967296 (untyped) cannot be converted to uint32",
		"-1 (type int64) cannot be converted to uint32",
		"4294967296 (type uint64) cannot be converted to uint32",
		"4294967296 (type int64) cannot be converted to uint32",
		"-9223372036854775809 (untyped) cannot be converted to int64",
		"9223372036854775808 (untyped) cannot be converted to int64",
		"9223372036854775808 (type uint64) cannot be converted to int64",
		"-1 (untyped) cannot be converted to uint64",
		"18446744073709551616 (untyped) cannot be converted to uint64",
		"-1 (type int64) cannot be converted to uint64",
		"-9223372036854775809 (untyped) cannot be converted to int",
		"9223372036854775808 (untyped) cannot be converted to int",
		"9223372036854775808 (type uint) cannot be converted to int",
		"-1 (untyped) cannot be converted to uint",
		"18446744073709551616 (untyped) cannot be converted to uint",
		"-1 (type int) cannot be converted to uint",
		"-1 (untyped) cannot be converted to uintptr",
		"18446744073709551616 (untyped) cannot be converted to uintptr",
		"-1 (type int) cannot be converted to uintptr",
		"1000000000000000000000000000000000000000 (untyped) cannot be converted to float32",

		"1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000 (untyped) cannot be converted to float64",
	} {
		expectError(t, dir, []string{fmt.Sprintf("int-err-%02d.go", i+1)}, e)
	}
}

func TestConvFloat(t *testing.T) {
	p, err := compilePackage("_test/src/conv", []string{"float.go"}, nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range []struct{ name, typ, val string }{
		{"Z0", "int32", "0"},
		{"Z1", "int32", "0"},
		{"A0", "int16", "-128"},
		{"A1", "int16", "-128"},
		{"B0", "int16", "128"},
		{"B1", "int16", "128"},
		{"C0", "float32", "2.000000"},
		{"C1", "float32", "2.000000"},
		{"D0", "float64", "1.999999"},
		{"D1", "float64", "2.000000"},
		{"D2", "float64", "2.000000"},
		{"E0", "complex64", "(2.000000+0.000000i)"},
		{"E1", "complex64", "(2.000000+0.000000i)"},
		{"F0", "complex128", "(1.999999+0.000000i)"},
		{"F1", "complex128", "(2.000000+0.000000i)"},
		{"G0", "int64", "6755399441055745"},
		{"G1", "uint64", "9223372036854777856"},
	} {
		n := p.Find(c.name).(*ast.Const)
		x := n.Init.(*ast.ConstValue)
		typ := x.Typ.(*ast.TypeDecl).Name

		if typ != c.typ {
			t.Errorf("unexpected type `%s` for `%s`\n", typ, n.Name)
		}
		s := valueToString(builtinType(x.Typ), x.Value)
		if s != c.val {
			t.Errorf("unexpected value `%s` for `%s`\n", s, n.Name)
		}
	}
}

func TestConvFloatErr(t *testing.T) {

	const dir = "_test/src/conv"
	for i, e := range []string{
		"1.100000 (untyped) cannot be converted to bool",
		"1.100000 (type float32) cannot be converted to bool",
		"1.100000 (untyped) cannot be converted to string",
		"1.100000 (type float64) cannot be converted to string",
		"1.100000 (untyped) cannot be converted to int",
		"1.100000 (type float64) cannot be converted to int",
		"9223372036854777856.000000 (type float64) cannot be converted to int64",
		"18446744073709555712.000000 (type float64) cannot be converted to uint64",
		"-9223372036854779904.000000 (type float64) cannot be converted to int64",
		"179769313486231589999999999999999999999999999999999999999999999999999999999999777816481292516925359185725128383638673141453317136668426630023755954681353297585823163889595831750739892328339287594669941410908927299736853716969618046803012291434577195764887115186790482212808716743819624705152288329332701528064.000000 (untyped) cannot be converted to float64",
	} {
		expectError(t, dir, []string{fmt.Sprintf("float-err-%02d.go", i+1)}, e)
	}
}
