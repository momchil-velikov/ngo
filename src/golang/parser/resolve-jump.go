package parser

import (
	"errors"
	"golang/ast"
)

// jumpResolver is a StmtVisitor, which determines the destination of `goto`,
// `break`, and `continue` statements and checks that:
// a) a goto statement does not jump into the scope of a variable, i.e. over a
//    declaration;
// b) a goto statement does not jump into a block, i.e. the destination label of
//    the jump statement is either in the same block or in an enclosing block;
// c) a break or continue statement is nested within `for`, `switch`, or
//    `select` statement
// d) a labeled break or continue statement refers to a label, which is that of
//    an enclosing `for`, `switch`, or `select` statement
type jumpResolver struct {
	fn      *ast.Func       // Function body being checked
	blk     *ast.Block      // Current block
	outer   ast.Stmt        // Enclosing `for`, `switch`, or `select` statement
	gotos   []*ast.GotoStmt // Pending "forward" gotos
	through bool            // `fallthrough` statement allowed
}

// Checks if scope INNER is the same or nested in scope OUTER.
func isNested(inner ast.Scope, outer ast.Scope) bool {
	for s := inner; s != nil; s = s.Parent() {
		if s == outer {
			return true
		}
	}
	return false
}

// Traverses a block. The SW parameter is true if this is the StatementList
// from an expressiopn switch statement.
func (ck *jumpResolver) visitStatementList(s *ast.Block, sw bool) error {
	b := ck.blk
	ck.blk = s
	defer func() { ck.blk = b }()
	if n := len(s.Body); n > 0 {
		for i := 0; i < n-1; i++ {
			if _, err := s.Body[i].TraverseStmt(ck); err != nil {
				return err
			}
		}
		ck.through = sw
		_, err := s.Body[n-1].TraverseStmt(ck)
		ck.through = false
		if err != nil {
			return err
		}
	}
	// Process the pending gotos list: select `goto` statements with destination
	// in this block and check they don't cross variable declaration
	// scope. Leave in the list only `goto` statements with destination in some
	// enclosing block
	for i := 0; i < len(ck.gotos); i++ {
		g := ck.gotos[i]
		l := ck.fn.FindLabel(g.Label)
		if l.Blk != s {
			continue
		}
		ck.gotos = append(ck.gotos[:i], ck.gotos[i+1:]...)
		for _, d := range s.Decls {
			v, ok := d.(*ast.Var)
			if !ok {
				continue
			}
			if g.Off < v.Off && v.Off < l.Off {
				return errors.New("goto jumps into the scope of " + v.Name)
			}
		}
	}
	return nil
}

func (ck *jumpResolver) VisitBlock(s *ast.Block) (ast.Stmt, error) {
	return nil, ck.visitStatementList(s, false)
}

func (ck *jumpResolver) VisitBreakStmt(s *ast.BreakStmt) (ast.Stmt, error) {
	if len(s.Label) > 0 {
		l := ck.fn.FindLabel(s.Label)
		if l == nil {
			return nil, errors.New("unknown label")
		}
		s.Dst = l.Stmt
	} else {
		s.Dst = ck.outer
	}
	var outer ast.Scope
	switch d := s.Dst.(type) {
	case *ast.ForStmt:
		outer = d
	case *ast.ForRangeStmt:
		outer = d
	case *ast.ExprSwitchStmt:
		outer = d
	case *ast.TypeSwitchStmt:
		outer = d
	case *ast.SelectStmt:
		outer = d
	default:
		err := errors.New(
			"FIXME:break label does not refer to a `for`, `switch`, or `select` statement")
		return nil, err
	}
	if !isNested(ck.blk, outer) {
		return nil, errors.New("break label does not refer to an enclosing statement")
	}
	return nil, nil
}

func (ck *jumpResolver) VisitContinueStmt(s *ast.ContinueStmt) (ast.Stmt, error) {
	if len(s.Label) > 0 {
		l := ck.fn.FindLabel(s.Label)
		if l == nil {
			return nil, errors.New("unknown label")
		}
		s.Dst = l.Stmt
	} else {
		s.Dst = ck.outer
	}
	var outer ast.Scope
	switch d := s.Dst.(type) {
	case *ast.ForStmt:
		outer = d
	case *ast.ForRangeStmt:
		outer = d
	default:
		return nil, errors.New("continue label does not refer to a `for` statement")
	}
	if !isNested(ck.blk, outer) {
		return nil, errors.New("continue label does not refer to an enclosing statement")
	}
	return nil, nil
}

func (ck *jumpResolver) VisitGotoStmt(s *ast.GotoStmt) (ast.Stmt, error) {
	l := ck.fn.FindLabel(s.Label)
	if l == nil {
		return nil, errors.New("unknown label")
	}
	s.Dst = l.Stmt
	dst := l.Blk
	if !isNested(ck.blk, dst) {
		return nil, errors.New("cross block goto jump")
	}
	// If it's a backward jump, it cannot jump into the scope of a
	// declaration.
	if s.Off > l.Off {
		return nil, nil
	}
	// Add the goto statement to the list of statements to be checked after
	// all statements in a block are traversed.
	ck.gotos = append(ck.gotos, s)
	return nil, nil
}

func (ck *jumpResolver) VisitIfStmt(s *ast.IfStmt) (ast.Stmt, error) {
	if _, err := s.Then.TraverseStmt(ck); err != nil {
		return nil, err
	}
	if s.Else == nil {
		return nil, nil
	}
	return s.Else.TraverseStmt(ck)
}

func (ck *jumpResolver) VisitForStmt(s *ast.ForStmt) (ast.Stmt, error) {
	o := ck.outer
	ck.outer = s
	defer func() { ck.outer = o }()
	return s.Blk.TraverseStmt(ck)
}

func (ck *jumpResolver) VisitForRangeStmt(s *ast.ForRangeStmt) (ast.Stmt, error) {
	o := ck.outer
	ck.outer = s
	defer func() { ck.outer = o }()
	return s.Blk.TraverseStmt(ck)
}

func (ck *jumpResolver) VisitExprSwitchStmt(s *ast.ExprSwitchStmt) (ast.Stmt, error) {
	o := ck.outer
	ck.outer = s
	defer func() { ck.outer = o }()
	if n := len(s.Cases); n > 0 {
		for i := 0; i < n; i++ {
			c := &s.Cases[i]
			if err := ck.visitStatementList(c.Blk, true); err != nil {
				return nil, err
			}
			if m := len(c.Blk.Body); m > 0 {
				if f, ok := c.Blk.Body[m-1].(*ast.FallthroughStmt); ok {
					if i == n-1 {
						return nil, errors.New("cannot fallthrough the last case")
					}
					f.Dst = s.Cases[i+1].Blk
				}
			}
		}
	}
	return nil, nil
}

func (ck *jumpResolver) VisitTypeSwitchStmt(s *ast.TypeSwitchStmt) (ast.Stmt, error) {
	o := ck.outer
	ck.outer = s
	defer func() { ck.outer = o }()
	for _, c := range s.Cases {
		if _, err := c.Blk.TraverseStmt(ck); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (ck *jumpResolver) VisitSelectStmt(s *ast.SelectStmt) (ast.Stmt, error) {
	o := ck.outer
	ck.outer = s
	defer func() { ck.outer = o }()
	for _, c := range s.Comms {
		if _, err := c.Blk.TraverseStmt(ck); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

// Nops
func (*jumpResolver) VisitError(*ast.Error) (*ast.Error, error) {
	panic("not reached")
}

func (*jumpResolver) VisitTypeDecl(*ast.TypeDecl) (ast.Stmt, error) {
	panic("not reached")
}

func (*jumpResolver) VisitTypeDeclGroup(*ast.TypeDeclGroup) (ast.Stmt, error) {
	panic("not reached")
}

func (*jumpResolver) VisitConstDecl(*ast.ConstDecl) (ast.Stmt, error) {
	panic("not reached")
}

func (*jumpResolver) VisitConstDeclGroup(*ast.ConstDeclGroup) (ast.Stmt, error) {
	panic("not reached")
}

func (*jumpResolver) VisitVarDecl(*ast.VarDecl) (ast.Stmt, error) {
	panic("not reached")
}

func (*jumpResolver) VisitVarDeclGroup(*ast.VarDeclGroup) (ast.Stmt, error) {
	panic("not reached")
}

func (*jumpResolver) VisitEmptyStmt(*ast.EmptyStmt) (ast.Stmt, error) {
	panic("not reached")
}

func (*jumpResolver) VisitLabel(*ast.Label) (ast.Stmt, error) {
	return nil, nil
}

func (*jumpResolver) VisitReturnStmt(*ast.ReturnStmt) (ast.Stmt, error) {
	return nil, nil
}

func (*jumpResolver) VisitGoStmt(s *ast.GoStmt) (ast.Stmt, error) {
	return nil, nil
}

func (ck *jumpResolver) VisitFallthroughStmt(s *ast.FallthroughStmt) (ast.Stmt, error) {
	if !ck.through {
		return nil, errors.New("misplaced fallthrough statement")
	}
	return nil, nil
}

func (*jumpResolver) VisitSendStmt(*ast.SendStmt) (ast.Stmt, error) {
	return nil, nil
}

func (*jumpResolver) VisitRecvStmt(*ast.RecvStmt) (ast.Stmt, error) {
	return nil, nil
}

func (*jumpResolver) VisitIncStmt(*ast.IncStmt) (ast.Stmt, error) {
	return nil, nil
}

func (*jumpResolver) VisitDecStmt(*ast.DecStmt) (ast.Stmt, error) {
	return nil, nil
}

func (*jumpResolver) VisitAssignStmt(*ast.AssignStmt) (ast.Stmt, error) {
	return nil, nil
}

func (*jumpResolver) VisitExprStmt(*ast.ExprStmt) (ast.Stmt, error) {
	return nil, nil
}

func (*jumpResolver) VisitDeferStmt(*ast.DeferStmt) (ast.Stmt, error) {
	return nil, nil
}
