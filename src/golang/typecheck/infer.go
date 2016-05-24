package typecheck

import "golang/ast"

type typeInferer struct {
	Pkg     *ast.Package
	File    *ast.File
	Files   []*ast.File
	Syms    []ast.Symbol
	TypeCtx ast.Type
	Done    map[ast.Symbol]struct{}
	Delay   []delayInfer
}

type delayInfer struct {
	File *ast.File
	Fn   func() error
}

func newInferer(p *ast.Package) *typeInferer {
	return &typeInferer{Pkg: p, Done: make(map[ast.Symbol]struct{})}
}

func inferTypes(pkg *ast.Package) error {
	ti := newInferer(pkg)

	for _, s := range pkg.Syms {
		var err error
		switch d := s.(type) {
		case *ast.TypeDecl:
			err = ti.inferTypeDecl(d)
		case *ast.Const:
			err = ti.inferConstDecl(d)
		case *ast.Var:
			err = ti.inferVarDecl(d)
		case *ast.FuncDecl:
			err = ti.inferFuncDecl(d)
		default:
			panic("not reached")
		}
		if err != nil {
			return err
		}
	}

	for n := len(ti.Delay); n > 0; n = len(ti.Delay) {
		d := ti.Delay[n-1]
		ti.Delay = ti.Delay[:n-1]

		ti.beginFile(d.File)
		err := d.Fn()
		ti.endFile()
		if err != nil {
			return err
		}
	}
	return nil
}

func (ti *typeInferer) delay(fn func() error) {
	ti.Delay = append(ti.Delay, delayInfer{File: ti.File, Fn: fn})
}

func (ti *typeInferer) inferTypeDecl(d *ast.TypeDecl) error {
	// No need to check things in a different package
	if d.File == nil || d.File.Pkg != ti.Pkg {
		return nil
	}
	// Check type.
	if err := ti.inferType(d.Type); err != nil {
		return err
	}
	// Check method declarations.
	for _, m := range d.Methods {
		if err := ti.inferFuncDecl(m); err != nil {
			return err
		}
	}
	for _, m := range d.PMethods {
		if err := ti.inferFuncDecl(m); err != nil {
			return err
		}
	}
	return nil
}

func (ti *typeInferer) inferConstDecl(c *ast.Const) error {
	// No need to check things in a different package
	if c.File == nil || c.File.Pkg != ti.Pkg {
		return nil
	}

	// Check for a type inference loop.
	if l := ti.checkLoop(c); l != nil {
		return &TypeInferLoop{Loop: l}
	}

	if _, ok := ti.Done[c]; ok {
		return nil
	}
	ti.Done[c] = struct{}{}

	ti.beginCheck(c, c.File)
	defer func() { ti.endCheck() }()

	x, err := ti.inferExpr(c.Init, c.Type)
	if err != nil {
		return err
	}
	c.Init = x
	if t := builtinType(x.Type()); t == nil {
		return &BadConstType{Off: c.Off, File: c.File, Type: x.Type()}
	} else if t == ast.BuiltinNilType {
		return &NotConst{Off: x.Position(), File: c.File, What: "`nil`"}
	}
	if c.Type == nil {
		c.Type = x.Type()
	}
	return nil
}

func (ti *typeInferer) inferVarDecl(v *ast.Var) error {
	// No need to check things in a different package
	if v.File == nil || v.File.Pkg != ti.Pkg {
		return nil
	}

	// Check for a type inference loop.
	if l := ti.checkLoop(v); l != nil {
		return &TypeInferLoop{Loop: l}
	}

	if _, ok := ti.Done[v]; ok {
		return nil
	}
	ti.Done[v] = struct{}{}

	ti.beginFile(v.File)
	defer func() { ti.endFile() }()

	if v.Type != nil {
		if err := ti.inferType(v.Type); err != nil {
			return err
		}
	}

	if v.Init == nil || len(v.Init.RHS) == 0 {
		return nil
	}

	ctx := v.Type
	if ctx == nil {
		ctx = ast.BuiltinDefault
	}
	if len(v.Init.LHS) > 1 && len(v.Init.RHS) == 1 {
		// If there is a single expression on the RHS, make all the variables
		// simultaneously depend on it. FIXME: multi-dependencies.
		for i := range v.Init.LHS {
			op := v.Init.LHS[i].(*ast.OperandName)
			ti.beginCheck(op.Decl, v.File)
		}
		x, err := ti.inferMultiValueExpr(v.Init.RHS[0], ctx)
		for range v.Init.LHS { // FIXME: don't loop
			ti.endCheck()
		}
		if err != nil {
			return err
		}
		v.Init.RHS[0] = x
		if v.Type == nil {
			// Assign types to the variables on the LHS.
			tp, ok := x.Type().(*ast.TupleType)
			if !ok || len(tp.Types) != len(v.Init.LHS) {
				return &BadMultiValueAssign{Off: x.Position(), File: v.File}
			}
			for i := range v.Init.LHS {
				op := v.Init.LHS[i].(*ast.OperandName)
				v := op.Decl.(*ast.Var)
				if t := tp.Types[i]; t == ast.BuiltinUntypedBool {
					v.Type = ast.BuiltinBool
				} else {
					v.Type = t
				}
				ti.Done[v] = struct{}{}
			}
		}
		return nil
	}

	// If there are multiple expressions on the RHS, they must be the same
	// number as the variables on the LHS.
	if len(v.Init.LHS) != len(v.Init.RHS) {
		return &BadMultiValueAssign{Off: v.Init.Off, File: v.File}
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

	// Infer the type of the initializer expression.
	ti.beginCheck(v, v.File)
	x, err := ti.inferExpr(v.Init.RHS[i], ctx)
	ti.endCheck()
	if err != nil {
		return err
	}
	v.Init.RHS[i] = x
	if v.Type == nil {
		// If the initializer expression is `nil`, the variable must be
		// declared with a type.
		if x.Type() == ast.BuiltinNilType {
			return &NilUse{Off: x.Position(), File: v.File}
		}
		v.Type = defaultType(x.Type())
	}
	return nil
}

func (ti *typeInferer) inferFuncDecl(fn *ast.FuncDecl) error {
	// No need to check things in a different package
	if fn.File == nil || fn.File.Pkg != ti.Pkg {
		return nil
	}

	ti.beginFile(fn.File)
	defer func() { ti.endFile() }()

	// Descend into the receiver type.
	if fn.Func.Recv != nil {
		if err := ti.inferType(fn.Func.Recv.Type); err != nil {
			return err
		}
	}
	// Descend into the signature.
	if err := ti.inferType(fn.Func.Sig); err != nil {
		return err
	}
	return nil
}

func (ti *typeInferer) inferType(t ast.Type) error {
	_, err := t.TraverseType(ti)
	return err
}

func (ti *typeInferer) beginFile(f *ast.File) {
	ti.Files = append(ti.Files, ti.File)
	ti.File = f
}

func (ti *typeInferer) endFile() {
	n := len(ti.Files)
	ti.File = ti.Files[n-1]
	ti.Files = ti.Files[:n-1]
}

func (ti *typeInferer) beginCheck(s ast.Symbol, f *ast.File) {
	ti.beginFile(f)
	ti.Syms = append(ti.Syms, s)
}

func (ti *typeInferer) endCheck() {
	ti.Syms = ti.Syms[:len(ti.Syms)-1]
	ti.endFile()
}

func (ti *typeInferer) checkLoop(sym ast.Symbol) []ast.Symbol {
	for i := len(ti.Syms) - 1; i >= 0; i-- {
		s := ti.Syms[i]
		if s == nil {
			return nil
		}
		if s == sym {
			return ti.Syms[i:]
		}
	}
	return nil
}

func (ti *typeInferer) inferMultiValueExpr(x ast.Expr, typ ast.Type) (ast.Expr, error) {
	typ, ti.TypeCtx = ti.TypeCtx, typ
	x, err := x.TraverseExpr(ti)
	typ, ti.TypeCtx = ti.TypeCtx, typ
	return x, err
}

func (ti *typeInferer) inferExpr(x ast.Expr, typ ast.Type) (ast.Expr, error) {
	x, err := ti.inferMultiValueExpr(x, typ)
	if err != nil {
		return nil, err
	}
	return x, ti.forceSingleValue(x)
}

func (ti *typeInferer) forceSingleValue(x ast.Expr) error {
	var t ast.Type
	if t = x.Type(); t == nil {
		return nil
	}
	tp, ok := t.(*ast.TupleType)
	if !ok {
		return nil
	}
	if tp.Strict {
		return &SingleValueContext{Off: x.Position(), File: ti.File}
	}
	t = tp.Types[0]
	switch x := x.(type) {
	case *ast.TypeAssertion:
		x.Typ = t
	case *ast.IndexExpr:
		x.Typ = t
	case *ast.UnaryExpr:
		x.Typ = t
	default:
		panic("not reached")
	}
	return nil
}

func (*typeInferer) VisitError(*ast.Error) (*ast.Error, error) {
	panic("not reached")
}

func (*typeInferer) VisitTypeName(*ast.QualifiedId) (ast.Type, error) {
	panic("not reached")
}

func (ti *typeInferer) VisitTypeDeclType(t *ast.TypeDecl) (ast.Type, error) {
	return t, nil
}

func (*typeInferer) VisitBuiltinType(t *ast.BuiltinType) (ast.Type, error) {
	return t, nil
}

func (ti *typeInferer) VisitArrayType(t *ast.ArrayType) (ast.Type, error) {
	if err := ti.inferType(t.Elt); err != nil {
		return nil, err
	}
	if t.Dim != nil {
		x, err := ti.inferExpr(t.Dim, nil)
		if err != nil {
			return nil, err
		}
		t.Dim = x
	}
	return t, nil
}

func (ti *typeInferer) VisitSliceType(t *ast.SliceType) (ast.Type, error) {
	if err := ti.inferType(t.Elt); err != nil {
		return nil, err
	}
	return t, nil
}

func (ti *typeInferer) VisitPtrType(t *ast.PtrType) (ast.Type, error) {
	if err := ti.inferType(t.Base); err != nil {
		return nil, err
	}
	return t, nil
}

func (ti *typeInferer) VisitMapType(t *ast.MapType) (ast.Type, error) {
	if err := ti.inferType(t.Key); err != nil {
		return nil, err
	}
	if err := ti.inferType(t.Elt); err != nil {
		return nil, err
	}
	return t, nil
}

func (ti *typeInferer) VisitChanType(t *ast.ChanType) (ast.Type, error) {
	if err := ti.inferType(t.Elt); err != nil {
		return nil, err
	}
	return t, nil
}

func (ti *typeInferer) VisitStructType(t *ast.StructType) (ast.Type, error) {
	for i := range t.Fields {
		fd := &t.Fields[i]
		if err := ti.inferType(fd.Type); err != nil {
			return nil, err
		}
	}
	return t, nil
}

func (*typeInferer) VisitTupleType(*ast.TupleType) (ast.Type, error) {
	panic("not reached")
}

func (ti *typeInferer) VisitFuncType(t *ast.FuncType) (ast.Type, error) {
	for i := range t.Params {
		p := &t.Params[i]
		if err := ti.inferType(p.Type); err != nil {
			return nil, err
		}
	}
	for i := range t.Returns {
		p := &t.Returns[i]
		if err := ti.inferType(p.Type); err != nil {
			return nil, err
		}
	}
	return t, nil
}

func (ti *typeInferer) VisitInterfaceType(t *ast.InterfaceType) (ast.Type, error) {
	for _, m := range t.Methods {
		if err := ti.inferType(m.Func.Sig); err != nil {
			return nil, err
		}
	}
	return t, nil
}

func (*typeInferer) VisitQualifiedId(*ast.QualifiedId) (ast.Expr, error) {
	panic("not reached")
}

func (ti *typeInferer) VisitConstValue(x *ast.ConstValue) (ast.Expr, error) {
	src := builtinType(x.Typ)
	if !src.IsUntyped() || ti.TypeCtx == nil {
		return x, nil
	}
	dst := ti.TypeCtx
	if dst == ast.BuiltinDefault {
		dst = defaultType(x.Type())
	}
	if t := builtinType(dst); t == nil {
		return nil, &BadConstType{Off: x.Off, File: ti.File, Type: dst}
	}
	c, err := Convert(dst, x)
	if err != nil {
		return nil, &ErrorPos{Off: x.Off, File: ti.File, Err: err}
	}
	c.Off = x.Off
	return c, nil
}

func (ti *typeInferer) VisitCompLiteral(x *ast.CompLiteral) (ast.Expr, error) {
	if x.Typ == nil {
		return nil, &MissingLiteralType{x.Off, ti.File}
	}
	if err := ti.inferType(x.Typ); err != nil {
		return nil, err
	}
	ti.delay(func() error { return ti.inferCompLiteralElts(x) })
	return x, nil
}

func (ti *typeInferer) inferCompLiteralElts(x *ast.CompLiteral) error {
	switch t := underlyingType(x.Typ).(type) {
	case *ast.ArrayType:
		return ti.inferArrayOrSliceLiteralElts(t.Elt, x)
	case *ast.SliceType:
		return ti.inferArrayOrSliceLiteralElts(t.Elt, x)
	case *ast.MapType:
		return ti.inferMapLiteralElts(t, x)
	case *ast.StructType:
		return ti.inferStructLiteralElts(t, x)
	default:
		return &BadLiteralType{Off: x.Off, File: ti.File, Type: t}
	}
}

func (ti *typeInferer) inferArrayOrSliceLiteralElts(
	etyp ast.Type, x *ast.CompLiteral) error {

	for _, elt := range x.Elts {
		if elt.Key != nil {
			k, err := ti.inferExpr(elt.Key, nil)
			if err != nil {
				return err
			}
			elt.Key = k
		}
		if c, ok := elt.Elt.(*ast.CompLiteral); ok {
			// Assign type to elements, which are composite literals with elided
			// type.
			if c.Typ == nil {
				c.Typ = etyp
			}
		}

		e, err := ti.inferExpr(elt.Elt, etyp)
		if err != nil {
			return err
		}
		elt.Elt = e
	}
	return nil
}

func (ti *typeInferer) inferMapLiteralElts(typ *ast.MapType, x *ast.CompLiteral) error {
	for _, elt := range x.Elts {
		if elt.Key != nil {
			if c, ok := elt.Key.(*ast.CompLiteral); ok {
				// Assign types to keys and elements, which are composite
				// literals with elided type.
				if c.Typ == nil {
					c.Typ = typ.Key
				}
			}
			k, err := ti.inferExpr(elt.Key, typ.Key)
			if err != nil {
				return err
			}
			elt.Key = k
		}
		if c, ok := elt.Elt.(*ast.CompLiteral); ok {
			if c.Typ == nil {
				c.Typ = typ.Elt
			}
		}
		e, err := ti.inferExpr(elt.Elt, typ.Elt)
		if err != nil {
			return err
		}
		elt.Elt = e
	}
	return nil
}

func (ti *typeInferer) inferStructLiteralElts(
	str *ast.StructType, x *ast.CompLiteral) error {

	if len(x.Elts) == 0 {
		return nil
	}

	if x.Elts[0].Key == nil {
		i, nf, ne := 0, len(str.Fields), len(x.Elts)
		for i < nf && i < ne {
			if x.Elts[i].Key != nil {
				return &MixedStructLiteral{Off: x.Off, File: ti.File}
			}
			y, err := ti.inferExpr(x.Elts[i].Elt, str.Fields[i].Type)
			if err != nil {
				return err
			}
			x.Elts[i].Elt = y
			i++
		}
		if nf != ne {
			return &FieldEltMismatch{Off: x.Off, File: ti.File}
		}
	} else {
		for _, elt := range x.Elts {
			if elt.Key == nil {
				return &MixedStructLiteral{Off: x.Off, File: ti.File}
			}
			id, ok := elt.Key.(*ast.QualifiedId)
			if !ok || len(id.Pkg) > 0 || id.Id == "_" {
				return &NotField{Off: elt.Key.Position(), File: ti.File}
			}
			f := findField(str, id.Id)
			if f == nil {
				return &NotFound{Off: id.Off, File: ti.File, What: "field", Name: id.Id}
			}
			y, err := ti.inferExpr(elt.Elt, f.Type)
			if err != nil {
				return err
			}
			elt.Elt = y
		}
	}
	return nil
}

func (ti *typeInferer) VisitOperandName(x *ast.OperandName) (ast.Expr, error) {
	switch d := x.Decl.(type) {
	case *ast.Var:
		if d.Type == nil {
			if err := ti.inferVarDecl(d); err != nil {
				return nil, err
			}
		}
		x.Typ = d.Type
	case *ast.Const:
		switch d.Init {
		case ast.BuiltinTrue, ast.BuiltinFalse:
			x.Typ = ast.BuiltinUntypedBool
		case ast.BuiltinIota:
			x.Typ = ast.BuiltinUntypedInt
		default:
			if d.Type == nil {
				if err := ti.inferConstDecl(d); err != nil {
					return nil, err
				}
			}
			x.Typ = d.Type
		}
		if src := builtinType(x.Typ); src.IsUntyped() && ti.TypeCtx != nil {
			dst := ti.TypeCtx
			if dst == ast.BuiltinDefault {
				dst = defaultType(x.Type())
			}
			if t := builtinType(dst); t == nil {
				return nil, &BadConstType{Off: x.Off, File: ti.File, Type: dst}
			}
			x.Typ = dst
		}
	case *ast.FuncDecl:
		x.Typ = d.Func.Sig
	default:
		panic("not reached")
	}
	if x.Typ == nil {
		panic("internal error")
	}
	return x, nil
}

func (ti *typeInferer) VisitCall(x *ast.Call) (ast.Expr, error) {
	// Check if we have a builtin function call.
	if op, ok := x.Func.(*ast.OperandName); ok {
		if d, ok := op.Decl.(*ast.FuncDecl); ok {
			switch d {
			case ast.BuiltinAppend:
				return ti.visitBuiltinAppend(x)
			case ast.BuiltinCap:
				return ti.visitBuiltinCap(x)
			case ast.BuiltinClose:
				return ti.visitBuiltinClose(x)
			case ast.BuiltinComplex:
				return ti.visitBuiltinComplex(x)
			case ast.BuiltinCopy:
				return ti.visitBuiltinCopy(x)
			case ast.BuiltinDelete:
				return ti.visitBuiltinDelete(x)
			case ast.BuiltinImag:
				return ti.visitBuiltinImag(x)
			case ast.BuiltinLen:
				return ti.visitBuiltinLen(x)
			case ast.BuiltinMake:
				return ti.visitBuiltinMake(x)
			case ast.BuiltinNew:
				return ti.visitBuiltinNew(x)
			case ast.BuiltinPanic:
				return ti.visitBuiltinPanic(x)
			case ast.BuiltinPrint:
				return ti.visitBuiltinPrint(x)
			case ast.BuiltinPrintln:
				return ti.visitBuiltinPrintln(x)
			case ast.BuiltinReal:
				return ti.visitBuiltinReal(x)
			case ast.BuiltinRecover:
				return ti.visitBuiltinRecover(x)
			}
		}
	}
	// Not a builtin call. Check the called expression.
	fn, err := ti.inferExpr(x.Func, nil)
	if err != nil {
		return nil, err
	}
	x.Func = fn
	ftyp, ok := fn.Type().(*ast.FuncType)
	if !ok {
		return nil, &NotFunc{Off: x.Off, File: ti.File, X: x}
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
		x.Typ = &ast.TupleType{Off: x.Off, Strict: true, Types: tp}
	}

	// Delay inferring the types of the argument expressions.
	ti.delay(func() error { return ti.inferArgs(x) })
	return x, nil
}

func (ti *typeInferer) inferArgs(x *ast.Call) error {
	// Check the special case `f(g(arguments-of-g))`.
	if len(x.Xs) == 1 && !x.Ell {
		if _, ok := x.Xs[0].(*ast.Call); ok {
			y, err := ti.inferMultiValueExpr(x.Xs[0], nil)
			if err != nil {
				return err
			}
			x.Xs[0] = y
			return nil
		}
	}

	// We need to match arguments to parameters, as parameter type might need
	// to provide type context for the argument expression.

	// Figure out the number of fixed (non-variadic) parameters.
	ftyp := x.Func.Type().(*ast.FuncType)
	if x.Ell && !ftyp.Var {
		return &BadVariadicCall{Off: x.Off, File: ti.File}
	}
	nparm, narg := len(ftyp.Params), len(x.Xs)
	nfix := nparm
	if (x.Ell || !ftyp.Var) && narg != nfix {
		// In call expressions using `...` or in calls to non-variadic
		// functions, the number of parameters and the number of arguments
		// must be the same.
		return &BadArgNumber{Off: x.Off, File: ti.File}
	}
	if ftyp.Var {
		nfix--
	}
	if x.Ell {
		narg--
	}
	if narg < nfix {
		// Not enough arguments for fixed parameters.
		return &BadArgNumber{Off: x.Off, File: ti.File}
	}

	// Infer types of the fixed argument expressions.
	for i := 0; i < nfix; i++ {
		y, err := ti.inferExpr(x.Xs[0], ftyp.Params[i].Type)
		if err != nil {
			return err
		}
		x.Xs[i] = y
	}

	// An argument expression which uses the `...` notation, cannot be of a
	// type that needs a type context.
	if x.Ell {
		y, err := ti.inferExpr(x.Xs[narg], nil)
		if err != nil {
			return err
		}
		x.Xs[narg] = y
	} else {
		// Infer the variadic arguments with the type context of the variadic
		// parameter type.
		for i := nfix; i < narg; i++ {
			y, err := ti.inferExpr(x.Xs[i], ftyp.Params[nfix].Type)
			if err != nil {
				return err
			}
			x.Xs[i] = y
		}
	}

	return nil
}

func (*typeInferer) visitBuiltinAppend(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*typeInferer) visitBuiltinCap(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*typeInferer) visitBuiltinClose(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*typeInferer) visitBuiltinComplex(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*typeInferer) visitBuiltinCopy(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*typeInferer) visitBuiltinDelete(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*typeInferer) visitBuiltinImag(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (ti *typeInferer) visitBuiltinLen(x *ast.Call) (ast.Expr, error) {
	x.Typ = ast.BuiltinInt

	ti.delay(func() error { return ti.inferBuiltinLenArg(x) })
	return x, nil
}

func (ti *typeInferer) inferBuiltinLenArg(x *ast.Call) error {
	if len(x.Xs) != 1 {
		return &BadArgNumber{Off: x.Off, File: ti.File}
	}
	y, err := ti.inferExpr(x.Xs[0], nil)
	if err != nil {
		return err
	}
	x.Xs[0] = y
	return nil
}

func (*typeInferer) visitBuiltinMake(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*typeInferer) visitBuiltinNew(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*typeInferer) visitBuiltinPanic(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*typeInferer) visitBuiltinPrint(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*typeInferer) visitBuiltinPrintln(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*typeInferer) visitBuiltinReal(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*typeInferer) visitBuiltinRecover(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (ti *typeInferer) VisitConversion(x *ast.Conversion) (ast.Expr, error) {
	if err := ti.inferType(x.Typ); err != nil {
		return nil, err
	}
	ti.delay(func() error { return ti.inferConversionArg(x) })
	return x, nil
}

func (ti *typeInferer) inferConversionArg(x *ast.Conversion) error {
	y, err := ti.inferExpr(x.X, x.Typ)
	if err != nil {
		return err
	}
	x.X = y
	return nil
}

func (ti *typeInferer) VisitMethodExpr(x *ast.MethodExpr) (ast.Expr, error) {
	if err := ti.inferType(x.RTyp); err != nil {
		return nil, err
	}
	// Find the method.
	m0, m1, vptr := findSelector(ti.File.Pkg, x.RTyp, x.Id)
	if (m0.M != nil || m0.F != nil) && (m1.M != nil || m1.F != nil) {
		return nil, &AmbiguousSelector{Off: x.Off, File: ti.File, Name: x.Id}
	}
	if m0.M == nil && m1.M == nil {
		return nil, &NotFound{Off: x.Off, File: ti.File, What: "method", Name: x.Id}
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
		return nil, &BadMethodExpr{Off: x.Off, File: ti.File}
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

func (ti *typeInferer) VisitParensExpr(*ast.ParensExpr) (ast.Expr, error) {
	panic("not reached")
}

func (ti *typeInferer) VisitFunc(x *ast.Func) (ast.Expr, error) {
	if err := ti.inferType(x.Sig); err != nil {
		return nil, err
	}
	return x, nil
}

func (ti *typeInferer) VisitTypeAssertion(x *ast.TypeAssertion) (ast.Expr, error) {
	if err := ti.inferType(x.ATyp); err != nil {
		return nil, err
	}
	x.Typ = &ast.TupleType{
		Off:    x.ATyp.Position(),
		Strict: false,
		Types:  []ast.Type{x.ATyp, ast.BuiltinUntypedBool},
	}
	ti.delay(func() error { return ti.inferTypeAssertionOperand(x) })
	return x, nil
}

func (ti *typeInferer) inferTypeAssertionOperand(x *ast.TypeAssertion) error {
	y, err := ti.inferExpr(x.X, nil)
	if err != nil {
		return err
	}
	x.X = y
	return nil
}

func (ti *typeInferer) VisitSelector(x *ast.Selector) (ast.Expr, error) {
	y, err := ti.inferExpr(x.X, nil)
	if err != nil {
		return nil, err
	}
	x.X = y
	s0, s1, _ := findSelector(ti.File.Pkg, x.X.Type(), x.Id)
	if s1.M != nil || s1.F != nil {
		return nil, &AmbiguousSelector{Off: x.Position(), File: ti.File, Name: x.Id}
	}
	if s0.M == nil && s0.F == nil {
		return nil, &NotFound{
			Off: x.Off, File: ti.File, What: "field or method", Name: x.Id}
	}
	if s0.M != nil {
		x.Typ = s0.M.Func.Sig
	} else {
		x.Typ = s0.F.Type
	}
	return x, nil
}

func (ti *typeInferer) VisitIndexExpr(x *ast.IndexExpr) (ast.Expr, error) {
	y, err := ti.inferExpr(x.X, nil)
	if err != nil {
		return nil, err
	}
	x.X = y
	ti.delay(func() error { return ti.inferIndexExprIndex(x) })

	// Get the type of the indexed expression.  If the type of the indexed
	// expression is a pointer, the pointed to type must be an array type.
	typ := underlyingType(x.X.Type())
	if ptr, ok := typ.(*ast.PtrType); ok {
		b, ok := underlyingType(ptr.Base).(*ast.ArrayType)
		if !ok {
			return nil, &BadIndexedType{Off: x.Off, File: ti.File, Type: typ}
		}
		typ = b
	}

	switch t := typ.(type) {
	case *ast.BuiltinType:
		if t.Kind != ast.BUILTIN_STRING && t.Kind != ast.BUILTIN_UNTYPED_STRING {
			return nil, &BadIndexedType{Off: x.Off, File: ti.File, Type: typ}
		}
		x.Typ = ast.BuiltinUint8
	case *ast.ArrayType:
		x.Typ = t.Elt
	case *ast.SliceType:
		x.Typ = t.Elt
	case *ast.MapType:
		x.Typ = &ast.TupleType{
			Off:    -1,
			Strict: false,
			Types:  []ast.Type{t.Elt, ast.BuiltinUntypedBool},
		}
	default:
		return nil, &BadIndexedType{Off: x.Off, File: ti.File, Type: typ}
	}
	return x, nil
}

func (ti *typeInferer) inferIndexExprIndex(x *ast.IndexExpr) error {
	// For maps, the type context for the index is the map key type, for other
	// indexed types, it's `int`.
	var ctx ast.Type
	if m, ok := x.X.Type().(*ast.MapType); ok {
		ctx = m.Key
	} else {
		ctx = ast.BuiltinInt
	}
	y, err := ti.inferExpr(x.I, ctx)
	if err != nil {
		return err
	}
	x.I = y
	return nil
}

func (ti *typeInferer) VisitSliceExpr(x *ast.SliceExpr) (ast.Expr, error) {
	y, err := ti.inferExpr(x.X, nil)
	if err != nil {
		return nil, err
	}
	x.X = y
	ti.delay(func() error { return ti.inferSliceExprIndices(x) })
	typ := underlyingType(x.X.Type())
	if ptr, ok := typ.(*ast.PtrType); ok {
		b, ok := underlyingType(ptr.Base).(*ast.ArrayType)
		if !ok {
			return nil, &BadIndexedType{Off: x.Off, File: ti.File, Type: typ}
		}
		typ = b
	}

	switch t := typ.(type) {
	case *ast.BuiltinType:
		if t.Kind != ast.BUILTIN_STRING && t.Kind != ast.BUILTIN_UNTYPED_STRING {
			return nil, &BadIndexedType{Off: x.Off, File: ti.File, Type: typ}
		}
		x.Typ = ast.BuiltinString
	case *ast.ArrayType:
		x.Typ = &ast.SliceType{Off: t.Off, Elt: t.Elt}
	case *ast.SliceType:
		x.Typ = x.X.Type()
	default:
		return nil, &BadIndexedType{Off: x.Off, File: ti.File, Type: typ}
	}
	return x, nil
}

func (ti *typeInferer) inferSliceExprIndices(x *ast.SliceExpr) error {
	if x.Lo != nil {
		y, err := ti.inferExpr(x.Lo, ast.BuiltinInt)
		if err != nil {
			return err
		}
		x.Lo = y
	}
	if x.Hi != nil {
		y, err := ti.inferExpr(x.Hi, ast.BuiltinInt)
		if err != nil {
			return err
		}
		x.Hi = y
	}
	if x.Cap != nil {
		y, err := ti.inferExpr(x.Cap, ast.BuiltinInt)
		if err != nil {
			return err
		}
		x.Cap = y
	}
	return nil
}

func (ti *typeInferer) VisitUnaryExpr(x *ast.UnaryExpr) (ast.Expr, error) {
	y, err := ti.inferExpr(x.X, nil)
	if err != nil {
		return nil, err
	}

	if y.Type() == nil {
		if ti.TypeCtx == nil {
			return x, nil
		}
		y, err = ti.inferExpr(y, ti.TypeCtx)
		if err != nil {
			return nil, err
		}
		if y.Type() == nil {
			panic("not reached")
		}
	}

	x.X = y
	switch x.Op {
	case '+':
		return ti.inferUnaryPlus(x)
	case '-':
		return ti.inferUnaryMinus(x)
	case '!':
		return ti.inferNot(x)
	case '^':
		return ti.inferComplement(x)
	case '*':
		return ti.inferIndirection(x)
	case '&':
		return ti.inferAddr(x)
	case ast.RECV:
		return ti.inferRecv(x)
	default:
		panic("not reached")
	}
}

func (ti *typeInferer) inferUnaryPlus(x *ast.UnaryExpr) (ast.Expr, error) {
	if !isArith(x.X) {
		return nil, &BadOperand{
			Off: x.X.Position(), File: ti.File, Op: '+', Expected: "arithmetic type",
			Type: x.X.Type()}
	}
	x.Typ = x.X.Type()
	return x, nil
}

func (ti *typeInferer) inferUnaryMinus(x *ast.UnaryExpr) (ast.Expr, error) {
	t := builtinType(x.X.Type())
	if t == nil || !t.IsArith() {
		return nil, &BadOperand{
			Off: x.X.Position(), File: ti.File, Op: '-',
			Expected: "arithmetic type", Type: x.X.Type(),
		}
	}
	x.Typ = x.X.Type()
	return x, nil
}

func (ti *typeInferer) inferNot(x *ast.UnaryExpr) (ast.Expr, error) {
	t := builtinType(x.X.Type())
	if t == nil || t.Kind != ast.BUILTIN_BOOL && t.Kind != ast.BUILTIN_UNTYPED_BOOL {
		return nil, &BadOperand{
			Off: x.X.Position(), File: ti.File, Op: '!', Expected: "boolean type",
			Type: x.X.Type()}
	}
	x.Typ = x.X.Type()
	return x, nil
}

func (ti *typeInferer) inferComplement(x *ast.UnaryExpr) (ast.Expr, error) {
	if t := builtinType(x.X.Type()); t == nil || !t.IsUntyped() && !t.IsInteger() {
		return nil, &BadOperand{
			Off: x.X.Position(), File: ti.File, Op: '^', Expected: "integer type",
			Type: x.X.Type()}
	}
	x.Typ = x.X.Type()
	return x, nil
}

func (ti *typeInferer) inferIndirection(x *ast.UnaryExpr) (ast.Expr, error) {
	ptr, ok := underlyingType(x.X.Type()).(*ast.PtrType)
	if !ok {
		return nil, &BadOperand{
			Off: x.X.Position(), File: ti.File, Op: '*', Expected: "pointer type",
			Type: x.X.Type()}
	}
	x.Typ = ptr.Base
	return x, nil
}

func (ti *typeInferer) inferAddr(x *ast.UnaryExpr) (ast.Expr, error) {
	if !isAddressable(x.X) {
		return nil, &BadOperand{
			Off: x.X.Position(), File: ti.File, Op: '&', Expected: "addressable"}
	}
	x.Typ = &ast.PtrType{Off: x.Off, Base: x.X.Type()}
	return x, nil
}

func (ti *typeInferer) inferRecv(x *ast.UnaryExpr) (ast.Expr, error) {
	ch, ok := underlyingType(x.X.Type()).(*ast.ChanType)
	if !ok || !ch.Recv {
		return nil, &BadOperand{
			Off: x.X.Position(), File: ti.File, Op: ast.RECV,
			Expected: "receivable channel", Type: x.X.Type()}
	}
	x.Typ = &ast.TupleType{Off: x.Off, Types: []ast.Type{ch.Elt, ast.BuiltinUntypedBool}}
	return x, nil
}

func (ti *typeInferer) VisitBinaryExpr(x *ast.BinaryExpr) (ast.Expr, error) {
	if x.Op == ast.SHL || x.Op == ast.SHR {
		return ti.inferShift(x)
	}
	if x.Op.IsComparison() {
		// FIXME: infer operand types
		x.Typ = ast.BuiltinUntypedBool
		ti.delay(func() error { return ti.inferComparisonExprOperands(x) })
		return x, nil
	}

	// Infer the operands without a type context first, as a type context from
	// one operand takes precedence over type context passed by the parent
	// expression.
	u, err := ti.inferExpr(x.X, nil)
	if err != nil {
		return nil, err
	}
	v, err := ti.inferExpr(x.Y, nil)
	if err != nil {
		return nil, err
	}

	// If one operand is untyped, convert it to the type of the other operand
	switch t0, t1 := u.Type(), v.Type(); {
	case isUntyped(t0) && !isUntyped(t1):
		if t0 != nil {
			u = &ast.Conversion{Off: u.Position(), Typ: t1, X: u}
		}
		if u, err = ti.inferExpr(u, t1); err != nil {
			return nil, err
		}
		if isUntyped(u.Type()) {
			panic("not reached")
		}
	case !isUntyped(t0) && isUntyped(t1):
		if t1 != nil {
			v = &ast.Conversion{Off: v.Position(), Typ: t0, X: v}
		}
		if v, err = ti.inferExpr(v, t0); err != nil {
			return nil, err
		}
		if isUntyped(v.Type()) {
			panic("not reached")
		}
	}

	// Force the type from the context, if the operands are non-constant, but
	// still untyped.
	if !ti.isConst(u) || !ti.isConst(v) {
		if isUntyped(u.Type()) && isUntyped(v.Type()) {
			if ti.TypeCtx == nil {
				return x, nil
			}
			if u, err = ti.inferExpr(u, ti.TypeCtx); err != nil {
				return nil, err
			}
			if v, err = ti.inferExpr(v, ti.TypeCtx); err != nil {
				return nil, err
			}
			if u.Type() == nil || v.Type() == nil {
				panic("not reached")
			}
		}
	}

	// Only builtin types make sense for non-comparison operations.
	utyp := builtinType(u.Type())
	if utyp == nil {
		return nil, &NotSupportedOperation{
			Off: x.Off, File: ti.File, Op: x.Op, Type: u.Type()}
	}
	vtyp := builtinType(v.Type())
	if vtyp == nil {
		return nil, &NotSupportedOperation{
			Off: x.Off, File: ti.File, Op: x.Op, Type: v.Type()}
	}

	// For untyped operands of a different kind, use the type later in the
	// sequence int, rune, float, complex, and only for operations that make
	// sense.
	if utyp.IsUntyped() || vtyp.IsUntyped() {
		// Both operands ought to be untyped here, or else we would have
		// already converted the untyped one to the type of the other.
		if !(utyp.IsUntyped() && vtyp.IsUntyped()) {
			panic("not reached")
		}
		switch x.Op {
		case '+':
			if !(utyp.IsArith() || utyp.Kind == ast.BUILTIN_UNTYPED_STRING) {
				return nil, &NotSupportedOperation{
					Off: x.Off, File: ti.File, Op: '+', Type: u.Type()}
			}
			if !(vtyp.IsArith() || vtyp.Kind == ast.BUILTIN_UNTYPED_STRING) {
				return nil, &NotSupportedOperation{
					Off: x.Off, File: ti.File, Op: '+', Type: v.Type()}
			}
			if utyp.Kind == ast.BUILTIN_UNTYPED_STRING ||
				vtyp.Kind == ast.BUILTIN_UNTYPED_STRING {
				if utyp.Kind != ast.BUILTIN_UNTYPED_STRING ||
					vtyp.Kind != ast.BUILTIN_UNTYPED_STRING {
					return nil, &BadBinaryOperands{Off: x.Off, File: ti.File, Op: '+',
						XType: u.Type(), YType: v.Type()}
				}
				x.Typ = ast.BuiltinUntypedString
			} else {
				x.Typ = promoteUntyped(utyp, vtyp)
			}
		case '-', '*', '/':
			if !utyp.IsArith() {
				return nil, &NotSupportedOperation{
					Off: x.Off, File: ti.File, Op: x.Op, Type: u.Type()}
			}
			if !vtyp.IsArith() {
				return nil, &NotSupportedOperation{
					Off: x.Off, File: ti.File, Op: x.Op, Type: v.Type()}
			}
			x.Typ = promoteUntyped(utyp, vtyp)
		case '%', '&', '|', '^', ast.ANDN:
			if utyp.Kind != ast.BUILTIN_UNTYPED_INT &&
				utyp.Kind != ast.BUILTIN_UNTYPED_RUNE {
				return nil, &NotSupportedOperation{
					Off: x.Off, File: ti.File, Op: x.Op, Type: u.Type()}
			}
			if vtyp.Kind != ast.BUILTIN_UNTYPED_INT &&
				vtyp.Kind != ast.BUILTIN_UNTYPED_RUNE {
				return nil, &NotSupportedOperation{
					Off: x.Off, File: ti.File, Op: x.Op, Type: v.Type()}
			}
			if utyp.Kind == ast.BUILTIN_UNTYPED_RUNE ||
				vtyp.Kind == ast.BUILTIN_UNTYPED_RUNE {
				x.Typ = ast.BuiltinUntypedRune
			} else {
				x.Typ = ast.BuiltinUntypedInt
			}
		case ast.AND, ast.OR:
			if utyp.Kind != ast.BUILTIN_UNTYPED_BOOL {
				return nil, &NotSupportedOperation{
					Off: x.Off, File: ti.File, Op: x.Op, Type: u.Type()}
			}
			if vtyp.Kind != ast.BUILTIN_UNTYPED_BOOL {
				return nil, &NotSupportedOperation{
					Off: x.Off, File: ti.File, Op: x.Op, Type: v.Type()}
			}
			x.Typ = ast.BuiltinUntypedBool
		}
	} else {
		// If the operands are typed, they should have the same type.
		if u.Type() != v.Type() {
			return nil, &BadBinaryOperands{
				Off: x.Off, File: ti.File, Op: x.Op, XType: u.Type(), YType: v.Type()}
		}
		// Check applicability of the operation to the operand type.
		switch x.Op {
		case '+':
			if !utyp.IsArith() && utyp.Kind != ast.BUILTIN_STRING {
				return nil, &NotSupportedOperation{
					Off: x.Off, File: ti.File, Op: '+', Type: u.Type()}
			}
		case '-', '*', '/':
			if !utyp.IsArith() {
				return nil, &NotSupportedOperation{
					Off: x.Off, File: ti.File, Op: x.Op, Type: u.Type()}
			}
		case '%', '&', '|', '^', ast.ANDN:
			if !utyp.IsInteger() {
				return nil, &NotSupportedOperation{
					Off: x.Off, File: ti.File, Op: x.Op, Type: u.Type()}
			}
		case ast.AND, ast.OR:
			if utyp.Kind != ast.BUILTIN_BOOL {
				return nil, &NotSupportedOperation{
					Off: x.Off, File: ti.File, Op: x.Op, Type: u.Type()}
			}
		}
		// The type of the expression is the type of the operands.
		x.Typ = u.Type()
	}
	x.X = u
	x.Y = v
	return x, nil
}

func (ti *typeInferer) inferShift(x *ast.BinaryExpr) (ast.Expr, error) {
	// Infer the left operand without a type context as the whole shift
	// expression might be a constant.
	u, err := ti.inferExpr(x.X, nil)
	if err != nil {
		return nil, err
	}
	ti.delay(func() error { return ti.inferShiftCount(x) })

	if u.Type() == nil || isUntyped(u.Type()) && !ti.isConst(x.Y) {
		if ti.TypeCtx == nil {
			return x, nil
		}
		if u, err = ti.inferExpr(u, ti.TypeCtx); err != nil {
			return nil, err
		}
		if u.Type() == nil {
			panic("not reached")
		}
	}

	// If both operands are constants, the whole expression is a possibly
	// untyped constant.
	if ti.isConst(u) && ti.isConst(x.Y) {
		switch t := builtinType(u.Type()); t.Kind {
		case ast.BUILTIN_UNTYPED_INT, ast.BUILTIN_UNTYPED_FLOAT,
			ast.BUILTIN_UNTYPED_COMPLEX:
			x.Typ = ast.BuiltinUntypedInt
		case ast.BUILTIN_UNTYPED_RUNE:
			x.Typ = ast.BuiltinUntypedRune
		default:
			if !t.IsInteger() {
				return nil, &BadOperand{
					Off: u.Position(), File: ti.File, Op: x.Op, Expected: "integer type",
					Type: u.Type()}
			}
			x.Typ = u.Type()
		}
		x.X = u
		return x, nil
	}

	// The left operand should have integer type.
	if t := builtinType(u.Type()); t == nil || !t.IsInteger() {
		return nil, &BadOperand{
			Off: u.Position(), File: ti.File, Op: x.Op, Expected: "integer type",
			Type: u.Type()}
	}

	x.X = u
	x.Typ = u.Type()
	return x, nil
}

func (ti *typeInferer) inferComparisonExprOperands(x *ast.BinaryExpr) error {
	u, err := ti.inferExpr(x.X, nil)
	if err != nil {
		return err
	}
	v, err := ti.inferExpr(x.Y, nil)
	if err != nil {
		return err
	}

	// If one operand is untyped, convert it to the type of the other operand
	switch t0, t1 := u.Type(), v.Type(); {
	case isUntyped(t0) && !isUntyped(t1):
		if t0 != nil {
			u = &ast.Conversion{Off: u.Position(), Typ: t1, X: u}
		}
		if u, err = ti.inferExpr(u, t1); err != nil {
			return err
		}
		if isUntyped(u.Type()) {
			panic("not reached")
		}
	case !isUntyped(t0) && isUntyped(t1):
		if t1 != nil {
			v = &ast.Conversion{Off: v.Position(), Typ: t0, X: v}
		}
		if v, err = ti.inferExpr(v, t0); err != nil {
			return err
		}
		if isUntyped(v.Type()) {
			panic("not reached")
		}
	}

	// Force the default type, if the operands are non-constant, but still
	// untyped.
	if !ti.isConst(u) || !ti.isConst(v) {
		if isUntyped(u.Type()) && isUntyped(v.Type()) {
			if u, err = ti.inferExpr(u, ast.BuiltinDefault); err != nil {
				return err
			}
			if v, err = ti.inferExpr(v, ast.BuiltinDefault); err != nil {
				return err
			}
			if u.Type() == nil || v.Type() == nil {
				panic("not reached")
			}
		}
	}

	x.X = u
	x.Y = v
	return nil
}

func (ti *typeInferer) inferShiftCount(x *ast.BinaryExpr) error {
	y, err := ti.inferExpr(x.Y, nil)
	if err != nil {
		return err
	}
	x.Y = y
	return nil
}

func (ti *typeInferer) isConst(x ast.Expr) bool {
	switch x := x.(type) {
	case *ast.ConstValue:
		return true
	case *ast.OperandName:
		_, ok := x.Decl.(*ast.Const)
		return ok
	case *ast.Call:
		return ti.isConstCall(x)
	case *ast.Conversion:
		if t := builtinType(x.Typ); t == nil {
			return false
		}
		return ti.isConst(x.X)
	case *ast.UnaryExpr:
		return x.Op != '*' && x.Op != '&' && x.Op != ast.RECV && ti.isConst(x.X)
	case *ast.BinaryExpr:
		return ti.isConst(x.X) && ti.isConst(x.Y)
	default:
		return false
	}
}

func (ti *typeInferer) isConstCall(x *ast.Call) bool {
	// Only builtin calls can possibly be constant expressions.
	op, ok := x.Func.(*ast.OperandName)
	if !ok {
		return false
	}
	d, ok := op.Decl.(*ast.FuncDecl)
	if !ok {
		return false
	}
	switch d {
	case ast.BuiltinLen, ast.BuiltinCap:
		if len(x.Xs) != 1 {
			return false
		}
		y, err := ti.inferExpr(x.Xs[0], nil)
		if err != nil {
			return false
		}
		x.Xs[0] = y
		if y.Type() == nil {
			return false
		}
		t := underlyingType(y.Type())
		if ptr, ok := t.(*ast.PtrType); ok {
			a, ok := underlyingType(ptr.Base).(*ast.ArrayType)
			if !ok {
				return false
			}
			t = a
		}
		switch t := underlyingType(t).(type) {
		case *ast.BuiltinType:
			return d == ast.BuiltinLen &&
				(t.Kind == ast.BUILTIN_STRING || t.Kind == ast.BUILTIN_UNTYPED_STRING) &&
				ti.isConst(y)
		case *ast.ArrayType:
			return !ti.hasSideEffects(y)
		default:
			return false
		}
	case ast.BuiltinComplex:
		return len(x.Xs) == 2 && ti.isConst(x.Xs[0]) && ti.isConst(x.Xs[1])
	case ast.BuiltinImag, ast.BuiltinReal:
		return len(x.Xs) == 1 && ti.isConst(x.Xs[0])
	}
	return false
}

func (ti *typeInferer) hasSideEffects(x ast.Expr) bool {
	switch x := x.(type) {
	case *ast.CompLiteral:
		for _, e := range x.Elts {
			if e.Key != nil && ti.hasSideEffects(e.Key) || ti.hasSideEffects(e.Elt) {
				return true
			}
		}
		return false
	case *ast.Call:
		return !ti.isConstCall(x)
	case *ast.Conversion:
		return ti.hasSideEffects(x.X)
	case *ast.TypeAssertion:
		return ti.hasSideEffects(x.X)
	case *ast.Selector:
		return ti.hasSideEffects(x.X)
	case *ast.IndexExpr:
		return ti.hasSideEffects(x.X) || ti.hasSideEffects(x.I)
	case *ast.SliceExpr:
		return ti.hasSideEffects(x.X) || ti.hasSideEffects(x.Lo) ||
			ti.hasSideEffects(x.Hi) || ti.hasSideEffects(x.Cap)
	case *ast.UnaryExpr:
		return x.Op == ast.RECV || ti.hasSideEffects(x.X)
	case *ast.BinaryExpr:
		return ti.hasSideEffects(x.X) || ti.hasSideEffects(x.Y)
	default:
		return false
	}
}
