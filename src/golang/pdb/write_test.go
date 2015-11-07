package pdb

import (
	"bytes"
	"golang/ast"
	"golang/scanner"
	"testing"
)

func keepEncoding(t *testing.T, fn func(*Encoder) error) []byte {
	buf := &bytes.Buffer{}
	lim := 0
	w := LimitWriter(buf, lim)
	enc := new(Encoder).Init(w)
	for err := fn(enc); err != nil; {
		if err != ErrorNoSpace {
			t.Error("expected I/O error")
		}
		lim++
		buf.Reset()
		w.N = lim
		err = fn(enc)
	}
	return buf.Bytes()
}

func TestWriteBuiltinType(t *testing.T) {
	ns := []string{
		"#nil", "bool", "byte",
		"uint8", "uint16", "uint32", "uint64",
		"int8", "int16", "int32", "int64", "rune",
		"float32", "float64",
		"complex64", "complex128",
		"uint", "int", "uintptr",
		"string",
	}

	tk := []byte{
		_NIL,
		_BOOL,
		_UINT8,
		_UINT8,
		_UINT16,
		_UINT32,
		_UINT64,
		_INT8,
		_INT16,
		_INT32,
		_INT64,
		_INT32,
		_FLOAT32,
		_FLOAT64,
		_COMPLEX64,
		_COMPLEX128,
		_UINT,
		_INT,
		_UINTPTR,
		_STRING,
	}

	pkg := &ast.Package{}

	for i, n := range ns {
		d := pkg.Lookup(n)
		if d == nil {
			t.Fatal("cannot look up", n)
		}
		_, ok := d.(*ast.TypeDecl)
		if !ok {
			t.Fatal(n, "is not a type declaration")
		}

		buf := keepEncoding(t, func(enc *Encoder) error { return writeDecl(enc, d) })
		exp := []byte{_TYPE_DECL, 0, 0, byte(len(n))}
		exp = append(exp, []byte(n)...)
		exp = append(exp, tk[i])
		expect_eq(t, "write builtin types", buf, exp)
	}
}

func TestWriteArrayType(t *testing.T) {
	typ := &ast.ArrayType{
		Dim: &ast.Literal{Value: []byte{'1', '0'}},
		Elt: &ast.BuiltinType{Kind: ast.BUILTIN_UINT8},
	}
	buf := keepEncoding(t, func(enc *Encoder) error { return writeType(enc, typ) })
	expect_eq(t, "write array type", buf, []byte{_ARRAY, 10, _UINT8})
}

func TestWriteSliceType(t *testing.T) {
	typ := &ast.SliceType{
		Elt: &ast.BuiltinType{Kind: ast.BUILTIN_UINT8},
	}
	buf := keepEncoding(t, func(enc *Encoder) error { return writeType(enc, typ) })
	expect_eq(t, "write slice type", buf, []byte{_SLICE, _UINT8})
}

func TestWritePtrType(t *testing.T) {
	typ := &ast.PtrType{
		Base: &ast.BuiltinType{Kind: ast.BUILTIN_UINT8},
	}
	buf := keepEncoding(t, func(enc *Encoder) error { return writeType(enc, typ) })
	expect_eq(t, "write ptr type", buf, []byte{_PTR, _UINT8})
}

func TestWriteMapType(t *testing.T) {
	typ := &ast.MapType{
		Key: &ast.BuiltinType{Kind: ast.BUILTIN_STRING},
		Elt: &ast.BuiltinType{Kind: ast.BUILTIN_UINT8},
	}
	buf := keepEncoding(t, func(enc *Encoder) error { return writeType(enc, typ) })
	expect_eq(t, "write map type", buf, []byte{_MAP, _STRING, _UINT8})
}

func TestWriteChanType(t *testing.T) {
	typ := &ast.ChanType{
		Send: true,
		Recv: true,
		Elt:  &ast.BuiltinType{Kind: ast.BUILTIN_UINT8},
	}
	buf := keepEncoding(t, func(enc *Encoder) error { return writeType(enc, typ) })
	expect_eq(t, "write chan type", buf, []byte{_CHAN, 3, _UINT8})
}

func TestWriteStructType1(t *testing.T) {
	typ := &ast.StructType{}
	buf := keepEncoding(t, func(enc *Encoder) error { return writeType(enc, typ) })
	expect_eq(t, "write struct type", buf, []byte{_STRUCT, 0})
}

func TestWriteStructType2(t *testing.T) {
	typ := &ast.StructType{
		Fields: []ast.Field{
			{Name: "a", Type: &ast.BuiltinType{Kind: ast.BUILTIN_BOOL}},
			{Name: "b", Type: &ast.BuiltinType{Kind: ast.BUILTIN_BOOL}, Tag: "xy"},
		},
	}
	buf := keepEncoding(t, func(enc *Encoder) error { return writeType(enc, typ) })
	expect_eq(t, "write struct type",
		buf,
		[]byte{
			_STRUCT,
			2,
			1, 'a', _BOOL, 0,
			1, 'b', _BOOL, 2, 'x', 'y',
		},
	)
}

func TestWriteFuncType1(t *testing.T) {
	typ := &ast.FuncType{}
	buf := keepEncoding(t, func(enc *Encoder) error { return writeType(enc, typ) })
	expect_eq(t, "write func type", buf, []byte{_FUNC, 0, 0, 0})
}

func TestWriteFuncType2(t *testing.T) {
	typ := &ast.FuncType{
		Params: []ast.Param{
			{Name: "x", Type: &ast.BuiltinType{Kind: ast.BUILTIN_STRING}},
			{Name: "y", Type: &ast.BuiltinType{Kind: ast.BUILTIN_STRING}},
		},
		Returns: []ast.Param{
			{Type: &ast.BuiltinType{Kind: ast.BUILTIN_STRING}},
		},
		Var: true,
	}
	buf := keepEncoding(t, func(enc *Encoder) error { return writeType(enc, typ) })
	expect_eq(t, "write func type",
		buf,
		[]byte{
			_FUNC,
			2,
			1, 'x', _STRING,
			1, 'y', _STRING,
			1,
			0, _STRING,
			1},
	)
}

func TestWriteIfaceType1(t *testing.T) {
	typ := &ast.InterfaceType{}
	buf := keepEncoding(t, func(enc *Encoder) error { return writeType(enc, typ) })
	expect_eq(t, "write iface type", buf, []byte{_IFACE, 0})
}

func TestWriteIfaceType2(t *testing.T) {
	typ := &ast.InterfaceType{
		Methods: []ast.MethodSpec{
			{Type: &ast.InterfaceType{}},
			{Name: "F", Type: &ast.FuncType{}},
		},
	}
	buf := keepEncoding(t, func(enc *Encoder) error { return writeType(enc, typ) })
	expect_eq(t, "write iface type",
		buf,
		[]byte{
			_IFACE,
			2,
			_IFACE, 0,
			_FUNC, 0, 0, 0, 1, 'F',
		},
	)
}

func TestWriteTypename1(t *testing.T) {
	typ := &ast.Typename{
		Decl: &ast.TypeDecl{
			Name: "S",
			Type: &ast.PtrType{Base: &ast.BuiltinType{Kind: ast.BUILTIN_FLOAT32}}},
	}
	buf := keepEncoding(t, func(enc *Encoder) error { return writeType(enc, typ) })
	expect_eq(t, "write typename",
		buf,
		[]byte{
			_TYPENAME,
			0,
			1, 'S',
		},
	)

	typ.Decl.File = &ast.File{}
	b := &bytes.Buffer{}
	enc := new(Encoder).Init(b)
	if err := writeType(enc, typ); err != nil {
		t.Fatal(err)
	}
	buf = b.Bytes()
	expect_eq(t, "write typename",
		buf,
		[]byte{
			_TYPENAME,
			1,
			1, 'S',
		},
	)

	typ.Decl.File.Pkg = &ast.Package{No: 42}
	b.Reset()
	if err := writeType(enc, typ); err != nil {
		t.Fatal(err)
	}
	buf = b.Bytes()
	expect_eq(t, "write typename",
		buf,
		[]byte{
			_TYPENAME,
			42,
			1, 'S',
		},
	)
}

func TestWriteVarDecl(t *testing.T) {
	v := &ast.Var{}
	buf := keepEncoding(t, func(e *Encoder) error { return writeDecl(e, v) })
	expect_eq(t, "var decl",
		buf,
		[]byte{
			_VAR_DECL,
			0, 0, // file, off
			0, // name
		},
	)

	v.Off = 12
	v.File = &ast.File{No: 42}
	v.Name = "xyz"
	v.Type = &ast.BuiltinType{Kind: ast.BUILTIN_INT32}
	buf = keepEncoding(t, func(e *Encoder) error { return writeDecl(e, v) })
	expect_eq(t, "var decl",
		buf,
		[]byte{
			_VAR_DECL,
			42, 12, // file, off
			3, 'x', 'y', 'z', // name
			_INT32,
		},
	)
}

func TestWriteConstDecl(t *testing.T) {
	c := &ast.Const{}
	buf := keepEncoding(t, func(e *Encoder) error { return writeDecl(e, c) })
	expect_eq(t, "const decl",
		buf,
		[]byte{
			_CONST_DECL,
			0, 0, // file, off
			0, // name
		},
	)

	c.Off = 12
	c.File = &ast.File{No: 42}
	c.Name = "xyz"
	c.Type = &ast.BuiltinType{Kind: ast.BUILTIN_INT32}
	buf = keepEncoding(t, func(e *Encoder) error { return writeDecl(e, c) })
	expect_eq(t, "const decl",
		buf,
		[]byte{
			_CONST_DECL,
			42, 12, // file, off
			3, 'x', 'y', 'z', // name
			_INT32,
		},
	)
}

func TestWriteFuncDecl(t *testing.T) {
	fn := &ast.FuncDecl{
		Off:  42,
		File: &ast.File{No: 11},
		Func: ast.Func{
			Sig: &ast.FuncType{},
		},
	}
	buf := keepEncoding(t, func(e *Encoder) error { return writeDecl(e, fn) })
	expect_eq(t, "func decl",
		buf,
		[]byte{
			_FUNC_DECL,
			11, 42, // file, off
			0,    // name
			_NIL, // receiver type
			_FUNC, 0, 0, 0,
		},
	)

	fn.Name = "F"
	fn.Func.Recv = &ast.Param{
		Type: &ast.PtrType{
			Base: &ast.Typename{
				Decl: &ast.TypeDecl{
					Name: "S",
					Type: &ast.StructType{},
				},
			},
		},
	}
	buf = keepEncoding(t, func(e *Encoder) error { return writeDecl(e, fn) })
	expect_eq(t, "func decl",
		buf,
		[]byte{
			_FUNC_DECL,
			11, 42, // file, off
			1, 'F', // name
			_PTR, _TYPENAME, 0, 1, 'S', // receiver type
			_FUNC, 0, 0, 0, // func type
		},
	)
}

func TestWriteFile1(t *testing.T) {
	pkg := &ast.Package{}
	file := &ast.File{
		Pkg: pkg,
		Name: "1234567890123456789012345678901234567890123456789012345678901234567890" +
			"123456789012345678901234567890123456789012345678901234567890.go",
	}

	_ = keepEncoding(t, func(e *Encoder) error { return writeFile(e, file) })

	file.Name = "xx.go"
	buf := keepEncoding(t, func(e *Encoder) error { return writeFile(e, file) })
	expect_eq(t, "empty file",
		buf,
		[]byte{
			5, 'x', 'x', '.', 'g', 'o',
			0, // no source map
		})
}

func TestWriteFile2(t *testing.T) {
	pkg := &ast.Package{}
	file := &ast.File{
		Pkg:    pkg,
		Name:   "xx.go",
		SrcMap: scanner.SourceMap{},
	}
	file.SrcMap.AddLine(1)
	file.SrcMap.AddLine(2)
	file.SrcMap.AddLine(3)

	buf := keepEncoding(t, func(e *Encoder) error { return writeFile(e, file) })
	expect_eq(t, "empty file",
		buf,
		[]byte{
			5, 'x', 'x', '.', 'g', 'o',
			1, 2, 3, 0, // source map
		})
}

func TestWritePackage1(t *testing.T) {
	pkg := &ast.Package{}

	buf := bytes.Buffer{}
	if err := Write(&buf, pkg); err != nil {
		t.Fatal(err)
	}

	exp := []byte{
		VERSION,
		0, // no name
		0, // no deps
		0, // no files
		0, // no decls
	}
	expect_eq(t, "empty pkg", buf.Bytes(), exp)
}

func TestWritePackage2(t *testing.T) {
	pkg := &ast.Package{
		Name:  "test",
		Deps:  make(map[string]*ast.Package),
		Decls: make(map[string]ast.Symbol),
	}
	deps := []*ast.Package{
		&ast.Package{Name: "depX", Sig: [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1}},
		&ast.Package{Name: "dep1", Sig: [20]byte{5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5}},
	}
	for _, d := range deps {
		pkg.Deps[d.Name] = d
	}

	d0 := &ast.Var{
		Off:  12,
		File: &ast.File{No: 42},
		Name: "xyz",
		Type: &ast.BuiltinType{Kind: ast.BUILTIN_INT32},
	}
	pkg.Decls[d0.Name] = d0

	d1 := &ast.Var{
		Off:  12,
		File: &ast.File{No: 42},
		Name: "Xyz",
		Type: &ast.BuiltinType{Kind: ast.BUILTIN_INT32},
	}
	pkg.Decls[d1.Name] = d1

	buf := keepEncoding(t, func(e *Encoder) error { return writePkg(e, pkg) })
	expect_eq(t, "pkg",
		buf,
		[]byte{
			VERSION,
			4, 't', 'e', 's', 't', // name
			2,                     // number of deps
			4, 'd', 'e', 'p', '1', // import path
			5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 0, 0, 0, 0, 0, 0, 0, 0, 0, // sig
			4, 'd', 'e', 'p', 'X', // import path
			1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, // sig
			0, // no files
			_VAR_DECL,
			42, 12, // file, off
			3, 'X', 'y', 'z', // name
			_INT32,
			_END, // end decls
		},
	)
}

func TestWritePackage3(t *testing.T) {
	pkg := &ast.Package{
		Name: "test",
		Files: []*ast.File{
			&ast.File{Name: "xx.go"},
			&ast.File{Name: "yy.go"},
		},
	}
	for _, f := range pkg.Files {
		f.SrcMap.AddLine(3)
		f.SrcMap.AddLine(14)
		f.SrcMap.AddLine(15)
	}

	buf := keepEncoding(t, func(e *Encoder) error { return writePkg(e, pkg) })
	expect_eq(t, "pkg",
		buf,
		[]byte{
			VERSION,
			4, 't', 'e', 's', 't', // name
			0, // deps
			2, // files
			5, 'x', 'x', '.', 'g', 'o',
			3, 14, 15, 0,
			5, 'y', 'y', '.', 'g', 'o',
			3, 14, 15, 0,
			0, // decls
		},
	)
}
