package typecheck

import (
	"golang/ast"
	"math/big"
)

type exprVerifier struct {
	Pkg   *ast.Package
	File  *ast.File
	Files []*ast.File
	Syms  []ast.Symbol
	Xs    []ast.Expr
	Done  map[*ast.Const]struct{}
	Iota  int
}

func verifyExprs(pkg *ast.Package) error {
	ev := exprVerifier{Pkg: pkg, Done: make(map[*ast.Const]struct{}), Iota: -1}

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
	if err := ev.checkType(d.Type); err != nil {
		return err
	}
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

	if c.Type == nil {
		panic("not reached")
	}

	// Check for a constant evaluation loop.
	// FIXME: not type inference loop.
	if l := ev.checkLoop(c); l != nil {
		return &TypeInferLoop{Loop: l}
	}

	// Check if we have already evaluated the value of this constant.
	if _, ok := ev.Done[c]; ok {
		return nil
	}
	ev.Done[c] = struct{}{}

	iota := ev.Iota
	ev.Iota = c.Iota
	ev.beginCheck(c, c.File)
	defer func() { ev.endCheck(); ev.Iota = iota }()

	x, err := ev.checkExpr(c.Init)
	if err != nil {
		return err
	}

	if _, ok := x.(*ast.ConstValue); !ok {
		return &NotConst{Off: x.Position(), File: c.File, What: "const initializer"}
	}

	if t := builtinType(c.Type); !t.IsUntyped() {
		if y, err := ev.isAssignable(c.Type, x); err != nil {
			return err
		} else {
			x = y
		}
	}
	c.Init = x

	return nil
}

func (ev *exprVerifier) checkVarDecl(v *ast.Var) error {
	// No need to check things in a different package
	if v.File == nil || v.File.Pkg != ev.Pkg {
		return nil
	}
	if v.Type == nil {
		panic("not reached")
	}

	ev.beginFile(v.File)
	defer func() { ev.endFile() }()

	// If there are no initializer expressions, check the variable
	// type. Otherwise, postpone type checking for after the initializer
	// expressions check, as it may modify the type, if the initialized
	// expression is an array literal with unspecified artay length.
	if v.Init == nil || len(v.Init.RHS) == 0 {
		return ev.checkType(v.Type)
	}

	if len(v.Init.LHS) > 1 && len(v.Init.RHS) == 1 {
		// Check the initializer and the type.
		ev.beginFile(v.File)
		x, err := ev.checkExpr(v.Init.RHS[0])
		ev.endFile()
		if err != nil {
			return err
		}
		v.Init.RHS[0] = x
		if err := ev.checkType(v.Type); err != nil {
			return err
		}

		// Check assignment compatibility of the initializer with the type.
		tp := x.Type().(*ast.TupleType)
		for i := range v.Init.LHS {
			op := v.Init.LHS[i].(*ast.OperandName)
			v := op.Decl.(*ast.Var)
			t := tp.Types[i]
			if ok, err := ev.isAssignableType(v.Type, t); err != nil {
				return err
			} else if !ok {
				return &NotAssignable{Off: v.Off, File: v.File, DType: v.Type, SType: t}
			}
		}
		return nil
	}

	// Find the index of the initializer expression, which corresponds to this
	// variable.
	i := 0
	for i = range v.Init.LHS {
		op := v.Init.LHS[i].(*ast.OperandName)
		if v == op.Decl {
			break
		}
	}

	// Check the initializer expression and the type.
	ev.beginFile(v.File)
	x, err := ev.checkExpr(v.Init.RHS[i])
	ev.endFile()
	if err != nil {
		return err
	}
	if err := ev.checkType(v.Type); err != nil {
		return err
	}

	// Convert initializer constants to the type of the variable.
	if _, ok := x.(*ast.ConstValue); ok {
		if x, err = ev.isAssignable(v.Type, x); err != nil {
			return err
		}
		v.Init.RHS[i] = x
		return nil
	}

	// Check assignment compatibility of the initialization expression with
	// the type.
	if ok, err := ev.isAssignableType(v.Type, x.Type()); err != nil {
		return err
	} else if !ok {
		return &NotAssignable{Off: v.Off, File: v.File, DType: v.Type, SType: x.Type()}
	}
	v.Init.RHS[i] = x
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
		if err := ev.checkType(fn.Func.Recv.Type); err != nil {
			return err
		}
	}
	// Check the signature.
	if err := ev.checkType(fn.Func.Sig); err != nil {
		return err
	}
	return nil
}

func (ev *exprVerifier) checkType(t ast.Type) error {
	_, err := t.TraverseType(ev)
	return err
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

func (ev *exprVerifier) checkExpr(x ast.Expr) (ast.Expr, error) {
	if l := ev.checkExprLoop(x); l != nil {
		return nil, &ExprLoop{Off: x.Position(), File: ev.File}
	}

	ev.Xs = append(ev.Xs, x)
	x, err := x.TraverseExpr(ev)
	ev.Xs = ev.Xs[:len(ev.Xs)-1]
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
	x, err := ev.checkExpr(t.Dim)
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
	v, err := ToInt(c)
	if err != nil {
		return nil, &ErrorPos{Off: x.Position(), File: ev.File, Err: err}
	}
	if v < 0 {
		return nil, &NegArrayLen{Off: c.Off, File: ev.File}
	}
	c.Typ = ast.BuiltinInt
	c.Value = ast.Int(v)
	return c, nil
}

func (ev *exprVerifier) VisitArrayType(t *ast.ArrayType) (ast.Type, error) {
	if err := ev.checkType(t.Elt); err != nil {
		return nil, err
	}
	if _, err := ev.checkArrayLength(t); err != nil {
		return nil, err
	}
	return t, nil
}

func (ev *exprVerifier) VisitSliceType(t *ast.SliceType) (ast.Type, error) {
	if err := ev.checkType(t.Elt); err != nil {
		return nil, err
	}
	return t, nil
}

func (ev *exprVerifier) VisitPtrType(t *ast.PtrType) (ast.Type, error) {
	if err := ev.checkType(t.Base); err != nil {
		return nil, err
	}
	return t, nil
}

func (ev *exprVerifier) VisitMapType(t *ast.MapType) (ast.Type, error) {
	if err := ev.checkType(t.Key); err != nil {
		return nil, err
	}
	if err := ev.checkType(t.Elt); err != nil {
		return nil, err
	}
	return t, nil
}

func (ev *exprVerifier) VisitChanType(t *ast.ChanType) (ast.Type, error) {
	if err := ev.checkType(t.Elt); err != nil {
		return nil, err
	}
	return t, nil
}

func (ev *exprVerifier) VisitStructType(t *ast.StructType) (ast.Type, error) {
	for i := range t.Fields {
		fd := &t.Fields[i]
		if err := ev.checkType(fd.Type); err != nil {
			return nil, err
		}
	}
	return t, nil
}

func (*exprVerifier) VisitTupleType(*ast.TupleType) (ast.Type, error) {
	panic("not reached")
}

func (ev *exprVerifier) VisitFuncType(t *ast.FuncType) (ast.Type, error) {
	for i := range t.Params {
		p := &t.Params[i]
		if err := ev.checkType(p.Type); err != nil {
			return nil, err
		}
	}
	for i := range t.Returns {
		p := &t.Returns[i]
		if err := ev.checkType(p.Type); err != nil {
			return nil, err
		}
	}
	return t, nil
}

func (ev *exprVerifier) VisitInterfaceType(t *ast.InterfaceType) (ast.Type, error) {
	for _, m := range t.Methods {
		if err := ev.checkType(m.Func.Sig); err != nil {
			return nil, err
		}
	}
	return t, nil
}

func (ev *exprVerifier) VisitQualifiedId(*ast.QualifiedId) (ast.Expr, error) {
	panic("not reached")
}

func (ev *exprVerifier) VisitConstValue(x *ast.ConstValue) (ast.Expr, error) {
	return x, nil
}

func (ev *exprVerifier) VisitCompLiteral(x *ast.CompLiteral) (ast.Expr, error) {
	// If we have an array composite literal with no dimension, check first
	// the literal value.
	if t, ok := underlyingType(x.Typ).(*ast.ArrayType); ok && t.Dim == nil {
		x, err := ev.checkCompLiteral(x.Typ, x)
		if err != nil {
			return nil, err
		}
		if err := ev.checkType(x.Typ); err != nil {
			return nil, err
		}
		return x, nil
	} else {
		if err := ev.checkType(x.Typ); err != nil {
			return nil, err
		}
		return ev.checkCompLiteral(x.Typ, x)
	}
}

// Checks the composite literal expression X of type TYP. The type may come
// from the literal or from the context, in either case it has been already
// checked.
func (ev *exprVerifier) checkCompLiteral(
	typ ast.Type, x *ast.CompLiteral) (*ast.CompLiteral, error) {
	switch t := underlyingType(typ).(type) {
	case *ast.ArrayType:
		return ev.checkArrayLiteral(t, x)
	case *ast.SliceType:
		return ev.checkSliceLiteral(t, x)
	case *ast.MapType:
		return ev.checkMapLiteral(t, x)
	case *ast.StructType:
		return ev.checkStructLiteral(t, x)
	default:
		panic("not reached")
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
		if elt.Key != nil {
			// Check index.
			k, err := ev.checkExpr(elt.Key)
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
			idx, err = ToInt(c)
			if err != nil {
				return nil, 0, &ErrorPos{Off: x.Off, File: ev.File, Err: err}
			}
			c.Typ, c.Value = ast.BuiltinInt, ast.Int(idx)
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
		e, err := ev.checkExpr(elt.Elt)
		if err != nil {
			return nil, 0, err
		}
		// Check element is assignable to array/slice element type.
		e, err = ev.isAssignable(etyp, e)
		if err != nil {
			return nil, 0, err
		}
		elt.Elt = e
	}
	return x, max + 1, nil
}

func (ev *exprVerifier) checkMapLiteral(
	t *ast.MapType, x *ast.CompLiteral) (*ast.CompLiteral, error) {

	n := 0
	for _, elt := range x.Elts {
		if elt.Key != nil {
			n++
			// Check key.
			k, err := ev.checkExpr(elt.Key)
			if err != nil {
				return nil, err
			}
			k, err = ev.isAssignable(t.Key, k)
			if err != nil {
				return nil, err
			}
			elt.Key = k
		}
		// Check element value.
		e, err := ev.checkExpr(elt.Elt)
		if err != nil {
			return nil, err
		}
		e, err = ev.isAssignable(t.Elt, e)
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
		k, _ := ToInt64(c)
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
		k, _ := ToUint64(c)
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
		k, _ := ToFloat(c)
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
		k, _ := ToComplex(c)
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
		k, _ := ToString(c)
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
		for i, n := 0, len(str.Fields); i < n; i++ {
			y, err := ev.checkExpr(x.Elts[i].Elt)
			if err != nil {
				return nil, err
			}
			y, err = ev.isAssignable(str.Fields[i].Type, y)
			if err != nil {
				return nil, err
			}
			x.Elts[i].Elt = y
		}
	} else {
		keys := make(map[string]struct{})
		for _, elt := range x.Elts {
			id := elt.Key.(*ast.QualifiedId)
			f := findField(str, id.Id)
			if _, ok := keys[id.Id]; ok {
				return nil, &DupLitField{Off: x.Off, File: ev.File, Name: id.Id}
			}
			keys[id.Id] = struct{}{}
			y, err := ev.checkExpr(elt.Elt)
			if err != nil {
				return nil, err
			}
			y, err = ev.isAssignable(f.Type, y)
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
		if t, ok := d.Type.(*ast.ArrayType); ok && t.Dim == nil {
			if err := ev.checkVarDecl(d); err != nil {
				return nil, err
			}
		}
		return x, nil
	case *ast.Const:
		var c *ast.ConstValue
		switch d.Init {
		case ast.BuiltinTrue, ast.BuiltinFalse:
			c = &ast.ConstValue{
				Off:   x.Off,
				Typ:   ast.BuiltinUntypedBool,
				Value: d.Init.(*ast.ConstValue).Value,
			}
		case ast.BuiltinIota:
			if ev.Iota == -1 {
				return nil, &BadIota{Off: x.Off, File: ev.File}
			}
			c = &ast.ConstValue{
				Off:   x.Off,
				Typ:   ast.BuiltinUntypedInt,
				Value: ast.UntypedInt{Int: big.NewInt(int64(ev.Iota))},
			}
		default:
			if err := ev.checkConstDecl(d); err != nil {
				return nil, err
			}
			c = &ast.ConstValue{
				Off:   x.Off,
				Typ:   d.Type,
				Value: d.Init.(*ast.ConstValue).Value,
			}
		}
		if dst := builtinType(x.Typ); dst != nil && !dst.IsUntyped() {
			c, err := Convert(x.Typ, c)
			if err != nil {
				return nil, &ErrorPos{Off: x.Off, File: ev.File, Err: err}
			}
			c.Off = x.Off
			return c, nil
		}
		return c, nil
	case *ast.FuncDecl:
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
	fn, err := ev.checkExpr(x.Func)
	if err != nil {
		return nil, err
	}
	x.Func = fn
	ftyp := fn.Type().(*ast.FuncType)

	// Type "argument" is not applicable to ordinary calls.
	if x.ATyp != nil {
		return nil, &BadTypeArg{Off: x.Off, File: ev.File}
	}

	// Check the special case `f(g(arguments-of-g))`.
	if len(x.Xs) == 1 && !x.Dots {
		if _, ok := x.Xs[0].(*ast.Call); ok {
			y, err := ev.checkExpr(x.Xs[0])
			if err != nil {
				return nil, err
			}
			x.Xs[0] = y

			if tp, ok := y.Type().(*ast.TupleType); ok {
				err = ev.checkArgumentTypes(x.Off, ftyp, tp.Types...)
			} else {
				err = ev.checkArgumentTypes(x.Off, ftyp, y.Type())
			}
			if err != nil {
				return nil, err
			}
			return x, nil
		}
	}

	// General case of a function call.
	for i := range x.Xs {
		y, err := ev.checkExpr(x.Xs[i])
		if err != nil {
			return nil, err
		}
		x.Xs[i] = y
	}
	if err := ev.checkArgumentExprs(ftyp, x.Xs, x.Dots); err != nil {
		return nil, err
	}
	return x, nil
}

func (ev *exprVerifier) checkArgumentTypes(
	off int, ftyp *ast.FuncType, ts ...ast.Type) error {

	nparm, narg := len(ftyp.Params), len(ts)
	nfix := nparm
	if !ftyp.Var && narg != nfix {
		// In calls to non-variadic functions, the number of parameters and
		// the number of arguments must be the same.
		return &BadArgNumber{Off: off, File: ev.File}
	}
	if ftyp.Var {
		nfix--
	}
	if narg < nfix {
		// Not enough arguments for fixed parameters.
		return &BadArgNumber{Off: off, File: ev.File}
	}
	// Check assignability to fixed parameter types.
	for i := 0; i < nfix; i++ {
		if ok, err := ev.isAssignableType(ftyp.Params[i].Type, ts[i]); err != nil {
			return err
		} else if !ok {
			return &NotAssignable{Off: off, File: ev.File, DType: ftyp.Params[i].Type,
				SType: ts[i]}

		}
	}
	// Check assignability to the variadic parameter type.
	for i := nfix; i < narg; i++ {
		if ok, err := ev.isAssignableType(ftyp.Params[nfix].Type, ts[i]); err != nil {
			return err
		} else if !ok {
			return &NotAssignable{Off: off, File: ev.File, DType: ftyp.Params[nfix].Type,
				SType: ts[i]}
		}
	}

	return nil
}

func (ev *exprVerifier) checkArgumentExprs(
	ftyp *ast.FuncType, xs []ast.Expr, dots bool) error {

	// Consistency between the number of parameters, number of arguments, use
	// of the `...` notation, and the variadicity of the function was checked
	// during type inference.
	nparm, narg := len(ftyp.Params), len(xs)
	nfix := nparm
	if ftyp.Var {
		nfix--
	}
	if dots {
		narg--
	}
	// Check assignability of the fixed argument expressions.
	for i := 0; i < nfix; i++ {
		y, err := ev.isAssignable(ftyp.Params[i].Type, xs[i])
		if err != nil {
			return err
		}
		xs[i] = y
	}
	// An argument expression which uses the `...` notation must be assignable
	// to `[]T`, where `T` is the variadic parameter type.
	if dots {
		if _, ok := underlyingType(xs[nfix].Type()).(*ast.SliceType); !ok {
			return &BadVariadicArg{Off: xs[nfix].Position(), File: ev.File}
		}
		dst := &ast.SliceType{
			Off: ftyp.Params[nfix].Type.Position(),
			Elt: ftyp.Params[nfix].Type,
		}
		ok, err := ev.isAssignableType(dst, xs[nfix].Type())
		if err != nil {
			return err
		}
		if !ok {
			return &NotAssignable{Off: xs[nfix].Position(), File: ev.File,
				DType: dst, SType: xs[nfix].Type()}
		}
	} else {
		// Check assignability of the rest of the arguments with the variadic
		// parameter type.
		for i := nfix; i < narg; i++ {
			y, err := ev.isAssignable(ftyp.Params[nfix].Type, xs[i])
			if err != nil {
				return err
			}
			xs[i] = y
		}
	}

	return nil
}

func (ev *exprVerifier) visitBuiltinAppend(x *ast.Call) (ast.Expr, error) {
	if x.ATyp != nil {
		return nil, &BadTypeArg{Off: x.ATyp.Position(), File: ev.File}
	}
	n := len(x.Xs)
	if n < 2 || x.Dots && n != 2 {
		return nil, &BadArgNumber{Off: x.Off, File: ev.File}
	}
	st := underlyingType(x.Typ).(*ast.SliceType)
	// Check the slice argument.
	y, err := ev.checkExpr(x.Xs[0])
	if err != nil {
		return nil, err
	}
	x.Xs[0] = y
	if x.Dots {
		y, err := ev.checkExpr(x.Xs[1])
		if err != nil {
			return nil, err
		}
		x.Xs[1] = y
		// As a special case, accept appending a string to a byte slice, using
		// the `...` notation.
		if st.Elt == ast.BuiltinUint8 {
			if t := builtinType(x.Xs[1].Type()); t != nil &&
				(t.Kind == ast.BUILTIN_STRING || t.Kind == ast.BUILTIN_UNTYPED_STRING) {
				return x, nil
			}
		}
		// Check the argument, passed via `...` is assignable to the slice type
		y, err = ev.isAssignable(st, x.Xs[1])
		if err != nil {
			return nil, err
		}
		x.Xs[1] = y
	} else {
		// Check argument expressions are assignable to slice element type.
		for i := 1; i < n; i++ {
			y, err := ev.checkExpr(x.Xs[i])
			if err != nil {
				return nil, err
			}
			y, err = ev.isAssignable(st.Elt, y)
			if err != nil {
				return nil, err
			}
			x.Xs[i] = y
		}
	}
	return x, nil
}

func (ev *exprVerifier) visitBuiltinCap(x *ast.Call) (ast.Expr, error) {
	if x.ATyp != nil {
		return nil, &BadTypeArg{Off: x.Off, File: ev.File}
	}
	if len(x.Xs) != 1 {
		return nil, &BadArgNumber{Off: x.Off, File: ev.File}
	}
	y, err := ev.checkExpr(x.Xs[0])
	if err != nil {
		return nil, err
	}
	x.Xs[0] = y
	typ := underlyingType(y.Type())
	if ptr, ok := typ.(*ast.PtrType); ok {
		b, ok := underlyingType(ptr.Base).(*ast.ArrayType)
		if !ok {
			return nil, &BadBuiltinArg{
				Off: x.Off, File: ev.File, Type: y.Type(), Func: "cap"}
		}
		typ = b
	}

	switch t := typ.(type) {
	case *ast.ArrayType:
		if !hasSideEffects(y) {
			return ev.checkArrayLength(t)
		}
	case *ast.SliceType, *ast.ChanType:
	default:
		return nil, &BadBuiltinArg{
			Off: x.Off, File: ev.File, Type: y.Type(), Func: "cap"}
	}
	return x, nil
}

func (ev *exprVerifier) visitBuiltinClose(x *ast.Call) (ast.Expr, error) {
	if x.ATyp != nil {
		return nil, &BadTypeArg{Off: x.Off, File: ev.File}
	}
	if len(x.Xs) != 1 {
		return nil, &BadArgNumber{Off: x.Off, File: ev.File}
	}
	y, err := ev.checkExpr(x.Xs[0])
	if err != nil {
		return nil, err
	}
	x.Xs[0] = y
	if ch, ok := underlyingType(y.Type()).(*ast.ChanType); !ok || !ch.Send {
		return nil, &BadBuiltinArg{
			Off: y.Position(), File: ev.File, Type: y.Type(), Func: "close"}
	}
	return x, nil
}

func (*exprVerifier) visitBuiltinComplex(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (ev *exprVerifier) visitBuiltinCopy(x *ast.Call) (ast.Expr, error) {
	if x.ATyp != nil {
		return nil, &BadTypeArg{Off: x.Off, File: ev.File}
	}
	if len(x.Xs) != 2 {
		return nil, &BadArgNumber{Off: x.Off, File: ev.File}
	}
	dst, err := ev.checkExpr(x.Xs[0])
	if err != nil {
		return nil, err
	}
	x.Xs[0] = dst
	src, err := ev.checkExpr(x.Xs[1])
	if err != nil {
		return nil, err
	}
	x.Xs[1] = src

	// The destination musy be a slice type.
	dt, ok := underlyingType(dst.Type()).(*ast.SliceType)
	if !ok {
		return nil, &BadBuiltinArg{
			Off: dst.Position(), File: ev.File, Type: dst.Type(), Func: "copy"}

	}
	if dt.Elt == ast.BuiltinUint8 {
		// As a special case, allow copying a string into a byte slice.
		if t := builtinType(src.Type()); t != nil &&
			(t.Kind == ast.BUILTIN_STRING || t.Kind == ast.BUILTIN_UNTYPED_STRING) {
			return x, nil
		}
	}
	st, ok := underlyingType(src.Type()).(*ast.SliceType)
	if !ok {
		return nil, &BadBuiltinArg{
			Off: src.Position(), File: ev.File, Type: src.Type(), Func: "copy"}
	}
	if ok, err := ev.identicalTypes(dt.Elt, st.Elt); err != nil {
		return nil, err
	} else if !ok {
		return nil, &NotAssignable{
			Off: src.Position(), File: ev.File, DType: dt, SType: src.Type()}
	}
	return x, nil
}

func (*exprVerifier) visitBuiltinDelete(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (ev *exprVerifier) visitBuiltinImag(x *ast.Call) (ast.Expr, error) {
	if x.ATyp != nil {
		return nil, &BadTypeArg{Off: x.Off, File: ev.File}
	}
	y, err := ev.checkExpr(x.Xs[0])
	if err != nil {
		return nil, err
	}
	x.Xs[0] = y
	if c, ok := y.(*ast.ConstValue); ok {
		d, err := Imag(c)
		if err != nil {
			// Impossible to happen due to argument type checks during type
			// inference.
			panic("not reached")
		}
		d.Off = c.Off
		return d, nil
	}
	return x, nil
}

func (ev *exprVerifier) visitBuiltinLen(x *ast.Call) (ast.Expr, error) {
	if x.ATyp != nil {
		return nil, &BadTypeArg{Off: x.Off, File: ev.File}
	}
	if len(x.Xs) != 1 {
		return nil, &BadArgNumber{Off: x.Off, File: ev.File}
	}
	y, err := ev.checkExpr(x.Xs[0])
	if err != nil {
		return nil, err
	}
	x.Xs[0] = y
	typ := underlyingType(y.Type())
	if ptr, ok := typ.(*ast.PtrType); ok {
		b, ok := underlyingType(ptr.Base).(*ast.ArrayType)
		if !ok {
			return nil, &BadBuiltinArg{
				Off: x.Off, File: ev.File, Type: y.Type(), Func: "len"}
		}
		typ = b
	}
	switch t := typ.(type) {
	case *ast.BuiltinType:
		if t.Kind != ast.BUILTIN_STRING && t.Kind != ast.BUILTIN_UNTYPED_STRING {
			return nil, &BadBuiltinArg{
				Off: x.Off, File: ev.File, Type: y.Type(), Func: "len"}
		}
		if c, ok := y.(*ast.ConstValue); ok {
			return &ast.ConstValue{
				Off:   x.Off,
				Typ:   ast.BuiltinInt,
				Value: ast.Int(len(c.Value.(ast.String))),
			}, nil
		}
	case *ast.ArrayType:
		if !hasSideEffects(y) {
			return ev.checkArrayLength(t)
		}
	case *ast.SliceType, *ast.ChanType, *ast.MapType:
	default:
		return nil, &BadBuiltinArg{
			Off: x.Off, File: ev.File, Type: y.Type(), Func: "len"}
	}
	return x, nil
}

func (ev *exprVerifier) visitBuiltinMake(x *ast.Call) (ast.Expr, error) {
	// Regardless of the type argument, all the expression arguments must be
	// integer or untyped. A constant argument must be non-negative and
	// representable by `int`.
	for i := range x.Xs {
		y, err := ev.checkExpr(x.Xs[i])
		if err != nil {
			return nil, err
		}
		if _, ok := y.(*ast.ConstValue); ok {
			z, err := ev.isAssignable(ast.BuiltinInt, y)
			if err != nil {
				return nil, err
			}
			if d := z.(*ast.ConstValue); int64(d.Value.(ast.Int)) < 0 {
				return nil, &NegMakeArg{Off: d.Off, File: ev.File}
			}
			x.Xs[i] = z
		} else {
			if t := builtinType(y.Type()); t == nil || !t.IsInteger() {
				return nil, &NotInteger{
					Off: y.Position(), File: ev.File, Type: y.Type(),
					What: "`make` argument"}
			}
			x.Xs[i] = y
		}
	}
	switch underlyingType(x.ATyp).(type) {
	case *ast.SliceType:
		// Making a slice should have at most two expression arguments.
		if len(x.Xs) > 2 {
			return nil, &BadArgNumber{Off: x.Off, File: ev.File}
		}
		// If the arguments are constant expressions, then the first (length)
		// must be no larger than the second(capacity).
		if len(x.Xs) == 2 {
			if c, ok := x.Xs[0].(*ast.ConstValue); ok {
				if d, ok := x.Xs[1].(*ast.ConstValue); ok {
					// The above assignability check converts constant values to `int`.
					len := c.Value.(ast.Int)
					cap := d.Value.(ast.Int)
					if len > cap {
						return nil, &LenExceedsCap{Off: x.Off, File: ev.File}
					}
				}
			}
		}
	case *ast.MapType, *ast.ChanType:
		// Making a map or a channel should have at most one expression
		// argument.
		if len(x.Xs) > 1 {
			return nil, &BadArgNumber{Off: x.Off, File: ev.File}
		}
	default:
		panic("not reached")
	}
	return x, nil
}

func (ev *exprVerifier) visitBuiltinNew(x *ast.Call) (ast.Expr, error) {
	if err := ev.checkType(x.ATyp); err != nil {
		return nil, err
	}
	if len(x.Xs) > 0 {
		return nil, &BadArgNumber{Off: x.Off, File: ev.File}
	}
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

func (ev *exprVerifier) visitBuiltinReal(x *ast.Call) (ast.Expr, error) {
	if x.ATyp != nil {
		return nil, &BadTypeArg{Off: x.Off, File: ev.File}
	}
	y, err := ev.checkExpr(x.Xs[0])
	if err != nil {
		return nil, err
	}
	x.Xs[0] = y
	if c, ok := y.(*ast.ConstValue); ok {
		d, err := Real(c)
		if err != nil {
			// Impossible to happen due to argument type checks during type
			// inference.
			panic("not reached")
		}
		d.Off = c.Off
		return d, nil
	}
	return x, nil
}

func (*exprVerifier) visitBuiltinRecover(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (ev *exprVerifier) VisitConversion(x *ast.Conversion) (ast.Expr, error) {
	if err := ev.checkType(x.Typ); err != nil {
		return nil, err
	}
	y, err := ev.checkExpr(x.X)
	if err != nil {
		return nil, err
	}

	// Check and perform constant conversion.
	if c, ok := y.(*ast.ConstValue); ok && !isEmptyInterfaceType(x.Typ) {
		c, err := Convert(x.Typ, c)
		if err != nil {
			return nil, &ErrorPos{Off: x.Off, File: ev.File, Err: err}
		}
		c.Off = x.Off
		return c, nil
	}

	// Check non-constant conversion.
	if ok, err := ev.isConvertible(x.Typ, y); err != nil {
		return nil, err
	} else if !ok {
		return nil, &NotConvertible{
			Off: y.Position(), File: ev.File, DType: x.Typ, SType: y.Type()}
	} else {
		x.X = y
		return x, nil
	}
}

func (ev *exprVerifier) VisitMethodExpr(x *ast.MethodExpr) (ast.Expr, error) {
	// The parser/resolver will allow only receiver types that look like
	// expressions, i.e.  `T` or `*T`, where `T` is a typename. Although see
	// https://github.com/golang/go/issues/9060
	if err := ev.checkType(x.RTyp); err != nil {
		return nil, err
	}
	return x, nil
}

func (ev *exprVerifier) VisitParensExpr(*ast.ParensExpr) (ast.Expr, error) {
	panic("not reached")
}

func (ev *exprVerifier) VisitFunc(x *ast.Func) (ast.Expr, error) {
	if err := ev.checkType(x.Sig); err != nil {
		return nil, err
	}
	return x, nil
}

func (ev *exprVerifier) VisitTypeAssertion(x *ast.TypeAssertion) (ast.Expr, error) {
	y, err := ev.checkExpr(x.X)
	if err != nil {
		return nil, err
	}
	x.X = y
	// Check the asserted expression has interface type.
	tx, ok := underlyingType(x.X.Type()).(*ast.InterfaceType)
	if !ok {
		return nil, &NotInterface{Off: x.Off, File: ev.File, Type: x.X.Type()}
	}
	if err := ev.checkType(x.ATyp); err != nil {
		return nil, err
	}
	// If asserting to a concrete type, said type must implement the interface
	// type of the expression.
	if _, ok := underlyingType(x.ATyp).(*ast.InterfaceType); !ok {
		if ok, err := ev.implements(x.ATyp, tx); err != nil {
			return nil, err
		} else if !ok {
			return nil, &DoesNotImplement{
				Off: x.Off, File: ev.File, Type: x.ATyp, Ifc: x.X.Type()}
		}
	}
	return x, nil
}

func (ev *exprVerifier) VisitSelector(x *ast.Selector) (ast.Expr, error) {
	y, err := ev.checkExpr(x.X)
	if err != nil {
		return nil, err
	}
	x.X = y
	return x, nil
}

func (ev *exprVerifier) VisitIndexExpr(x *ast.IndexExpr) (ast.Expr, error) {
	y, err := ev.checkExpr(x.X)
	if err != nil {
		return nil, err
	}
	x.X = y

	y, err = ev.checkExpr(x.I)
	if err != nil {
		return nil, err
	}
	x.I = y

	// Get the type of the indexed expression.  If the type of the indexed
	// expression is a pointer, the pointed to type must be an array type.
	typ := underlyingType(x.X.Type())
	if ptr, ok := typ.(*ast.PtrType); ok {
		typ = underlyingType(ptr.Base)
	}

	switch t := typ.(type) {
	case *ast.BuiltinType:
		return ev.checkStringIndexExpr(x)
	case *ast.ArrayType:
		return ev.checkArrayIndexExpr(x, t)
	case *ast.SliceType:
		return ev.checkSliceIndexExpr(x, t)
	case *ast.MapType:
		return ev.checkMapIndexExpr(x, t)
	default:
		panic("not reached")
	}
}

func (ev *exprVerifier) checkIndexValue(x ast.Expr) (int64, bool, error) {
	c, ok := x.(*ast.ConstValue)
	if !ok {
		// If the indexed expression is not of a map type, the index
		// expression must be of integer type.
		if t := builtinType(x.Type()); t == nil || !t.IsInteger() {
			return 0, false, &NotInteger{
				Off: x.Position(), File: ev.File, Type: x.Type(), What: "index"}
		}
		return 0, false, nil
	}
	// If the index expression is a constant, it must be representable by
	// `int` and non-negative.
	idx, err := ToInt(c)
	if err != nil {
		return 0, false, &ErrorPos{Off: c.Off, File: ev.File, Err: err}
	}
	if idx < 0 {
		return 0, false, &IndexOutOfBounds{Off: c.Off, File: ev.File, Idx: idx}
	}
	return idx, true, nil
}

func (ev *exprVerifier) checkStringIndexExpr(x *ast.IndexExpr) (ast.Expr, error) {
	// If the type of the indexed expression is `string``, then the type of
	// the index expression is `byte`.
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

	// Check the index is integer and non-negative
	_, _, err := ev.checkIndexValue(x.I)
	if err != nil {
		return nil, err
	}
	return x, nil
}

func (ev *exprVerifier) checkMapIndexExpr(
	x *ast.IndexExpr, t *ast.MapType) (ast.Expr, error) {

	// The type of the index expression must be assignable to the map key
	// type.
	ok, err := ev.isAssignableType(t.Key, x.I.Type())
	if err != nil {
		return nil, err
	} else if !ok {
		return nil, &NotAssignable{
			Off: x.I.Position(), File: ev.File, DType: t.Key, SType: x.I.Type()}
	}
	return x, nil
}

func (ev *exprVerifier) VisitSliceExpr(x *ast.SliceExpr) (ast.Expr, error) {
	y, err := ev.checkExpr(x.X)
	if err != nil {
		return nil, err
	}
	x.X = y

	if x.Lo == nil {
		x.Lo = &ast.ConstValue{Off: -1, Typ: ast.BuiltinInt, Value: ast.Int(0)}
	} else {
		y, err = ev.checkExpr(x.Lo)
		if err != nil {
			return nil, err
		}
		x.Lo = y
	}
	if x.Hi != nil {
		y, err = ev.checkExpr(x.Hi)
		if err != nil {
			return nil, err
		}
		x.Hi = y
	}
	if x.Cap != nil {
		y, err = ev.checkExpr(x.Cap)
		if err != nil {
			return nil, err
		}
		x.Cap = y
	}

	typ := underlyingType(x.X.Type())
	if ptr, ok := typ.(*ast.PtrType); ok {
		typ = underlyingType(ptr.Base)
	}

	switch t := typ.(type) {
	case *ast.BuiltinType:
		return ev.checkStringSliceExpr(x)
	case *ast.ArrayType:
		return ev.checkArraySliceExpr(x, t)
	case *ast.SliceType:
		return ev.checkSliceSliceExpr(x)
	default:
		panic("not reached")
	}
}

func (ev *exprVerifier) checkStringSliceExpr(x *ast.SliceExpr) (ast.Expr, error) {
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
	// Check array operand is addressable.
	if !isAddressable(x.X) {
		return nil, &NotAddressable{Off: x.X.Position(), File: ev.File}
	}
	return x, nil
}

func (ev *exprVerifier) checkSliceSliceExpr(x *ast.SliceExpr) (ast.Expr, error) {

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
	y, err := ev.checkExpr(x.X)
	if err != nil {
		return nil, err
	}

	c, ok := y.(*ast.ConstValue)
	if !ok {
		x.X = y
		return x, nil
	}

	var d *ast.ConstValue
	switch x.Op {
	case '+':
		return c, nil
	case '-':
		d, err = Minus(c)
	case '!':
		d, err = Not(c)
	case '^':
		d, err = Complement(c)
	default:
		panic("not reached")
	}
	if err != nil {
		return nil, &ErrorPos{Off: c.Off, File: ev.File, Err: err}
	}
	d.Off = x.Off
	return d, nil
}

func (ev *exprVerifier) VisitBinaryExpr(x *ast.BinaryExpr) (ast.Expr, error) {
	if x.Op == ast.SHL || x.Op == ast.SHR {
		return ev.checkShift(x)
	}

	u, err := ev.checkExpr(x.X)
	if err != nil {
		return nil, err
	}
	v, err := ev.checkExpr(x.Y)
	if err != nil {
		return nil, err
	}

	// Evaluate constant binary expressions.
	if u, ok := u.(*ast.ConstValue); ok {
		if v, ok := v.(*ast.ConstValue); ok {
			var c *ast.ConstValue
			var err error
			switch x.Op {
			case ast.LT, ast.GT, ast.EQ, ast.NE, ast.LE, ast.GE:
				c, err = Compare(u, v, x.Op)
			case '+':
				c, err = Add(u, v)
			case '-':
				c, err = Sub(u, v)
			case '*':
				c, err = Mul(u, v)
			case '/':
				c, err = Div(u, v)
			case '%':
				c, err = Rem(u, v)
			case '&', '|', '^', ast.ANDN:
				c, err = Bit(x.Op, u, v)
			case ast.AND, ast.OR:
				c, err = Logical(x.Op, u, v)
			default:
				panic("not reached")
			}
			if err != nil {
				return nil, &ErrorPos{Off: x.Off, File: ev.File, Err: err}
			}
			c.Off = x.Off
			return c, nil
		}
	}

	// Check compatibility of operand types. For comparison operators, one of
	// the operands must be assignable to the type of the other operand. For
	// other operators, the operand types must be identical, which is checked
	// during type inference.
	if x.Op.IsComparison() {
		if ok, err := ev.isAssignableType(u.Type(), v.Type()); err != nil {
			return nil, err
		} else if !ok {
			if ok, err := ev.isAssignableType(v.Type(), u.Type()); err != nil {
				return nil, err
			} else if !ok {
				return nil, &BadCompareOperands{
					Off: x.Off, File: ev.File, XType: u.Type(), YType: v.Type()}
			}
		}
	}

	// Check applicability of a comparison operation to the operand
	// types. Other operations are checked durting type inference.
	utyp, vtyp := underlyingType(u.Type()), underlyingType(v.Type())
	switch x.Op {
	case ast.EQ, ast.NE:
		// If one of the operands is `nil` and the other is not comparable to
		// `nil`, then the above check for assignability would have failed,
		// therefore no need to check it, except for the special case of
		// comparing `nil` to `nil`.
		if utyp == ast.BuiltinNilType && vtyp == ast.BuiltinNilType {
			return nil, &NotSupportedOperation{
				Off: x.Off, File: ev.File, Op: x.Op, Type: u.Type()}
			// Check equality comparison when one of the operands is of an
			// interface type and the other isn't.
		} else if _, ok := utyp.(*ast.InterfaceType); ok {
			if _, ok := vtyp.(*ast.InterfaceType); !ok && !isEqualityComparable(vtyp) {
				return nil, &NotComparable{
					Off: v.Position(), File: ev.File, Type: v.Type()}
			}
		} else if _, ok := vtyp.(*ast.InterfaceType); ok {
			if _, ok := utyp.(*ast.InterfaceType); !ok && !isEqualityComparable(utyp) {
				return nil, &NotComparable{
					Off: u.Position(), File: ev.File, Type: u.Type()}
			}
			// Comparability check with only one of the operands is
			// sufficient, following the above assignability check.
		} else if !isEqualityComparable(utyp) {
			return nil, &NotComparable{Off: x.Off, File: ev.File, Type: u.Type()}
		}
	case '<', '>', ast.LE, ast.GE:
		// Ordering check with only one of the operands is sufficient,
		// following the assignability check above.
		if !isOrdered(utyp) {
			return nil, &NotOrdered{Off: u.Position(), File: ev.File, Type: u.Type()}
		}
	}

	x.X = u
	x.Y = v
	return x, nil
}

func (ev *exprVerifier) checkShift(x *ast.BinaryExpr) (ast.Expr, error) {
	u, err := ev.checkExpr(x.X)
	if err != nil {
		return nil, err
	}
	v, err := ev.checkExpr(x.Y)
	if err != nil {
		return nil, err
	}

	// Evaluate a constant shift expression.
	if u, ok := u.(*ast.ConstValue); ok {
		if s, ok := v.(*ast.ConstValue); ok {
			c, err := Shift(u, s, x.Op)
			if err != nil {
				return nil, &ErrorPos{Off: x.Off, File: ev.File, Err: err}
			}
			return c, nil
		}
	}

	// The right operand should have unsigned integer type.  The left operand
	// should have integer type, which is checked in during type inference.
	if t := builtinType(v.Type()); t == nil || !t.IsInteger() || t.IsSigned() {
		return nil, &BadShiftCount{Off: v.Position(), File: ev.File}
	}
	x.X = u
	x.Y = v
	return x, nil
}

// Returns the length of the array type T.
func (ev *exprVerifier) getArrayLength(t *ast.ArrayType) (int64, error) {
	var err error
	c, ok := t.Dim.(*ast.ConstValue)
	if !ok || isUntyped(c.Typ) {
		if c, err = ev.checkArrayLength(t); err != nil {
			return 0, err
		}
	}
	return int64(c.Value.(ast.Int)), nil
}

// Returns true if types S and T are identical.
// https://golang.org/ref/spec#Type_identity
func (ev *exprVerifier) identicalTypes(s ast.Type, t ast.Type) (bool, error) {
	if s == t {
		// "Two named types are identical if their type names originate in the
		// same TypeSpec". Also fastpath for comparing a type to itself.
		return true, nil
	}

	switch s := s.(type) {
	case *ast.TypeDecl:
		// T is either an unnamed type, or does not originate from the same
		// TypeSpec as S.
		return false, nil
	case *ast.BuiltinType:
		if t, ok := t.(*ast.BuiltinType); ok {
			return s.Kind == t.Kind, nil
		}
		return false, nil
	case *ast.ArrayType:
		if t, ok := t.(*ast.ArrayType); ok {
			ns, err := ev.getArrayLength(s)
			if err != nil {
				return false, err
			}
			nt, err := ev.getArrayLength(t)
			if err != nil {
				return false, err
			}
			if ns == nt {
				return ev.identicalTypes(s.Elt, t.Elt)
			}
		}
		return false, nil
	case *ast.SliceType:
		if t, ok := t.(*ast.SliceType); ok {
			return ev.identicalTypes(s.Elt, t.Elt)
		}
		return false, nil
	case *ast.PtrType:
		if t, ok := t.(*ast.PtrType); ok {
			return ev.identicalTypes(s.Base, t.Base)
		}
		return false, nil
	case *ast.MapType:
		if t, ok := t.(*ast.MapType); ok {
			if ok, err := ev.identicalTypes(s.Key, t.Key); err != nil || !ok {
				return false, err
			}
			return ev.identicalTypes(s.Elt, t.Elt)
		}
		return false, nil
	case *ast.ChanType:
		if t, ok := t.(*ast.ChanType); ok {
			if s.Send == t.Send && s.Recv == t.Recv {
				return ev.identicalTypes(s.Elt, t.Elt)
			}
		}
		return false, nil
	case *ast.StructType:
		if t, ok := t.(*ast.StructType); !ok {
			return false, nil
		} else if len(s.Fields) != len(t.Fields) {
			return false, nil
		} else {
			for i := range s.Fields {
				sf, tf := &s.Fields[i], &t.Fields[i]
				// "corresponding fields [must] have the same names, [...],
				// and identical tags. Two anonymous fields are considered to
				// have the same name."
				if sf.Name != tf.Name || sf.Tag != tf.Tag {
					return false, nil
				}
				// "Lower-case field names from different packages are always
				// different."
				if s.File == nil || t.File == nil || s.File.Pkg != t.File.Pkg &&
					!ast.IsExported(sf.Name) {
					return false, nil
				}
				// "corresponding fields have [...] identical types"
				if ok, err := ev.identicalTypes(sf.Type, tf.Type); err != nil || !ok {
					return false, err
				}
			}
			return true, nil
		}
	case *ast.FuncType:
		if t, ok := t.(*ast.FuncType); !ok {
			return false, nil
		} else if len(s.Params) != len(t.Params) || len(s.Returns) != len(t.Returns) ||
			s.Var != t.Var {
			// "the same number of parameters and result values, [...], and
			// either both functions are variadic or neither is."
			return false, nil
		} else {
			// "corresponding parameter and result types are identical"
			for i := range s.Params {
				ok, err := ev.identicalTypes(s.Params[i].Type, t.Params[i].Type)
				if err != nil || !ok {
					return false, err
				}
			}
			for i := range s.Returns {
				ok, err := ev.identicalTypes(s.Returns[i].Type, t.Returns[i].Type)
				if err != nil || !ok {
					return false, err
				}
			}
			return true, nil
		}
	case *ast.InterfaceType:
		if t, ok := t.(*ast.InterfaceType); !ok {
			return false, nil
		} else {
			// "Interface types are identical if they have the same set of methods ...
			ss, st := ifaceMethodSet(s), ifaceMethodSet(t)
			if len(ss) != len(st) {
				return false, nil
			}
			for name, ms := range ss {
				mt := st[name]
				// "... with the same name
				if mt == nil {
					return false, nil
				}
				// "Lower-case method names from different packages are always
				// different."
				if ms.File == nil || mt.File == nil || ms.File.Pkg != mt.File.Pkg &&
					!ast.IsExported(ms.Name) {
					return false, nil
				}
				//  "... "and identical function types."
				ok, err := ev.identicalTypes(ms.Func.Sig, mt.Func.Sig)
				if err != nil || !ok {
					return false, err
				}
			}
			return true, nil
		}
	default:
		panic("not reached")
	}
}

// Returns true if TYP implements IFC.
func (ev *exprVerifier) implements(typ ast.Type, ifc *ast.InterfaceType) (bool, error) {
	// Get the receiver base type.
	ptr := false
	if p, ok := typ.(*ast.PtrType); ok {
		ptr = true
		typ = p.Base
	}
	dcl, ok := typ.(*ast.TypeDecl)
	if !ok {
		// If the given type is not a (pointer to a) type name, the only
		// implemented interface is `interface{}`.
		return isEmptyInterfaceType(ifc), nil
	}

	if impl, ok := dcl.Type.(*ast.InterfaceType); ok {
		if ptr {
			return false, nil
		}
		return ev.ifaceImplements(impl, ifc)
	} else {
		return ev.typenameImplements(ptr, dcl, ifc)
	}
}

func (ev *exprVerifier) ifaceImplements(
	impl *ast.InterfaceType, ifc *ast.InterfaceType) (bool, error) {

	implSet := ifaceMethodSet(impl)
	for _, fn := range ifaceMethodSet(ifc) {
		m := implSet[fn.Name]
		if m == nil {
			return false, nil
		}
		if ok, err := ev.identicalTypes(m.Func.Sig, fn.Func.Sig); err != nil || !ok {
			return false, err
		}
	}
	return true, nil
}

func (ev *exprVerifier) typenameImplements(
	ptr bool, dcl *ast.TypeDecl, ifc *ast.InterfaceType) (bool, error) {

	for _, fn := range ifaceMethodSet(ifc) {
		// Look for an implementation among the methods with value receivers.
		i := 0
		for ; i < len(dcl.Methods); i++ {
			m := dcl.Methods[i]
			if fn.Name == m.Name {
				ok, err := ev.identicalTypes(fn.Func.Sig, m.Func.Sig)
				if err != nil || !ok {
					return false, err
				}
				break
			}
		}
		// Do not consider pointer receiver methods if the original type
		// wasn't a pointer type.
		if !ptr {
			if i == len(dcl.Methods) {
				// No method name matches.
				return false, nil
			}
			continue
		}
		// Look for an implementation among the methods with pointer
		// receivers.
		for i = 0; i < len(dcl.PMethods); i++ {
			m := dcl.PMethods[i]
			if fn.Name == m.Name {
				ok, err := ev.identicalTypes(fn.Func.Sig, m.Func.Sig)
				if err != nil || !ok {
					return false, err
				}
				break
			}
		}
		if i == len(dcl.PMethods) {
			// No method name matches.
			return false, nil
		}
	}
	return true, nil
}

// Checks whether the expression X is assignable to a variable of type DST. If
// the expression is an untyped constant, returns it after converson to DST,
// otherwise returns it unmodified.  Returns an error if the expression is not
// assignable to DST, or if it is an untyped constant, which cannot be
// converted to DST.
func (ev *exprVerifier) isAssignable(dst ast.Type, x ast.Expr) (ast.Expr, error) {
	src := x.Type()
	if c, ok := x.(*ast.ConstValue); ok && isUntyped(src) {
		// "x is an untyped constant representable by a value of type T."
		// Every constant is representable by a value of type ``interface{}`.
		if isEmptyInterfaceType(dst) {
			return x, nil
		}
		c, err := Convert(dst, c)
		if err != nil {
			return nil, &ErrorPos{Off: x.Position(), File: ev.File, Err: err}
		}
		c.Off = x.Position()
		return c, nil
	}

	if ok, err := ev.isAssignableType(dst, src); err != nil {
		return nil, err
	} else if !ok {
		return nil, &NotAssignable{
			Off: x.Position(), File: ev.File, DType: dst, SType: x.Type()}
	}
	return x, nil
}

// Returns true, if a value of type SRC is assignable to a variable of type DST.
func (ev *exprVerifier) isAssignableType(dst ast.Type, src ast.Type) (bool, error) {
	// "x's type is identical to T."
	if ok, err := ev.identicalTypes(dst, src); err != nil || ok {
		return ok, err
	}

	usrc, udst := underlyingType(src), underlyingType(dst)
	// Allow untyped boolean to be assigned to any type, which has `bool` as
	// its underlying type.
	if udst == ast.BuiltinBool && src == ast.BuiltinUntypedBool {
		return true, nil
	}

	// "x's type V and T have identical underlying types and at least one of V
	// or T is not a named type."
	if !isNamed(src) || !isNamed(dst) {
		if ok, err := ev.identicalTypes(usrc, udst); err != nil || ok {
			return ok, err
		}
	}
	// "T is an interface type and x implements T.
	if ifc, ok := udst.(*ast.InterfaceType); ok {
		if ok, err := ev.implements(src, ifc); err != nil || ok {
			return ok, err
		}
	}
	// "x is a bidirectional channel value, T is a channel type, x's type V
	// and T have identical element types, and at least one of V or T is not a
	// named type.
	if ch, ok := usrc.(*ast.ChanType); ok && ch.Send && ch.Recv {
		if dch, ok := udst.(*ast.ChanType); ok && (!isNamed(src) || !isNamed(dst)) {
			return ev.identicalTypes(ch.Elt, dch.Elt)
		}
	}
	// "x is the predeclared identifier nil and T is a pointer, function,
	// slice, map, channel, or interface type."
	if usrc == ast.BuiltinNilType {
		switch udst.(type) {
		case *ast.PtrType, *ast.FuncType, *ast.SliceType, *ast.MapType, *ast.ChanType,
			*ast.InterfaceType:
			return true, nil
		}
	}
	return false, nil
}

// Checks whether a conversion of a non-constant expression is allowed.
// https://golang.org/ref/spec#Conversions
func (ev *exprVerifier) isConvertible(typ ast.Type, x ast.Expr) (bool, error) {
	// "x is assignable to T."
	if _, err := ev.isAssignable(typ, x); err == nil {
		return true, nil
	}

	// "x's type and T have identical underlying types."
	ut, ux := underlyingType(typ), underlyingType(x.Type())
	if ok, err := ev.identicalTypes(ut, ux); err != nil {
		return false, err
	} else if ok {
		return true, nil
	}

	// "x's type and T are unnamed pointer types and their pointer base types"
	// "have identical underlying types."
	if pt, ok := typ.(*ast.PtrType); ok {
		if px, ok := x.Type().(*ast.PtrType); ok {
			t, tx := underlyingType(pt.Base), underlyingType(px.Base)
			return ev.identicalTypes(t, tx)
		}
		return false, nil
	}

	// "x is an integer or a slice of bytes or runes and T is a string type."
	if ut == ast.BuiltinString {
		if s, ok := ux.(*ast.SliceType); ok {
			e := builtinType(s.Elt)
			return e == ast.BuiltinUint8 || e == ast.BuiltinInt32, nil
		} else {
			i := builtinType(ux)
			return i != nil && i.IsInteger(), nil
		}
	}

	// "x is a string and T is a slice of bytes or runes."
	if ux == ast.BuiltinString {
		if s, ok := ut.(*ast.SliceType); ok {
			e := builtinType(s.Elt)
			return e == ast.BuiltinUint8 || e == ast.BuiltinInt32, nil
		}
		return false, nil
	}

	// Rest of allowed conversions are all between builtin types.
	bt, bx := builtinType(ut), builtinType(ux)
	if bt == nil || bx == nil {
		return false, nil
	}

	// "x's type and T are both integer ..."
	return bt.IsInteger() && bx.IsInteger() ||
		// " ...or floating point types."
		bt.IsFloat() && bx.IsFloat() ||
		// "x's type and T are both complex types."
		bt.IsComplex() && bx.IsComplex(), nil
}
