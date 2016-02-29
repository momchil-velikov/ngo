package pdb

import (
	"golang/ast"
	"io"
	"lib/sort"
)

const VERSION = 1

type Writer struct {
	Encoder
	extra map[string]*ast.TypeDecl
}

func Write(w io.Writer, pkg *ast.Package) error {
	wr := &Writer{}
	wr.Encoder.Init(w)
	wr.extra = make(map[string]*ast.TypeDecl)
	return wr.writePkg(pkg)
}

func (w *Writer) writePkg(pkg *ast.Package) error {
	// Version
	if err := w.WriteNum(VERSION); err != nil {
		return err
	}
	// Package name
	if err := w.WriteString(pkg.Name); err != nil {
		return err
	}

	// Dependencies
	if err := w.WriteNum(uint64(len(pkg.Deps))); err != nil {
		return err
	}
	for _, i := range pkg.Deps {
		if err := w.WriteString(i.Path); err != nil {
			return err
		}
		if err := w.WriteBytes(i.Pkg.Sig[:]); err != nil {
			return err
		}
	}

	// Source files
	if err := w.WriteNum(uint64(len(pkg.Files))); err != nil {
		return err
	}
	for i, f := range pkg.Files {
		f.No = i + 1
		if err := w.writeFile(f); err != nil {
			return err
		}
	}

	// Exported declarations
	ss := make([]sort.StringKey, len(pkg.Syms))
	ss = ss[:0]
	for n, d := range pkg.Syms {
		if !d.IsExported() {
			continue
		}
		ss = append(ss, sort.StringKey{Key: n, Value: d})
	}
	sort.StringKeySlice(ss).Quicksort()
	for i := range ss {
		if err := w.writeDecl(pkg, ss[i].Value.(ast.Symbol)); err != nil {
			return err
		}
	}

	// Non-exported typenames, which nevertheless must be output.
	for len(w.extra) > 0 {
		ss = ss[:0]
		for _, t := range w.extra {
			ss = append(ss, sort.StringKey{Key: t.Name, Value: t})
		}
		w.extra = make(map[string]*ast.TypeDecl)
		sort.StringKeySlice(ss).Quicksort()
		for i := range ss {
			if err := w.writeDecl(pkg, ss[i].Value.(ast.Symbol)); err != nil {
				return err
			}
		}
	}

	// End of declarations
	if err := w.WriteByte(_END); err != nil {
		return err
	}
	return nil
}

func (w *Writer) writeFile(file *ast.File) error {
	// Output file name
	if err := w.WriteString(file.Name); err != nil {
		return err
	}
	// Output source map.
	for i, n := 0, file.SrcMap.LineCount(); i < n; i++ {
		_, len := file.SrcMap.LineExtent(i)
		if err := w.WriteNum(uint64(len)); err != nil {
			return err
		}
	}
	// Terminate source map with a zero.
	if err := w.WriteNum(0); err != nil {
		return err
	}
	return nil
}

const (
	_END byte = iota
	_VAR_DECL
	_CONST_DECL
	_FUNC_DECL
	_TYPE_DECL
)

func (w *Writer) writeDecl(pkg *ast.Package, d ast.Symbol) error {
	var (
		k  byte
		t  ast.Type
		fn int
	)
	name := d.Id()
	off, file := d.DeclaredAt()
	if file != nil {
		fn = file.No
	}
	switch d := d.(type) {
	case *ast.FuncDecl:
		if err := w.WriteByte(_FUNC_DECL); err != nil {
			return err
		}
		if err := w.WriteNum(uint64(fn)); err != nil {
			return err
		}
		if err := w.WriteNum(uint64(off)); err != nil {
			return err
		}
		if err := w.WriteString(name); err != nil {
			return err
		}
		var err error
		if d.Func.Recv == nil {
			err = w.WriteByte(_VOID)
		} else {
			err = w.writeType(pkg, d.Func.Recv.Type)
		}
		if err == nil {
			err = w.writeType(pkg, d.Func.Sig)
		}
		return err
	case *ast.Var:
		k = _VAR_DECL
		t = d.Type
	case *ast.Const:
		k = _CONST_DECL
		t = d.Type
	case *ast.TypeDecl:
		k = _TYPE_DECL
		t = d.Type
	default:
		panic("not reached")
	}
	return w.writedcl(pkg, k, fn, off, name, t)
}

func (w *Writer) writedcl(
	pkg *ast.Package, k byte, file, off int, name string, t ast.Type) error {

	if err := w.WriteByte(k); err != nil {
		return err
	}
	if err := w.WriteNum(uint64(file)); err != nil {
		return err
	}
	if err := w.WriteNum(uint64(off)); err != nil {
		return err
	}
	if err := w.WriteString(name); err != nil {
		return err
	}
	return w.writeType(pkg, t)
}

const (
	_VOID byte = iota
	_NIL
	_BOOL
	_UINT8
	_UINT16
	_UINT32
	_UINT64
	_INT8
	_INT16
	_INT32
	_INT64
	_FLOAT32
	_FLOAT64
	_COMPLEX64
	_COMPLEX128
	_UINT
	_INT
	_UINTPTR
	_STRING
	_ARRAY
	_SLICE
	_STRUCT
	_PTR
	_FUNC
	_IFACE
	_MAP
	_CHAN
	_TYPENAME
)

const (
	_CHAN_CAN_SEND = 1
	_CHAN_CAN_RECV = 2
)

func (w *Writer) writeType(pkg *ast.Package, t ast.Type) error {
	if t == nil {
		return w.WriteByte(_VOID)
	}
	switch t := t.(type) {
	case *ast.BuiltinType:
		return w.writeBuiltinType(t.Kind)
	case *ast.TypeDecl:
		return w.writeTypename(pkg, t)
	case *ast.ArrayType:
		if err := w.WriteByte(_ARRAY); err != nil {
			return err
		}
		if err := w.WriteNum(10 /* FIXME: t.Dim */); err != nil {
			return err
		}
		return w.writeType(pkg, t.Elt)
	case *ast.SliceType:
		if err := w.WriteByte(_SLICE); err != nil {
			return err
		}
		return w.writeType(pkg, t.Elt)
	case *ast.PtrType:
		if err := w.WriteByte(_PTR); err != nil {
			return err
		}
		return w.writeType(pkg, t.Base)
	case *ast.MapType:
		if err := w.WriteByte(_MAP); err != nil {
			return err
		}
		if err := w.writeType(pkg, t.Key); err != nil {
			return err
		}
		return w.writeType(pkg, t.Elt)
	case *ast.ChanType:
		b := byte(0)
		if t.Send {
			b = _CHAN_CAN_SEND
		}
		if t.Recv {
			b |= _CHAN_CAN_RECV
		}
		if err := w.WriteByte(_CHAN); err != nil {
			return err
		}
		if err := w.WriteByte(b); err != nil {
			return err
		}
		return w.writeType(pkg, t.Elt)
	case *ast.StructType:
		return w.writeStructType(pkg, t)
	case *ast.FuncType:
		return w.writeFuncType(pkg, t)
	case *ast.InterfaceType:
		return w.writeIfaceType(pkg, t)
	default:
		panic("not reached")
	}
}

func (w *Writer) writeBuiltinType(k int) error {
	var b byte
	switch k {
	case ast.BUILTIN_BOOL:
		b = _BOOL
	case ast.BUILTIN_UINT8:
		b = _UINT8
	case ast.BUILTIN_UINT16:
		b = _UINT16
	case ast.BUILTIN_UINT32:
		b = _UINT32
	case ast.BUILTIN_UINT64:
		b = _UINT64
	case ast.BUILTIN_INT8:
		b = _INT8
	case ast.BUILTIN_INT16:
		b = _INT16
	case ast.BUILTIN_INT32:
		b = _INT32
	case ast.BUILTIN_INT64:
		b = _INT64
	case ast.BUILTIN_FLOAT32:
		b = _FLOAT32
	case ast.BUILTIN_FLOAT64:
		b = _FLOAT64
	case ast.BUILTIN_COMPLEX64:
		b = _COMPLEX64
	case ast.BUILTIN_COMPLEX128:
		b = _COMPLEX128
	case ast.BUILTIN_UINT:
		b = _UINT
	case ast.BUILTIN_INT:
		b = _INT
	case ast.BUILTIN_UINTPTR:
		b = _UINTPTR
	case ast.BUILTIN_STRING:
		b = _STRING
	default:
		panic("not reached")
	}
	return w.WriteByte(b)
}

func (w *Writer) findImportNo(pkg *ast.Package, imp *ast.Package) int {
	if pkg == imp {
		return 1
	}
	for i := range pkg.Deps {
		if pkg.Deps[i].Pkg == imp {
			return i + 2
		}
	}
	panic("not reached")
}

func (w *Writer) writeTypename(pkg *ast.Package, t *ast.TypeDecl) error {
	pno := 0
	if f := t.File; f != nil {
		pno = w.findImportNo(pkg, f.Pkg)
	}
	if pno == 1 {
		// If we happen to output a non-exported type name (because it was
		// mentioned by some exported declaration, e.g. it is the return type
		// of an exported function) force output of this type declaration.
		if !t.IsExported() {
			w.extra[t.Name] = t
		}
	}
	if err := w.WriteByte(_TYPENAME); err != nil {
		return err
	}
	if err := w.WriteNum(uint64(pno)); err != nil {
		return err
	}
	return w.WriteString(t.Name)
}

func (w *Writer) writeStructType(pkg *ast.Package, t *ast.StructType) error {
	if err := w.WriteByte(_STRUCT); err != nil {
		return err
	}
	fn := int(0)
	if t.File != nil {
		fn = t.File.No
	}
	if err := w.WriteNum(uint64(fn)); err != nil {
		return err
	}
	if err := w.WriteNum(uint64(len(t.Fields))); err != nil {
		return err
	}
	for i := range t.Fields {
		f := &t.Fields[i]
		if err := w.WriteNum(uint64(f.Off)); err != nil {
			return err
		}
		if err := w.WriteString(f.Name); err != nil {
			return err
		}
		if err := w.writeType(pkg, f.Type); err != nil {
			return err
		}
		if err := w.WriteString(f.Tag); err != nil {
			return err
		}
	}
	return nil
}

func (w *Writer) writeParams(pkg *ast.Package, ps []ast.Param) error {
	if err := w.WriteNum(uint64(len(ps))); err != nil {
		return err
	}
	for i := range ps {
		p := &ps[i]
		if err := w.WriteString(p.Name); err != nil {
			return err
		}
		if err := w.writeType(pkg, p.Type); err != nil {
			return err
		}
	}
	return nil
}

func (w *Writer) writeFuncType(pkg *ast.Package, t *ast.FuncType) error {
	if err := w.WriteByte(_FUNC); err != nil {
		return err
	}
	if err := w.writeParams(pkg, t.Params); err != nil {
		return err
	}
	if err := w.writeParams(pkg, t.Returns); err != nil {
		return err
	}
	b := byte(0)
	if t.Var {
		b = 1
	}
	return w.WriteByte(b)
}

func (w *Writer) writeIfaceType(pkg *ast.Package, t *ast.InterfaceType) error {
	if err := w.WriteByte(_IFACE); err != nil {
		return err
	}
	if err := w.WriteNum(uint64(len(t.Embedded))); err != nil {
		return err
	}
	for _, d := range t.Embedded {
		if err := w.writeType(pkg, d); err != nil {
			return err
		}
	}
	if err := w.WriteNum(uint64(len(t.Methods))); err != nil {
		return err
	}
	for _, f := range t.Methods {
		if err := w.writeDecl(pkg, f); err != nil {
			return err
		}
	}
	return nil
}
