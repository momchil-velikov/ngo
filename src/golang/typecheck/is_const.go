package typecheck

import "golang/ast"

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
		if t := builtinType(x.Typ); t == nil || t.Kind <= ast.BUILTIN_NIL_TYPE {
			return false
		}
		return ti.isConst(x.X)
	case *ast.UnaryExpr:
		return x.Op != '*' && x.Op != '&' && ti.isConst(x.X)
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
		// Constant argument to `len` or `cap` can only be a constant string.
		if ti.isConst(x.Xs[0]) {
			return true
		}
		// lit, ok := x.Xs[0].(*ast.CompLiteral)
		// if !ok || lit.Typ == nil {
		// 	return false
		// }
		// t := underlyingType(lit.Typ)
		// if ptr, ok := t.(*ast.PtrType); ok {
		// 	t = underlyingType(ptr.Base)
		// }
		// arr, ok := t.(*ast.ArrayType)
		// if !ok {
		// 	return false
		// }
	case ast.BuiltinComplex:
		return len(x.Xs) == 2 && ti.isConst(x.Xs[0]) && ti.isConst(x.Xs[1])
	case ast.BuiltinImag, ast.BuiltinReal:
		return len(x.Xs) == 1 && ti.isConst(x.Xs[0])
	}
	return false
}
