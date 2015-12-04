package ast

import "golang/scanner"

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
func (TypeDecl) typ()      {}
func (BuiltinType) typ()   {}
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
func (Var) expr()           {}
func (Const) expr()         {}

type Stmt interface {
	Node
	stmt()
}

func (Error) stmt()           {}
func (TypeDecl) stmt()        {}
func (TypeDeclGroup) stmt()   {}
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
func (RecvStmt) stmt()        {}
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

// The PackageLocator interface abstracts mapping an import path to an actual
// package.
type PackageLocator interface {
	FindPackage(string) (*Package, error)
}

// Package
type Package struct {
	Path  string             // Absolute path of the package directory
	Name  string             // Last component of the path name or "main"
	Sig   [20]byte           // SHA-1 signature of something unrelated here
	Files []*File            // Source files of the package
	Decls map[string]Symbol  // Package-level declarations
	Deps  map[string]*Import // Map of package dependencies
	Init  []*AssignStmt      // Init statements for package-level variables
}

// Imported package
type Import struct {
	No  int
	Pkg *Package
	Sig [20]byte
}

type UnresolvedPackage struct {
	Path  string
	Name  string
	Files []*UnresolvedFile
}

// Source file
type File struct {
	Off     int               // Position of the "package" keyword
	No      int               // Sequence number
	Pkg     *Package          // Owner package
	Imports []*ImportDecl     // Import declarations
	Name    string            // File name
	SrcMap  scanner.SourceMap // Map between source offsets and line/column numbers
	Decls   map[string]Symbol // File scope declarations
}

type UnresolvedFile struct {
	Off      int // position of the "package" keyword
	Pkg      *UnresolvedPackage
	PkgName  string
	Imports  []*ImportDecl
	Decls    []Decl
	Comments []Comment
	Name     string
	SrcMap   scanner.SourceMap
}

// Import declaration
type ImportDecl struct {
	Off  int
	Name string
	Path []byte
	File *File
	Pkg  *Package
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
	File *File
	Name string
	Type Type
}

type TypeDeclGroup struct {
	Types []*TypeDecl
}

type Const struct {
	Off  int
	File *File
	Name string
	Type Type
	Init Expr
	Iota int
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
	File *File
	Name string
	Type Type
	Init *AssignStmt
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
	Decl *FuncDecl
	Recv *Param
	Sig  *FuncType
	Blk  *Block
}

type FuncDecl struct {
	Off  int
	File *File
	Name string
	Func Func
}

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

const (
	BUILTIN_NIL = iota
	BUILTIN_BOOL
	BUILTIN_UINT8
	BUILTIN_UINT16
	BUILTIN_UINT32
	BUILTIN_UINT64
	BUILTIN_INT8
	BUILTIN_INT16
	BUILTIN_INT32
	BUILTIN_INT64
	BUILTIN_FLOAT32
	BUILTIN_FLOAT64
	BUILTIN_COMPLEX64
	BUILTIN_COMPLEX128
	BUILTIN_UINT
	BUILTIN_INT
	BUILTIN_UINTPTR
	BUILTIN_STRING
)

type BuiltinType struct {
	Kind int
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
	Tag  string
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
	Up    Scope
	Decls map[string]Symbol
	Body  []Stmt
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

type RecvStmt struct {
	Op   uint
	X, Y Expr
	Rcv  Expr
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
