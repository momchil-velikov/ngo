package ast

import s "golang/scanner"

type Node interface {
	Format(*FormatContext, uint)
}

type Decl interface {
	Node
	decl()
}

type Type interface {
	Node
	typ()
}

type Expr interface {
	Node
	expr()
}

type Stmt interface {
	Node
	stmt()
}

// Source file
type File struct {
	Package string
	Imports []Decl
	Decls   []Decl
}

type DeclGroup struct {
	Kind  uint
	Decls []Decl
}

func (d DeclGroup) decl() {}
func (d DeclGroup) stmt() {}

type Import struct {
	Name string
	Path []byte
}

func (d Import) decl() {}

// Universal error node
type Error struct {
}

func (e Error) decl() {}
func (e Error) typ()  {}
func (e Error) expr() {}
func (e Error) stmt() {}

//
// Declarations
//
type TypeDecl struct {
	Name string
	Type Type
}

func (d TypeDecl) decl() {}
func (d TypeDecl) stmt() {}

type ConstDecl struct {
	Names  []string
	Type   Type
	Values []Expr
}

func (d ConstDecl) decl() {}
func (d ConstDecl) stmt() {}

type VarDecl struct {
	Names []string
	Type  Type
	Init  []Expr
}

func (v VarDecl) decl() {}
func (v VarDecl) stmt() {}

type Receiver struct {
	Name string
	Type Type
}

type FuncDecl struct {
	Name string
	Recv *Receiver
	Sig  *FuncType
	Blk  *Block
}

func (d FuncDecl) decl() {}

//
// Expressions
//

// Precedence table for binary expressions.
var opPrec = map[uint]uint{
	'*': 5, '/': 5, '%': 5, s.SHL: 5, s.SHR: 5, '&': 5, s.ANDN: 5,
	'+': 4, '-': 4, '|': 4, '^': 4,
	s.EQ: 3, s.NE: 3, s.LT: 3, s.LE: 3, s.GT: 3, s.GE: 3,
	s.AND: 2,
	s.OR:  1,
}

type Literal struct {
	Kind  uint
	Value []byte
}

func (e Literal) expr() {}

type Element struct {
	Key   Expr
	Value Expr
}

type CompLiteral struct {
	Type Type
	Elts []*Element
}

func (e CompLiteral) expr() {}

type Call struct {
	Func Expr
	Type Type
	Xs   []Expr
	Ell  bool
}

func (e Call) expr() {}

type Conversion struct {
	Type Type
	X    Expr
}

func (e Conversion) expr() {}

type MethodExpr struct {
	Type Type
	Id   string
}

func (e MethodExpr) expr() {}

type ParensExpr struct {
	X Expr
}

func (e ParensExpr) expr() {}

type FuncLiteral struct {
	Sig *FuncType
	Blk *Block
}

func (e FuncLiteral) expr() {}

type TypeAssertion struct {
	Type Type
	X    Expr
}

func (e TypeAssertion) expr() {}

type Selector struct {
	X  Expr
	Id string
}

func (e Selector) expr() {}

type IndexExpr struct {
	X, I Expr
}

func (e IndexExpr) expr() {}

type SliceExpr struct {
	X           Expr
	Lo, Hi, Cap Expr
}

func (e SliceExpr) expr() {}

type UnaryExpr struct {
	Op uint
	X  Expr
}

func (ex UnaryExpr) expr() {}

type BinaryExpr struct {
	Op   uint
	X, Y Expr
}

func (ex BinaryExpr) expr() {}

//
// Types
//
type QualId struct {
	Pkg, Id string
}

func (t QualId) typ()  {}
func (t QualId) expr() {}

type ArrayType struct {
	Dim Expr
	Elt Type
}

func (t ArrayType) typ() {}

type SliceType struct {
	Elt Type
}

func (t SliceType) typ() {}

type PtrType struct {
	Base Type
}

func (t PtrType) typ() {}

type MapType struct {
	Key, Elt Type
}

func (t MapType) typ() {}

type ChanType struct {
	Send, Recv bool
	Elt        Type
}

func (t ChanType) typ() {}

type FieldDecl struct {
	Names []string
	Type  Type
	Tag   []byte
}

type StructType struct {
	Fields []*FieldDecl
}

func (t StructType) typ() {}

type ParamDecl struct {
	Names []string
	Type  Type
	Var   bool
}

type FuncType struct {
	Params  []*ParamDecl
	Returns []*ParamDecl
}

func (f FuncType) typ() {}

type MethodSpec struct {
	Name string
	Sig  *FuncType
}

type InterfaceType struct {
	Embed   []*QualId
	Methods []*MethodSpec
}

func (t InterfaceType) typ() {}

//
// Statements
//
type EmptyStmt struct{}

func (e EmptyStmt) stmt() {}

type Block struct {
	Body []Stmt
}

func (b Block) stmt() {}

type LabeledStmt struct {
	Label string
	Stmt  Stmt
}

func (l LabeledStmt) stmt() {}

type GoStmt struct {
	X Expr
}

func (b GoStmt) stmt() {}

type ReturnStmt struct {
	Xs []Expr
}

func (b ReturnStmt) stmt() {}

type BreakStmt struct {
	Label string
}

func (b BreakStmt) stmt() {}

type ContinueStmt struct {
	Label string
}

func (b ContinueStmt) stmt() {}

type GotoStmt struct {
	Label string
}

func (b GotoStmt) stmt() {}

type FallthroughStmt struct{}

func (b FallthroughStmt) stmt() {}

type SendStmt struct {
	Ch Expr
	X  Expr
}

func (b SendStmt) stmt() {}

type IncStmt struct {
	X Expr
}

func (b IncStmt) stmt() {}

type DecStmt struct {
	X Expr
}

func (b DecStmt) stmt() {}

type AssignStmt struct {
	Op  uint
	LHS []Expr
	RHS []Expr
}

func (a AssignStmt) stmt() {}

type ExprStmt struct {
	X Expr
}

func (e ExprStmt) stmt() {}

type IfStmt struct {
	Init Stmt
	Cond Expr
	Then *Block
	Else Stmt
}

func (i *IfStmt) stmt() {}

type ForStmt struct {
	Init Stmt
	Cond Expr
	Post Stmt
	Blk  *Block
}

func (f ForStmt) stmt() {}

type ForRangeStmt struct {
	Op    uint
	LHS   []Expr
	Range Expr
	Blk   *Block
}

func (f ForRangeStmt) stmt() {}

type DeferStmt struct {
	X Expr
}

func (d DeferStmt) stmt() {}

type ExprCaseClause struct {
	Xs   []Expr
	Body []Stmt
}

type ExprSwitchStmt struct {
	Init  Stmt
	X     Expr
	Cases []ExprCaseClause
}

func (d ExprSwitchStmt) stmt() {}

type TypeCaseClause struct {
	Types []Type
	Body  []Stmt
}

type TypeSwitchStmt struct {
	Init  Stmt
	Id    string
	X     Expr
	Cases []TypeCaseClause
}

func (d TypeSwitchStmt) stmt() {}

type CommClause struct {
	Comm Stmt
	Body []Stmt
}

type SelectStmt struct {
	Comms []CommClause
}

func (s SelectStmt) stmt() {}
