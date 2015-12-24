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
		Decls: make(map[string]ast.Symbol),
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
	file := &ast.File{Name: name, Decls: make(map[string]ast.Symbol)}
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
		pkg.Decls[name] = v
	case _CONST_DECL:
		c := &ast.Const{Off: off, File: file, Name: name, Type: typ}
		pkg.Decls[name] = c
	case _FUNC_DECL:
		var rcv *ast.Param
		if typ != ast.BuiltinNilType {
			rcv = &ast.Param{Type: typ}
		}
		typ, err = r.readType(pkg)
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
		if rcv != nil {
			base, err := receiverBaseType(rcv.Type)
			if err != nil {
				return false, err
			}
			if base == rcv.Type {
				base.Methods = append(base.Methods, f)
			} else {
				base.PMethods = append(base.PMethods, f)
			}
		}
	case _TYPE_DECL:
		var t *ast.TypeDecl
		// Check if the type declaration was created by an earlier encounter
		// with its type name.
		if sym, ok := pkg.Decls[name]; !ok {
			t = &ast.TypeDecl{}
			pkg.Decls[name] = t
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

func (r *Reader) readType(pkg *ast.Package) (ast.Type, error) {
	tag, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	switch tag {
	case _NIL:
		return ast.BuiltinNilType, nil
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
		if _, err := r.ReadNum(); err != nil {
			return nil, err
		}
		t, err := r.readType(pkg)
		if err != nil {
			return nil, err
		}
		return &ast.ArrayType{
			Dim: &ast.Literal{Kind: scanner.INTEGER, Value: []byte("10")},
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
		pkg.Decls[name] = dcl
		r.fwd++
		return dcl, nil
	} else if dcl, ok := sym.(*ast.TypeDecl); !ok {
		return nil, BadFile
	} else {
		return dcl, nil
	}
}

func (r *Reader) readStructType(pkg *ast.Package) (*ast.StructType, error) {
	n, err := r.ReadNum()
	if err != nil {
		return nil, err
	}
	fs := make([]ast.Field, n)
	for i := range fs {
		f := &fs[i]
		var err error
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
	return &ast.StructType{Fields: fs}, nil
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
	n, err := r.ReadNum()
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
		m.Type, err = r.readType(pkg)
		if err != nil {
			return nil, err
		}
		if _, ok := m.Type.(*ast.FuncType); ok {
			m.Name, err = r.ReadString()
			if err != nil {
				return nil, err
			}
		}
	}
	return &ast.InterfaceType{Methods: ms}, nil
}
