package pdb

import (
	"golang/ast"
	"golang/scanner"
	"io"
)

type errorBadVersion struct{}

func (errorBadVersion) Error() string { return "version mismatch" }

type errorBadFile struct{}

func (errorBadFile) Error() string { return "file format error" }

var (
	BadVersion errorBadVersion
	BadFile    errorBadFile
)

func Read(r io.Reader, loc ast.PackageLocator) (*ast.Package, error) {
	dec := Decoder{}
	dec.Init(r)
	return readPkg(&dec, loc)
}

func readPkg(dec *Decoder, loc ast.PackageLocator) (*ast.Package, error) {
	// Version
	n, err := dec.ReadNum()
	if err != nil {
		return nil, err
	}
	if n != VERSION {
		return nil, BadVersion
	}

	pkg := &ast.Package{
		Deps:  make(map[string]*ast.Import),
		Decls: make(map[string]ast.Symbol),
	}

	// Package name
	pkg.Name, err = dec.ReadString()
	if err != nil {
		return nil, err
	}

	// Dependencies
	n, err = dec.ReadNum()
	if err != nil {
		return nil, err
	}
	for i := 0; i < int(n); i++ {
		imp := &ast.Import{No: i + 2}
		path, err := dec.ReadString()
		if err != nil {
			return nil, err
		}
		imp.Pkg, err = loc.FindPackage(path)
		if err != nil {
			return nil, err
		}
		err = dec.ReadBytes(imp.Sig[:])
		if err != nil {
			return nil, err
		}
		pkg.Deps[path] = imp
	}

	// Source files
	n, err = dec.ReadNum()
	if err != nil {
		return nil, err
	}
	pkg.Files = make([]*ast.File, n)
	for i := range pkg.Files {
		pkg.Files[i], err = readFile(dec)
		if err != nil {
			return nil, err
		}
		pkg.Files[i].No = i + 1
	}

	// Declarations.
	ok, err := readDecl(dec, pkg)
	for err == nil && ok {
		ok, err = readDecl(dec, pkg)
	}
	if err != nil {
		return nil, err
	}

	// FIXME: read signature
	return pkg, nil
}

func readFile(dec *Decoder) (*ast.File, error) {
	name, err := dec.ReadString()
	if err != nil {
		return nil, err
	}
	file := &ast.File{Name: name, Decls: make(map[string]ast.Symbol)}
	for {
		n, err := dec.ReadNum()
		if err != nil {
			return nil, err
		}
		if n == 0 {
			return file, nil
		}
		file.SrcMap.AddLine(int(n))
	}

}

func readDecl(dec *Decoder, pkg *ast.Package) (bool, error) {
	kind, err := dec.ReadByte()
	if err != nil {
		return false, err
	}
	if kind == _END {
		return false, nil
	}
	n, err := dec.ReadNum()
	if err != nil {
		return false, err
	}
	fno := int(n)
	if fno > len(pkg.Files) {
		return false, BadFile
	}
	var file *ast.File
	if fno > 0 {
		file = pkg.Files[fno-1]
	}
	n, err = dec.ReadNum()
	if err != nil {
		return false, err
	}
	off := int(n)
	name, err := dec.ReadString()
	if err != nil {
		return false, err
	}
	if len(name) == 0 {
		return false, BadFile
	}
	typ, err := readType(dec, pkg)
	if err != nil {
		return false, err
	}
	switch kind {
	case _VAR_DECL:
		v := &ast.Var{Off: off, File: file, Name: name, Type: typ}
		pkg.Decls[name] = v
	case _CONST_DECL:
		c := &ast.Const{Off: off, File: file, Name: name, Type: typ}
		pkg.Decls[name] = c
	case _FUNC_DECL:
		var rcv *ast.Param
		if typ != ast.BuiltinNil {
			rcv = &ast.Param{Type: typ}
		}
		typ, err = readType(dec, pkg)
		if err != nil {
			return false, err
		}
		ft, ok := typ.(*ast.FuncType)
		if !ok {
			return false, BadFile
		}
		f := &ast.FuncDecl{
			Off:  off,
			File: file,
			Name: name,
			Func: ast.Func{Recv: rcv, Sig: ft},
		}
		f.Func.Decl = f
		pkg.Decls[name] = f
	case _TYPE_DECL:
		t := &ast.TypeDecl{Off: off, File: file, Name: name, Type: typ}
		pkg.Decls[name] = t
	default:
		return false, BadFile
	}
	return true, nil
}

func readType(dec *Decoder, pkg *ast.Package) (ast.Type, error) {
	tag, err := dec.ReadByte()
	if err != nil {
		return nil, err
	}
	switch tag {
	case _NIL:
		return ast.BuiltinNil, nil
	case _BOOL:
		return ast.BuiltinBool, nil
	case _UINT8:
		return ast.BuiltinUint8, nil
	case _UINT16:
		return ast.BuiltinUint16, nil
	case _UINT32:
		return ast.BuiltinUint32, nil
	case _UINT64:
		return ast.BuiltinUint64, nil
	case _INT8:
		return ast.BuiltinInt8, nil
	case _INT16:
		return ast.BuiltinInt16, nil
	case _INT32:
		return ast.BuiltinInt32, nil
	case _INT64:
		return ast.BuiltinInt64, nil
	case _FLOAT32:
		return ast.BuiltinFloat32, nil
	case _FLOAT64:
		return ast.BuiltinFloat64, nil
	case _COMPLEX64:
		return ast.BuiltinComplex64, nil
	case _COMPLEX128:
		return ast.BuiltinComplex128, nil
	case _UINT:
		return ast.BuiltinUint, nil
	case _INT:
		return ast.BuiltinInt, nil
	case _UINTPTR:
		return ast.BuiltinUintptr, nil
	case _STRING:
		return ast.BuiltinString, nil
	case _ARRAY:
		// FIXME: real dimension
		if _, err := dec.ReadNum(); err != nil {
			return nil, err
		}
		t, err := readType(dec, pkg)
		if err != nil {
			return nil, err
		}
		return &ast.ArrayType{
			Dim: &ast.Literal{Kind: scanner.INTEGER, Value: []byte("10")},
			Elt: t,
		}, nil
	case _SLICE:
		t, err := readType(dec, pkg)
		if err != nil {
			return nil, err
		}
		return &ast.SliceType{Elt: t}, nil
	case _STRUCT:
		return readStructType(dec, pkg)
	case _PTR:
		t, err := readType(dec, pkg)
		if err != nil {
			return nil, err
		}
		return &ast.PtrType{Base: t}, nil
	case _FUNC:
		return readFuncType(dec, pkg)
	case _IFACE:
		return readIfaceType(dec, pkg)
	case _MAP:
		k, err := readType(dec, pkg)
		if err != nil {
			return nil, err
		}
		e, err := readType(dec, pkg)
		if err != nil {
			return nil, err
		}
		return &ast.MapType{Key: k, Elt: e}, nil
	case _CHAN:
		b, err := dec.ReadByte()
		if err != nil {
			return nil, err
		}
		t, err := readType(dec, pkg)
		if err != nil {
			return nil, err
		}
		return &ast.ChanType{
			Send: (b & _CHAN_CAN_SEND) != 0,
			Recv: (b & _CHAN_CAN_RECV) != 0,
			Elt:  t,
		}, nil
	case _TYPENAME:
		return readTypename(dec, pkg)
	default:
		panic("not reached")
	}
}

func findScopeByNo(pkg *ast.Package, no int) ast.Scope {
	switch no {
	case 0:
		return ast.UniverseScope
	case 1:
		return pkg
	default:
		for _, i := range pkg.Deps {
			if i.No == no {
				return i.Pkg
			}
		}
	}
	panic("not reached")
}

func readTypename(dec *Decoder, pkg *ast.Package) (*ast.TypeDecl, error) {
	n, err := dec.ReadNum()
	if err != nil {
		return nil, err
	}
	if n >= uint64(len(pkg.Deps)+2) {
		return nil, BadFile
	}
	name, err := dec.ReadString()
	if err != nil {
		return nil, err
	}
	s := findScopeByNo(pkg, int(n))
	if sym := s.Find(name); sym == nil {
		return nil, BadFile
	} else if dcl, ok := sym.(*ast.TypeDecl); !ok {
		return nil, BadFile
	} else {
		return dcl, nil
	}
}

func readStructType(dec *Decoder, pkg *ast.Package) (*ast.StructType, error) {
	n, err := dec.ReadNum()
	if err != nil {
		return nil, err
	}
	fs := make([]ast.Field, n)
	for i := range fs {
		f := &fs[i]
		var err error
		if f.Name, err = dec.ReadString(); err != nil {
			return nil, err
		}
		if f.Type, err = readType(dec, pkg); err != nil {
			return nil, err
		}
		if f.Tag, err = dec.ReadString(); err != nil {
			return nil, err
		}
	}
	return &ast.StructType{Fields: fs}, nil
}

func readParams(dec *Decoder, pkg *ast.Package) ([]ast.Param, error) {
	n, err := dec.ReadNum()
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, nil
	}
	ps := make([]ast.Param, n)
	for i := range ps {
		p := &ps[i]
		p.Name, err = dec.ReadString()
		if err != nil {
			return nil, err
		}
		p.Type, err = readType(dec, pkg)
		if err != nil {
			return nil, err
		}
	}
	return ps, nil
}

func readFuncType(dec *Decoder, pkg *ast.Package) (*ast.FuncType, error) {
	ps, err := readParams(dec, pkg)
	if err != nil {
		return nil, err
	}
	rs, err := readParams(dec, pkg)
	if err != nil {
		return nil, err
	}
	b, err := dec.ReadByte()
	if err != nil {
		return nil, err
	}
	return &ast.FuncType{Params: ps, Returns: rs, Var: b == 1}, nil
}

func readIfaceType(dec *Decoder, pkg *ast.Package) (*ast.InterfaceType, error) {
	n, err := dec.ReadNum()
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return &ast.InterfaceType{}, nil
	}
	ms := make([]ast.MethodSpec, n)
	for i := range ms {
		var err error
		m := &ms[i]
		m.Type, err = readType(dec, pkg)
		if err != nil {
			return nil, err
		}
		if _, ok := m.Type.(*ast.FuncType); ok {
			m.Name, err = dec.ReadString()
			if err != nil {
				return nil, err
			}
		}
	}
	return &ast.InterfaceType{Methods: ms}, nil
}
