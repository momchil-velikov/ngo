package typecheck

import "golang/ast"

type pkgTypesCheck struct {
	Files []*ast.File
	File  *ast.File
	Syms  []ast.Symbol
	Xs    []ast.Expr
	Done  map[ast.Symbol]struct{}
}

func checkPkgLevelTypes(pkg *ast.Package) error {
	ck := &pkgTypesCheck{Done: make(map[ast.Symbol]struct{})}
	for _, s := range pkg.Syms {
		_, ck.File = s.DeclaredAt()
		var err error
		switch d := s.(type) {
		case *ast.TypeDecl:
			err = ck.checkTypeDecl(d)
		case *ast.Const:
			err = ck.checkConstDeclType(d)
		case *ast.Var:
			err = ck.checkVarDeclType(d)
		case *ast.FuncDecl:
			err = ck.checkFuncDeclType(d)
		default:
			panic("not reached")
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (ck *pkgTypesCheck) checkTypeDecl(d *ast.TypeDecl) error {
	// No need to check things in a different package
	_, f := d.DeclaredAt()
	if f == nil || f.Pkg != ck.File.Pkg {
		return nil
	}
	// Check for a typechecking loop
	if l := ck.checkTypeLoop(d); l != nil {
		return &TypeCheckLoop{Off: d.Off, File: ck.File, Loop: l}
	}
	// Check if we have already processed this type declaration.
	if _, ok := ck.Done[d]; ok {
		return nil
	}
	// Descend into the type literal
	ck.Done[d] = struct{}{}
	ck.beginCheckType(d, f)
	t, err := d.Type.TraverseType(ck)
	ck.endCheckType()
	if err != nil {
		return err
	}
	if err := checkMethodUniqueness(d); err != nil {
		return err
	}
	d.Type = t
	return nil
}

func (ck *pkgTypesCheck) checkConstDeclType(c *ast.Const) error {
	if c.Type == nil {
		return nil
	}
	t, ok := unnamedType(c.Type).(*ast.BuiltinType)
	if !ok || t.Kind == ast.BUILTIN_NIL_TYPE {
		_, f := c.DeclaredAt()
		return &BadConstType{Off: c.Off, File: f, Type: c.Type}
	}
	return nil
}

func (ck *pkgTypesCheck) checkVarDeclType(v *ast.Var) error {
	if v.Type == nil {
		return nil
	}
	_, f := v.DeclaredAt()
	if f == nil {
		return nil
	}
	ck.beginCheckType(nil, f)
	t, err := v.Type.TraverseType(ck)
	ck.endCheckType()
	if err != nil {
		return err
	}
	v.Type = t
	return nil
}

func (ck *pkgTypesCheck) checkFuncDeclType(fn *ast.FuncDecl) error {
	// There's nothing to check here about a possible receiver.
	_, f := fn.DeclaredAt()
	if f == nil {
		return nil
	}
	ck.beginCheckType(nil, f)
	t, err := fn.Func.Sig.TraverseType(ck)
	ck.endCheckType()
	if err != nil {
		return err
	}
	fn.Func.Sig = t.(*ast.FuncType)
	return nil
}

func (ck *pkgTypesCheck) beginCheckType(s ast.Symbol, f *ast.File) {
	ck.Files = append(ck.Files, ck.File)
	ck.File = f
	ck.Syms = append(ck.Syms, s)
}

func (ck *pkgTypesCheck) endCheckType() {
	n := len(ck.Files)
	ck.File = ck.Files[n-1]
	ck.Files = ck.Files[:n-1]
	ck.Syms = ck.Syms[:len(ck.Syms)-1]
}

func (ck *pkgTypesCheck) breakEmbedChain() {
	ck.Syms = append(ck.Syms, nil)
}

func (ck *pkgTypesCheck) restoreEmbedChain() {
	ck.Syms = ck.Syms[:len(ck.Syms)-1]
}

func (ck *pkgTypesCheck) checkTypeLoop(sym ast.Symbol) []ast.Symbol {
	for i := len(ck.Syms); i > 0; i-- {
		s := ck.Syms[i-1]
		if s == nil {
			return nil
		}
		if s == sym {
			return ck.Syms[i-1:]
		}
	}
	return nil
}

// Returns the first type literal in a chain of TypeNames.
func unnamedType(typ ast.Type) ast.Type {
	for t, ok := typ.(*ast.TypeDecl); ok; t, ok = typ.(*ast.TypeDecl) {
		typ = t.Type
	}
	return typ
}

// Check if a type can be compared with `==` and `!=`.
//
// IMPORTANT: Callers must have ensured that type is not an invalid recursive
// type.
func isEqualityComparable(t ast.Type) bool {
	switch t := t.(type) {
	case *ast.BuiltinType, *ast.PtrType, *ast.ChanType, *ast.InterfaceType:
		return true
	case *ast.SliceType, *ast.MapType, *ast.FuncType:
		return false
	case *ast.ArrayType:
		return isEqualityComparable(t.Elt)
	case *ast.StructType:
		for i := range t.Fields {
			if !isEqualityComparable(t.Fields[i].Type) {
				return false
			}
		}
		return true
	case *ast.TypeDecl:
		return isEqualityComparable(t.Type)
	default:
		panic("not reached")
	}
}

// Checks for uniqueness of method names in the method set of the typename
// DCL.
// IMPORTANT: Callers must have ensured that type is not an invalid recursive
// type.
type methodSet map[string]*ast.FuncDecl

func checkMethodUniqueness(dcl *ast.TypeDecl) error {
	// Check for uniqueness of the method names in an interface type,
	// including embedded interfaces.
	if iface, ok := dcl.Type.(*ast.InterfaceType); ok {
		return checkIfaceMethodUniqueness(make(methodSet), dcl, iface)
	}

	// For non-interface types, check methods, declared on this typename.
	set := make(methodSet)
	for i := range dcl.Methods {
		f := dcl.Methods[i]
		g := set[f.Name]
		if g != nil {
			return &DupMethodName{M0: f, M1: g}
		}
		set[f.Name] = f
	}
	for i := range dcl.PMethods {
		f := dcl.PMethods[i]
		g := set[f.Name]
		if g != nil {
			return &DupMethodName{M0: f, M1: g}
		}
		set[f.Name] = f
	}

	// For struct types, additionally check for conflict with field names.
	if str, ok := unnamedType(dcl.Type).(*ast.StructType); ok {
		for i := range str.Fields {
			name := fieldName(&str.Fields[i])
			g := set[name]
			if g != nil {
				return &DupFieldMethodName{M: g, S: dcl}
			}
		}
	}
	return nil
}

// Checks for uniqueness of method names in the method set of the typename
// TOP. The parameter's DCL and IFACE denote either the top interface type
// declaration (same as TOP), or an embedded interface type.

// IMPORTANT: Callers must have ensured that type is not an invalid recursive
// type.
func checkIfaceMethodUniqueness(
	set methodSet, dcl *ast.TypeDecl, iface *ast.InterfaceType) error {

	for _, t := range iface.Embedded {
		ifc := unnamedType(t).(*ast.InterfaceType)
		if err := checkIfaceMethodUniqueness(set, dcl, ifc); err != nil {
			return err
		}

	}
	for _, m := range iface.Methods {
		if d, ok := set[m.Name]; ok {
			return &DupIfaceMethodName{
				Decl:  dcl,
				Off0:  m.Off,
				File0: m.File,
				Off1:  d.Off,
				File1: d.File,
				Name:  m.Name,
			}
		}
		set[m.Name] = m
	}
	return nil
}

func (*pkgTypesCheck) VisitError(*ast.Error) (*ast.Error, error) {
	panic("not reached")
}

func (*pkgTypesCheck) VisitTypeName(*ast.QualifiedId) (ast.Type, error) {
	panic("not reached")
}

func (*pkgTypesCheck) VisitTypeVar(*ast.TypeVar) (ast.Type, error) {
	panic("not reached")
}

func (ck *pkgTypesCheck) VisitTypeDeclType(td *ast.TypeDecl) (ast.Type, error) {
	if err := ck.checkTypeDecl(td); err != nil {
		return nil, err
	}
	return td, nil
}

func (*pkgTypesCheck) VisitBuiltinType(*ast.BuiltinType) (ast.Type, error) {
	panic("not reached")
}

func (ck *pkgTypesCheck) VisitArrayType(t *ast.ArrayType) (ast.Type, error) {
	elt, err := t.Elt.TraverseType(ck)
	if err != nil {
		return nil, err
	}
	t.Elt = elt
	return t, nil
}

func (ck *pkgTypesCheck) VisitSliceType(t *ast.SliceType) (ast.Type, error) {
	ck.breakEmbedChain()
	elt, err := t.Elt.TraverseType(ck)
	ck.restoreEmbedChain()
	if err != nil {
		return nil, err
	}
	t.Elt = elt
	return t, nil
}

func (ck *pkgTypesCheck) VisitPtrType(t *ast.PtrType) (ast.Type, error) {
	ck.breakEmbedChain()
	base, err := t.Base.TraverseType(ck)
	ck.restoreEmbedChain()
	if err != nil {
		return nil, err
	}
	t.Base = base
	return t, nil
}

func (ck *pkgTypesCheck) VisitMapType(t *ast.MapType) (ast.Type, error) {
	key, err := t.Key.TraverseType(ck)
	if err != nil {
		return nil, err
	}
	if !isEqualityComparable(key) {
		return nil, &BadMapKey{Off: t.Position(), File: ck.File}
	}
	ck.breakEmbedChain()
	elt, err := t.Elt.TraverseType(ck)
	ck.restoreEmbedChain()
	if err != nil {
		return nil, err
	}
	t.Key = key
	t.Elt = elt
	return t, nil
}

func (ck *pkgTypesCheck) VisitChanType(t *ast.ChanType) (ast.Type, error) {
	ck.breakEmbedChain()
	elt, err := t.Elt.TraverseType(ck)
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

func (ck *pkgTypesCheck) VisitStructType(t *ast.StructType) (ast.Type, error) {
	// Check field types.
	for i := range t.Fields {
		fd := &t.Fields[i]
		t, err := fd.Type.TraverseType(ck)
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
	return t, nil
}

func (*pkgTypesCheck) VisitTupleType(*ast.TupleType) (ast.Type, error) {
	panic("not reached")
}

func (ck *pkgTypesCheck) VisitFuncType(t *ast.FuncType) (ast.Type, error) {
	ck.breakEmbedChain()
	defer ck.restoreEmbedChain()
	for i := range t.Params {
		p := &t.Params[i]
		t, err := p.Type.TraverseType(ck)
		if err != nil {
			return nil, err
		}
		p.Type = t
	}
	for i := range t.Returns {
		p := &t.Returns[i]
		t, err := p.Type.TraverseType(ck)
		if err != nil {
			return nil, err
		}
		p.Type = t
	}
	return t, nil
}

func (ck *pkgTypesCheck) VisitInterfaceType(typ *ast.InterfaceType) (ast.Type, error) {
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
		t, err := m.Func.Sig.TraverseType(ck)
		if err != nil {
			return nil, err
		}
		m.Func.Sig = t.(*ast.FuncType)
	}
	return typ, nil
}
