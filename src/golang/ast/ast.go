package ast

import s "golang/scanner"

type Node interface {
	Format(*FormatContext, uint)
}

type Decl interface {
	Node
	decl()
}

func (Error) decl()          {}
func (TypeDecl) decl()       {}
func (TypeDeclGroup) decl()  {}
func (ConstDecl) decl()      {}
func (ConstDeclGroup) decl() {}
func (VarDecl) decl()        {}
func (VarDeclGroup) decl()   {}
func (FuncDecl) decl()       {}

type Type interface {
	Node
	typ()
}

func (Error) typ()         {}
func (QualifiedId) typ()   {}
func (ArrayType) typ()     {}
func (SliceType) typ()     {}
func (PtrType) typ()       {}
func (MapType) typ()       {}
func (ChanType) typ()      {}
func (StructType) typ()    {}
func (FuncType) typ()      {}
func (InterfaceType) typ() {}

type Expr interface {
	Node
	expr()
}

func (Error) expr()         {}
func (Literal) expr()       {}
func (CompLiteral) expr()   {}
func (Call) expr()          {}
func (Conversion) expr()    {}
func (MethodExpr) expr()    {}
func (ParensExpr) expr()    {}
func (Func) expr()          {}
func (TypeAssertion) expr() {}
func (Selector) expr()      {}
func (IndexExpr) expr()     {}
func (SliceExpr) expr()     {}
func (UnaryExpr) expr()     {}
func (BinaryExpr) expr()    {}
func (QualifiedId) expr()   {}

type Stmt interface {
	Node
	stmt()
}

func (Error) stmt()           {}
func (TypeDecl) stmt()        {}
func (ConstDecl) stmt()       {}
func (ConstDeclGroup) stmt()  {}
func (VarDecl) stmt()         {}
func (VarDeclGroup) stmt()    {}
func (EmptyStmt) stmt()       {}
func (Block) stmt()           {}
func (LabeledStmt) stmt()     {}
func (GoStmt) stmt()          {}
func (ReturnStmt) stmt()      {}
func (BreakStmt) stmt()       {}
func (ContinueStmt) stmt()    {}
func (GotoStmt) stmt()        {}
func (FallthroughStmt) stmt() {}
func (SendStmt) stmt()        {}
func (IncStmt) stmt()         {}
func (DecStmt) stmt()         {}
func (AssignStmt) stmt()      {}
func (ExprStmt) stmt()        {}
func (IfStmt) stmt()          {}
func (ForStmt) stmt()         {}
func (ForRangeStmt) stmt()    {}
func (DeferStmt) stmt()       {}
func (ExprSwitchStmt) stmt()  {}
func (TypeSwitchStmt) stmt()  {}
func (SelectStmt) stmt()      {}

// Source comment.
type Comment struct {
	Off  int
	Text []byte
}

func (c *Comment) IsLineComment() bool {
	return c.Text != nil && c.Text[0] == '/' && c.Text[1] == '/'
}

// Package
type UnresolvedPackage struct {
	Path, Name string
	Files      []*UnresolvedFile
	Imports    map[string]*UnresolvedPackage
	Mark       int
}

// Source file
type UnresolvedFile struct {
	Off      int // position of the "package" keyword
	Pkg      *UnresolvedPackage
	PkgName  string
	Imports  []*ImportDecl
	Decls    []Decl
	Comments []Comment
	Name     string
	SrcMap   s.SourceMap
}

// Import declaration
type ImportDecl struct {
	Off  int
	Name string
	Path []byte
}

// Universal error node
type Error struct {
	Off int
}

//
// Declarations
//

type Ident struct {
	Off int
	Id  string
}

type TypeDecl struct {
	Off  int
	Name string
	Type Type
}

type TypeDeclGroup struct {
	Types []*TypeDecl
}

func (d TypeDeclGroup) stmt() {}

type Const struct {
	Off  int
	Name string
	Type Type
	Init Expr
}

type ConstDecl struct {
	Names  []*Const
	Type   Type
	Values []Expr
}

type ConstDeclGroup struct {
	Consts []*ConstDecl
}

type Var struct {
	Off  int
	Name string
	Type Type
	Init Expr
}

type VarDecl struct {
	Names []*Var
	Type  Type
	Init  []Expr
}

type VarDeclGroup struct {
	Vars []*VarDecl
}

type Func struct {
	Off  int
	Recv *Param
	Sig  *FuncType
	Blk  *Block
}

type FuncDecl struct {
	Off  int
	Name string
	Func Func
}

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
	Off   int
	Kind  uint
	Value []byte
}

type Element struct {
	Key   Expr
	Value Expr
}

type CompLiteral struct {
	Type Type
	Elts []*Element
}

type Call struct {
	Func Expr
	Type Type
	Xs   []Expr
	Ell  bool
}

type Conversion struct {
	Type Type
	X    Expr
}

type MethodExpr struct {
	Type Type
	Id   string
}

type ParensExpr struct {
	Off int
	X   Expr
}

type TypeAssertion struct {
	Type Type
	X    Expr
}

type Selector struct {
	X  Expr
	Id string
}

type IndexExpr struct {
	X, I Expr
}

type SliceExpr struct {
	X           Expr
	Lo, Hi, Cap Expr
}

type UnaryExpr struct {
	Off int
	Op  uint
	X   Expr
}

type BinaryExpr struct {
	Op   uint
	X, Y Expr
}

//
// Types
//
type QualifiedId struct {
	Off     int
	Pkg, Id string
}

type ArrayType struct {
	Off int
	Dim Expr
	Elt Type
}

type SliceType struct {
	Off int
	Elt Type
}

type PtrType struct {
	Off  int
	Base Type
}

type MapType struct {
	Off      int
	Key, Elt Type
}

type ChanType struct {
	Off        int
	Send, Recv bool
	Elt        Type
}

type Field struct {
	Off  int
	Name string
	Type Type
	Tag  []byte
	Anon bool
}

type StructType struct {
	Off    int
	Fields []Field
}

type Param struct {
	Off  int
	Name string
	Type Type
}

type FuncType struct {
	Off     int
	Params  []Param
	Returns []Param
	Var     bool
}

type MethodSpec struct {
	Off  int
	Name string
	Type Type
}

type InterfaceType struct {
	Off     int
	Methods []MethodSpec
}

//
// Statements
//
type EmptyStmt struct {
	Off int
}

type Block struct {
	Body []Stmt
}

type LabeledStmt struct {
	Off   int
	Label string
	Stmt  Stmt
}

type GoStmt struct {
	Off int
	X   Expr
}

type ReturnStmt struct {
	Off int
	Xs  []Expr
}

type BreakStmt struct {
	Off   int
	Label string
}

type ContinueStmt struct {
	Off   int
	Label string
}

type GotoStmt struct {
	Off   int
	Label string
}

type FallthroughStmt struct {
	Off int
}

type SendStmt struct {
	Ch Expr
	X  Expr
}

type IncStmt struct {
	X Expr
}

type DecStmt struct {
	X Expr
}

type AssignStmt struct {
	Op  uint
	LHS []Expr
	RHS []Expr
}

type ExprStmt struct {
	X Expr
}

type IfStmt struct {
	Off  int
	Init Stmt
	Cond Expr
	Then *Block
	Else Stmt
}

type ForStmt struct {
	Off  int
	Init Stmt
	Cond Expr
	Post Stmt
	Blk  *Block
}

type ForRangeStmt struct {
	Off   int
	Op    uint
	LHS   []Expr
	Range Expr
	Blk   *Block
}

type DeferStmt struct {
	Off int
	X   Expr
}

type ExprCaseClause struct {
	Xs  []Expr
	Blk *Block
}

type ExprSwitchStmt struct {
	Off   int // position of the `switch` keyword
	Init  Stmt
	X     Expr
	Cases []ExprCaseClause
}

type TypeCaseClause struct {
	Types []Type
	Blk   *Block
}

type TypeSwitchStmt struct {
	Off   int // position of the `switch` keyword
	Init  Stmt
	Id    string
	X     Expr
	Cases []TypeCaseClause
}

type CommClause struct {
	Comm Stmt
	Blk  *Block
}

type SelectStmt struct {
	Off     int // position of the `select` keyword
	in, End int // positions of the opening and the closing braces
	Comms   []CommClause
}
