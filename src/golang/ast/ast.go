package ast

import "golang/scanner"

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

// Source comment.
type Comment struct {
	Off  int
	Text []byte
}

func (c *Comment) IsLineComment() bool {
	return c.Text != nil && c.Text[0] == '/' && c.Text[1] == '/'
}

// Package
type Package struct {
	Path, Name string
	Files      []*File
	Imports    map[string]*Package
	Mark       int
}

// Source file
type File struct {
	Off      int // position of the "package" keyword
	Pkg      *Package
	PkgName  string
	Imports  []*Import
	Decls    []Decl
	Comments []Comment
	Name     string
	SrcMap   scanner.SourceMap
}

// Import declaration
type Import struct {
	Off  int
	Name string
	Path []byte
}

// Universal error node
type Error struct {
	Off int
}

func (e Error) decl() {}
func (e Error) typ()  {}
func (e Error) expr() {}
func (e Error) stmt() {}

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

func (d TypeDecl) decl() {}
func (d TypeDecl) stmt() {}

type TypeDeclGroup struct {
	Types []*TypeDecl
}

func (d TypeDeclGroup) decl() {}
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

func (d ConstDecl) decl() {}
func (d ConstDecl) stmt() {}

type ConstDeclGroup struct {
	Consts []*ConstDecl
}

func (d ConstDeclGroup) decl() {}
func (d ConstDeclGroup) stmt() {}

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

func (v VarDecl) decl() {}
func (v VarDecl) stmt() {}

type VarDeclGroup struct {
	Vars []*VarDecl
}

func (d VarDeclGroup) decl() {}
func (d VarDeclGroup) stmt() {}

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

func (d FuncDecl) decl() {}

//
// Expressions
//

// Precedence table for binary expressions.
var opPrec = map[uint]uint{
	'*':          5,
	'/':          5,
	'%':          5,
	scanner.SHL:  5,
	scanner.SHR:  5,
	'&':          5,
	scanner.ANDN: 5,
	'+':          4,
	'-':          4,
	'|':          4,
	'^':          4,
	scanner.EQ:   3,
	scanner.NE:   3,
	scanner.LT:   3,
	scanner.LE:   3,
	scanner.GT:   3,
	scanner.GE:   3,
	scanner.AND:  2,
	scanner.OR:   1,
}

type Literal struct {
	Off   int
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
	Off int
	X   Expr
}

func (e ParensExpr) expr() {}

func (e Func) expr() {}

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
	Off int
	Op  uint
	X   Expr
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
type QualifiedId struct {
	Off     int
	Pkg, Id string
}

func (t QualifiedId) typ()  {}
func (t QualifiedId) expr() {}

type ArrayType struct {
	Off int
	Dim Expr
	Elt Type
}

func (t ArrayType) typ() {}

type SliceType struct {
	Off int
	Elt Type
}

func (t SliceType) typ() {}

type PtrType struct {
	Off  int
	Base Type
}

func (t PtrType) typ() {}

type MapType struct {
	Off      int
	Key, Elt Type
}

func (t MapType) typ() {}

type ChanType struct {
	Off        int
	Send, Recv bool
	Elt        Type
}

func (t ChanType) typ() {}

type Field struct {
	Off  int
	Name string
	Type Type
	Tag  string
	Anon bool
}

type StructType struct {
	Off    int
	Fields []Field
}

func (t StructType) typ() {}

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

func (f FuncType) typ() {}

type MethodSpec struct {
	Off  int
	Name string
	Type Type
}

type InterfaceType struct {
	Off     int
	Methods []*MethodSpec
}

func (t InterfaceType) typ() {}

//
// Statements
//
type EmptyStmt struct {
	Off int
}

func (e EmptyStmt) stmt() {}

type Block struct {
	Body []Stmt
}

func (b Block) stmt() {}

type LabeledStmt struct {
	Off   int
	Label string
	Stmt  Stmt
}

func (l LabeledStmt) stmt() {}

type GoStmt struct {
	Off int
	X   Expr
}

func (b GoStmt) stmt() {}

type ReturnStmt struct {
	Off int
	Xs  []Expr
}

func (b ReturnStmt) stmt() {}

type BreakStmt struct {
	Off   int
	Label string
}

func (b BreakStmt) stmt() {}

type ContinueStmt struct {
	Off   int
	Label string
}

func (b ContinueStmt) stmt() {}

type GotoStmt struct {
	Off   int
	Label string
}

func (b GotoStmt) stmt() {}

type FallthroughStmt struct {
	Off int
}

func (b FallthroughStmt) stmt() {}

type SendStmt struct {
	Ch Expr
	X  Expr
}

func (b SendStmt) stmt() {}

type RecvStmt struct {
	Op   uint
	X, Y Expr
	Rcv  Expr
}

func (b RecvStmt) stmt() {}

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
	Off  int
	Init Stmt
	Cond Expr
	Then *Block
	Else Stmt
}

func (i *IfStmt) stmt() {}

type ForStmt struct {
	Off  int
	Init Stmt
	Cond Expr
	Post Stmt
	Blk  *Block
}

func (f ForStmt) stmt() {}

type ForRangeStmt struct {
	Off   int
	Op    uint
	LHS   []Expr
	Range Expr
	Blk   *Block
}

func (f ForRangeStmt) stmt() {}

type DeferStmt struct {
	Off int
	X   Expr
}

func (d DeferStmt) stmt() {}

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

func (d ExprSwitchStmt) stmt() {}

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

func (d TypeSwitchStmt) stmt() {}

type CommClause struct {
	Comm Stmt
	Blk  *Block
}

type SelectStmt struct {
	Off     int // position of the `select` keyword
	in, End int // positions of the opening and the closing braces
	Comms   []CommClause
}

func (s SelectStmt) stmt() {}
