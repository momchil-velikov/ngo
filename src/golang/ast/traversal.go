package ast

type TypeVisitor interface {
	VisitError(*Error) (*Error, error)

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
}

type ExprVisitor interface {
	VisitError(*Error) (*Error, error)

	VisitOperandName(*QualifiedId) (Expr, error)

	VisitBuiltinConst(*BuiltinConst) (Expr, error)
	VisitBuiltinFunc(*BuiltinFunc) (Expr, error)
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
}

type StmtVisitor interface {
	VisitError(*Error) (*Error, error)

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
	VisitLabel(*Label) (Stmt, error)
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

// Types
func (e *Error) TraverseType(v TypeVisitor) (Type, error) {
	return v.VisitError(e)
}
func (t *QualifiedId) TraverseType(v TypeVisitor) (Type, error) {
	return v.VisitTypeName(t)
}

func (t *TypeDecl) TraverseType(v TypeVisitor) (Type, error) {
	return v.VisitTypeDeclType(t)
}

func (t *BuiltinType) TraverseType(v TypeVisitor) (Type, error) {
	return v.VisitBuiltinType(t)
}

func (t *ArrayType) TraverseType(v TypeVisitor) (Type, error) {
	return v.VisitArrayType(t)
}

func (t *SliceType) TraverseType(v TypeVisitor) (Type, error) {
	return v.VisitSliceType(t)
}

func (t *PtrType) TraverseType(v TypeVisitor) (Type, error) {
	return v.VisitPtrType(t)
}

func (t *MapType) TraverseType(v TypeVisitor) (Type, error) {
	return v.VisitMapType(t)
}

func (t *ChanType) TraverseType(v TypeVisitor) (Type, error) {
	return v.VisitChanType(t)
}

func (t *StructType) TraverseType(v TypeVisitor) (Type, error) {
	return v.VisitStructType(t)
}

func (t *FuncType) TraverseType(v TypeVisitor) (Type, error) {
	return v.VisitFuncType(t)
}

func (t *InterfaceType) TraverseType(v TypeVisitor) (Type, error) {
	return v.VisitInterfaceType(t)
}

// Expressions
func (e *Error) TraverseExpr(v ExprVisitor) (Expr, error) {
	return v.VisitError(e)
}

func (x *QualifiedId) TraverseExpr(v ExprVisitor) (Expr, error) {
	return v.VisitOperandName(x)
}

func (x *BuiltinConst) TraverseExpr(v ExprVisitor) (Expr, error) {
	return v.VisitBuiltinConst(x)
}

func (x *BuiltinFunc) TraverseExpr(v ExprVisitor) (Expr, error) {
	return v.VisitBuiltinFunc(x)
}

func (x *Literal) TraverseExpr(v ExprVisitor) (Expr, error) {
	return v.VisitLiteral(x)
}

func (x *CompLiteral) TraverseExpr(v ExprVisitor) (Expr, error) {
	return v.VisitCompLiteral(x)
}

func (x *Call) TraverseExpr(v ExprVisitor) (Expr, error) {
	return v.VisitCall(x)
}

func (x *Conversion) TraverseExpr(v ExprVisitor) (Expr, error) {
	return v.VisitConversion(x)
}

func (x *MethodExpr) TraverseExpr(v ExprVisitor) (Expr, error) {
	return v.VisitMethodExpr(x)
}

func (x *ParensExpr) TraverseExpr(v ExprVisitor) (Expr, error) {
	return v.VisitParensExpr(x)
}

func (x *Func) TraverseExpr(v ExprVisitor) (Expr, error) {
	return v.VisitFunc(x)
}

func (x *TypeAssertion) TraverseExpr(v ExprVisitor) (Expr, error) {
	return v.VisitTypeAssertion(x)
}

func (x *Selector) TraverseExpr(v ExprVisitor) (Expr, error) {
	return v.VisitSelector(x)
}

func (x *IndexExpr) TraverseExpr(v ExprVisitor) (Expr, error) {
	return v.VisitIndexExpr(x)
}

func (x *SliceExpr) TraverseExpr(v ExprVisitor) (Expr, error) {
	return v.VisitSliceExpr(x)
}

func (x *UnaryExpr) TraverseExpr(v ExprVisitor) (Expr, error) {
	return v.VisitUnaryExpr(x)
}

func (x *BinaryExpr) TraverseExpr(v ExprVisitor) (Expr, error) {
	return v.VisitBinaryExpr(x)
}

func (x *Var) TraverseExpr(v ExprVisitor) (Expr, error) {
	return v.VisitVar(x)
}

func (x *Const) TraverseExpr(v ExprVisitor) (Expr, error) {
	return v.VisitConst(x)
}

// Statements
func (e *Error) TraverseStmt(v StmtVisitor) (Stmt, error) {
	return v.VisitError(e)
}

func (s *TypeDecl) TraverseStmt(v StmtVisitor) (Stmt, error) {
	return v.VisitTypeDecl(s)
}

func (s *TypeDeclGroup) TraverseStmt(v StmtVisitor) (Stmt, error) {
	return v.VisitTypeDeclGroup(s)
}

func (s *ConstDecl) TraverseStmt(v StmtVisitor) (Stmt, error) {
	return v.VisitConstDecl(s)
}

func (s *ConstDeclGroup) TraverseStmt(v StmtVisitor) (Stmt, error) {
	return v.VisitConstDeclGroup(s)
}

func (s *VarDecl) TraverseStmt(v StmtVisitor) (Stmt, error) {
	return v.VisitVarDecl(s)
}

func (s *VarDeclGroup) TraverseStmt(v StmtVisitor) (Stmt, error) {
	return v.VisitVarDeclGroup(s)
}

func (s *EmptyStmt) TraverseStmt(v StmtVisitor) (Stmt, error) {
	return v.VisitEmptyStmt(s)
}

func (s *Block) TraverseStmt(v StmtVisitor) (Stmt, error) {
	return v.VisitBlock(s)
}

func (s *Label) TraverseStmt(v StmtVisitor) (Stmt, error) {
	return v.VisitLabel(s)
}

func (s *GoStmt) TraverseStmt(v StmtVisitor) (Stmt, error) {
	return v.VisitGoStmt(s)
}

func (s *ReturnStmt) TraverseStmt(v StmtVisitor) (Stmt, error) {
	return v.VisitReturnStmt(s)
}

func (s *BreakStmt) TraverseStmt(v StmtVisitor) (Stmt, error) {
	return v.VisitBreakStmt(s)
}

func (s *ContinueStmt) TraverseStmt(v StmtVisitor) (Stmt, error) {
	return v.VisitContinueStmt(s)
}

func (s *GotoStmt) TraverseStmt(v StmtVisitor) (Stmt, error) {
	return v.VisitGotoStmt(s)
}

func (s *FallthroughStmt) TraverseStmt(v StmtVisitor) (Stmt, error) {
	return v.VisitFallthroughStmt(s)
}

func (s *SendStmt) TraverseStmt(v StmtVisitor) (Stmt, error) {
	return v.VisitSendStmt(s)
}

func (s *RecvStmt) TraverseStmt(v StmtVisitor) (Stmt, error) {
	return v.VisitRecvStmt(s)
}

func (s *IncStmt) TraverseStmt(v StmtVisitor) (Stmt, error) {
	return v.VisitIncStmt(s)
}

func (s *DecStmt) TraverseStmt(v StmtVisitor) (Stmt, error) {
	return v.VisitDecStmt(s)
}

func (s *AssignStmt) TraverseStmt(v StmtVisitor) (Stmt, error) {
	return v.VisitAssignStmt(s)
}

func (s *ExprStmt) TraverseStmt(v StmtVisitor) (Stmt, error) {
	return v.VisitExprStmt(s)
}

func (s *IfStmt) TraverseStmt(v StmtVisitor) (Stmt, error) {
	return v.VisitIfStmt(s)
}

func (s *ForStmt) TraverseStmt(v StmtVisitor) (Stmt, error) {
	return v.VisitForStmt(s)
}

func (s *ForRangeStmt) TraverseStmt(v StmtVisitor) (Stmt, error) {
	return v.VisitForRangeStmt(s)
}

func (s *DeferStmt) TraverseStmt(v StmtVisitor) (Stmt, error) {
	return v.VisitDeferStmt(s)
}

func (s *ExprSwitchStmt) TraverseStmt(v StmtVisitor) (Stmt, error) {
	return v.VisitExprSwitchStmt(s)
}

func (s *TypeSwitchStmt) TraverseStmt(v StmtVisitor) (Stmt, error) {
	return v.VisitTypeSwitchStmt(s)
}

func (s *SelectStmt) TraverseStmt(v StmtVisitor) (Stmt, error) {
	return v.VisitSelectStmt(s)
}
