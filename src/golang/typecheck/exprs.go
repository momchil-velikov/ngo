package typecheck

import "golang/ast"

func (ck *typeckPhase0) VisitQualifiedId(*ast.QualifiedId) (ast.Expr, error) {
	panic("not reached")
}

func (*typeckPhase0) VisitConstValue(x *ast.ConstValue) (ast.Expr, error) {
	return x, nil
}

func (*typeckPhase0) VisitCompLiteral(x *ast.CompLiteral) (ast.Expr, error) {
	return x, nil
}

func (ck *typeckPhase0) VisitOperandName(x *ast.OperandName) (ast.Expr, error) {
	switch d := x.Decl.(type) {
	case *ast.Var:
		err := ck.checkVarDecl(d)
		if err != nil {
			return nil, err
		}
		x.Typ = d.Type
	case *ast.Const:
		err := ck.checkConstDecl(d)
		if err != nil {
			return nil, err
		}
		x.Typ = d.Type
	case *ast.FuncDecl:
		if d.Func.Recv != nil {
			panic("method declaration should never be an OperandName")
		}
		err := ck.checkFuncDecl(d)
		if err != nil {
			return nil, err
		}
		x.Typ = d.Func.Sig
	default:
		panic("not reached")
	}
	return x, nil
}

func (ck *typeckPhase0) VisitCall(x *ast.Call) (ast.Expr, error) {
	// Check arguments.
	if x.ATyp != nil {
		t, err := ck.checkType(x.ATyp)
		if err != nil {
			return nil, err
		}
		x.ATyp = t
	}
	for i := range x.Xs {
		y, err := ck.checkExpr(x.Xs[i])
		if err != nil {
			return nil, err
		}
		x.Xs[i] = y
	}
	// Check if we have a builtin function call.
	if op, ok := x.Func.(*ast.OperandName); ok {
		if d, ok := op.Decl.(*ast.FuncDecl); ok {
			switch d {
			case ast.BuiltinAppend:
				return ck.visitBuiltinAppend(x)
			case ast.BuiltinCap:
				return ck.visitBuiltinCap(x)
			case ast.BuiltinClose:
				return ck.visitBuiltinClose(x)
			case ast.BuiltinComplex:
				return ck.visitBuiltinComplex(x)
			case ast.BuiltinCopy:
				return ck.visitBuiltinCopy(x)
			case ast.BuiltinDelete:
				return ck.visitBuiltinDelete(x)
			case ast.BuiltinImag:
				return ck.visitBuiltinImag(x)
			case ast.BuiltinLen:
				return ck.visitBuiltinLen(x)
			case ast.BuiltinMake:
				return ck.visitBuiltinMake(x)
			case ast.BuiltinNew:
				return ck.visitBuiltinNew(x)
			case ast.BuiltinPanic:
				return ck.visitBuiltinPanic(x)
			case ast.BuiltinPrint:
				return ck.visitBuiltinPrint(x)
			case ast.BuiltinPrintln:
				return ck.visitBuiltinPrintln(x)
			case ast.BuiltinReal:
				return ck.visitBuiltinReal(x)
			case ast.BuiltinRecover:
				return ck.visitBuiltinRecover(x)
			}
		}
	}
	// Not a builtin call. Check the called expression.
	fn, err := ck.checkExpr(x.Func)
	if err != nil {
		return nil, err
	}
	x.Func = fn

	// If the type of the called object is unknown, create a new type variable
	// for the call expression.
	t := unnamedType(fn.Type())
	if _, ok := t.(*ast.TypeVar); ok {
		x.Typ = &ast.TypeVar{Off: x.Off, File: ck.File}
		return x, nil
	}
	ftyp, ok := t.(*ast.FuncType)
	if !ok {
		return nil, &NotFunc{Off: x.Off, File: ck.File, X: x}
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
	return x, nil
}

func (*typeckPhase0) visitBuiltinAppend(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*typeckPhase0) visitBuiltinCap(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*typeckPhase0) visitBuiltinClose(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*typeckPhase0) visitBuiltinComplex(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*typeckPhase0) visitBuiltinCopy(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*typeckPhase0) visitBuiltinDelete(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*typeckPhase0) visitBuiltinImag(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (ck *typeckPhase0) visitBuiltinLen(x *ast.Call) (ast.Expr, error) {
	x.Typ = ast.BuiltinInt
	return x, nil
}

func (*typeckPhase0) visitBuiltinMake(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*typeckPhase0) visitBuiltinNew(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*typeckPhase0) visitBuiltinPanic(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*typeckPhase0) visitBuiltinPrint(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*typeckPhase0) visitBuiltinPrintln(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*typeckPhase0) visitBuiltinReal(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*typeckPhase0) visitBuiltinRecover(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (ck *typeckPhase0) VisitConversion(x *ast.Conversion) (ast.Expr, error) {
	t, err := ck.checkType(x.Typ)
	if err != nil {
		return nil, err
	}
	x.Typ = t
	y, err := ck.checkExpr(x.X)
	if err != nil {
		return nil, err
	}
	x.X = y
	return x, nil
}

func (ck *typeckPhase0) VisitMethodExpr(x *ast.MethodExpr) (ast.Expr, error) {
	// Does not check if the receiver type has the form `T` or `*T`, where `T`
	// is a typename. See https://github.com/golang/go/issues/9060

	// Find the method.
	m0, m1, vptr := findFieldOrMethod(x.RTyp, x.Id)
	if (m0.M != nil || m0.F != nil) && (m1.M != nil || m1.F != nil) {
		return nil, &AmbiguousSelector{Off: x.Off, File: ck.File, Name: x.Id}
	}
	if m0.M == nil && m1.M == nil {
		return nil, &NotFound{Off: x.Off, File: ck.File, What: "method", Name: x.Id}
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
		return nil, &BadMethodExpr{Off: x.Off, File: ck.File}
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

func (ck *typeckPhase0) VisitParensExpr(*ast.ParensExpr) (ast.Expr, error) {
	panic("not reached")
}

func (*typeckPhase0) VisitFunc(x *ast.Func) (ast.Expr, error) {
	return x, nil
}

func (ck *typeckPhase0) VisitTypeAssertion(x *ast.TypeAssertion) (ast.Expr, error) {
	if x.ATyp == nil {
		return nil, &BadTypeAssertion{Off: x.Off, File: ck.File}
	}
	t, err := ck.checkType(x.ATyp)
	if err != nil {
		return nil, err
	}
	x.ATyp = t
	x.Typ = &ast.TupleType{
		Off:    x.ATyp.Position(),
		Strict: false,
		Type:   []ast.Type{x.ATyp, ast.BuiltinBool},
	}
	return x, nil
}

func (ck *typeckPhase0) VisitSelector(x *ast.Selector) (ast.Expr, error) {
	y, err := ck.checkExpr(x.X)
	if err != nil {
		return nil, err
	}
	x.X = y
	s0, s1, _ := findSelector(x.X.Type(), x.Id)
	if s1.M != nil || s1.F != nil {
		return nil, &AmbiguousSelector{Off: x.Position(), File: ck.File, Name: x.Id}
	}
	if s0.M == nil && s0.F == nil {
		// Do not report an error now, because the expression type may be
		// unknown.
		x.Typ = &ast.TypeVar{Off: x.Off, File: ck.File}
		return x, nil
	}
	if s0.M != nil {
		x.Typ = s0.M.Func.Sig
	} else {
		x.Typ = s0.F.Type
	}
	return x, nil
}

func (ck *typeckPhase0) VisitIndexExpr(*ast.IndexExpr) (ast.Expr, error) {
	return nil, nil
}

func (ck *typeckPhase0) VisitSliceExpr(*ast.SliceExpr) (ast.Expr, error) {
	return nil, nil
}

func (ck *typeckPhase0) VisitUnaryExpr(x *ast.UnaryExpr) (ast.Expr, error) {
	return x, nil
}

func (ck *typeckPhase0) VisitBinaryExpr(*ast.BinaryExpr) (ast.Expr, error) {
	return nil, nil
}

func (ck *typeckPhase1) VisitQualifiedId(*ast.QualifiedId) (ast.Expr, error) {
	panic("not reached")
}

func (*typeckPhase1) VisitConstValue(x *ast.ConstValue) (ast.Expr, error) {
	return x, nil
}

func (*typeckPhase1) VisitCompLiteral(x *ast.CompLiteral) (ast.Expr, error) {
	return x, nil
}

func (ck *typeckPhase1) VisitOperandName(x *ast.OperandName) (ast.Expr, error) {
	c, ok := x.Decl.(*ast.Const)
	if !ok {
		return x, nil
	}
	if err := ck.checkConstDecl(c); err != nil {
		return nil, err
	}
	return c.Init.(*ast.ConstValue), nil
}

func (ck *typeckPhase1) VisitCall(x *ast.Call) (ast.Expr, error) {
	// Check if we have a builtin function call.
	if op, ok := x.Func.(*ast.OperandName); ok {
		if d, ok := op.Decl.(*ast.FuncDecl); ok {
			switch d {
			case ast.BuiltinAppend:
				return ck.visitBuiltinAppend(x)
			case ast.BuiltinCap:
				return ck.visitBuiltinCap(x)
			case ast.BuiltinClose:
				return ck.visitBuiltinClose(x)
			case ast.BuiltinComplex:
				return ck.visitBuiltinComplex(x)
			case ast.BuiltinCopy:
				return ck.visitBuiltinCopy(x)
			case ast.BuiltinDelete:
				return ck.visitBuiltinDelete(x)
			case ast.BuiltinImag:
				return ck.visitBuiltinImag(x)
			case ast.BuiltinLen:
				return ck.visitBuiltinLen(x)
			case ast.BuiltinMake:
				return ck.visitBuiltinMake(x)
			case ast.BuiltinNew:
				return ck.visitBuiltinNew(x)
			case ast.BuiltinPanic:
				return ck.visitBuiltinPanic(x)
			case ast.BuiltinPrint:
				return ck.visitBuiltinPrint(x)
			case ast.BuiltinPrintln:
				return ck.visitBuiltinPrintln(x)
			case ast.BuiltinReal:
				return ck.visitBuiltinReal(x)
			case ast.BuiltinRecover:
				return ck.visitBuiltinRecover(x)
			}
		}
	}
	// Not a builtin call. Check the called expression.
	fn, err := ck.checkExpr(x.Func)
	if err != nil {
		return nil, err
	}
	x.Func = fn

	// FIXME: check parameters

	// If the type of the expression is type variable, bind it to a concrete
	// type.
	if tv, ok := x.Typ.(*ast.TypeVar); ok {
		ftyp, ok := unnamedType(fn.Type()).(*ast.FuncType)
		if !ok {
			return nil, &NotFunc{Off: x.Off, File: ck.File, X: x}
		}
		// FIXME: refactor here
		switch len(ftyp.Returns) {
		case 0:
			// If the function has no returns, set the call expression type to
			// `void`.
			tv.Type = ast.BuiltinVoidType
		case 1:
			// If the function has a single return, set the type of the call
			// expression to that return's type.
			tv.Type = ftyp.Returns[0].Type
		default:
			// If the called function has multiple return values, set the type
			// of the call expression to a new TupleType, containing the
			// returns' types.
			tp := make([]ast.Type, len(ftyp.Returns))
			for i := range tp {
				tp[i] = ftyp.Returns[i].Type
			}
			tv.Type = &ast.TupleType{Off: x.Off, Strict: true, Type: tp}
		}
		x.Typ = tv.Type
	}
	return x, nil
}

func (*typeckPhase1) visitBuiltinAppend(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*typeckPhase1) visitBuiltinCap(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*typeckPhase1) visitBuiltinClose(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*typeckPhase1) visitBuiltinComplex(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*typeckPhase1) visitBuiltinCopy(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*typeckPhase1) visitBuiltinDelete(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*typeckPhase1) visitBuiltinImag(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (ck *typeckPhase1) visitBuiltinLen(x *ast.Call) (ast.Expr, error) {
	if x.ATyp != nil {
		return nil, &BadTypeArg{Off: x.Off, File: ck.File}
	}
	if len(x.Xs) != 1 {
		return nil, &BadArgNumber{Off: x.Off, File: ck.File}
	}
	y, err := ck.checkExpr(x.Xs[0])
	if err != nil {
		return nil, err
	}
	x.Xs[0] = y
	if a, ok := unnamedType(y.Type()).(*ast.ArrayType); ok {
		return ck.checkExpr(a.Dim)
	}
	return x, nil
}

func (*typeckPhase1) visitBuiltinMake(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*typeckPhase1) visitBuiltinNew(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*typeckPhase1) visitBuiltinPanic(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*typeckPhase1) visitBuiltinPrint(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*typeckPhase1) visitBuiltinPrintln(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*typeckPhase1) visitBuiltinReal(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (*typeckPhase1) visitBuiltinRecover(x *ast.Call) (ast.Expr, error) {
	return x, nil
}

func (ck *typeckPhase1) VisitConversion(x *ast.Conversion) (ast.Expr, error) {
	t, err := ck.checkType(x.Typ)
	if err != nil {
		return nil, err
	}
	x.Typ = t
	y, err := ck.checkExpr(x.X)
	if err != nil {
		return nil, err
	}
	x.X = y

	if c, ok := y.(*ast.ConstValue); ok {
		src := builtinType(c.Typ)
		dst, ok := unnamedType(x.Typ).(*ast.BuiltinType)
		if !ok {
			// return nil, &BadConversion{
			// 	Off: x.Off, File: ck.File, Dst: dst, Src: src, Val: c.Value}
			return nil, &BadConstType{Off: x.Off, File: ck.File, Type: x.Typ}
		}
		v := convertConst(dst, builtinType(c.Typ), c.Value)
		if v == nil {
			return nil, &BadConversion{
				Off: x.Off, File: ck.File, Dst: dst, Src: src, Val: c.Value}
		}
		return &ast.ConstValue{Off: x.Off, Typ: x.Typ, Value: v}, nil
	}

	// FIXME: check conversion is valid
	return x, nil
}

func (ck *typeckPhase1) VisitMethodExpr(x *ast.MethodExpr) (ast.Expr, error) {
	return x, nil
}

func (ck *typeckPhase1) VisitParensExpr(*ast.ParensExpr) (ast.Expr, error) {
	panic("not reached")
}

func (*typeckPhase1) VisitFunc(x *ast.Func) (ast.Expr, error) {
	return x, nil
}

func (ck *typeckPhase1) VisitTypeAssertion(x *ast.TypeAssertion) (ast.Expr, error) {
	return x, nil
}

func (ck *typeckPhase1) VisitSelector(x *ast.Selector) (ast.Expr, error) {
	y, err := ck.checkExpr(x.X)
	if err != nil {
		return nil, err
	}
	x.X = y
	s0, s1, _ := findSelector(x.X.Type(), x.Id)
	if s1.M != nil || s1.F != nil {
		return nil, &AmbiguousSelector{Off: x.Position(), File: ck.File, Name: x.Id}
	}
	if s0.M == nil && s0.F == nil {
		return nil, &NotFound{
			Off: x.Off, File: ck.File, What: "field or method", Name: x.Id}
	}
	if tv, ok := x.Typ.(*ast.TypeVar); ok {
		if s0.M != nil {
			tv.Type = s0.M.Func.Sig
		} else {
			tv.Type = s0.F.Type
		}
		x.Typ = tv.Type
	}
	return x, nil
}

func (ck *typeckPhase1) VisitIndexExpr(*ast.IndexExpr) (ast.Expr, error) {
	return nil, nil
}

func (ck *typeckPhase1) VisitSliceExpr(*ast.SliceExpr) (ast.Expr, error) {
	return nil, nil
}

func (ck *typeckPhase1) VisitUnaryExpr(x *ast.UnaryExpr) (ast.Expr, error) {
	y, err := ck.checkExpr(x.X)
	if err != nil {
		return nil, err
	}
	x.X = y

	c, ok := y.(*ast.ConstValue)
	if !ok {
		return x, nil
	}
	if x.Op == '-' {
		v := minus(builtinType(c.Typ), c.Value)
		if v == nil {
			return nil, &BadOperand{Off: x.Off, File: ck.File, Op: "unary minus"}
		}
		return &ast.ConstValue{Off: x.Off, Typ: c.Typ, Value: v}, nil
	}
	return x, nil
}

func (ck *typeckPhase1) VisitBinaryExpr(*ast.BinaryExpr) (ast.Expr, error) {
	return nil, nil
}
