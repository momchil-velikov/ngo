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
	if !ok || c.Typ != ast.BuiltinUntypedBool || !bool(b) {
		t.Fatal("expected untyped boolean value `true` after evaluation of `A`")
	}

	B := p.Find("B").(*ast.Const)
	c, ok = B.Init.(*ast.ConstValue)
	if !ok {
		t.Fatal("initializer of `B` is not a ConstValue")
	}
	b, ok = c.Value.(ast.Bool)
	if !ok || c.Typ != ast.BuiltinBool || !bool(b) {
		t.Fatal("expected typed boolean value `true` after evaluation of `B`")
	}
}

func TestConvBoolErr(t *testing.T) {
	expectError(t, "_test/src/conv", []string{"bool-err-1.go"},
		"true (`untyped bool`) cannot be converted to `int`")
	expectError(t, "_test/src/conv", []string{"bool-err-2.go"},
		"`*int` is not a valid constant type")
	expectError(t, "_test/src/conv", []string{"bool-err-3.go"},
		"true (`untyped bool`) cannot be converted to `string`")
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
		{"f32", "float32", "-128.0"},
		{"f64", "float64", "-128.0"},
		{"g32", "float32", "128.0"},
		{"g64", "float64", "128.0"},
		{"h32", "float32", "-128.0"},
		{"h64", "float64", "128.0"},
		{"p32", "float32", "16777216.0"},
		{"p64", "float64", "4503599627370497.0"},
		{"q32", "float32", "2147483648.0"},
		{"q64", "float64", "9223372036854775808.0"},
		{"c64", "complex64", "(-128.0 + 0.0i)"},
		{"c128", "complex128", "(-128.0 + 0.0i)"},
		{"d64", "complex64", "(128.0 + 0.0i)"},
		{"d128", "complex128", "(128.0 + 0.0i)"},
		{"e64", "complex64", "(-128.0 + 0.0i)"},
		{"e128", "complex128", "(128.0 + 0.0i)"},
	} {
		n := p.Find(c.name).(*ast.Const)
		x := n.Init.(*ast.ConstValue)

		if name := x.Typ.String(); name != c.typ {
			t.Errorf("unexpected type `%s` for `%s`\n", name, n.Name)
		}
		if s := x.String(); s != c.val {
			t.Errorf("unexpected value `%s` for `%s`\n", s, n.Name)
		}
	}
}

func TestConvIntErr(t *testing.T) {
	const dir = "_test/src/conv"
	for i, e := range []string{
		"1 (`untyped int`) cannot be converted to `bool`",
		"1 (`uint32`) cannot be converted to `bool`",
		"-129 (`untyped int`) cannot be converted to `int8`",
		"128 (`untyped int`) cannot be converted to `int8`",
		"-129 (`int32`) cannot be converted to `int8`",
		"128 (`uint32`) cannot be converted to `int8`",
		"128 (`int32`) cannot be converted to `int8`",
		"-1 (`untyped int`) cannot be converted to `uint8`",
		"256 (`untyped int`) cannot be converted to `uint8`",
		"-1 (`int32`) cannot be converted to `uint8`",
		"256 (`uint32`) cannot be converted to `uint8`",
		"256 (`int32`) cannot be converted to `uint8`",
		"-32769 (`untyped int`) cannot be converted to `int16`",
		"32768 (`untyped int`) cannot be converted to `int16`",
		"-32769 (`int32`) cannot be converted to `int16`",
		"32768 (`uint32`) cannot be converted to `int16`",
		"32768 (`int32`) cannot be converted to `int16`",
		"-1 (`untyped int`) cannot be converted to `uint16`",
		"65536 (`untyped int`) cannot be converted to `uint16`",
		"-1 (`int32`) cannot be converted to `uint16`",
		"65536 (`uint32`) cannot be converted to `uint16`",
		"65536 (`int32`) cannot be converted to `uint16`",
		"-2147483649 (`untyped int`) cannot be converted to `int32`",
		"2147483648 (`untyped int`) cannot be converted to `int32`",
		"-2147483649 (`int64`) cannot be converted to `int32`",
		"2147483648 (`uint32`) cannot be converted to `int32`",
		"2147483648 (`int64`) cannot be converted to `int32`",
		"-1 (`untyped int`) cannot be converted to `uint32`",
		"4294967296 (`untyped int`) cannot be converted to `uint32`",
		"-1 (`int64`) cannot be converted to `uint32`",
		"4294967296 (`uint64`) cannot be converted to `uint32`",
		"4294967296 (`int64`) cannot be converted to `uint32`",
		"-9223372036854775809 (`untyped int`) cannot be converted to `int64`",
		"9223372036854775808 (`untyped int`) cannot be converted to `int64`",
		"9223372036854775808 (`uint64`) cannot be converted to `int64`",
		"-1 (`untyped int`) cannot be converted to `uint64`",
		"18446744073709551616 (`untyped int`) cannot be converted to `uint64`",
		"-1 (`int64`) cannot be converted to `uint64`",
		"-9223372036854775809 (`untyped int`) cannot be converted to `int`",
		"9223372036854775808 (`untyped int`) cannot be converted to `int`",
		"9223372036854775808 (`uint`) cannot be converted to `int`",
		"-1 (`untyped int`) cannot be converted to `uint`",
		"18446744073709551616 (`untyped int`) cannot be converted to `uint`",
		"-1 (`int`) cannot be converted to `uint`",
		"-1 (`untyped int`) cannot be converted to `uintptr`",
		"18446744073709551616 (`untyped int`) cannot be converted to `uintptr`",
		"-1 (`int`) cannot be converted to `uintptr`",
		"1000000000000000000000000000000000000000 (`untyped int`) cannot be converted to `float32`",

		"1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000 (`untyped int`) cannot be converted to `float64`",
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
		{"C0", "float32", "2.0"},
		{"C1", "float32", "2.0"},
		{"D0", "float64", "1.999999"},
		{"D1", "float64", "2.0"},
		{"D2", "float64", "2.0"},
		{"E0", "complex64", "(2.0 + 0.0i)"},
		{"E1", "complex64", "(2.0 + 0.0i)"},
		{"F0", "complex128", "(1.999999 + 0.0i)"},
		{"F1", "complex128", "(2.0 + 0.0i)"},
		{"G0", "int64", "6755399441055745"},
		{"G1", "uint64", "9223372036854777856"},
	} {
		n := p.Find(c.name).(*ast.Const)
		x := n.Init.(*ast.ConstValue)

		if name := x.Typ.String(); name != c.typ {
			t.Errorf("unexpected type `%s` for `%s`\n", name, n.Name)
		}

		if s := x.String(); s != c.val {
			t.Errorf("unexpected value `%s` for `%s`\n", s, n.Name)
		}
	}
}

func TestConvFloatErr(t *testing.T) {

	const dir = "_test/src/conv"
	for i, e := range []string{
		"1.1 (`untyped float`) cannot be converted to `bool`",
		"1.1 (`float32`) cannot be converted to `bool`",
		"1.1 (`untyped float`) cannot be converted to `string`",
		"1.1 (`float64`) cannot be converted to `string`",
		"1.1 (`untyped float`) cannot be converted to `int`",
		"1.1 (`float64`) cannot be converted to `int`",
		"9223372036854777856.0 (`float64`) cannot be converted to `int64`",
		"18446744073709555712.0 (`float64`) cannot be converted to `uint64`",
		"-9223372036854779904.0 (`float64`) cannot be converted to `int64`",
		"179769313486231589999999999999999999999999999999999999999999999999999999999999777816481292516925359185725128383638673141453317136668426630023755954681353297585823163889595831750739892328339287594669941410908927299736853716969618046803012291434577195764887115186790482212808716743819624705152288329332701528064.0 (`untyped float`) cannot be converted to `float64`",
	} {
		expectError(t, dir, []string{fmt.Sprintf("float-err-%02d.go", i+1)}, e)
	}
}
