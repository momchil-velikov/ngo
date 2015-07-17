package ast

import s "golang/scanner"

type Decl interface {
	Formatter
	decl()
}

type TypeSpec interface {
	Formatter
	typeSpec()
}

type Expr interface {
	Formatter
	expr()
}

type Stmt interface {
	Formatter
	stmt()
}

// Source file
type File struct {
	PackageName string
	Imports     []Import
	Decls       []Decl
}

type Import struct {
	Name string
	Path string
}

// Universal error node
type Error struct {
}

func (e Error) decl()     {}
func (e Error) typeSpec() {}
func (e Error) expr()     {}
func (e Error) stmt()     {}

// Declarations
type TypeDecl struct {
	Name string
	Type TypeSpec
}

func (d TypeDecl) decl() {}
func (d TypeDecl) stmt() {}

type TypeGroup struct {
	Decls []*TypeDecl
}

func (d TypeGroup) decl() {}
func (d TypeGroup) stmt() {}

type ConstDecl struct {
	Names  []string
	Type   TypeSpec
	Values []Expr
}

func (d ConstDecl) decl() {}
func (d ConstDecl) stmt() {}

type ConstGroup struct {
	Decls []*ConstDecl
}

func (d ConstGroup) decl() {}
func (d ConstGroup) stmt() {}

type VarDecl struct {
	Names []string
	Type  TypeSpec
	Init  []Expr
}

func (v VarDecl) decl() {}
func (v VarDecl) stmt() {}

type VarGroup struct {
	Decls []*VarDecl
}

func (d VarGroup) decl() {}
func (d VarGroup) stmt() {}

type Receiver struct {
	Name string
	Type TypeSpec
}

type FuncDecl struct {
	Name string
	Recv *Receiver
	Sig  *FuncType
	Body *Block
}

func (d FuncDecl) decl() {}

// Expressions

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
	Value string
}

func (e Literal) expr() {}

type Element struct {
	Key   Expr
	Value Expr
}

type CompLiteral struct {
	Type TypeSpec
	Elts []*Element
}

func (e CompLiteral) expr() {}

type Call struct {
	Func     Expr
	Type     TypeSpec
	Args     []Expr
	Ellipsis bool
}

func (e Call) expr() {}

type Conversion struct {
	Type TypeSpec
	Arg  Expr
}

func (e Conversion) expr() {}

type MethodExpr struct {
	Type TypeSpec
	Id   string
}

func (e MethodExpr) expr() {}

type FuncLiteral struct {
	Sig  *FuncType
	Body *Block
}

func (e FuncLiteral) expr() {}

type TypeAssertion struct {
	Type TypeSpec
	Arg  Expr
}

func (e TypeAssertion) expr() {}

type Selector struct {
	Arg Expr
	Id  string
}

func (e Selector) expr() {}

type IndexExpr struct {
	Array Expr
	Idx   Expr
}

func (e IndexExpr) expr() {}

type SliceExpr struct {
	Array          Expr
	Low, High, Cap Expr
}

func (e SliceExpr) expr() {}

type UnaryExpr struct {
	Op  uint
	Arg Expr
}

func (ex UnaryExpr) expr() {}

type BinaryExpr struct {
	Op         uint
	Arg0, Arg1 Expr
}

func (ex BinaryExpr) expr() {}

// Types
type QualId struct {
	Pkg, Id string
}

func (t QualId) typeSpec() {}
func (t QualId) expr()     {}

type ArrayType struct {
	Dim     Expr
	EltType TypeSpec
}

func (t ArrayType) typeSpec() {}

type SliceType struct {
	EltType TypeSpec
}

func (t SliceType) typeSpec() {}

type PtrType struct {
	Base TypeSpec
}

func (t PtrType) typeSpec() {}

type MapType struct {
	KeyType, EltType TypeSpec
}

func (t MapType) typeSpec() {}

type ChanType struct {
	Send, Recv bool
	EltType    TypeSpec
}

func (t ChanType) typeSpec() {}

type FieldDecl struct {
	Names []string
	Type  TypeSpec
	Tag   string
}

type StructType struct {
	Fields []*FieldDecl
}

func (t StructType) typeSpec() {}

type ParamDecl struct {
	Name     string
	Type     TypeSpec
	Variadic bool
}

type FuncType struct {
	Params  []*ParamDecl
	Returns []*ParamDecl
}

func (f FuncType) typeSpec() {}

type MethodSpec struct {
	Name string
	Sig  *FuncType
}

type InterfaceType struct {
	Embed   []*QualId
	Methods []*MethodSpec
}

func (t InterfaceType) typeSpec() {}

// Statements

type EmptyStmt struct{}

func (e EmptyStmt) stmt() {}

type Block struct {
	Stmts []Stmt
}

func (b Block) stmt() {}

type GoStmt struct {
	Ex Expr
}

func (b GoStmt) stmt() {}

type ReturnStmt struct {
	Exs []Expr
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
	Ex Expr
}

func (b SendStmt) stmt() {}

type IncStmt struct {
	Ex Expr
}

func (b IncStmt) stmt() {}

type DecStmt struct {
	Ex Expr
}

func (b DecStmt) stmt() {}

type AssignStmt struct {
	Op  uint
	LHS []Expr
	RHS []Expr
}

func (a AssignStmt) stmt() {}

type ExprStmt struct {
	Ex Expr
}

func (e ExprStmt) stmt() {}

type IfStmt struct {
	S    Stmt
	Ex   Expr
	Then *Block
	Else Stmt
}

func (i *IfStmt) stmt() {}

type ForStmt struct {
	Init Stmt
	Cond Expr
	Post Stmt
	Body *Block
}

func (f ForStmt) stmt() {}

type ForRangeStmt struct {
	Op   uint
	LHS  []Expr
	Ex   Expr
	Body *Block
}

func (f ForRangeStmt) stmt() {}

type DeferStmt struct {
	Ex Expr
}

func (d DeferStmt) stmt() {}
