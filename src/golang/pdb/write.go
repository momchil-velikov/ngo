package pdb

import (
	"golang/ast"
	"io"
	"lib/sort"
)

const VERSION = 1

func Write(w io.Writer, pkg *ast.Package) error {
	enc := Encoder{}
	enc.Init(w)
	return writePkg(&enc, pkg)
}

func writePkg(enc *Encoder, pkg *ast.Package) error {
	// Version
	if err := enc.WriteNum(VERSION); err != nil {
		return err
	}
	// Package name
	if err := enc.WriteString(pkg.Name); err != nil {
		return err
	}

	// Dependencies
	ss := make([]sort.StringKey, len(pkg.Deps))
	ss = ss[:0]
	for k, v := range pkg.Deps {
		ss = append(ss, sort.StringKey{Key: k, Value: v})
	}
	sort.StringKeySlice(ss).Quicksort()
	if err := enc.WriteNum(uint64(len(ss))); err != nil {
		return err
	}
	for i, n := 0, len(ss); i < n; i++ {
		if err := enc.WriteString(ss[i].Key); err != nil {
			return err
		}
		p := ss[i].Value.(*ast.Package)
		p.No = i + 2
		if err := enc.WriteBytes(p.Sig[:]); err != nil {
			return err
		}
	}

	// Source files
	ss = ss[:0]
	for _, f := range pkg.Files {
		ss = append(ss, sort.StringKey{Key: f.Name, Value: f})
	}
	sort.StringKeySlice(ss).Quicksort()
	if err := enc.WriteNum(uint64(len(ss))); err != nil {
		return err
	}
	for i := range ss {
		f := ss[i].Value.(*ast.File)
		f.No = i + 1
		if err := writeFile(enc, f); err != nil {
			return err
		}
	}

	// Declarations
	ss = ss[:0]
	for n, d := range pkg.Decls {
		if !d.IsExported() {
			continue
		}
		ss = append(ss, sort.StringKey{Key: n, Value: d})
	}
	sort.StringKeySlice(ss).Quicksort()
	for i := range ss {
		if err := writeDecl(enc, ss[i].Value.(ast.Symbol)); err != nil {
			return err
		}
	}

	// End of declarations
	if err := enc.WriteByte(_END); err != nil {
		return err
	}
	return nil
}

func writeFile(enc *Encoder, file *ast.File) error {
	// Output file name
	if err := enc.WriteString(file.Name); err != nil {
		return err
	}
	// Output source map.
	for i, n := 0, file.SrcMap.LineCount(); i < n; i++ {
		_, len := file.SrcMap.LineExtent(i)
		if err := enc.WriteNum(uint64(len)); err != nil {
			return err
		}
	}
	// Terminate source map with a zero.
	if err := enc.WriteNum(0); err != nil {
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

func writeDecl(enc *Encoder, d ast.Symbol) error {
	var (
		k  byte
		t  ast.Type
		fn int
	)
	name, off, file := d.DeclaredAt()
	if file != nil {
		fn = file.No
	}
	switch d := d.(type) {
	case *ast.FuncDecl:
		if err := enc.WriteByte(_FUNC_DECL); err != nil {
			return err
		}
		if err := enc.WriteNum(uint64(file.No)); err != nil {
			return err
		}
		if err := enc.WriteNum(uint64(off)); err != nil {
			return err
		}
		if err := enc.WriteString(name); err != nil {
			return err
		}
		var err error
		if d.Func.Recv == nil {
			err = enc.WriteByte(_NIL)
		} else {
			err = writeType(enc, d.Func.Recv.Type)
		}
		if err == nil {
			err = writeType(enc, d.Func.Sig)
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
	return writedcl(enc, k, fn, off, name, t)
}

func writedcl(enc *Encoder, k byte, file, off int, name string, t ast.Type) error {
	if err := enc.WriteByte(k); err != nil {
		return err
	}
	if err := enc.WriteNum(uint64(file)); err != nil {
		return err
	}
	if err := enc.WriteNum(uint64(off)); err != nil {
		return err
	}
	if err := enc.WriteString(name); err != nil {
		return err
	}
	return writeType(enc, t)
}

const (
	_NIL byte = iota
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

func writeType(enc *Encoder, t ast.Type) error {
	switch t := t.(type) {
	case *ast.BuiltinType:
		return writeBuiltinType(enc, t.Kind)
	case *ast.Typename:
		return writeTypename(enc, t)
	case *ast.ArrayType:
		if err := enc.WriteByte(_ARRAY); err != nil {
			return err
		}
		if err := enc.WriteNum(10 /* FIXME: t.Dim */); err != nil {
			return err
		}
		return writeType(enc, t.Elt)
	case *ast.SliceType:
		if err := enc.WriteByte(_SLICE); err != nil {
			return err
		}
		return writeType(enc, t.Elt)
	case *ast.PtrType:
		if err := enc.WriteByte(_PTR); err != nil {
			return err
		}
		return writeType(enc, t.Base)
	case *ast.MapType:
		if err := enc.WriteByte(_MAP); err != nil {
			return err
		}
		if err := writeType(enc, t.Key); err != nil {
			return err
		}
		return writeType(enc, t.Elt)
	case *ast.ChanType:
		b := byte(0)
		if t.Send {
			b = _CHAN_CAN_SEND
		}
		if t.Recv {
			b |= _CHAN_CAN_RECV
		}
		if err := enc.WriteByte(_CHAN); err != nil {
			return err
		}
		if err := enc.WriteByte(b); err != nil {
			return err
		}
		return writeType(enc, t.Elt)
	case *ast.StructType:
		return writeStructType(enc, t)
	case *ast.FuncType:
		return writeFuncType(enc, t)
	case *ast.InterfaceType:
		return writeIfaceType(enc, t)
	}
	return nil
}

func writeBuiltinType(enc *Encoder, k int) error {
	var b byte
	switch k {
	case ast.BUILTIN_NIL:
		b = _NIL
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
	return enc.WriteByte(b)
}

func writeTypename(enc *Encoder, t *ast.Typename) error {
	var p int
	if f := t.Decl.File; f == nil {
		p = 0
	} else if pkg := f.Pkg; pkg == nil {
		p = 1
	} else {
		p = pkg.No
	}
	if err := enc.WriteByte(_TYPENAME); err != nil {
		return err
	}
	if err := enc.WriteNum(uint64(p)); err != nil {
		return err
	}
	return enc.WriteString(t.Decl.Name)
}

func writeStructType(enc *Encoder, t *ast.StructType) error {
	if err := enc.WriteByte(_STRUCT); err != nil {
		return err
	}
	if err := enc.WriteNum(uint64(len(t.Fields))); err != nil {
		return err
	}
	for i := range t.Fields {
		f := &t.Fields[i]
		if err := enc.WriteString(f.Name); err != nil {
			return err
		}
		if err := writeType(enc, f.Type); err != nil {
			return err
		}
		if err := enc.WriteString(f.Tag); err != nil {
			return err
		}
	}
	return nil
}

func writeParams(enc *Encoder, ps []ast.Param) error {
	if err := enc.WriteNum(uint64(len(ps))); err != nil {
		return err
	}
	for i := range ps {
		p := &ps[i]
		if err := enc.WriteString(p.Name); err != nil {
			return err
		}
		if err := writeType(enc, p.Type); err != nil {
			return err
		}
	}
	return nil
}

func writeFuncType(enc *Encoder, t *ast.FuncType) error {
	if err := enc.WriteByte(_FUNC); err != nil {
		return err
	}
	if err := writeParams(enc, t.Params); err != nil {
		return err
	}
	if err := writeParams(enc, t.Returns); err != nil {
		return err
	}
	b := byte(0)
	if t.Var {
		b = 1
	}
	return enc.WriteByte(b)
}

func writeIfaceType(enc *Encoder, t *ast.InterfaceType) error {
	if err := enc.WriteByte(_IFACE); err != nil {
		return err
	}
	if err := enc.WriteNum(uint64(len(t.Methods))); err != nil {
		return err
	}
	for i := range t.Methods {
		var err error
		m := &t.Methods[i]
		if len(m.Name) > 0 {
			err = writeFuncType(enc, m.Type.(*ast.FuncType))
			if err == nil {
				err = enc.WriteString(m.Name)
			}
		} else {
			err = writeType(enc, m.Type)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
