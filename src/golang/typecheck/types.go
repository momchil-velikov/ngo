package typecheck

import "golang/ast"

type typeVerifier struct {
	Pkg   *ast.Package
	File  *ast.File
	Files []*ast.File
	Types []*ast.TypeDecl
	Syms  []ast.Symbol
	Xs    []ast.Expr
	Done  map[string]struct{}
}

func verifyTypes(pkg *ast.Package) error {
	tv := typeVerifier{
		Pkg:  pkg,
		Done: make(map[string]struct{}),
	}

	for _, s := range pkg.Syms {
		var err error
		switch d := s.(type) {
		case *ast.TypeDecl:
			err = tv.checkTypeDecl(d)
		case *ast.Const:
			err = tv.checkConstDecl(d)
		case *ast.Var:
			err = tv.checkVarDecl(d)
		case *ast.FuncDecl:
			err = tv.checkFuncDecl(d)
		default:
			panic("not reached")
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (tv *typeVerifier) checkTypeDecl(d *ast.TypeDecl) error {
	// No need to check things in a different package
	if d.File == nil || d.File.Pkg != tv.Pkg {
		return nil
	}
	// Check for a typechecking loop
	if l := tv.checkTypeLoop(d); l != nil {
		return &TypeCheckLoop{Off: d.Off, File: d.File, Loop: l}
	}
	// Check if we have already processed this type declaration.
	if _, ok := tv.Done[d.Name]; ok {
		return nil
	}
	tv.Done[d.Name] = struct{}{}

	// Descend into the type literal
	tv.beginCheckType(d, d.File)
	err := tv.checkType(d.Type)
	tv.endCheckType()
	if err != nil {
		return err
	}
	// Check method declarations.
	for _, m := range d.Methods {
		if err := tv.checkFuncDecl(m); err != nil {
			return err
		}
	}
	for _, m := range d.PMethods {
		if err := tv.checkFuncDecl(m); err != nil {
			return err
		}
	}
	if err := checkMethodUniqueness(d); err != nil {
		return err
	}
	return nil
}

func (tv *typeVerifier) checkConstDecl(c *ast.Const) error {
	// No need to check things in a different package
	if c.File == nil || c.File.Pkg != tv.Pkg {
		return nil
	}

	if c.Type != nil {
		t, ok := unnamedType(c.Type).(*ast.BuiltinType)
		if !ok || t.Kind == ast.BUILTIN_NIL_TYPE {
			return &BadConstType{Off: c.Off, File: c.File, Type: c.Type}
		}
	}

	if c.Init != nil {
		tv.beginFile(c.File)
		err := tv.checkExpr(c.Init)
		tv.endFile()
		return err
	}

	return nil
}

func (tv *typeVerifier) checkVarDecl(v *ast.Var) error {
	// No need to check things in a different package
	if v.File == nil || v.File.Pkg != tv.Pkg {
		return nil
	}

	tv.beginFile(v.File)
	defer func() { tv.endFile() }()

	// Check the declared type of the variables.
	if v.Type != nil {
		// FIXME: All the variables share this type, maybe not check it
		// multiple times.
		if err := tv.checkType(v.Type); err != nil {
			return err
		}
	}

	if v.Init != nil {
		for _, x := range v.Init.RHS {
			if err := tv.checkExpr(x); err != nil {
				return err
			}
		}
	}

	return nil
}

func (tv *typeVerifier) checkFuncDecl(fn *ast.FuncDecl) error {
	// No need to check things in a different package
	if fn.File == nil || fn.File.Pkg != tv.Pkg {
		return nil
	}

	tv.beginFile(fn.File)
	defer func() { tv.endFile() }()

	// Check the receiver type.
	if fn.Func.Recv != nil {
		if err := tv.checkType(fn.Func.Recv.Type); err != nil {
			return err
		}
	}
	// Check the signature.
	if err := tv.checkType(fn.Func.Sig); err != nil {
		return err
	}

	return nil
}

func (tv *typeVerifier) checkType(t ast.Type) error {
	_, err := t.TraverseType(tv)
	return err
}

func (tv *typeVerifier) checkExpr(x ast.Expr) error {
	_, err := x.TraverseExpr(tv)
	return err
}

func (tv *typeVerifier) beginFile(f *ast.File) {
	tv.Files = append(tv.Files, tv.File)
	tv.File = f
}

func (tv *typeVerifier) endFile() {
	n := len(tv.Files)
	tv.File = tv.Files[n-1]
	tv.Files = tv.Files[:n-1]
}

func (tv *typeVerifier) beginCheckType(t *ast.TypeDecl, f *ast.File) {
	tv.beginFile(f)
	tv.Types = append(tv.Types, t)
}

func (tv *typeVerifier) endCheckType() {
	tv.Types = tv.Types[:len(tv.Types)-1]
	tv.endFile()
}

func (tv *typeVerifier) breakEmbedChain() {
	tv.Types = append(tv.Types, nil)
}

func (tv *typeVerifier) restoreEmbedChain() {
	tv.Types = tv.Types[:len(tv.Types)-1]
}

func (tv *typeVerifier) checkTypeLoop(t *ast.TypeDecl) []*ast.TypeDecl {
	for i := len(tv.Types) - 1; i >= 0; i-- {
		s := tv.Types[i]
		if s == nil {
			return nil
		}
		if s == t {
			return tv.Types[i:]
		}
	}
	return nil
}

// Types traversal

func (*typeVerifier) VisitError(*ast.Error) (*ast.Error, error) {
	panic("not reached")
}

func (*typeVerifier) VisitTypeName(*ast.QualifiedId) (ast.Type, error) {
	panic("not reached")
}

func (tv *typeVerifier) VisitTypeDeclType(td *ast.TypeDecl) (ast.Type, error) {
	if err := tv.checkTypeDecl(td); err != nil {
		return nil, err
	}
	return td, nil
}

func (*typeVerifier) VisitBuiltinType(t *ast.BuiltinType) (ast.Type, error) {
	return t, nil
}

func (tv *typeVerifier) VisitArrayType(t *ast.ArrayType) (ast.Type, error) {
	if err := tv.checkType(t.Elt); err != nil {
		return nil, err
	}
	if t.Dim != nil {
		if err := tv.checkExpr(t.Dim); err != nil {
			return nil, err
		}
	}
	return t, nil
}

func (tv *typeVerifier) VisitSliceType(t *ast.SliceType) (ast.Type, error) {
	tv.breakEmbedChain()
	err := tv.checkType(t.Elt)
	tv.restoreEmbedChain()
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (tv *typeVerifier) VisitPtrType(t *ast.PtrType) (ast.Type, error) {
	tv.breakEmbedChain()
	err := tv.checkType(t.Base)
	tv.restoreEmbedChain()
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (tv *typeVerifier) VisitMapType(t *ast.MapType) (ast.Type, error) {
	if err := tv.checkType(t.Key); err != nil {
		return nil, err
	}
	if !isEqualityComparable(t.Key) {
		return nil, &BadMapKey{Off: t.Position(), File: tv.File}
	}
	tv.breakEmbedChain()
	err := tv.checkType(t.Elt)
	tv.restoreEmbedChain()
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (tv *typeVerifier) VisitChanType(t *ast.ChanType) (ast.Type, error) {
	tv.breakEmbedChain()
	err := tv.checkType(t.Elt)
	tv.restoreEmbedChain()
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (tv *typeVerifier) VisitStructType(t *ast.StructType) (ast.Type, error) {
	// Check field types.
	for i := range t.Fields {
		fd := &t.Fields[i]
		if err := tv.checkType(fd.Type); err != nil {
			return nil, err
		}
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
				return nil, &DupFieldName{Field: &t.Fields[i], File: tv.File}
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
					Off: f.Off, File: tv.File, What: "pointer to interface type"}
			}
		}
		if _, ok := unnamedType(dcl.Type).(*ast.PtrType); ok {
			return nil, &BadAnonType{
				Off: f.Off, File: tv.File, What: "pointer type"}
		}
	}
	return t, nil
}

func (*typeVerifier) VisitTupleType(*ast.TupleType) (ast.Type, error) {
	panic("not reached")
}

func (tv *typeVerifier) VisitFuncType(t *ast.FuncType) (ast.Type, error) {
	tv.breakEmbedChain()
	defer tv.restoreEmbedChain()
	for i := range t.Params {
		if err := tv.checkType(t.Params[i].Type); err != nil {
			return nil, err
		}
	}
	for i := range t.Returns {
		if err := tv.checkType(t.Returns[i].Type); err != nil {
			return nil, err
		}
	}
	return t, nil
}

func (tv *typeVerifier) VisitInterfaceType(t *ast.InterfaceType) (ast.Type, error) {
	for i := range t.Embedded {
		// parser/resolver guarantee we have a TypeDecl
		d := t.Embedded[i].(*ast.TypeDecl)
		if _, ok := unnamedType(d).(*ast.InterfaceType); !ok {
			return nil, &BadEmbed{Type: d}
		}
		if err := tv.checkTypeDecl(d); err != nil {
			return nil, err
		}
	}
	for _, m := range t.Methods {
		if err := tv.checkType(m.Func.Sig); err != nil {
			return nil, err
		}
	}
	return t, nil
}

func (tv *typeVerifier) VisitQualifiedId(x *ast.QualifiedId) (ast.Expr, error) {
	// The only such case is for key in a struct composite literal.
	return x, nil
}

func (*typeVerifier) VisitConstValue(x *ast.ConstValue) (ast.Expr, error) {
	return x, nil
}

func (tv *typeVerifier) VisitCompLiteral(x *ast.CompLiteral) (ast.Expr, error) {
	if x.Typ != nil {
		if err := tv.checkType(x.Typ); err != nil {
			return nil, err
		}
	}

	for _, elt := range x.Elts {
		// Check index.
		if elt.Key != nil {
			if err := tv.checkExpr(elt.Key); err != nil {
				return nil, err
			}
		}
		// Check element value.
		if err := tv.checkExpr(elt.Elt); err != nil {
			return nil, err
		}
	}

	return x, nil
}

func (tv *typeVerifier) VisitOperandName(x *ast.OperandName) (ast.Expr, error) {
	return x, nil
}

func (tv *typeVerifier) VisitCall(x *ast.Call) (ast.Expr, error) {
	if x.ATyp != nil {
		if err := tv.checkType(x.ATyp); err != nil {
			return nil, err
		}
	}
	if err := tv.checkExpr(x.Func); err != nil {
		return nil, err
	}
	for _, x := range x.Xs {
		if err := tv.checkExpr(x); err != nil {
			return nil, err
		}
	}
	return x, nil
}

func (tv *typeVerifier) VisitConversion(x *ast.Conversion) (ast.Expr, error) {
	if err := tv.checkType(x.Typ); err != nil {
		return nil, err
	}
	if err := tv.checkExpr(x.X); err != nil {
		return nil, err
	}
	return x, nil
}

func (tv *typeVerifier) VisitMethodExpr(x *ast.MethodExpr) (ast.Expr, error) {
	return x, tv.checkType(x.RTyp)
}

func (tv *typeVerifier) VisitParensExpr(*ast.ParensExpr) (ast.Expr, error) {
	panic("not reached")
}

func (tv *typeVerifier) VisitFunc(x *ast.Func) (ast.Expr, error) {
	return x, tv.checkType(x.Sig)
}

func (tv *typeVerifier) VisitTypeAssertion(x *ast.TypeAssertion) (ast.Expr, error) {
	if x.ATyp == nil {
		return nil, &BadTypeAssertion{Off: x.Off, File: tv.File}
	}
	if err := tv.checkType(x.ATyp); err != nil {
		return nil, err
	}
	if err := tv.checkExpr(x.X); err != nil {
		return nil, err
	}
	return x, nil
}

func (tv *typeVerifier) VisitSelector(x *ast.Selector) (ast.Expr, error) {
	return x, tv.checkExpr(x.X)
}

func (tv *typeVerifier) VisitIndexExpr(x *ast.IndexExpr) (ast.Expr, error) {
	if err := tv.checkExpr(x.X); err != nil {
		return nil, err
	}
	if err := tv.checkExpr(x.I); err != nil {
		return nil, err
	}
	return x, nil
}

func (tv *typeVerifier) VisitSliceExpr(x *ast.SliceExpr) (ast.Expr, error) {
	if err := tv.checkExpr(x.X); err != nil {
		return nil, err
	}
	if x.Lo != nil {
		if err := tv.checkExpr(x.Lo); err != nil {
			return nil, err
		}
	}
	if x.Hi != nil {
		if err := tv.checkExpr(x.Hi); err != nil {
			return nil, err
		}
	}
	if x.Cap != nil {
		if err := tv.checkExpr(x.Cap); err != nil {
			return nil, err
		}
	}
	return x, nil
}

func (tv *typeVerifier) VisitUnaryExpr(x *ast.UnaryExpr) (ast.Expr, error) {
	return x, tv.checkExpr(x.X)
}

func (tv *typeVerifier) VisitBinaryExpr(x *ast.BinaryExpr) (ast.Expr, error) {
	if err := tv.checkExpr(x.X); err != nil {
		return nil, err
	}
	if err := tv.checkExpr(x.Y); err != nil {
		return nil, err
	}
	return x, nil
}
