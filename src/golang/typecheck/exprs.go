package typecheck

import (
	"golang/ast"
	"math/big"
)

type exprVerifier struct {
	Pkg     *ast.Package
	File    *ast.File
	Files   []*ast.File
	Syms    []ast.Symbol
	Xs      []ast.Expr
	Done    map[ast.Symbol]struct{}
	TypeCtx ast.Type
	Iota    int
}

func verifyExprs(pkg *ast.Package) error {
	ev := exprVerifier{Pkg: pkg, Done: make(map[ast.Symbol]struct{}), Iota: -1}

	for _, s := range pkg.Syms {
		var err error
		switch d := s.(type) {
		case *ast.TypeDecl:
			err = ev.checkTypeDecl(d)
		case *ast.Const:
			err = ev.checkConstDecl(d)
		case *ast.Var:
			err = ev.checkVarDecl(d)
		case *ast.FuncDecl:
			err = ev.checkFuncDecl(d)
		default:
			panic("not reached")
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (ev *exprVerifier) checkTypeDecl(d *ast.TypeDecl) error {
	// No need to check things in a different package
	if d.File == nil || d.File.Pkg != ev.Pkg {
		return nil
	}
	// Check type.
	t, err := ev.checkType(d.Type)
	if err != nil {
		return err
	}
	d.Type = t
	// Check method declarations.
	for _, m := range d.Methods {
		if err := ev.checkFuncDecl(m); err != nil {
			return err
		}
	}
	for _, m := range d.PMethods {
		if err := ev.checkFuncDecl(m); err != nil {
			return err
		}
	}
	return nil
}

func (ev *exprVerifier) checkConstDecl(c *ast.Const) error {
	// No need to check things in a different package
	if c.File == nil || c.File.Pkg != ev.Pkg {
		return nil
	}
	// Check for a constant definition loop.
	if l := ev.checkLoop(c); l != nil {
		return &TypeInferLoop{Loop: l}
	}
	// Check if we have already processed this const declaration.
	if _, ok := ev.Done[c]; ok {
		return nil
	}
	ev.Done[c] = struct{}{}

	iota := ev.Iota
	ev.Iota = c.Iota
	ev.beginCheck(c, c.File)
	x, err := ev.checkExpr(c.Init, nil)
	ev.endCheck()
	ev.Iota = iota
	if err != nil {
		return err
	}
	c.Init = x
	if c.Type == nil {
		c.Type = x.Type()
	} else if y, ok := isAssignable(c.Type, x); ok {
		c.Init = y
	} else {
		return &NotAssignable{Off: x.Position(), File: c.File, Type: c.Type, X: x}
	}

	if _, ok := x.(*ast.ConstValue); !ok {
		return &NotConst{Off: c.Init.Position(), File: c.File, What: "const initializer"}
	}

	return nil
}

func (ev *exprVerifier) checkVarDecl(v *ast.Var) error {
	// No need to check things in a different package
	if v.File == nil || v.File.Pkg != ev.Pkg {
		return nil
	}
	// Check for a variable initialization loop.
	if l := ev.checkLoop(v); l != nil {
		return &TypeInferLoop{Loop: l}
	}

	if _, ok := ev.Done[v]; ok {
		return nil
	}

	ev.beginFile(v.File)
	defer func() { ev.endFile() }()

	// Check the declared type of the variables.
	if v.Type != nil {
		t, err := ev.checkType(v.Type)
		if err != nil {
			return err
		}
		// All the variables share this type.
		if v.Init == nil {
			v.Type = t
			ev.Done[v] = struct{}{}
		} else {
			for i := range v.Init.LHS {
				op := v.Init.LHS[i].(*ast.OperandName)
				v := op.Decl.(*ast.Var)
				v.Type = t
				ev.Done[v] = struct{}{}
			}
		}
	}

	if v.Init == nil || len(v.Init.RHS) == 0 {
		return nil
	}

	var tctx ast.Type
	if v.Type == nil {
		tctx = ast.BuiltinDefault
	} else {
		tctx = v.Type
	}

	if len(v.Init.LHS) > 1 && len(v.Init.RHS) == 1 {
		// If there is a single expression on the RHS, make all the variables
		// simultaneously depend on it. FIXME: multi-dependencies.
		for i := range v.Init.LHS {
			op := v.Init.LHS[i].(*ast.OperandName)
			ev.beginCheck(op.Decl, v.File)
		}
		x, err := ev.checkExpr(v.Init.RHS[0], tctx)
		for range v.Init.LHS {
			ev.endCheck()
		}
		if err != nil {
			return err
		}
		v.Init.RHS[0] = x
		// FIXME: don't loop
		// Assign types to the variables on the LHS and check assignment
		// compatibility of the initialization expression with the declared
		// type.
		tp, ok := x.Type().(*ast.TupleType)
		if !ok || len(tp.Type) != len(v.Init.LHS) {
			return &BadMultiValueAssign{Off: x.Position(), File: v.File}
		}
		for i := range v.Init.LHS {
			op := v.Init.LHS[i].(*ast.OperandName)
			v := op.Decl.(*ast.Var)
			if v.Type == nil {
				v.Type = tp.Type[i]
				ev.Done[v] = struct{}{}
			} else if !isAssignableType(v.Type, tp.Type[i]) {
				return &NotAssignable{Off: v.Off, File: v.File, Type: v.Type}
			}
		}
		return nil
	}

	// If there are multiple expressions on the RHS they must be single-valued
	// and the same number as the variables on the LHS.
	if len(v.Init.LHS) != len(v.Init.RHS) {
		return &BadMultiValueAssign{Off: v.Init.Off, File: v.File}
	}

	i := 0
	for i = range v.Init.LHS {
		op := v.Init.LHS[i].(*ast.OperandName)
		if v == op.Decl {
			break
		}
	}

	// Check the initializer expression.
	ev.beginCheck(v, v.File)
	x, err := ev.checkExpr(v.Init.RHS[i], tctx)
	ev.endCheck()
	if err != nil {
		return err
	}

	// Handle constant initializers.
	if c, ok := x.(*ast.ConstValue); ok && c.Typ == nil {
		// If checking resulted in untyped constant, convert it either to the
		// declared type, or to the default type.
		if v.Type == nil {
			x, err = ev.checkExpr(x, ast.BuiltinDefault)
			if err != nil {
				return err
			}
			v.Type = x.Type()
		} else if x, ok = isAssignable(v.Type, x); !ok {
			return &NotAssignable{Off: c.Off, File: v.File, Type: v.Type}
		}
		v.Init.RHS[i] = x
		ev.Done[v] = struct{}{}
		return nil
	}

	// If the initializer expression is `nil`, the variable must be declared
	// with a type.
	if x.Type() == ast.BuiltinNilType {
		if v.Type == nil {
			return &NilUse{Off: x.Position(), File: v.File}
		}
	}

	t := singleValueType(x.Type())
	if t == nil && x.Type() != nil {
		return &SingleValueContext{Off: x.Position(), File: v.File}
	}
	if v.Type == nil {
		v.Type = t
	} else if _, ok := isAssignable(v.Type, x); !ok {
		return &NotAssignable{Off: v.Off, File: v.File, Type: v.Type, X: x}
	}
	v.Init.RHS[i] = x
	ev.Done[v] = struct{}{}
	return nil
}

func (ev *exprVerifier) checkFuncDecl(fn *ast.FuncDecl) error {
	// No need to check things in a different package
	if fn.File == nil || fn.File.Pkg != ev.Pkg {
		return nil
	}

	ev.beginFile(fn.File)
	defer func() { ev.endFile() }()

	// Check the receiver type.
	if fn.Func.Recv != nil {
		t, err := ev.checkType(fn.Func.Recv.Type)
		if err != nil {
			return err
		}
		fn.Func.Recv.Type = t
	}
	// Check the signature.
	t, err := ev.checkType(fn.Func.Sig)
	if err != nil {
		return err
	}
	fn.Func.Sig = t.(*ast.FuncType)
	return nil
}

func (ev *exprVerifier) checkType(t ast.Type) (ast.Type, error) {
	return t.TraverseType(ev)
}

func (ev *exprVerifier) beginFile(f *ast.File) {
	ev.Files = append(ev.Files, ev.File)
	ev.File = f
}

func (ev *exprVerifier) endFile() {
	n := len(ev.Files)
	ev.File = ev.Files[n-1]
	ev.Files = ev.Files[:n-1]
}

func (ev *exprVerifier) beginCheck(s ast.Symbol, f *ast.File) {
	ev.beginFile(f)
	ev.Syms = append(ev.Syms, s)
}

func (ev *exprVerifier) endCheck() {
	ev.Syms = ev.Syms[:len(ev.Syms)-1]
	ev.endFile()
}

func (ev *exprVerifier) breakEvalChain() {
	ev.Syms = append(ev.Syms, nil)
}

func (ev *exprVerifier) restoreEvalChain() {
	ev.Syms = ev.Syms[:len(ev.Syms)-1]
}

func (ev *exprVerifier) checkLoop(sym ast.Symbol) []ast.Symbol {
	for i := len(ev.Syms) - 1; i >= 0; i-- {
		s := ev.Syms[i]
		if s == nil {
			return nil
		}
		if s == sym {
			return ev.Syms[i:]
		}
	}
	return nil
}

func (ev *exprVerifier) checkExprLoop(x ast.Expr) []ast.Expr {
	for i := len(ev.Xs) - 1; i >= 0; i-- {
		if x == ev.Xs[i] {
			return ev.Xs[i:]
		}
	}
	return nil
}

func (ev *exprVerifier) checkExpr(x ast.Expr, typ ast.Type) (ast.Expr, error) {
	if l := ev.checkExprLoop(x); l != nil {
		return nil, &ExprLoop{Off: x.Position(), File: ev.File}
	}

	typ, ev.TypeCtx = ev.TypeCtx, typ
	ev.Xs = append(ev.Xs, x)
	x, err := x.TraverseExpr(ev)
	ev.Xs = ev.Xs[:len(ev.Xs)-1]
	typ, ev.TypeCtx = ev.TypeCtx, typ
	return x, err
}

func (*exprVerifier) VisitError(*ast.Error) (*ast.Error, error) {
	panic("not reached")
}

func (*exprVerifier) VisitTypeName(*ast.QualifiedId) (ast.Type, error) {
	panic("not reached")
}

func (ev *exprVerifier) VisitTypeDeclType(t *ast.TypeDecl) (ast.Type, error) {
	return t, nil
}

func (*exprVerifier) VisitBuiltinType(t *ast.BuiltinType) (ast.Type, error) {
	return t, nil
}

func (ev *exprVerifier) checkArrayLength(t *ast.ArrayType) (*ast.ConstValue, error) {
	if t.Dim == nil {
		return nil, &BadUnspecArrayLen{Off: t.Position(), File: ev.File}
	}
	x, err := ev.checkExpr(t.Dim, nil)
	if err != nil {
		return nil, err
	}
	t.Dim = x
	c, ok := x.(*ast.ConstValue)
	if !ok {
		return nil, &NotConst{Off: x.Position(), File: ev.File, What: "array length"}
	}
	if c.Typ == ast.BuiltinInt {
		return c, nil
	}
	v := convertConst(ast.BuiltinInt, builtinType(c.Typ), c.Value)
	if v == nil {
		return nil, &BadConstConversion{
			Off: x.Position(), File: ev.File, Dst: ast.BuiltinInt, Src: c}
	}
	if int64(v.(ast.Int)) < 0 {
		return nil, &NegArrayLen{Off: c.Off, File: ev.File}
	}
	c.Typ = ast.BuiltinInt
	c.Value = v
	return c, nil
}

func (ev *exprVerifier) VisitArrayType(t *ast.ArrayType) (ast.Type, error) {
	elt, err := ev.checkType(t.Elt)
	if err != nil {
		return nil, err
	}
	t.Elt = elt
	_, err = ev.checkArrayLength(t)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (ev *exprVerifier) VisitSliceType(t *ast.SliceType) (ast.Type, error) {
	elt, err := ev.checkType(t.Elt)
	if err != nil {
		return nil, err
	}
	t.Elt = elt
	return t, nil
}

func (ev *exprVerifier) VisitPtrType(t *ast.PtrType) (ast.Type, error) {
	base, err := ev.checkType(t.Base)
	if err != nil {
		return nil, err
	}
	t.Base = base
	return t, nil
}

func (ev *exprVerifier) VisitMapType(t *ast.MapType) (ast.Type, error) {
	key, err := ev.checkType(t.Key)
	if err != nil {
		return nil, err
	}
	elt, err := ev.checkType(t.Elt)
	if err != nil {
		return nil, err
	}
	t.Key = key
	t.Elt = elt
	return t, nil
}

func (ev *exprVerifier) VisitChanType(t *ast.ChanType) (ast.Type, error) {
	elt, err := ev.checkType(t.Elt)
	if err != nil {
		return nil, err
	}
	t.Elt = elt
	return t, nil
}

func (ev *exprVerifier) VisitStructType(t *ast.StructType) (ast.Type, error) {
	// Check field types.
	for i := range t.Fields {
		fd := &t.Fields[i]
		t, err := ev.checkType(fd.Type)
		if err != nil {
			return nil, err
		}
		fd.Type = t
	}
	return t, nil
}

func (*exprVerifier) VisitTupleType(*ast.TupleType) (ast.Type, error) {
	panic("not reached")
}

func (ev *exprVerifier) VisitFuncType(t *ast.FuncType) (ast.Type, error) {
	for i := range t.Params {
		p := &t.Params[i]
		t, err := ev.checkType(p.Type)
		if err != nil {
			return nil, err
		}
		p.Type = t
	}
	for i := range t.Returns {
		p := &t.Returns[i]
		t, err := ev.checkType(p.Type)
		if err != nil {
			return nil, err
		}
		p.Type = t
	}
	return t, nil
}

func (ev *exprVerifier) VisitInterfaceType(typ *ast.InterfaceType) (ast.Type, error) {
	for _, m := range typ.Methods {
		t, err := ev.checkType(m.Func.Sig)
		if err != nil {
			return nil, err
		}
		m.Func.Sig = t.(*ast.FuncType)
	}
	return typ, nil
}

func (ev *exprVerifier) VisitQualifiedId(*ast.QualifiedId) (ast.Expr, error) {
	panic("not reached")
}

func (ev *exprVerifier) VisitConstValue(x *ast.ConstValue) (ast.Expr, error) {
	if x.Typ != nil || ev.TypeCtx == nil {
		return x, nil
	}

	var src, dst *ast.BuiltinType
	src = builtinType(x.Typ)
	if ev.TypeCtx == ast.BuiltinDefault {
		dst = builtinType(defaultType(x))
	} else {
		dst = builtinType(ev.TypeCtx)
	}
	v := convertConst(dst, src, x.Value)
	if v == nil {
		return nil, &BadConstConversion{Off: x.Off, File: ev.File, Dst: dst, Src: x}
	}
	return &ast.ConstValue{Off: x.Off, Typ: dst, Value: v}, nil
}

func (ev *exprVerifier) VisitCompLiteral(x *ast.CompLiteral) (ast.Expr, error) {
	if x.Typ == nil {
		return nil, &MissingLiteralType{x.Off, ev.File}
	}

	// If we have an array composite literal with no dimension, check first
	// the literal value.
	if t, ok := unnamedType(x.Typ).(*ast.ArrayType); ok && t.Dim == nil {
		x, err := ev.checkCompLiteral(x.Typ, x)
		if err != nil {
			return nil, err
		}
		t, err := ev.checkType(x.Typ)
		if err != nil {
			return nil, err
		}
		x.Typ = t
		return x, nil
	} else {
		t, err := ev.checkType(x.Typ)
		if err != nil {
			return nil, err
		}
		x.Typ = t
		return ev.checkCompLiteral(x.Typ, x)
	}
}

// Checks the composite literal expression X of type TYP. The type may come
// from the literal or from the context, in either case it has been already
// checked.
func (ev *exprVerifier) checkCompLiteral(
	typ ast.Type, x *ast.CompLiteral) (*ast.CompLiteral, error) {
	switch t := unnamedType(typ).(type) {
	case *ast.ArrayType:
		return ev.checkArrayLiteral(t, x)
	case *ast.SliceType:
		return ev.checkSliceLiteral(t, x)
	case *ast.MapType:
		return ev.checkMapLiteral(t, x)
	case *ast.StructType:
		return ev.checkStructLiteral(t, x)
	default:
		return nil, &BadLiteralType{Off: x.Off, File: ev.File, Type: t}
	}
}

func (ev *exprVerifier) checkArrayLiteral(
	t *ast.ArrayType, x *ast.CompLiteral) (*ast.CompLiteral, error) {

	n := int64(0x7fffffff) // FIXME
	if t.Dim != nil {
		c, err := ev.checkArrayLength(t)
		if err != nil {
			return nil, err
		}
		n = int64(c.Value.(ast.Int))
	}
	x, n, err := ev.checkArrayOrSliceLiteral(n, t.Elt, x)
	if err != nil {
		return nil, err
	}
	// If the array type has no dimension, set it to the length of the literal.
	if t.Dim == nil {
		// FIXME: check dimension fits in target `int`
		t.Dim = &ast.ConstValue{Off: x.Off, Typ: ast.BuiltinInt, Value: ast.Int(n)}
	}
	return x, nil
}

func (ev *exprVerifier) checkSliceLiteral(
	t *ast.SliceType, x *ast.CompLiteral) (*ast.CompLiteral, error) {
	x, _, err := ev.checkArrayOrSliceLiteral(int64(0x7fffffff), t.Elt, x)
	return x, err
}

func (ev *exprVerifier) checkArrayOrSliceLiteral(
	n int64, etyp ast.Type, x *ast.CompLiteral) (*ast.CompLiteral, int64, error) {

	// The element type itself cannot be an array of unspecified length.
	if t, ok := etyp.(*ast.ArrayType); ok && t.Dim == nil {
		return nil, 0, &BadUnspecArrayLen{Off: t.Off, File: ev.File}
	}

	keys := make(map[int64]struct{})
	idx := int64(0)
	max := int64(0)
	for _, elt := range x.Elts {
		// Assign type to elements, which are composite literals with elided
		// type.
		if c, ok := elt.Elt.(*ast.CompLiteral); ok {
			if c.Typ == nil {
				c.Typ = etyp
			}
		}
		if elt.Key != nil {
			// Check index.
			k, err := ev.checkExpr(elt.Key, nil)
			if err != nil {
				return nil, 0, err
			}
			elt.Key = k
			// Convert the index to `int`.
			c, ok := k.(*ast.ConstValue)
			if !ok {
				return nil, 0, &BadArraySize{
					Off: elt.Key.Position(), File: ev.File, What: "index"}
			}
			src := builtinType(c.Typ)
			v := convertConst(ast.BuiltinInt, src, c.Value)
			if v == nil {
				return nil, 0, &BadConstConversion{
					Off: x.Off, File: ev.File, Dst: ast.BuiltinInt, Src: c}
			}
			c.Typ = ast.BuiltinInt
			c.Value = v
			idx = int64(v.(ast.Int))
		}
		// Check for duplicate index.
		if _, ok := keys[idx]; ok {
			return nil, 0, &DupLitIndex{Off: x.Off, File: ev.File, Idx: idx}
		}
		keys[idx] = struct{}{}
		// Check the index is within bounds.
		if idx < 0 || idx >= n {
			off := elt.Elt.Position()
			if elt.Key != nil {
				off = elt.Key.Position()
			}
			return nil, 0, &IndexOutOfBounds{Off: off, File: ev.File, Idx: idx}
		}
		// Update maximum index.
		if idx > max {
			max = idx
		}
		idx++
		// Check element value.
		e, err := ev.checkExpr(elt.Elt, etyp)
		if err != nil {
			return nil, 0, err
		}
		elt.Elt = e
		// FIXME: check element is assignable to array/slice element type.
	}
	return x, max + 1, nil
}

func (ev *exprVerifier) checkMapLiteral(
	t *ast.MapType, x *ast.CompLiteral) (*ast.CompLiteral, error) {

	n := 0
	for _, elt := range x.Elts {
		// Assign types to keys and elements, which are composite literals
		// with elided type.
		if elt.Key != nil {
			n++
			if c, ok := elt.Key.(*ast.CompLiteral); ok {
				if c.Typ == nil {
					c.Typ = t.Key
				}
			}
			// Check key.
			k, err := ev.checkExpr(elt.Key, t.Key)
			if err != nil {
				return nil, err
			}
			elt.Key = k
		}
		if c, ok := elt.Elt.(*ast.CompLiteral); ok {
			if c.Typ == nil {
				c.Typ = t.Elt
			}
		}
		// Check element value.
		e, err := ev.checkExpr(elt.Elt, t.Elt)
		if err != nil {
			return nil, err
		}
		elt.Elt = e
	}
	// Every element must have a key.
	if n > 0 && n != len(x.Elts) {
		return nil, &MissingMapKey{Off: x.Off, File: ev.File}
	}

	// Check for duplicate constant keys.
	k := builtinType(t.Key)
	if k == nil {
		return x, nil
	}
	switch {
	case k.Kind == ast.BUILTIN_BOOL:
		return ev.checkUniqBoolKeys(x)
	case k.IsInteger():
		if k.IsSigned() {
			return ev.checkUniqIntKeys(x)
		} else {
			return ev.checkUniqUintKeys(x)
		}
	case k.Kind == ast.BUILTIN_FLOAT32 || k.Kind == ast.BUILTIN_FLOAT64:
		return ev.checkUniqFloatKeys(x)
	case k.Kind == ast.BUILTIN_COMPLEX64 || k.Kind == ast.BUILTIN_COMPLEX128:
		return ev.checkUniqComplexKeys(x)
	case k.Kind == ast.BUILTIN_STRING:
		return ev.checkUniqStringKeys(x)
	}
	return x, nil
}

func (ev *exprVerifier) checkUniqBoolKeys(x *ast.CompLiteral) (*ast.CompLiteral, error) {
	t, f := false, false
	for i := range x.Elts {
		c, ok := x.Elts[i].Key.(*ast.ConstValue)
		if !ok {
			continue
		}
		b := bool(c.Value.(ast.Bool))
		if b && t || !b && f {
			return nil, &DupLitKey{Off: x.Off, File: ev.File, Key: b}
		}
		if b {
			t = true
		} else {
			f = true
		}
	}
	return x, nil
}

func (ev *exprVerifier) checkUniqIntKeys(x *ast.CompLiteral) (*ast.CompLiteral, error) {
	keys := make(map[int64]struct{})
	for i := range x.Elts {
		c, ok := x.Elts[i].Key.(*ast.ConstValue)
		if !ok {
			continue
		}
		// The conversion must succeed due to a previous check for assignment
		// compatibility with the map key type.
		v := convertConst(ast.BuiltinInt64, builtinType(c.Typ), c.Value)
		k := int64(v.(ast.Int))
		if _, ok := keys[k]; ok {
			return nil, &DupLitKey{Off: x.Off, File: ev.File, Key: k}
		}
		keys[k] = struct{}{}
	}
	return x, nil
}

func (ev *exprVerifier) checkUniqUintKeys(x *ast.CompLiteral) (*ast.CompLiteral, error) {
	keys := make(map[uint64]struct{})
	for i := range x.Elts {
		c, ok := x.Elts[i].Key.(*ast.ConstValue)
		if !ok {
			continue
		}
		// The conversion must succeed due to a previous check for assignment
		// compatibility with the map key type.
		v := convertConst(ast.BuiltinUint64, builtinType(c.Typ), c.Value)
		k := uint64(v.(ast.Int))
		if _, ok := keys[k]; ok {
			return nil, &DupLitKey{Off: x.Off, File: ev.File, Key: k}
		}
		keys[k] = struct{}{}
	}
	return x, nil
}

func (ev *exprVerifier) checkUniqFloatKeys(x *ast.CompLiteral) (*ast.CompLiteral, error) {
	keys := make(map[float64]struct{})
	for i := range x.Elts {
		c, ok := x.Elts[i].Key.(*ast.ConstValue)
		if !ok {
			continue
		}
		// The conversion must succeed due to a previous check for assignment
		// compatibility with the map key type.
		v := convertConst(ast.BuiltinFloat64, builtinType(c.Typ), c.Value)
		k := float64(v.(ast.Float))
		if _, ok := keys[k]; ok {
			return nil, &DupLitKey{Off: x.Off, File: ev.File, Key: k}
		}
		keys[k] = struct{}{}
	}
	return x, nil
}

func (ev *exprVerifier) checkUniqComplexKeys(
	x *ast.CompLiteral) (*ast.CompLiteral, error) {

	keys := make(map[complex128]struct{})
	for i := range x.Elts {
		c, ok := x.Elts[i].Key.(*ast.ConstValue)
		if !ok {
			continue
		}
		// The conversion must succeed due to a previous check for assignment
		// compatibility with the map key type.
		v := convertConst(ast.BuiltinComplex128, builtinType(c.Typ), c.Value)
		k := complex128(v.(ast.Complex))
		if _, ok := keys[k]; ok {
			return nil, &DupLitKey{Off: x.Off, File: ev.File, Key: k}
		}
		keys[k] = struct{}{}
	}
	return x, nil
}

func (ev *exprVerifier) checkUniqStringKeys(x *ast.CompLiteral) (*ast.CompLiteral, error) {
	keys := make(map[string]struct{})
	for i := range x.Elts {
		c, ok := x.Elts[i].Key.(*ast.ConstValue)
		if !ok {
			continue
		}
		// The conversion must succeed due to a previous check for assignment
		// compatibility with the map key type.
		v := convertConst(ast.BuiltinString, builtinType(c.Typ), c.Value)
		k := string(v.(ast.String))
		if _, ok := keys[k]; ok {
			return nil, &DupLitKey{Off: x.Off, File: ev.File, Key: k}
		}
		keys[k] = struct{}{}
	}
	return x, nil
}

func (ev *exprVerifier) checkStructLiteral(
	str *ast.StructType, x *ast.CompLiteral) (*ast.CompLiteral, error) {

	if len(x.Elts) == 0 {
		return x, nil
	}

	if x.Elts[0].Key == nil {
		i, nf, ne := 0, len(str.Fields), len(x.Elts)
		for i < nf && i < ne {
			if x.Elts[i].Key != nil {
				return nil, &MixedStructLiteral{Off: x.Off, File: ev.File}
			}
			y, err := ev.checkExpr(x.Elts[i].Elt, str.Fields[i].Type)
			if err != nil {
				return nil, err
			}
			x.Elts[i].Elt = y
			i++
		}
		if nf != ne {
			return nil, &FieldEltMismatch{Off: x.Off, File: ev.File}
		}
	} else {
		keys := make(map[string]struct{})
		for _, elt := range x.Elts {
			if elt.Key == nil {
				return nil, &MixedStructLiteral{Off: x.Off, File: ev.File}
			}
			id, ok := elt.Key.(*ast.QualifiedId)
			if !ok || len(id.Pkg) > 0 || id.Id == "_" {
				return nil, &NotField{Off: elt.Key.Position(), File: ev.File}
			}
			f := findField(str, id.Id)
			if f == nil {
				return nil, &NotFound{
					Off: id.Off, File: ev.File, What: "field", Name: id.Id}
			}
			if _, ok := keys[id.Id]; ok {
				return nil, &DupLitField{Off: x.Off, File: ev.File, Name: id.Id}
			}
			keys[id.Id] = struct{}{}

			y, err := ev.checkExpr(elt.Elt, f.Type)
			if err != nil {
				return nil, err
			}
			elt.Elt = y
		}
	}
	return x, nil
}

func (ev *exprVerifier) VisitOperandName(x *ast.OperandName) (ast.Expr, error) {
	switch d := x.Decl.(type) {
	case *ast.Var:
		if d.Type == nil {
			if err := ev.checkVarDecl(d); err != nil {
				return nil, err
			}
		}
		x.Typ = d.Type
		return x, nil
	case *ast.Const:
		switch d.Init {
		case ast.BuiltinTrue, ast.BuiltinFalse:
			return &ast.ConstValue{Off: x.Off, Value: d.Init.(*ast.ConstValue).Value}, nil
		case ast.BuiltinIota:
			if ev.Iota == -1 {
				return nil, &BadIota{Off: x.Off, File: ev.File}
			}
			c := &ast.ConstValue{
				Off:   x.Off,
				Value: ast.UntypedInt{Int: big.NewInt(int64(ev.Iota))},
			}
			return ev.VisitConstValue(c)
		}
		if err := ev.checkConstDecl(d); err != nil {
			return nil, err
		}
		c := &ast.ConstValue{
			Off:   x.Off,
			Typ:   d.Type,
			Value: d.Init.(*ast.ConstValue).Value,
		}
		return ev.VisitConstValue(c)
	case *ast.FuncDecl:
		x.Typ = d.Func.Sig
		return x, nil
	default:
		panic("not reached")
	}
}

func (ev *exprVerifier) VisitCall(x *ast.Call) (ast.Expr, error) {
	// Check if we have a builtin function call.
	if op, ok := x.Func.(*ast.OperandName); ok {
		if d, ok := op.Decl.(*ast.FuncDecl); ok {
			switch d {
			case ast.BuiltinAppend:
				return ev.visitBuiltinAppend(x)
			case ast.BuiltinCap:
				return ev.visitBuiltinCap(x)
			case ast.BuiltinClose:
				return ev.visitBuiltinClose(x)
			case ast.BuiltinComplex:
				return ev.visitBuiltinComplex(x)
			case ast.BuiltinCopy:
				return ev.visitBuiltinCopy(x)
			case ast.BuiltinDelete:
				return ev.visitBuiltinDelete(x)
			case ast.BuiltinImag:
				return ev.visitBuiltinImag(x)
			case ast.BuiltinLen:
				return ev.visitBuiltinLen(x)
			case ast.BuiltinMake:
				return ev.visitBuiltinMake(x)
			case ast.BuiltinNew:
				return ev.visitBuiltinNew(x)
			case ast.BuiltinPanic:
				return ev.visitBuiltinPanic(x)
			case ast.BuiltinPrint:
				return ev.visitBuiltinPrint(x)
			case ast.BuiltinPrintln:
				return ev.visitBuiltinPrintln(x)
			case ast.BuiltinReal:
				return ev.visitBuiltinReal(x)
			case ast.BuiltinRecover:
				return ev.visitBuiltinRecover(x)
			}
		}
	}
	// Not a builtin call. Check the called expression.
	fn, err := ev.checkExpr(x.Func, nil)
	if err != nil {
		return nil, err
	}
	x.Func = fn
	ftyp, ok := fn.Type().(*ast.FuncType)
	if !ok {
		return nil, &NotFunc{Off: x.Off, File: ev.File, X: x}
	}
	switch len(ftyp.Returns) {
	case 0:
		// If the function has no returns, set the call expression type to
		// `void`.
		x.Typ = ast.BuiltinVoidType
	case 1:
		// If the function has a single return, set the type of the call
		// expression to that return's type.
		x.Typ = ftyp.Returns[0].Type
	default:
		// If the called function has multiple return values, set the type of
		// the call expression to a new TupleType, containing the returns'
		// types.
		tp := make([]ast.Type, len(ftyp.Returns))
		for i := range tp {
			tp[i] = ftyp.Returns[i].Type
		}
		x.Typ = &ast.TupleType{Off: x.Off, Strict: true, Type: tp}
	}

	// FIXME: check arguments match parameters
	return x, nil
}

func (*exprVerifier) visitBuiltinAppend(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*exprVerifier) visitBuiltinCap(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*exprVerifier) visitBuiltinClose(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*exprVerifier) visitBuiltinComplex(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*exprVerifier) visitBuiltinCopy(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*exprVerifier) visitBuiltinDelete(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*exprVerifier) visitBuiltinImag(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (ev *exprVerifier) visitBuiltinLen(x *ast.Call) (ast.Expr, error) {
	if x.ATyp != nil {
		return nil, &BadTypeArg{Off: x.Off, File: ev.File}
	}
	if len(x.Xs) != 1 {
		return nil, &BadArgNumber{Off: x.Off, File: ev.File}
	}
	y, err := ev.checkExpr(x.Xs[0], nil)
	if err != nil {
		return nil, err
	}
	x.Xs[0] = y
	if a, ok := unnamedType(y.Type()).(*ast.ArrayType); ok {
		c, err := ev.checkArrayLength(a)
		if err != nil {
			return nil, err
		}
		return c, nil
	}
	x.Typ = ast.BuiltinInt
	return x, nil
}

func (*exprVerifier) visitBuiltinMake(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*exprVerifier) visitBuiltinNew(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*exprVerifier) visitBuiltinPanic(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*exprVerifier) visitBuiltinPrint(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*exprVerifier) visitBuiltinPrintln(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*exprVerifier) visitBuiltinReal(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*exprVerifier) visitBuiltinRecover(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (ev *exprVerifier) VisitConversion(x *ast.Conversion) (ast.Expr, error) {
	t, err := ev.checkType(x.Typ)
	if err != nil {
		return nil, err
	}
	x.Typ = t
	y, err := ev.checkExpr(x.X, x.Typ)
	if err != nil {
		return nil, err
	}

	if c, ok := y.(*ast.ConstValue); ok {
		dst, ok := unnamedType(x.Typ).(*ast.BuiltinType)
		if !ok {
			return nil, &BadConstType{Off: x.Off, File: ev.File, Type: x.Typ}
		}
		v := convertConst(dst, builtinType(c.Typ), c.Value)
		if v == nil {
			return nil, &BadConstConversion{Off: x.Off, File: ev.File, Dst: dst, Src: c}
		}
		return &ast.ConstValue{Off: x.Off, Typ: x.Typ, Value: v}, nil
	}

	x.X = y

	// FIXME: check conversion is valid
	return x, nil
}

func (ev *exprVerifier) VisitMethodExpr(x *ast.MethodExpr) (ast.Expr, error) {
	// The parser/resolver will allow only receiver types that look like
	// expressions, i.e.  `T` or `*T`, where `T` is a typename. Although see
	// https://github.com/golang/go/issues/9060
	t, err := ev.checkType(x.RTyp)
	if err != nil {
		return nil, err
	}
	x.RTyp = t

	// Find the method.
	m0, m1, vptr := findSelector(ev.File.Pkg, x.RTyp, x.Id)
	if (m0.M != nil || m0.F != nil) && (m1.M != nil || m1.F != nil) {
		return nil, &AmbiguousSelector{Off: x.Off, File: ev.File, Name: x.Id}
	}
	if m0.M == nil && m1.M == nil {
		return nil, &NotFound{Off: x.Off, File: ev.File, What: "method", Name: x.Id}
	}
	// See whether the method has a pointer or value receiver.
	m := m0.M
	rptr := false
	if rcv := m.Func.Recv; rcv != nil {
		// The receiver can be NIL only if the method returned is from an
		// interface type declaration
		if _, ok := rcv.Type.(*ast.PtrType); ok {
			rptr = true
		}
	}
	if rptr && !vptr {
		// The combination of pointer receiver method and a value type in the
		// method expression is invalid.
		return nil, &BadMethodExpr{Off: x.Off, File: ev.File}
	}
	// Construct the function type of the method expression.
	ftyp := &ast.FuncType{
		Off:     -1,
		Params:  make([]ast.Param, 1+len(m.Func.Sig.Params)),
		Returns: m.Func.Sig.Returns,
		Var:     m.Func.Sig.Var,
	}
	copy(ftyp.Params[1:], m.Func.Sig.Params)
	ftyp.Params[0].Type = x.RTyp
	x.Typ = ftyp
	return x, nil
}

func (ev *exprVerifier) VisitParensExpr(*ast.ParensExpr) (ast.Expr, error) {
	panic("not reached")
}

func (ev *exprVerifier) VisitFunc(x *ast.Func) (ast.Expr, error) {
	t, err := ev.checkType(x.Sig)
	if err != nil {
		return nil, err
	}
	x.Sig = t.(*ast.FuncType)
	return x, nil
}

func (ev *exprVerifier) VisitTypeAssertion(x *ast.TypeAssertion) (ast.Expr, error) {
	if x.ATyp == nil {
		return nil, &BadTypeAssertion{Off: x.Off, File: ev.File}
	}
	t, err := ev.checkType(x.ATyp)
	if err != nil {
		return nil, err
	}
	x.ATyp = t
	x.Typ = &ast.TupleType{
		Off:    x.ATyp.Position(),
		Strict: false,
		Type:   []ast.Type{x.ATyp, ast.BuiltinBool},
	}
	y, err := ev.checkExpr(x.X, nil)
	if err != nil {
		return nil, err
	}
	x.X = y
	// FIXME: x.X must be interface type
	return x, nil
}

func (ev *exprVerifier) VisitSelector(x *ast.Selector) (ast.Expr, error) {
	y, err := ev.checkExpr(x.X, nil)
	if err != nil {
		return nil, err
	}
	x.X = y
	s0, s1, _ := findSelector(ev.File.Pkg, x.X.Type(), x.Id)
	if s1.M != nil || s1.F != nil {
		return nil, &AmbiguousSelector{Off: x.Position(), File: ev.File, Name: x.Id}
	}
	if s0.M == nil && s0.F == nil {
		return nil, &NotFound{
			Off: x.Off, File: ev.File, What: "field or method", Name: x.Id}
	}
	if s0.M != nil {
		x.Typ = s0.M.Func.Sig
	} else {
		x.Typ = s0.F.Type
	}
	return x, nil
}

func (ev *exprVerifier) VisitIndexExpr(x *ast.IndexExpr) (ast.Expr, error) {
	y, err := ev.checkExpr(x.X, nil)
	if err != nil {
		return nil, err
	}
	x.X = y

	// For maps, the type context for the index is the map key type, for other
	// indexed types, it's `int`.
	var ctx ast.Type
	if m, ok := x.X.Type().(*ast.MapType); ok {
		ctx = m.Key
	} else {
		ctx = ast.BuiltinInt
	}
	y, err = ev.checkExpr(x.I, ctx)
	if err != nil {
		return nil, err
	}
	x.I = y

	// Get the type of the indexed expression.  If the type of the indexed
	// expression is a pointer, the pointed to type must be an array type.
	typ := defaultType(x.X)
	if ptr, ok := typ.(*ast.PtrType); ok {
		b, ok := unnamedType(ptr.Base).(*ast.ArrayType)
		if !ok {
			return nil, &BadIndexedType{Off: x.Off, File: ev.File, Type: typ}
		}
		typ = b
	}

	switch t := typ.(type) {
	case *ast.BuiltinType:
		if t.Kind != ast.BUILTIN_STRING {
			return nil, &BadIndexedType{Off: x.Off, File: ev.File, Type: typ}
		}
		return ev.checkStringIndexExpr(x)
	case *ast.ArrayType:
		return ev.checkArrayIndexExpr(x, t)
	case *ast.SliceType:
		return ev.checkSliceIndexExpr(x, t)
	case *ast.MapType:
		return ev.checkMapIndexExpr(x, t)
	default:
		return nil, &BadIndexedType{Off: x.Off, File: ev.File, Type: typ}
	}
}

func (ev *exprVerifier) checkIndexValue(x ast.Expr) (int64, bool, error) {
	c, ok := x.(*ast.ConstValue)
	if !ok {
		// If the indexed expression is not of a map type, the index
		// expression must be of integer type.
		if t := builtinType(x.Type()); t == nil || !t.IsInteger() {
			return 0, false, &NotInteger{Off: x.Position(), File: ev.File, Type: x.Type()}
		}
		return 0, false, nil
	}

	// If the index expression is a constant, it must be representable by
	// `int` and non-negative.
	idx, ok := toInt(c)
	if !ok {
		return 0, false, &BadConstConversion{
			Off: c.Off, File: ev.File, Dst: ast.BuiltinInt, Src: c}
	}

	if idx < 0 {
		return 0, false, &IndexOutOfBounds{Off: c.Off, File: ev.File, Idx: idx}
	}

	return idx, true, nil
}

func (ev *exprVerifier) checkStringIndexExpr(x *ast.IndexExpr) (ast.Expr, error) {
	// If the type of the indexed expression is `string``, then the type of
	// the index expression is `byte`.
	x.Typ = ast.BuiltinUint8
	idx, idxIsConst, err := ev.checkIndexValue(x.I)
	if err != nil {
		return nil, err
	}

	// Indexing a constant string with a constant index is NOT a constant
	// expression, but still have to check the index is within bounds.
	if c, ok := x.X.(*ast.ConstValue); ok && idxIsConst {
		s := string(c.Value.(ast.String))
		if idx >= int64(len(s)) {
			return nil, &IndexOutOfBounds{Off: x.I.Position(), File: ev.File, Idx: idx}
		}
	}

	return x, nil
}

func (ev *exprVerifier) checkArrayIndexExpr(
	x *ast.IndexExpr, t *ast.ArrayType) (ast.Expr, error) {

	// The type of the expression is the array element type.
	x.Typ = t.Elt

	// Check the index is integer and, if constant, within array bounds.
	idx, idxIsConst, err := ev.checkIndexValue(x.I)
	if err != nil {
		return nil, err
	}

	if idxIsConst {
		c, err := ev.checkArrayLength(t)
		if err != nil {
			return nil, err
		}
		n := int64(c.Value.(ast.Int))
		if idx >= n {
			return nil, &IndexOutOfBounds{Off: x.I.Position(), File: ev.File, Idx: idx}
		}
	}

	return x, nil
}

func (ev *exprVerifier) checkSliceIndexExpr(
	x *ast.IndexExpr, t *ast.SliceType) (ast.Expr, error) {

	// The type of the expression is the array element type.
	x.Typ = t.Elt

	// Check the index is integer and non-negative
	_, _, err := ev.checkIndexValue(x.I)
	if err != nil {
		return nil, err
	}

	return x, nil
}

func (ev *exprVerifier) checkMapIndexExpr(
	x *ast.IndexExpr, t *ast.MapType) (ast.Expr, error) {

	// If the type of the indexed expression is a map, then the type of the
	// expression is a non-strict pair of map element type and bool.
	x.Typ = &ast.TupleType{
		Off:    -1,
		Strict: false,
		Type:   []ast.Type{t.Elt, ast.BuiltinBool},
	}
	// FIXME: The type of the index expression must be assignable to the map
	// key type.
	return x, nil
}

func (ev *exprVerifier) VisitSliceExpr(x *ast.SliceExpr) (ast.Expr, error) {
	y, err := ev.checkExpr(x.X, nil)
	if err != nil {
		return nil, err
	}
	x.X = y
	if x.Lo == nil {
		x.Lo = &ast.ConstValue{Off: -1, Typ: ast.BuiltinInt, Value: ast.Int(0)}
	} else {
		y, err = ev.checkExpr(x.Lo, ast.BuiltinInt)
		if err != nil {
			return nil, err
		}
		x.Lo = y
	}
	if x.Hi != nil {
		y, err = ev.checkExpr(x.Hi, ast.BuiltinInt)
		if err != nil {
			return nil, err
		}
		x.Hi = y
	}
	if x.Cap != nil {
		y, err = ev.checkExpr(x.Cap, ast.BuiltinInt)
		if err != nil {
			return nil, err
		}
		x.Cap = y
	}

	typ := defaultType(x.X)
	if ptr, ok := typ.(*ast.PtrType); ok {
		b, ok := unnamedType(ptr.Base).(*ast.ArrayType)
		if !ok {
			return nil, &BadIndexedType{Off: x.Off, File: ev.File, Type: typ}
		}
		typ = b

	}

	switch t := typ.(type) {
	case *ast.BuiltinType:
		if t.Kind != ast.BUILTIN_STRING {
			return nil, &BadIndexedType{Off: x.Off, File: ev.File, Type: typ}
		}
		x, err := ev.checkStringSliceExpr(x)
		return x, err
	case *ast.ArrayType:
		return ev.checkArraySliceExpr(x, t)
	case *ast.SliceType:
		return ev.checkSliceSliceExpr(x)
	default:
		return nil, &BadIndexedType{Off: x.Off, File: ev.File, Type: typ}
	}
}

func (ev *exprVerifier) checkStringSliceExpr(x *ast.SliceExpr) (ast.Expr, error) {
	// If the type of the indexed expression is `string``, then the type of
	// the slice expression is a non-constant string.
	x.Typ = ast.BuiltinString

	// Check index types
	var (
		err                  error
		lo, hi               int64
		loIsConst, hiIsConst bool
	)
	if x.Lo != nil {
		lo, loIsConst, err = ev.checkIndexValue(x.Lo)
		if err != nil {
			return nil, err
		}
	}
	if x.Hi != nil {
		hi, hiIsConst, err = ev.checkIndexValue(x.Hi)
		if err != nil {
			return nil, err
		}
	}

	// Check indices are within bounds.
	if c, ok := x.X.(*ast.ConstValue); ok {
		s := string(c.Value.(ast.String))
		n := int64(len(s))
		if loIsConst && (lo > n || hiIsConst && lo > hi) {
			return nil, &IndexOutOfBounds{Off: x.Lo.Position(), File: ev.File, Idx: lo}
		}
		if hiIsConst && hi > n {
			return nil, &IndexOutOfBounds{Off: x.Hi.Position(), File: ev.File, Idx: hi}
		}
	}
	if x.Cap != nil {
		return nil, &BadSliceExpr{Off: x.Off, File: ev.File}
	}

	return x, nil
}

func (ev *exprVerifier) checkArraySliceExpr(
	x *ast.SliceExpr, t *ast.ArrayType) (ast.Expr, error) {

	// Slicing an array produces a slice of the array element type.
	x.Typ = &ast.SliceType{Off: t.Off, Elt: t.Elt}

	// Check index types.
	var (
		err                              error
		lo, hi, cap                      int64
		loIsConst, hiIsConst, capIsConst bool
	)
	if x.Lo != nil {
		lo, loIsConst, err = ev.checkIndexValue(x.Lo)
		if err != nil {
			return nil, err
		}
	}
	if x.Hi != nil {
		hi, hiIsConst, err = ev.checkIndexValue(x.Hi)
		if err != nil {
			return nil, err
		}
	}
	if x.Cap != nil {
		cap, capIsConst, err = ev.checkIndexValue(x.Cap)
		if err != nil {
			return nil, err
		}
	}

	// Check indices are within bounds.
	c, err := ev.checkArrayLength(t)
	if err != nil {
		return nil, err
	}
	n := int64(c.Value.(ast.Int))

	if loIsConst && (lo > n || hiIsConst && lo > hi || capIsConst && lo > cap) {
		return nil, &IndexOutOfBounds{Off: x.Lo.Position(), File: ev.File, Idx: lo}
	}
	if hiIsConst && (hi > n || capIsConst && hi > cap) {
		return nil, &IndexOutOfBounds{Off: x.Hi.Position(), File: ev.File, Idx: hi}
	}
	if capIsConst && cap > n {
		return nil, &IndexOutOfBounds{Off: x.Hi.Position(), File: ev.File, Idx: cap}
	}
	return x, nil
}

func (ev *exprVerifier) checkSliceSliceExpr(x *ast.SliceExpr) (ast.Expr, error) {
	// Slicing a slice produces a slice of the same type.
	x.Typ = x.X.Type()

	// Check index types.
	var (
		err                              error
		lo, hi, cap                      int64
		loIsConst, hiIsConst, capIsConst bool
	)
	if x.Lo != nil {
		lo, loIsConst, err = ev.checkIndexValue(x.Lo)
		if err != nil {
			return nil, err
		}
	}
	if x.Hi != nil {
		hi, hiIsConst, err = ev.checkIndexValue(x.Hi)
		if err != nil {
			return nil, err
		}
	}
	if x.Cap != nil {
		cap, capIsConst, err = ev.checkIndexValue(x.Cap)
		if err != nil {
			return nil, err
		}
	}

	// Check indices are within bounds.
	if loIsConst && (hiIsConst && lo > hi || capIsConst && lo > cap) {
		return nil, &IndexOutOfBounds{Off: x.Lo.Position(), File: ev.File, Idx: lo}
	}
	if hiIsConst && capIsConst && hi > cap {
		return nil, &IndexOutOfBounds{Off: x.Hi.Position(), File: ev.File, Idx: hi}
	}
	return x, nil
}

func (ev *exprVerifier) VisitUnaryExpr(x *ast.UnaryExpr) (ast.Expr, error) {
	y, err := ev.checkExpr(x.X, nil)
	if err != nil {
		return nil, err
	}

	if _, ok := y.(*ast.ConstValue); !ok && y.Type() == nil {
		y, err = ev.checkExpr(y, ev.TypeCtx)
		if err != nil {
			return nil, err
		}
	}

	switch x.Op {
	case '+':
		return ev.checkUnaryPlus(x, y)
	case '-':
		return ev.checkUnaryMinus(x, y)
	case '!':
		return ev.checkNot(x, y)
	case '^':
		return ev.checkComplement(x, y)
	case '*':
		return ev.checkIndirection(x, y)
	case '&':
		return ev.checkAddr(x, y)
	case ast.RECV:
		return ev.checkRecv(x, y)
	default:
		panic("not reached")
	}
}

func (ev *exprVerifier) checkUnaryPlus(x *ast.UnaryExpr, y ast.Expr) (ast.Expr, error) {
	if definitelyNotArith(y) {
		return nil, &BadOperand{Off: y.Position(), File: ev.File, Op: '+'}
	}
	if _, ok := y.(*ast.ConstValue); ok {
		return y, nil
	} else {
		x.X = y
		x.Typ = y.Type()
		return x, nil
	}
}

func (ev *exprVerifier) checkUnaryMinus(x *ast.UnaryExpr, y ast.Expr) (ast.Expr, error) {
	if definitelyNotArith(y) {
		return nil, &BadOperand{Off: y.Position(), File: ev.File, Op: '-'}
	}
	if c, ok := y.(*ast.ConstValue); ok {
		v := minus(builtinType(c.Typ), c.Value)
		if v == nil {
			return nil, &BadOperand{Off: c.Off, File: ev.File, Op: '-'}
		}
		return &ast.ConstValue{Off: x.Off, Typ: c.Typ, Value: v}, nil
	} else {
		x.X = y
		x.Typ = y.Type()
		return x, nil
	}
}

func (ev *exprVerifier) checkNot(x *ast.UnaryExpr, y ast.Expr) (ast.Expr, error) {
	if y.Type() != nil {
		if t := builtinType(y.Type()); t == nil || t.Kind != ast.BUILTIN_BOOL {
			return nil, &BadOperand{Off: y.Position(), File: ev.File, Op: '!'}
		}
	}
	if c, ok := y.(*ast.ConstValue); ok {
		v, ok := c.Value.(ast.Bool)
		if !ok {
			return nil, &BadOperand{Off: y.Position(), File: ev.File, Op: '!'}
		}
		return &ast.ConstValue{Off: x.Off, Typ: c.Typ, Value: !v}, nil
	} else {
		x.X = y
		x.Typ = ast.BuiltinBool
		return x, nil
	}
}

func (ev *exprVerifier) checkComplement(x *ast.UnaryExpr, y ast.Expr) (ast.Expr, error) {
	if y.Type() != nil {
		if t := builtinType(y.Type()); t == nil || !t.IsInteger() {
			return nil, &BadOperand{Off: y.Position(), File: ev.File, Op: '^'}
		}
	}
	if c, ok := y.(*ast.ConstValue); ok {
		v := complement(builtinType(c.Typ), c.Value)
		if v == nil {
			return nil, &BadOperand{Off: c.Off, File: ev.File, Op: '^'}
		}
		return &ast.ConstValue{Off: x.Off, Typ: c.Typ, Value: v}, nil
	} else {
		x.X = y
		x.Typ = y.Type()
		return x, nil
	}
}

func (ev *exprVerifier) checkIndirection(x *ast.UnaryExpr, y ast.Expr) (ast.Expr, error) {
	ptr, ok := unnamedType(y.Type()).(*ast.PtrType)
	if !ok {
		return nil, &BadOperand{Off: y.Position(), File: ev.File, Op: '*'}
	}
	x.X = y
	x.Typ = ptr.Base
	return x, nil
}

func (ev *exprVerifier) checkAddr(x *ast.UnaryExpr, y ast.Expr) (ast.Expr, error) {
	if !isAddressable(y) {
		return nil, &BadOperand{Off: y.Position(), File: ev.File, Op: '&'}
	}
	x.X = y
	x.Typ = &ast.PtrType{Off: x.Off, Base: y.Type()}
	return x, nil
}

func (ev *exprVerifier) checkRecv(x *ast.UnaryExpr, y ast.Expr) (ast.Expr, error) {
	ch, ok := unnamedType(y.Type()).(*ast.ChanType)
	if !ok || !ch.Recv {
		return nil, &BadOperand{Off: y.Position(), File: ev.File, Op: ast.RECV}
	}
	x.X = y
	x.Typ = &ast.TupleType{Off: x.Off, Type: []ast.Type{ch.Elt, ast.BuiltinBool}}
	return x, nil
}

func (ev *exprVerifier) VisitBinaryExpr(x *ast.BinaryExpr) (ast.Expr, error) {
	if x.Op == ast.SHL || x.Op == ast.SHR {
		return ev.checkShift(x)
	}

	// Check the operands without a type context first, as a type context from
	// one operand takes precedene over type context passed by the parent
	// expression.
	u, err := ev.checkExpr(x.X, nil)
	if err != nil {
		return nil, err
	}
	v, err := ev.checkExpr(x.Y, nil)
	if err != nil {
		return nil, err
	}

	switch t0, t1 := u.Type(), v.Type(); {
	case t0 == nil && t1 != nil:
		// If one operand is untyped, convert it to the type of the other operand
		if u, err = ev.checkExpr(u, t1); err != nil {
			return nil, err
		}
	case t0 != nil && t1 == nil:
		if v, err = ev.checkExpr(v, t0); err != nil {
			return nil, err
		}
	}

	// Evaluate constant binary expressions.
	if u, ok := u.(*ast.ConstValue); ok {
		if v, ok := v.(*ast.ConstValue); ok {
			switch x.Op {
			case ast.LT, ast.GT, ast.EQ, ast.NE, ast.LE, ast.GE:
				v, err := compare(u, v, x.Op)
				if err != nil {
					return nil, &ErrorPos{Off: x.Off, File: ev.File, Err: err}
				}
				return &ast.ConstValue{Off: x.Off, Value: v}, nil
			}
		}
	}

	if u.Type() == nil && v.Type() == nil {
		// If both operands are without a type, retry with a type context,
		// either parent type context if not nil, or with default type
		// context.
		ctx := ev.TypeCtx
		if ctx == nil {
			ctx = ast.BuiltinDefault
		}
		if u, err = ev.checkExpr(u, ctx); err != nil {
			return nil, err
		}
		if v, err = ev.checkExpr(v, ctx); err != nil {
			return nil, err
		}
	}
	x.X = u
	x.Y = v

	switch x.Op {
	case ast.LT, ast.GT, ast.EQ, ast.NE, ast.LE, ast.GE:
		// For comparison operators, one of the operands must be assignable to
		// the type of the other operand.
		// FIXME: check assignability
		x.Typ = ast.BuiltinBool
	default:
		// For other operators, the operand types must be identical.
		// FIXME: check type identity.
		x.Typ = x.X.Type()
	}
	return x, nil
}

func (ev *exprVerifier) checkShift(x *ast.BinaryExpr) (ast.Expr, error) {
	// Check the left operand without a type context as the whole shift
	// expression might be a constant.
	u, err := ev.checkExpr(x.X, nil)
	if err != nil {
		return nil, err
	}
	// Check the right operand. Convert untyped constants at this point to
	// `uint64`.
	v, err := ev.checkExpr(x.Y, ast.BuiltinUint64)
	if err != nil {
		return nil, err
	}
	// Evaluate a constant shift expression.
	if u, ok := u.(*ast.ConstValue); ok {
		if s, ok := v.(*ast.ConstValue); ok {
			v, err := shift(u, s, x.Op)
			if err != nil {
				return nil, &ErrorPos{Off: x.Off, File: ev.File, Err: err}
			}
			return &ast.ConstValue{Off: x.Off, Typ: u.Typ, Value: v}, nil
		}
	}
	// The left operand is not a constant. Retry with a type context.
	if u.Type() == nil && ev.TypeCtx != nil {
		if u, err = ev.checkExpr(x.X, ev.TypeCtx); err != nil {
			return nil, err
		}
	}
	// Return if we cannot yet determine the type of the left operand.
	if u.Type() == nil {
		return x, nil
	}
	// The left operand should have integer type.
	if t := builtinType(u.Type()); t == nil || !t.IsInteger() {
		return nil, &BadOperand{Off: u.Position(), File: ev.File, Op: x.Op}
	}
	// The right operand should have unsigned integer type.
	if t := builtinType(v.Type()); t == nil || !t.IsInteger() || t.IsSigned() {
		return nil, &BadShiftCount{Off: v.Position(), File: ev.File}
	}
	x.X = u
	x.Y = v
	x.Typ = u.Type()
	return x, nil
}
