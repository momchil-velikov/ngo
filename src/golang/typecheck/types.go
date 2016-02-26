package typecheck

import "golang/ast"

func (*typeckPhase0) VisitError(*ast.Error) (*ast.Error, error) {
	panic("not reached")
}

func (*typeckPhase0) VisitTypeName(*ast.QualifiedId) (ast.Type, error) {
	panic("not reached")
}

func (*typeckPhase0) VisitTypeVar(*ast.TypeVar) (ast.Type, error) {
	panic("not reached")
}

func (ck *typeckPhase0) VisitTypeDeclType(td *ast.TypeDecl) (ast.Type, error) {
	if err := ck.checkTypeDecl(td); err != nil {
		return nil, err
	}
	return td, nil
}

func (*typeckPhase0) VisitBuiltinType(*ast.BuiltinType) (ast.Type, error) {
	panic("not reached")
}

func (ck *typeckPhase0) VisitArrayType(t *ast.ArrayType) (ast.Type, error) {
	ck.breakEmbedChain()
	d, err := ck.checkExpr(t.Dim)
	ck.restoreEmbedChain()
	if err != nil {
		return nil, err
	}
	t.Dim = d
	elt, err := ck.checkType(t.Elt)
	if err != nil {
		return nil, err
	}
	t.Elt = elt
	return t, nil
}

func (ck *typeckPhase0) VisitSliceType(t *ast.SliceType) (ast.Type, error) {
	ck.breakEmbedChain()
	elt, err := ck.checkType(t.Elt)
	ck.restoreEmbedChain()
	if err != nil {
		return nil, err
	}
	t.Elt = elt
	return t, nil
}

func (ck *typeckPhase0) VisitPtrType(t *ast.PtrType) (ast.Type, error) {
	ck.breakEmbedChain()
	base, err := ck.checkType(t.Base)
	ck.restoreEmbedChain()
	if err != nil {
		return nil, err
	}
	t.Base = base
	return t, nil
}

func (ck *typeckPhase0) VisitMapType(t *ast.MapType) (ast.Type, error) {
	key, err := ck.checkType(t.Key)
	if err != nil {
		return nil, err
	}
	if !isEqualityComparable(key) {
		return nil, &BadMapKey{Off: t.Position(), File: ck.File}
	}
	ck.breakEmbedChain()
	elt, err := ck.checkType(t.Elt)
	ck.restoreEmbedChain()
	if err != nil {
		return nil, err
	}
	t.Key = key
	t.Elt = elt
	return t, nil
}

func (ck *typeckPhase0) VisitChanType(t *ast.ChanType) (ast.Type, error) {
	ck.breakEmbedChain()
	elt, err := ck.checkType(t.Elt)
	ck.restoreEmbedChain()
	if err != nil {
		return nil, err
	}
	t.Elt = elt
	return t, nil
}

func fieldName(f *ast.Field) string {
	if len(f.Name) > 0 {
		return f.Name
	}
	typ := f.Type
	if t, ok := f.Type.(*ast.PtrType); ok {
		typ = t.Base
	}
	// Parser guarantees that this type assertion will not fail.
	t := typ.(*ast.TypeDecl)
	return t.Name
}

func (ck *typeckPhase0) VisitStructType(t *ast.StructType) (ast.Type, error) {
	// Check field types.
	for i := range t.Fields {
		fd := &t.Fields[i]
		t, err := ck.checkType(fd.Type)
		if err != nil {
			return nil, err
		}
		fd.Type = t
	}
	// Check for uniqueness of field names.
	for i := range t.Fields {
		name := fieldName(&t.Fields[i])
		if name == "_" {
			continue
		}
		for j := i + 1; j < len(t.Fields); j++ {
			other := fieldName(&t.Fields[j])
			if name == other {
				return nil, &DupFieldName{Field: &t.Fields[i], File: ck.File}
			}
		}
	}
	// Check restrictions on the anonymous field types. The parser already
	// guarantees that they are in the form `T` or `*T`. Check the `T` is not
	// a pointer type, and, in the case of `*T`, that `T` is not an interface
	// or pointer type.
	for i := range t.Fields {
		f := &t.Fields[i]
		if len(f.Name) > 0 {
			continue
		}
		dcl, _ := f.Type.(*ast.TypeDecl)
		if ptr, ok := f.Type.(*ast.PtrType); ok {
			dcl = ptr.Base.(*ast.TypeDecl)
			if _, ok := unnamedType(dcl.Type).(*ast.InterfaceType); ok {
				return nil, &BadAnonType{
					Off: f.Off, File: ck.File, What: "pointer to interface type"}
			}
		}
		if _, ok := unnamedType(dcl.Type).(*ast.PtrType); ok {
			return nil, &BadAnonType{
				Off: f.Off, File: ck.File, What: "pointer type"}
		}
	}
	return t, nil
}

func (*typeckPhase0) VisitTupleType(*ast.TupleType) (ast.Type, error) {
	panic("not reached")
}

func (ck *typeckPhase0) VisitFuncType(t *ast.FuncType) (ast.Type, error) {
	ck.breakEmbedChain()
	defer ck.restoreEmbedChain()
	for i := range t.Params {
		p := &t.Params[i]
		t, err := ck.checkType(p.Type)
		if err != nil {
			return nil, err
		}
		p.Type = t
	}
	for i := range t.Returns {
		p := &t.Returns[i]
		t, err := ck.checkType(p.Type)
		if err != nil {
			return nil, err
		}
		p.Type = t
	}
	return t, nil
}

func (ck *typeckPhase0) VisitInterfaceType(typ *ast.InterfaceType) (ast.Type, error) {
	for i := range typ.Embedded {
		// parser/resolver guarantee we have a TypeDecl
		d := typ.Embedded[i].(*ast.TypeDecl)
		if _, ok := unnamedType(d).(*ast.InterfaceType); !ok {
			return nil, &BadEmbed{Type: d}
		}
		if err := ck.checkTypeDecl(d); err != nil {
			return nil, err
		}
	}
	for _, m := range typ.Methods {
		t, err := ck.checkType(m.Func.Sig)
		if err != nil {
			return nil, err
		}
		m.Func.Sig = t.(*ast.FuncType)
	}
	return typ, nil
}

func (*typeckPhase1) VisitError(*ast.Error) (*ast.Error, error) {
	panic("not reached")
}

func (*typeckPhase1) VisitTypeName(*ast.QualifiedId) (ast.Type, error) {
	panic("not reached")
}

func (ck *typeckPhase1) VisitTypeVar(tv *ast.TypeVar) (ast.Type, error) {
	if tv.Type == nil {
		panic("not reached")
	}
	if ck.checkTypeVarLoop(tv) != nil {
		return nil, &TypeInferLoop{Off: tv.Off, File: tv.File}
	}
	ck.TVars = append(ck.TVars, tv)
	t, err := ck.checkType(tv.Type)
	ck.TVars = ck.TVars[:len(ck.TVars)-1]
	if err != nil {
		return nil, err
	}
	tv.Type = t
	return t, nil
}

func (ck *typeckPhase1) VisitTypeDeclType(td *ast.TypeDecl) (ast.Type, error) {
	if err := ck.checkTypeDecl(td); err != nil {
		return nil, err
	}
	return td, nil
}

func (*typeckPhase1) VisitBuiltinType(*ast.BuiltinType) (ast.Type, error) {
	panic("not reached")
}

func (ck *typeckPhase1) VisitArrayType(t *ast.ArrayType) (ast.Type, error) {
	x, err := ck.checkExpr(t.Dim)
	if err != nil {
		return nil, err
	}
	t.Dim = x
	c, ok := x.(*ast.ConstValue)
	if !ok {
		return nil, &NotConst{Off: x.Position(), File: ck.File, What: "array dimension"}
	}
	v := convertConst(ast.BuiltinInt, builtinType(c.Typ), c.Value)
	if v == nil {
		return nil, &BadConversion{
			Off: x.Position(), File: ck.File,
			Dst: ast.BuiltinInt, Src: builtinType(c.Typ), Val: c.Value,
		}
	}
	if int64(v.(ast.Int)) < 0 {
		return nil, &NegArrayLen{Off: c.Off, File: ck.File}
	}
	c.Typ = ast.BuiltinInt
	c.Value = v
	elt, err := ck.checkType(t.Elt)
	if err != nil {
		return nil, err
	}
	t.Elt = elt
	return t, nil
}

func (ck *typeckPhase1) VisitSliceType(t *ast.SliceType) (ast.Type, error) {
	elt, err := ck.checkType(t.Elt)
	if err != nil {
		return nil, err
	}
	t.Elt = elt
	return t, nil
}

func (ck *typeckPhase1) VisitPtrType(t *ast.PtrType) (ast.Type, error) {
	base, err := ck.checkType(t.Base)
	if err != nil {
		return nil, err
	}
	t.Base = base
	return t, nil
}

func (ck *typeckPhase1) VisitMapType(t *ast.MapType) (ast.Type, error) {
	key, err := ck.checkType(t.Key)
	if err != nil {
		return nil, err
	}
	elt, err := ck.checkType(t.Elt)
	if err != nil {
		return nil, err
	}
	t.Key = key
	t.Elt = elt
	return t, nil
}

func (ck *typeckPhase1) VisitChanType(t *ast.ChanType) (ast.Type, error) {
	elt, err := ck.checkType(t.Elt)
	if err != nil {
		return nil, err
	}
	t.Elt = elt
	return t, nil
}

func (ck *typeckPhase1) VisitStructType(t *ast.StructType) (ast.Type, error) {
	for i := range t.Fields {
		fd := &t.Fields[i]
		t, err := ck.checkType(fd.Type)
		if err != nil {
			return nil, err
		}
		fd.Type = t
	}
	return t, nil
}

func (*typeckPhase1) VisitTupleType(*ast.TupleType) (ast.Type, error) {
	panic("not reached")
}

func (ck *typeckPhase1) VisitFuncType(t *ast.FuncType) (ast.Type, error) {
	for i := range t.Params {
		p := &t.Params[i]
		t, err := ck.checkType(p.Type)
		if err != nil {
			return nil, err
		}
		p.Type = t
	}
	for i := range t.Returns {
		p := &t.Returns[i]
		t, err := ck.checkType(p.Type)
		if err != nil {
			return nil, err
		}
		p.Type = t
	}
	return t, nil
}

func (ck *typeckPhase1) VisitInterfaceType(typ *ast.InterfaceType) (ast.Type, error) {
	for i := range typ.Embedded {
		d := typ.Embedded[i].(*ast.TypeDecl)
		if err := ck.checkTypeDecl(d); err != nil {
			return nil, err
		}
	}
	for _, m := range typ.Methods {
		t, err := ck.checkType(m.Func.Sig)
		if err != nil {
			return nil, err
		}
		m.Func.Sig = t.(*ast.FuncType)
	}
	return typ, nil
}
