package pdb

import (
	"bytes"
	"golang/ast"
	"golang/scanner"
	"io"
	"reflect"
	"testing"
)

func encode(t *testing.T, fn func(*Writer) error) ([]byte, error) {
	buf := &bytes.Buffer{}
	w := &Writer{}
	w.Encoder.Init(buf)
	err := fn(w)
	return buf.Bytes(), err
}

func decode(t *testing.T, buf []byte, fn func(*Reader)) {
	r := &Reader{}
	r.Decoder.Init(bytes.NewReader(buf))
	fn(r)
}

func keepEncoding(t *testing.T, fn func(*Writer) error) []byte {
	buf := &bytes.Buffer{}
	lim := 0
	lw := LimitWriter(buf, lim)
	w := &Writer{}
	w.Encoder.Init(lw)
	for err := fn(w); err != nil; {
		if err != ErrorNoSpace {
			t.Fatal("expected I/O error, got", err)
		}
		lim++
		buf.Reset()
		lw.N = lim
		err = fn(w)
	}
	return buf.Bytes()
}

func keepDecoding(t *testing.T, bs []byte, fn func(*Reader) error) {
	buf := bytes.NewReader(bs)
	lim := int64(0)
	lr := &io.LimitedReader{R: buf, N: lim}
	r := &Reader{}
	r.Decoder.Init(lr)
	for err := fn(r); err != nil; {
		if err != io.EOF {
			t.Fatal("expected I/O error, got", err)
		}
		lim++
		buf.Seek(0, 0)
		lr.N = lim
		err = fn(r)
	}
}

func decodeType(t *testing.T, buf []byte) ast.Type {
	var typ ast.Type
	keepDecoding(t, buf, func(r *Reader) error {
		var err error
		typ, err = r.readType(nil)
		return err
	})
	return typ
}

func TestWriteBuiltinType(t *testing.T) {
	ts := []ast.Type{
		ast.BuiltinNil,
		ast.BuiltinBool,
		ast.BuiltinUint8,
		ast.BuiltinUint16,
		ast.BuiltinUint32,
		ast.BuiltinUint64,
		ast.BuiltinInt8,
		ast.BuiltinInt16,
		ast.BuiltinInt32,
		ast.BuiltinInt64,
		ast.BuiltinFloat32,
		ast.BuiltinFloat64,
		ast.BuiltinComplex64,
		ast.BuiltinComplex128,
		ast.BuiltinUint,
		ast.BuiltinInt,
		ast.BuiltinUintptr,
		ast.BuiltinString,
	}

	tk := []byte{
		_NIL,
		_BOOL,
		_UINT8,
		_UINT16,
		_UINT32,
		_UINT64,
		_INT8,
		_INT16,
		_INT32,
		_INT64,
		_FLOAT32,
		_FLOAT64,
		_COMPLEX64,
		_COMPLEX128,
		_UINT,
		_INT,
		_UINTPTR,
		_STRING,
	}

	exp := []byte{0}
	for i, typ := range ts {
		buf := keepEncoding(t, func(w *Writer) error { return w.writeType(nil, typ) })
		exp[0] = tk[i]
		expect_eq(t, "write builtin types", buf, exp)

		if dtyp := decodeType(t, buf); dtyp != typ {
			t.Error("read builtin type: incorrect decoding")
		}
	}
}

func TestWriteArrayType(t *testing.T) {
	typ := &ast.ArrayType{
		Dim: &ast.Literal{Kind: scanner.INTEGER, Value: []byte{'1', '0'}},
		Elt: &ast.BuiltinType{Kind: ast.BUILTIN_UINT8},
	}
	buf := keepEncoding(t, func(w *Writer) error { return w.writeType(nil, typ) })
	expect_eq(t, "write array type", buf, []byte{_ARRAY, 10, _UINT8})

	if dtyp := decodeType(t, buf); !reflect.DeepEqual(typ, dtyp) {
		t.Error("read array: types not equal")
	}
}

func TestWriteSliceType(t *testing.T) {
	typ := &ast.SliceType{
		Elt: &ast.BuiltinType{Kind: ast.BUILTIN_UINT8},
	}
	buf := keepEncoding(t, func(w *Writer) error { return w.writeType(nil, typ) })
	expect_eq(t, "write slice type", buf, []byte{_SLICE, _UINT8})

	var dtyp ast.Type
	keepDecoding(t, buf, func(r *Reader) error {
		var err error
		dtyp, err = r.readType(nil)
		return err
	})
	if !reflect.DeepEqual(typ, dtyp) {
		t.Error("read slice: types not equal")
	}
}

func TestWritePtrType(t *testing.T) {
	typ := &ast.PtrType{
		Base: &ast.BuiltinType{Kind: ast.BUILTIN_UINT8},
	}
	buf := keepEncoding(t, func(w *Writer) error { return w.writeType(nil, typ) })
	expect_eq(t, "write ptr type", buf, []byte{_PTR, _UINT8})

	var dtyp ast.Type
	keepDecoding(t, buf, func(r *Reader) error {
		var err error
		dtyp, err = r.readType(nil)
		return err
	})
	if !reflect.DeepEqual(typ, dtyp) {
		t.Error("read ptr: types not equal")
	}
}

func TestWriteMapType(t *testing.T) {
	typ := &ast.MapType{
		Key: &ast.BuiltinType{Kind: ast.BUILTIN_STRING},
		Elt: &ast.BuiltinType{Kind: ast.BUILTIN_UINT8},
	}
	buf := keepEncoding(t, func(w *Writer) error { return w.writeType(nil, typ) })
	expect_eq(t, "write map type", buf, []byte{_MAP, _STRING, _UINT8})

	var dtyp ast.Type
	keepDecoding(t, buf, func(r *Reader) error {
		var err error
		dtyp, err = r.readType(nil)
		return err
	})
	if !reflect.DeepEqual(typ, dtyp) {
		t.Error("read map: types not equal")
	}
}

func TestWriteChanType(t *testing.T) {
	typ := &ast.ChanType{
		Send: true,
		Recv: true,
		Elt:  &ast.BuiltinType{Kind: ast.BUILTIN_UINT8},
	}
	buf := keepEncoding(t, func(w *Writer) error { return w.writeType(nil, typ) })
	expect_eq(t, "write chan type", buf, []byte{_CHAN, 3, _UINT8})

	var dtyp ast.Type
	keepDecoding(t, buf, func(r *Reader) error {
		var err error
		dtyp, err = r.readType(nil)
		return err
	})
	if !reflect.DeepEqual(typ, dtyp) {
		t.Error("read chan: types not equal")
	}
}

func TestWriteStructType1(t *testing.T) {
	typ := &ast.StructType{}
	buf := keepEncoding(t, func(w *Writer) error { return w.writeType(nil, typ) })
	expect_eq(t, "write struct type", buf, []byte{_STRUCT, 0})
}

func TestWriteStructType2(t *testing.T) {
	typ := &ast.StructType{
		Fields: []ast.Field{
			{Name: "a", Type: &ast.BuiltinType{Kind: ast.BUILTIN_BOOL}},
			{Name: "b", Type: &ast.BuiltinType{Kind: ast.BUILTIN_BOOL}, Tag: "xy"},
		},
	}
	buf := keepEncoding(t, func(w *Writer) error { return w.writeType(nil, typ) })
	expect_eq(t, "write struct type",
		buf,
		[]byte{
			_STRUCT,
			2,
			1, 'a', _BOOL, 0,
			1, 'b', _BOOL, 2, 'x', 'y',
		},
	)

	var dtyp ast.Type
	keepDecoding(t, buf, func(r *Reader) error {
		var err error
		dtyp, err = r.readType(nil)
		return err
	})
	if !reflect.DeepEqual(typ, dtyp) {
		t.Error("read struct: types not equal")
	}
}

func TestWriteFuncType1(t *testing.T) {
	typ := &ast.FuncType{}
	buf := keepEncoding(t, func(w *Writer) error { return w.writeType(nil, typ) })
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
	buf := keepEncoding(t, func(w *Writer) error { return w.writeType(nil, typ) })
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

	var dtyp ast.Type
	keepDecoding(t, buf, func(r *Reader) error {
		var err error
		dtyp, err = r.readType(nil)
		return err
	})
	if !reflect.DeepEqual(typ, dtyp) {
		t.Error("read func: types not equal")
	}
}

func TestWriteIfaceType1(t *testing.T) {
	typ := &ast.InterfaceType{}
	buf := keepEncoding(t, func(w *Writer) error { return w.writeType(nil, typ) })
	expect_eq(t, "write iface type", buf, []byte{_IFACE, 0})
}

func TestWriteIfaceType2(t *testing.T) {
	typ := &ast.InterfaceType{
		Methods: []ast.MethodSpec{
			{Type: &ast.InterfaceType{}},
			{Name: "F", Type: &ast.FuncType{}},
		},
	}
	buf := keepEncoding(t, func(w *Writer) error { return w.writeType(nil, typ) })
	expect_eq(t, "write iface type",
		buf,
		[]byte{
			_IFACE,
			2,
			_IFACE, 0,
			_FUNC, 0, 0, 0, 1, 'F',
		},
	)

	var dtyp ast.Type
	keepDecoding(t, buf, func(r *Reader) error {
		var err error
		dtyp, err = r.readType(nil)
		return err
	})
	if !reflect.DeepEqual(typ, dtyp) {
		t.Error("read iface: types not equal")
	}
}

func TestWriteTypename(t *testing.T) {
	flt32 := ast.UniverseScope.Find("float32").(*ast.TypeDecl)
	pkgA := &ast.Package{
		Path:  "x/y/a",
		Name:  "a",
		Decls: make(map[string]ast.Symbol),
	}
	files := []*ast.File{
		{No: 1, Name: "a.go", Pkg: pkgA},
		{No: 2, Name: "b.go", Pkg: pkgA},
	}
	pkgA.Files = files
	typA := &ast.TypeDecl{
		Off:  13,
		Name: "A",
		File: pkgA.Files[0],
		Type: &ast.PtrType{Base: flt32},
	}
	pkgA.Decls["A"] = typA

	varX := &ast.Var{
		Off:  97,
		Name: "X",
		File: pkgA.Files[1],
		Type: typA,
	}
	pkgA.Decls["X"] = varX

	buf, err := encode(t, func(e *Writer) error {
		return e.writeDecl(pkgA, varX)
	})
	if err != nil {
		t.Fatal(err)
	}
	expect_eq(t, "write typename",
		buf,
		[]byte{
			_VAR_DECL,
			2, 97,
			1, 'X',
			_TYPENAME, 1, 1, 'A',
		},
	)
	buf[4] = 'Y'
	buf[8] = 'X'
	decode(t, buf, func(r *Reader) {
		_, err := r.readDecl(pkgA)
		if err == nil || err != BadFile {
			t.Error("expecting BadFile error")
		}
	})

	buf = keepEncoding(t, func(w *Writer) error { return w.writeDecl(pkgA, typA) })
	expect_eq(t, "write typename",
		buf,
		[]byte{
			_TYPE_DECL,
			1, 13,
			1, 'A',
			_PTR, _TYPENAME, 0, 7, 'f', 'l', 'o', 'a', 't', '3', '2',
		},
	)

	buf[7] = 2
	decode(t, buf, func(r *Reader) {
		_, err := r.readDecl(pkgA)
		if err == nil || err != BadFile {
			t.Error("expecting BadFile error")
		}
	})
	buf[7] = 0

	buf[4] = 'B'
	decode(t, buf, func(r *Reader) {
		_, err := r.readDecl(pkgA)
		if err != nil {
			t.Fatal(err)
		}
	})
	if sym, ok := pkgA.Decls["B"]; !ok {
		t.Error("declaration not found in package")
	} else if dcl, ok := sym.(*ast.TypeDecl); !ok {
		t.Error("wrong declaration type: expecting TypeDecl")
	} else if dcl.Name != "B" || dcl.Off != 13 || dcl.File != pkgA.Files[0] ||
		!reflect.DeepEqual(dcl.Type, typA.Type) {
		t.Error("declaration differ:", dcl.Name, dcl.Off, dcl.File)
	}

	buf = keepEncoding(t, func(w *Writer) error { return w.writeType(pkgA, typA) })
	expect_eq(t, "write typename",
		buf,
		[]byte{
			_TYPENAME,
			1,
			1, 'A',
		},
	)
	var dtyp ast.Type
	decode(t, buf, func(r *Reader) {
		typ, err := r.readType(pkgA)
		if err != nil {
			t.Fatal(err)
		}
		dtyp = typ
	})
	if !reflect.DeepEqual(dtyp, typA) {
		t.Error("decoded typenames differ")
	}

	pkgB := &ast.Package{
		Path:  "x/y/b",
		Name:  "b",
		Decls: make(map[string]ast.Symbol),
		Deps:  []*ast.Import{&ast.Import{Path: "a", Pkg: pkgA}},
	}
	files = []*ast.File{
		{No: 1, Name: "c.go", Pkg: pkgB},
		{No: 2, Name: "d.go", Pkg: pkgB},
	}
	pkgB.Files = files

	buf = keepEncoding(t, func(w *Writer) error { return w.writeType(pkgB, typA) })
	expect_eq(t, "write typename",
		buf,
		[]byte{
			_TYPENAME,
			2,
			1, 'A',
		},
	)
	keepDecoding(t, buf, func(r *Reader) error {
		var err error
		dtyp, err = r.readType(pkgB)
		return err
	})
	if tn, ok := dtyp.(*ast.TypeDecl); !ok {
		t.Error("wrong decode: expecting type declaration")
	} else if !reflect.DeepEqual(tn, typA) {
		t.Error("decoded typenames differ")
	}
}

func TestWriteTypeDecl(t *testing.T) {
	pkg := &ast.Package{Decls: make(map[string]ast.Symbol)}
	d := &ast.TypeDecl{
		Name: "S",
		Type: &ast.PtrType{Base: &ast.BuiltinType{Kind: ast.BUILTIN_FLOAT32}},
	}
	buf := keepEncoding(t, func(e *Writer) error { return e.writeDecl(nil, d) })
	expect_eq(t, "type decl",
		buf,
		[]byte{
			_TYPE_DECL,
			0, 0, // file, off
			1, 'S', // name
			_PTR, _FLOAT32,
		},
	)

	buf[0] = _TYPE_DECL + 1
	decode(t, buf, func(r *Reader) {
		_, err := r.readDecl(pkg)
		if err == nil || err != BadFile {
			t.Error("expecting BadFile error")
		}
	})
	buf[0] = _TYPE_DECL

	keepDecoding(t, buf, func(r *Reader) error {
		_, err := r.readDecl(pkg)
		return err
	})
	if sym, ok := pkg.Decls["S"]; !ok {
		t.Error("declaration not found in package")
	} else if dd, ok := sym.(*ast.TypeDecl); !ok {
		t.Error("incorrect declaration type; expected TypeDecl")
	} else if !reflect.DeepEqual(d, dd) {
		t.Error("declarations not equal")
	}
}

func TestWriteTypeDecl1(t *testing.T) {
	pkg := &ast.Package{
		Name:  "pkg",
		Decls: make(map[string]ast.Symbol),
	}
	file := &ast.File{
		Pkg:   pkg,
		Name:  "pkg.go",
		Decls: make(map[string]ast.Symbol),
	}
	pkg.Files = append(pkg.Files, file)

	Z := &ast.TypeDecl{
		File: file,
		Name: "Z",
		Type: ast.BuiltinInt,
	}

	F := &ast.FuncDecl{
		File: file,
		Name: "F",
		Func: ast.Func{
			Recv: &ast.Param{Type: Z},
			Sig:  &ast.FuncType{},
			Up:   file,
		},
	}
	F.Func.Decl = F
	pkg.Declare("Z", Z)
	pkg.Declare("F", F)

	buf := bytes.Buffer{}
	if err := Write(&buf, pkg); err != nil {
		t.Fatal(err)
	}
	_, err := Read(bytes.NewReader(buf.Bytes()), nil)
	if err != nil {
		t.Error(err)
	}
}

func TestWriteTypeDecl2(t *testing.T) {
	pkg := &ast.Package{
		Name:  "pkg",
		Decls: make(map[string]ast.Symbol),
	}
	file := &ast.File{
		Pkg:   pkg,
		Name:  "pkg.go",
		Decls: make(map[string]ast.Symbol),
	}
	pkg.Files = append(pkg.Files, file)

	Z := &ast.TypeDecl{
		File: file,
		Name: "Z",
		Type: ast.BuiltinInt,
	}

	S := &ast.TypeDecl{
		File: file,
		Name: "S",
		Type: &ast.StructType{
			Fields: []ast.Field{
				{Name: "x", Type: Z},
			},
		},
	}
	pkg.Declare(Z.Name, Z)
	pkg.Declare(S.Name, S)

	buf := bytes.Buffer{}
	if err := Write(&buf, pkg); err != nil {
		t.Fatal(err)
	}
	_, err := Read(bytes.NewReader(buf.Bytes()), nil)
	if err != nil {
		t.Error(err)
	}
}

func TestWriteTypeDecl3(t *testing.T) {
	pkg := &ast.Package{
		Name:  "pkg",
		Decls: make(map[string]ast.Symbol),
	}
	file := &ast.File{
		Pkg:   pkg,
		Name:  "pkg.go",
		Decls: make(map[string]ast.Symbol),
	}
	pkg.Files = append(pkg.Files, file)

	str := &ast.StructType{
		Fields: []ast.Field{
			{Name: "x"},
		},
	}

	S := &ast.TypeDecl{File: file, Name: "S", Type: str}
	str.Fields[0].Type = S
	pkg.Declare(S.Name, S)

	buf := bytes.Buffer{}
	if err := Write(&buf, pkg); err != nil {
		t.Fatal(err)
	}
	_, err := Read(bytes.NewReader(buf.Bytes()), nil)
	if err != nil {
		t.Error(err)
	}
}

func TestWriteVarDecl(t *testing.T) {
	v := &ast.Var{}
	buf := keepEncoding(t, func(e *Writer) error { return e.writeDecl(nil, v) })
	expect_eq(t, "var decl",
		buf,
		[]byte{
			_VAR_DECL,
			0, 0, // file, off
			0, // name
			0, // type
		},
	)

	pkg := &ast.Package{
		Files: []*ast.File{&ast.File{No: 1}},
		Decls: make(map[string]ast.Symbol),
	}
	decode(t, buf, func(r *Reader) {
		ok, err := r.readDecl(pkg)
		if ok || err == nil || err != BadFile {
			t.Error("unnamed decl: expecing BadFile")
		}
	})

	v.Name = "xyz"
	v.File = &ast.File{No: 2}
	buf, err := encode(t, func(e *Writer) error {
		return e.writeDecl(pkg, v)
	})
	if err != nil {
		t.Fatal(err)
	}
	decode(t, buf, func(r *Reader) {
		ok, err := r.readDecl(pkg)
		if ok || err == nil || err != BadFile {
			t.Error("bogus file index: expecting BadFile")
		}
	})

	v.Off = 12
	v.File = pkg.Files[0]
	v.Type = &ast.BuiltinType{Kind: ast.BUILTIN_INT32}
	buf = keepEncoding(t, func(e *Writer) error { return e.writeDecl(nil, v) })
	expect_eq(t, "var decl",
		buf,
		[]byte{
			_VAR_DECL,
			1, 12, // file, off
			3, 'x', 'y', 'z', // name
			_INT32,
		},
	)
	keepDecoding(t, buf, func(r *Reader) error {
		_, err := r.readDecl(pkg)
		return err
	})
	if sym, ok := pkg.Decls["xyz"]; !ok {
		t.Error("declaration not found in package")
	} else if dv, ok := sym.(*ast.Var); !ok {
		t.Error("incorrect declaration type; expected Var")
	} else if !reflect.DeepEqual(v, dv) {
		t.Error("declarations not equal")
	}
}

func TestWriteConstDecl(t *testing.T) {
	c := &ast.Const{}
	buf := keepEncoding(t, func(e *Writer) error { return e.writeDecl(nil, c) })
	expect_eq(t, "const decl",
		buf,
		[]byte{
			_CONST_DECL,
			0, 0, // file, off
			0, // name
			0, // type
		},
	)

	c.Off = 12
	c.File = &ast.File{No: 42}
	c.Name = "xyz"
	c.Type = &ast.BuiltinType{Kind: ast.BUILTIN_INT32}
	buf = keepEncoding(t, func(e *Writer) error { return e.writeDecl(nil, c) })
	expect_eq(t, "const decl",
		buf,
		[]byte{
			_CONST_DECL,
			42, 12, // file, off
			3, 'x', 'y', 'z', // name
			_INT32,
		},
	)
	c.File = nil
	buf = keepEncoding(t, func(e *Writer) error { return e.writeDecl(nil, c) })
	expect_eq(t, "const decl",
		buf,
		[]byte{
			_CONST_DECL,
			0, 12, // file, off
			3, 'x', 'y', 'z', // name
			_INT32,
		},
	)
	pkg := &ast.Package{Decls: make(map[string]ast.Symbol)}
	keepDecoding(t, buf, func(r *Reader) error {
		_, err := r.readDecl(pkg)
		return err
	})
	if sym, ok := pkg.Decls["xyz"]; !ok {
		t.Error("declaration not found in package")
	} else if dc, ok := sym.(*ast.Const); !ok {
		t.Error("incorrect declaration type; expected Const")
	} else if !reflect.DeepEqual(c, dc) {
		t.Error("declarations not equal")
	}
}

func TestWriteFuncDecl(t *testing.T) {
	pkg := &ast.Package{
		Path:  "x/y/a",
		Name:  "a",
		Decls: make(map[string]ast.Symbol),
	}
	files := []*ast.File{
		{No: 1, Name: "a.go", Pkg: pkg},
		{No: 2, Name: "b.go", Pkg: pkg},
	}
	pkg.Files = files
	typA := &ast.TypeDecl{
		Name: "A",
		File: pkg.Files[0],
		Type: &ast.PtrType{Base: ast.BuiltinFloat32},
	}
	pkg.Decls["A"] = typA

	fn := &ast.FuncDecl{
		Off:  42,
		Name: "Fn",
		File: pkg.Files[1],
		Func: ast.Func{
			Sig: &ast.FuncType{},
		},
	}
	buf := keepEncoding(t, func(e *Writer) error { return e.writeDecl(pkg, fn) })
	expect_eq(t, "func decl",
		buf,
		[]byte{
			_FUNC_DECL,
			2, 42, // file, off
			2, 'F', 'n', // name
			_NIL, // receiver type
			_FUNC, 0, 0, 0,
		},
	)

	buf[5] = 'F'
	keepDecoding(t, buf, func(r *Reader) error {
		_, err := r.readDecl(pkg)
		return err
	})
	if sym, ok := pkg.Decls["FF"]; !ok {
		t.Error("declaration not found in package")
	} else if dfn, ok := sym.(*ast.FuncDecl); !ok {
		t.Error("incorrect declaration type; expected FuncDecl")
	} else {
		if dfn.Off != fn.Off || dfn.Name != "FF" || dfn.File != pkg.Files[1] ||
			dfn.Func.Recv != nil || !reflect.DeepEqual(dfn.Func.Sig, fn.Func.Sig) {
			t.Error("declarations not equal")
		}
	}

	typS := &ast.TypeDecl{
		Name: "S",
		File: pkg.Files[0],
		Type: &ast.PtrType{Base: ast.BuiltinFloat32},
	}
	pkg.Decls["S"] = typS

	fn.Func.Recv = &ast.Param{Type: &ast.PtrType{Base: typS}}
	buf = keepEncoding(t, func(e *Writer) error { return e.writeDecl(pkg, fn) })
	expect_eq(t, "func decl",
		buf,
		[]byte{
			_FUNC_DECL,
			2, 42, // file, off
			2, 'F', 'n', // name
			_PTR, _TYPENAME, 1, 1, 'S', // receiver type
			_FUNC, 0, 0, 0, // func type
		},
	)
	copy(buf[11:], []byte{_TYPENAME, 1, 1, 'S'})
	decode(t, buf, func(r *Reader) {
		_, err := r.readDecl(pkg)
		if err == nil || err != BadFile {
			t.Error("expecting BadFile error")
		}
	})

	buf[5] = 'F'
	copy(buf[11:], []byte{_FUNC, 0, 0, 0})
	keepDecoding(t, buf, func(r *Reader) error {
		var err error
		_, err = r.readDecl(pkg)
		return err
	})
	if sym, ok := pkg.Decls["FF"]; !ok {
		t.Error("declaration not found in package")
	} else if dfn, ok := sym.(*ast.FuncDecl); !ok {
		t.Error("incorrect declaration type; expected FuncDecl")
	} else {
		if dfn.Off != fn.Off || dfn.Name != "FF" || dfn.File != pkg.Files[1] ||
			!reflect.DeepEqual(dfn.Func.Recv, fn.Func.Recv) ||
			!reflect.DeepEqual(dfn.Func.Sig, fn.Func.Sig) {
			t.Error("declarations not equal")
		}
	}
}

func TestWriteFile1(t *testing.T) {
	pkg := &ast.Package{}
	file := &ast.File{
		Pkg: pkg,
		Name: "1234567890123456789012345678901234567890123456789012345678901234567890" +
			"123456789012345678901234567890123456789012345678901234567890.go",
	}

	_ = keepEncoding(t, func(e *Writer) error { return e.writeFile(file) })

	file.Name = "xx.go"
	buf := keepEncoding(t, func(e *Writer) error { return e.writeFile(file) })
	expect_eq(t, "empty file",
		buf,
		[]byte{
			5, 'x', 'x', '.', 'g', 'o',
			0, // no source map
		})

	var dfile *ast.File
	keepDecoding(t, buf, func(r *Reader) error {
		var err error
		dfile, err = r.readFile()
		return err
	})
	if file.Name != dfile.Name || dfile.SrcMap.LineCount() != 0 {
		t.Error("read file: files not equal")
	}
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

	buf := keepEncoding(t, func(e *Writer) error { return e.writeFile(file) })
	expect_eq(t, "empty file",
		buf,
		[]byte{
			5, 'x', 'x', '.', 'g', 'o',
			1, 2, 3, 0, // source map
		})
	var dfile *ast.File
	keepDecoding(t, buf, func(r *Reader) error {
		var err error
		dfile, err = r.readFile()
		return err
	})
	if file.Name != dfile.Name || dfile.SrcMap.LineCount() != 3 {
		t.Error("read file: files not equal")
	}
	if _, k := dfile.SrcMap.LineExtent(0); k != 1 {
		t.Error("read file: line 1, wrong length")
	}
	if _, k := dfile.SrcMap.LineExtent(1); k != 2 {
		t.Error("read file: line 1, wrong length")
	}
	if _, k := dfile.SrcMap.LineExtent(2); k != 3 {
		t.Error("read file: line 1, wrong length")
	}
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

	buf := keepEncoding(t, func(e *Writer) error { return e.writePkg(pkg) })
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

type MockPackageLocator struct {
	pkgs map[string]*ast.Package
}

func (loc *MockPackageLocator) FindPackage(path string) (*ast.Package, error) {
	return loc.pkgs[path], nil
}

func TestWritePackage3(t *testing.T) {
	loc := &MockPackageLocator{pkgs: make(map[string]*ast.Package)}
	pkg := &ast.Package{
		Name: "test",
		Files: []*ast.File{
			&ast.File{Name: "xx.go"},
			&ast.File{Name: "yy.go"},
		},
		Decls: make(map[string]ast.Symbol),
	}
	dep1 := &ast.Package{Name: "dep1", Sig: [20]byte{5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5}}
	depX := &ast.Package{Name: "depX", Sig: [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1}}

	pkg.Deps = []*ast.Import{{Path: "dep1", Pkg: dep1}, {Path: "depX", Pkg: depX}}
	for _, d := range pkg.Deps {
		loc.pkgs[d.Pkg.Name] = d.Pkg
	}

	d0 := &ast.Var{
		Off:  12,
		File: pkg.Files[0],
		Name: "xyz",
		Type: &ast.BuiltinType{Kind: ast.BUILTIN_INT32},
	}
	pkg.Decls[d0.Name] = d0

	d1 := &ast.Var{
		Off:  12,
		File: pkg.Files[1],
		Name: "Xyz",
		Type: &ast.BuiltinType{Kind: ast.BUILTIN_INT32},
	}
	pkg.Decls[d1.Name] = d1

	buf := keepEncoding(t, func(e *Writer) error { return e.writePkg(pkg) })
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
			2, // files
			5, 'x', 'x', '.', 'g', 'o',
			0,
			5, 'y', 'y', '.', 'g', 'o',
			0,
			_VAR_DECL,
			2, 12, // file, off
			3, 'X', 'y', 'z', // name
			_INT32,
			_END, // end decls
		},
	)

	keepDecoding(t, buf, func(r *Reader) error {
		_, err := r.readPkg(loc)
		return err
	})

	dpkg, err := Read(bytes.NewReader(buf), loc)
	if err != nil {
		t.Error(err)
	}
	if dpkg.Name != pkg.Name {
		t.Error("package names different")
	}
	if len(dpkg.Deps) != 2 {
		t.Fatal("expected exactly two dependencies")
	}
	if len(dpkg.Deps) != len(pkg.Deps) {
		t.Error("incorrect number of depdendencies")
	}
	if dpkg.Deps[0].Path != "dep1" || dpkg.Deps[1].Path != "depX" {
		t.Error("unexpected dependency name")
	}
	if !reflect.DeepEqual(dpkg.Deps[0].Sig, dep1.Sig) ||
		!reflect.DeepEqual(dpkg.Deps[1].Sig, depX.Sig) {
		t.Error("incorrect signature read")
	}
	if len(dpkg.Decls) != 1 {
		t.Error("expecting only one exported decl")
	}
	if dpkg.Decls["Xyz"] == nil {
		t.Error("unexpected declaration name")
	}
	if dpkg.Files[0].Pkg != dpkg || dpkg.Files[1].Pkg != dpkg {
		t.Error("files do not point oot their package")
	}
}
