package ast

type Visitor interface {
	VisitError(*Error) (*Error, error)

	// Types
	VisitTypeName(*QualifiedId) (Type, error)
	VisitTypeDeclType(*TypeDecl) (Type, error)

	VisitBuiltinType(*BuiltinType) (Type, error)
	VisitArrayType(*ArrayType) (Type, error)
	VisitSliceType(*SliceType) (Type, error)
	VisitPtrType(*PtrType) (Type, error)
	VisitMapType(*MapType) (Type, error)
	VisitChanType(*ChanType) (Type, error)
	VisitStructType(*StructType) (Type, error)
	VisitFuncType(*FuncType) (Type, error)
	VisitInterfaceType(*InterfaceType) (Type, error)

	// Expressions
	VisitOperandName(*QualifiedId) (Expr, error)

	VisitLiteral(*Literal) (Expr, error)
	VisitCompLiteral(*CompLiteral) (Expr, error)
	VisitCall(*Call) (Expr, error)
	VisitConversion(*Conversion) (Expr, error)
	VisitMethodExpr(*MethodExpr) (Expr, error)
	VisitParensExpr(*ParensExpr) (Expr, error)
	VisitFunc(*Func) (Expr, error)
	VisitTypeAssertion(*TypeAssertion) (Expr, error)
	VisitSelector(*Selector) (Expr, error)
	VisitIndexExpr(*IndexExpr) (Expr, error)
	VisitSliceExpr(*SliceExpr) (Expr, error)
	VisitUnaryExpr(*UnaryExpr) (Expr, error)
	VisitBinaryExpr(*BinaryExpr) (Expr, error)
	VisitVar(*Var) (Expr, error)
	VisitConst(*Const) (Expr, error)

	// Block-level declarations
	VisitTypeDecl(*TypeDecl) (Stmt, error)
	VisitTypeDeclGroup(*TypeDeclGroup) (Stmt, error)
	VisitConstDecl(*ConstDecl) (Stmt, error)
	VisitConstDeclGroup(*ConstDeclGroup) (Stmt, error)
	VisitVarDecl(*VarDecl) (Stmt, error)
	VisitVarDeclGroup(*VarDeclGroup) (Stmt, error)

	// Statements
	VisitEmptyStmt(*EmptyStmt) (Stmt, error)
	VisitBlock(*Block) (Stmt, error)
	VisitLabeledStmt(*LabeledStmt) (Stmt, error)
	VisitGoStmt(*GoStmt) (Stmt, error)
	VisitReturnStmt(*ReturnStmt) (Stmt, error)
	VisitBreakStmt(*BreakStmt) (Stmt, error)
	VisitContinueStmt(*ContinueStmt) (Stmt, error)
	VisitGotoStmt(*GotoStmt) (Stmt, error)
	VisitFallthroughStmt(*FallthroughStmt) (Stmt, error)
	VisitSendStmt(*SendStmt) (Stmt, error)
	VisitRecvStmt(*RecvStmt) (Stmt, error)
	VisitIncStmt(*IncStmt) (Stmt, error)
	VisitDecStmt(*DecStmt) (Stmt, error)
	VisitAssignStmt(*AssignStmt) (Stmt, error)
	VisitExprStmt(*ExprStmt) (Stmt, error)
	VisitIfStmt(*IfStmt) (Stmt, error)
	VisitForStmt(*ForStmt) (Stmt, error)
	VisitForRangeStmt(*ForRangeStmt) (Stmt, error)
	VisitDeferStmt(*DeferStmt) (Stmt, error)
	VisitExprSwitchStmt(*ExprSwitchStmt) (Stmt, error)
	VisitTypeSwitchStmt(*TypeSwitchStmt) (Stmt, error)
	VisitSelectStmt(*SelectStmt) (Stmt, error)
}

// Error node
func (e *Error) TraverseDecl(v Visitor) (Decl, error) {
	return v.VisitError(e)
}

func (e *Error) TraverseType(v Visitor) (Type, error) {
	return v.VisitError(e)
}

func (e *Error) TraverseExpr(v Visitor) (Expr, error) {
	return v.VisitError(e)
}

func (e *Error) TraverseStmt(v Visitor) (Stmt, error) {
	return v.VisitError(e)
}

// Types
func (t *QualifiedId) TraverseType(v Visitor) (Type, error) {
	return v.VisitTypeName(t)
}

func (t *TypeDecl) TraverseType(v Visitor) (Type, error) {
	return v.VisitTypeDeclType(t)
}

func (t *BuiltinType) TraverseType(v Visitor) (Type, error) {
	return v.VisitBuiltinType(t)
}

func (t *ArrayType) TraverseType(v Visitor) (Type, error) {
	return v.VisitArrayType(t)
}

func (t *SliceType) TraverseType(v Visitor) (Type, error) {
	return v.VisitSliceType(t)
}

func (t *PtrType) TraverseType(v Visitor) (Type, error) {
	return v.VisitPtrType(t)
}

func (t *MapType) TraverseType(v Visitor) (Type, error) {
	return v.VisitMapType(t)
}

func (t *ChanType) TraverseType(v Visitor) (Type, error) {
	return v.VisitChanType(t)
}

func (t *StructType) TraverseType(v Visitor) (Type, error) {
	return v.VisitStructType(t)
}

func (t *FuncType) TraverseType(v Visitor) (Type, error) {
	return v.VisitFuncType(t)
}

func (t *InterfaceType) TraverseType(v Visitor) (Type, error) {
	return v.VisitInterfaceType(t)
}

// Expressions
func (x *QualifiedId) TraverseExpr(v Visitor) (Expr, error) {
	return v.VisitOperandName(x)
}

func (x *Literal) TraverseExpr(v Visitor) (Expr, error) {
	return v.VisitLiteral(x)
}

func (x *CompLiteral) TraverseExpr(v Visitor) (Expr, error) {
	return v.VisitCompLiteral(x)
}

func (x *Call) TraverseExpr(v Visitor) (Expr, error) {
	return v.VisitCall(x)
}

func (x *Conversion) TraverseExpr(v Visitor) (Expr, error) {
	return v.VisitConversion(x)
}

func (x *MethodExpr) TraverseExpr(v Visitor) (Expr, error) {
	return v.VisitMethodExpr(x)
}

func (x *ParensExpr) TraverseExpr(v Visitor) (Expr, error) {
	return v.VisitParensExpr(x)
}

func (x *Func) TraverseExpr(v Visitor) (Expr, error) {
	return v.VisitFunc(x)
}

func (x *TypeAssertion) TraverseExpr(v Visitor) (Expr, error) {
	return v.VisitTypeAssertion(x)
}

func (x *Selector) TraverseExpr(v Visitor) (Expr, error) {
	return v.VisitSelector(x)
}

func (x *IndexExpr) TraverseExpr(v Visitor) (Expr, error) {
	return v.VisitIndexExpr(x)
}

func (x *SliceExpr) TraverseExpr(v Visitor) (Expr, error) {
	return v.VisitSliceExpr(x)
}

func (x *UnaryExpr) TraverseExpr(v Visitor) (Expr, error) {
	return v.VisitUnaryExpr(x)
}

func (x *BinaryExpr) TraverseExpr(v Visitor) (Expr, error) {
	return v.VisitBinaryExpr(x)
}

func (x *Var) TraverseExpr(v Visitor) (Expr, error) {
	return v.VisitVar(x)
}

func (x *Const) TraverseExpr(v Visitor) (Expr, error) {
	return v.VisitConst(x)
}

// Statements
func (s *TypeDecl) TraverseStmt(v Visitor) (Stmt, error) {
	return v.VisitTypeDecl(s)
}

func (s *TypeDeclGroup) TraverseStmt(v Visitor) (Stmt, error) {
	return v.VisitTypeDeclGroup(s)
}

func (s *ConstDecl) TraverseStmt(v Visitor) (Stmt, error) {
	return v.VisitConstDecl(s)
}

func (s *ConstDeclGroup) TraverseStmt(v Visitor) (Stmt, error) {
	return v.VisitConstDeclGroup(s)
}

func (s *VarDecl) TraverseStmt(v Visitor) (Stmt, error) {
	return v.VisitVarDecl(s)
}

func (s *VarDeclGroup) TraverseStmt(v Visitor) (Stmt, error) {
	return v.VisitVarDeclGroup(s)
}

func (s *EmptyStmt) TraverseStmt(v Visitor) (Stmt, error) {
	return v.VisitEmptyStmt(s)
}

func (s *Block) TraverseStmt(v Visitor) (Stmt, error) {
	return v.VisitBlock(s)
}

func (s *LabeledStmt) TraverseStmt(v Visitor) (Stmt, error) {
	return v.VisitLabeledStmt(s)
}

func (s *GoStmt) TraverseStmt(v Visitor) (Stmt, error) {
	return v.VisitGoStmt(s)
}

func (s *ReturnStmt) TraverseStmt(v Visitor) (Stmt, error) {
	return v.VisitReturnStmt(s)
}

func (s *BreakStmt) TraverseStmt(v Visitor) (Stmt, error) {
	return v.VisitBreakStmt(s)
}

func (s *ContinueStmt) TraverseStmt(v Visitor) (Stmt, error) {
	return v.VisitContinueStmt(s)
}

func (s *GotoStmt) TraverseStmt(v Visitor) (Stmt, error) {
	return v.VisitGotoStmt(s)
}

func (s *FallthroughStmt) TraverseStmt(v Visitor) (Stmt, error) {
	return v.VisitFallthroughStmt(s)
}

func (s *SendStmt) TraverseStmt(v Visitor) (Stmt, error) {
	return v.VisitSendStmt(s)
}

func (s *RecvStmt) TraverseStmt(v Visitor) (Stmt, error) {
	return v.VisitRecvStmt(s)
}

func (s *IncStmt) TraverseStmt(v Visitor) (Stmt, error) {
	return v.VisitIncStmt(s)
}

func (s *DecStmt) TraverseStmt(v Visitor) (Stmt, error) {
	return v.VisitDecStmt(s)
}

func (s *AssignStmt) TraverseStmt(v Visitor) (Stmt, error) {
	return v.VisitAssignStmt(s)
}

func (s *ExprStmt) TraverseStmt(v Visitor) (Stmt, error) {
	return v.VisitExprStmt(s)
}

func (s *IfStmt) TraverseStmt(v Visitor) (Stmt, error) {
	return v.VisitIfStmt(s)
}

func (s *ForStmt) TraverseStmt(v Visitor) (Stmt, error) {
	return v.VisitForStmt(s)
}

func (s *ForRangeStmt) TraverseStmt(v Visitor) (Stmt, error) {
	return v.VisitForRangeStmt(s)
}

func (s *DeferStmt) TraverseStmt(v Visitor) (Stmt, error) {
	return v.VisitDeferStmt(s)
}

func (s *ExprSwitchStmt) TraverseStmt(v Visitor) (Stmt, error) {
	return v.VisitExprSwitchStmt(s)
}

func (s *TypeSwitchStmt) TraverseStmt(v Visitor) (Stmt, error) {
	return v.VisitTypeSwitchStmt(s)
}

func (s *SelectStmt) TraverseStmt(v Visitor) (Stmt, error) {
	return v.VisitSelectStmt(s)
}
