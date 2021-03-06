package pdb

import (
	"golang/ast"
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

type Reader struct {
	Decoder
	fwd int
}

func Read(r io.Reader, loc ast.PackageLocator) (*ast.Package, error) {
	rd := Reader{}
	rd.Decoder.Init(r)
	return rd.readPkg(loc)
}

func (r *Reader) readPkg(loc ast.PackageLocator) (*ast.Package, error) {
	// Version
	n, err := r.ReadNum()
	if err != nil {
		return nil, err
	}
	if n != VERSION {
		return nil, BadVersion
	}

	pkg := &ast.Package{
		Syms: make(map[string]ast.Symbol),
	}

	// Package name
	pkg.Name, err = r.ReadString()
	if err != nil {
		return nil, err
	}

	// Dependencies
	n, err = r.ReadNum()
	if err != nil {
		return nil, err
	}
	for i := 0; i < int(n); i++ {
		path, err := r.ReadString()
		if err != nil {
			return nil, err
		}
		p, err := loc.FindPackage(path)
		if err != nil {
			return nil, err
		}
		imp := &ast.Import{Path: path, Pkg: p}
		err = r.ReadBytes(imp.Sig[:])
		if err != nil {
			return nil, err
		}
		pkg.Deps = append(pkg.Deps, imp)
	}

	// Source files
	n, err = r.ReadNum()
	if err != nil {
		return nil, err
	}
	pkg.Files = make([]*ast.File, n)
	for i := range pkg.Files {
		pkg.Files[i], err = r.readFile()
		if err != nil {
			return nil, err
		}
		pkg.Files[i].No = i + 1
		pkg.Files[i].Pkg = pkg
	}

	// Declarations.
	ok, err := r.readDecl(pkg)
	for err == nil && ok {
		ok, err = r.readDecl(pkg)
	}
	if err != nil {
		return nil, err
	}

	// Check that each "forward" type name has a corresponding type
	// declaration.
	if r.fwd != 0 {
		return nil, BadFile
	}

	// FIXME: read signature
	return pkg, nil
}

func (r *Reader) readFile() (*ast.File, error) {
	name, err := r.ReadString()
	if err != nil {
		return nil, err
	}
	file := &ast.File{Name: name, Syms: make(map[string]ast.Symbol)}
	for {
		n, err := r.ReadNum()
		if err != nil {
			return nil, err
		}
		if n == 0 {
			return file, nil
		}
		file.SrcMap.AddLine(int(n))
	}

}

func receiverBaseType(typ ast.Type) (*ast.TypeDecl, error) {
	if ptr, ok := typ.(*ast.PtrType); ok {
		typ = ptr.Base
	}
	if dcl, ok := typ.(*ast.TypeDecl); ok {
		return dcl, nil
	}
	return nil, BadFile
}

func (r *Reader) readDecl(pkg *ast.Package) (bool, error) {
	kind, err := r.ReadByte()
	if err != nil {
		return false, err
	}
	if kind == _END {
		return false, nil
	}
	if kind == _FUNC_DECL {
		f, err := r.readFuncDecl(pkg)
		if err != nil {
			return false, err
		}
		pkg.Syms[f.Name] = f
		return true, nil
	}
	n, err := r.ReadNum()
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
	n, err = r.ReadNum()
	if err != nil {
		return false, err
	}
	off := int(n)
	name, err := r.ReadString()
	if err != nil {
		return false, err
	}
	if len(name) == 0 {
		return false, BadFile
	}
	typ, err := r.readType(pkg)
	if err != nil {
		return false, err
	}
	switch kind {
	case _VAR_DECL:
		v := &ast.Var{Off: off, File: file, Name: name, Type: typ}
		pkg.Syms[name] = v
	case _CONST_DECL:
		c := &ast.Const{Off: off, File: file, Name: name, Type: typ}
		pkg.Syms[name] = c
	case _TYPE_DECL:
		var t *ast.TypeDecl
		// Check if the type declaration was created by an earlier encounter
		// with its type name.
		if sym, ok := pkg.Syms[name]; !ok {
			t = &ast.TypeDecl{}
			pkg.Syms[name] = t
		} else if td, ok := sym.(*ast.TypeDecl); !ok {
			return false, BadFile
		} else {
			r.fwd--
			t = td
		}
		*t = ast.TypeDecl{Off: off, File: file, Name: name, Type: typ}
	default:
		return false, BadFile
	}
	return true, nil
}

func (r *Reader) readFuncDecl(pkg *ast.Package) (*ast.FuncDecl, error) {
	n, err := r.ReadNum()
	if err != nil {
		return nil, err
	}
	fno := int(n)
	if fno > len(pkg.Files) {
		return nil, BadFile
	}
	var file *ast.File
	if fno > 0 {
		file = pkg.Files[fno-1]
	}
	n, err = r.ReadNum()
	if err != nil {
		return nil, err
	}
	off := int(n)
	name, err := r.ReadString()
	if err != nil {
		return nil, err
	}
	if len(name) == 0 {
		return nil, BadFile
	}
	typ, err := r.readType(pkg)
	if err != nil {
		return nil, err
	}
	var rcv *ast.Param
	if typ != nil {
		rcv = &ast.Param{Type: typ}
	}
	typ, err = r.readType(pkg)
	if err != nil {
		return nil, err
	}
	ft, ok := typ.(*ast.FuncType)
	if !ok {
		return nil, BadFile
	}
	f := &ast.FuncDecl{
		Off:  off,
		File: file,
		Name: name,
		Func: ast.Func{Recv: rcv, Sig: ft},
	}
	f.Func.Decl = f
	if rcv != nil {
		base, err := receiverBaseType(rcv.Type)
		if err != nil {
			return nil, err
		}
		if base == rcv.Type {
			base.Methods = append(base.Methods, f)
		} else {
			base.PMethods = append(base.PMethods, f)
		}
	}
	return f, nil
}

func (r *Reader) readType(pkg *ast.Package) (ast.Type, error) {
	tag, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	switch tag {
	case _VOID:
		return nil, nil
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
	case _UNTYPED_BOOL:
		return ast.BuiltinUntypedBool, nil
	case _UNTYPED_RUNE:
		return ast.BuiltinUntypedRune, nil
	case _UNTYPED_INT:
		return ast.BuiltinUntypedInt, nil
	case _UNTYPED_FLOAT:
		return ast.BuiltinUntypedFloat, nil
	case _UNTYPED_COMPLEX:
		return ast.BuiltinUntypedComplex, nil
	case _UNTYPED_STRING:
		return ast.BuiltinUntypedString, nil
	case _ARRAY:
		// FIXME: real dimension
		if _, err := r.ReadNum(); err != nil {
			return nil, err
		}
		t, err := r.readType(pkg)
		if err != nil {
			return nil, err
		}
		return &ast.ArrayType{
			Dim: &ast.ConstValue{Typ: ast.BuiltinInt, Value: ast.Int(10)},
			Elt: t,
		}, nil
	case _SLICE:
		t, err := r.readType(pkg)
		if err != nil {
			return nil, err
		}
		return &ast.SliceType{Elt: t}, nil
	case _STRUCT:
		return r.readStructType(pkg)
	case _PTR:
		t, err := r.readType(pkg)
		if err != nil {
			return nil, err
		}
		return &ast.PtrType{Base: t}, nil
	case _FUNC:
		return r.readFuncType(pkg)
	case _IFACE:
		return r.readIfaceType(pkg)
	case _MAP:
		k, err := r.readType(pkg)
		if err != nil {
			return nil, err
		}
		e, err := r.readType(pkg)
		if err != nil {
			return nil, err
		}
		return &ast.MapType{Key: k, Elt: e}, nil
	case _CHAN:
		b, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		t, err := r.readType(pkg)
		if err != nil {
			return nil, err
		}
		return &ast.ChanType{
			Send: (b & _CHAN_CAN_SEND) != 0,
			Recv: (b & _CHAN_CAN_RECV) != 0,
			Elt:  t,
		}, nil
	case _TYPENAME:
		return r.readTypename(pkg)
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
		return pkg.Deps[no-2].Pkg
	}
}

func (r *Reader) readTypename(pkg *ast.Package) (*ast.TypeDecl, error) {
	n, err := r.ReadNum()
	if err != nil {
		return nil, err
	}
	if n >= uint64(len(pkg.Deps)+2) {
		return nil, BadFile
	}
	name, err := r.ReadString()
	if err != nil {
		return nil, err
	}
	s := findScopeByNo(pkg, int(n))
	if sym := s.Find(name); sym == nil {
		if int(n) != 1 {
			return nil, BadFile
		}
		// In the current package, if the type declaration is not found,
		// create it. This allows a type name to refer to a type declaration
		// that is not yet read.
		dcl := &ast.TypeDecl{Name: name}
		pkg.Syms[name] = dcl
		r.fwd++
		return dcl, nil
	} else if dcl, ok := sym.(*ast.TypeDecl); !ok {
		return nil, BadFile
	} else {
		return dcl, nil
	}
}

func (r *Reader) readStructType(pkg *ast.Package) (*ast.StructType, error) {
	fno, err := r.ReadNum()
	if err != nil {
		return nil, err
	}
	if int(fno) > len(pkg.Files) {
		return nil, BadFile
	}
	var file *ast.File
	if fno > 0 {
		file = pkg.Files[fno-1]
	}
	n, err := r.ReadNum()
	if err != nil {
		return nil, err
	}
	fs := make([]ast.Field, n)
	for i := range fs {
		f := &fs[i]
		off, err := r.ReadNum()
		if err != nil {
			return nil, err
		}
		f.Off = int(off)
		if f.Name, err = r.ReadString(); err != nil {
			return nil, err
		}
		if f.Type, err = r.readType(pkg); err != nil {
			return nil, err
		}
		if f.Tag, err = r.ReadString(); err != nil {
			return nil, err
		}
	}
	return &ast.StructType{File: file, Fields: fs}, nil
}

func (r *Reader) readParams(pkg *ast.Package) ([]ast.Param, error) {
	n, err := r.ReadNum()
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, nil
	}
	ps := make([]ast.Param, n)
	for i := range ps {
		p := &ps[i]
		p.Name, err = r.ReadString()
		if err != nil {
			return nil, err
		}
		p.Type, err = r.readType(pkg)
		if err != nil {
			return nil, err
		}
	}
	return ps, nil
}

func (r *Reader) readFuncType(pkg *ast.Package) (*ast.FuncType, error) {
	ps, err := r.readParams(pkg)
	if err != nil {
		return nil, err
	}
	rs, err := r.readParams(pkg)
	if err != nil {
		return nil, err
	}
	b, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	return &ast.FuncType{Params: ps, Returns: rs, Var: b == 1}, nil
}

func (r *Reader) readIfaceType(pkg *ast.Package) (*ast.InterfaceType, error) {
	ni, err := r.ReadNum()
	if err != nil {
		return nil, err
	}
	var ifs []ast.Type
	if ni > 0 {
		ifs = make([]ast.Type, ni)
	}
	for i := range ifs {
		t, err := r.readType(pkg)
		if err != nil {
			return nil, err
		}
		d, ok := t.(*ast.TypeDecl)
		if !ok {
			return nil, &BadFile
		}
		ifs[i] = d
	}
	nm, err := r.ReadNum()
	if err != nil {
		return nil, err
	}
	ms := make([]*ast.FuncDecl, nm)
	for i := range ms {
		kind, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		if kind != _FUNC_DECL {
			return nil, BadFile
		}
		f, err := r.readFuncDecl(pkg)
		if err != nil {
			return nil, err
		}
		ms[i] = f
	}
	return &ast.InterfaceType{Embedded: ifs, Methods: ms}, nil
}
