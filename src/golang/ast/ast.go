package ast

import (
	"golang/scanner"
	"math/big"
)

type Node interface {
	Format(*FormatContext, uint)
	Position() int
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
	TraverseType(TypeVisitor) (Type, error)
	typ()
}

func (Error) typ()         {}
func (QualifiedId) typ()   {}
func (TypeDecl) typ()      {}
func (TypeVar) typ()       {}
func (BuiltinType) typ()   {}
func (ArrayType) typ()     {}
func (SliceType) typ()     {}
func (PtrType) typ()       {}
func (MapType) typ()       {}
func (ChanType) typ()      {}
func (StructType) typ()    {}
func (TupleType) typ()     {}
func (FuncType) typ()      {}
func (InterfaceType) typ() {}

type Expr interface {
	Node
	TraverseExpr(ExprVisitor) (Expr, error)
	Type() Type
	expr()
}

func (Error) expr()         {}
func (ConstValue) expr()    {}
func (OperandName) expr()   {}
func (QualifiedId) expr()   {}
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

type Stmt interface {
	Node
	TraverseStmt(StmtVisitor) (Stmt, error)
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
func (Label) stmt()           {}
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

// Constant value
type Value interface {
	value()
}

func (BuiltinValue) value()   {}
func (Bool) value()           {}
func (Rune) value()           {}
func (UntypedInt) value()     {}
func (Int) value()            {}
func (UntypedFloat) value()   {}
func (Float) value()          {}
func (UntypedComplex) value() {}
func (Complex) value()        {}
func (String) value()         {}

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
	Path  string            // Absolute path of the package directory
	Name  string            // Last component of the path name or "main"
	Sig   [20]byte          // SHA-1 signature of something unrelated here
	Files []*File           // Source files of the package
	Syms  map[string]Symbol // Package-level declarations
	Deps  []*Import         // Package dependencies, ordered by import path
}

// Imported package
type Import struct {
	Path string
	Pkg  *Package
	Sig  [20]byte
}

// Source file
type File struct {
	No       int               // Sequence number
	Off      int               // Position of the "package" keyword
	Pkg      *Package          // Owner package
	PkgName  string            // Package name
	Imports  []*ImportDecl     // Import declarations
	Name     string            // File name
	SrcMap   scanner.SourceMap // Map between source offsets and line/column numbers
	Comments []Comment
	Decls    []Decl
	Syms     map[string]Symbol // File scope declarations
	Init     []*AssignStmt     // Init statements for package-level variables
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

func (e Error) Position() int { return e.Off }
func (Error) Type() Type      { return nil }

//
// Declarations
//

type TypeDecl struct {
	Off      int // position of the identifier
	File     *File
	Name     string
	Type     Type
	Methods  []*FuncDecl
	PMethods []*FuncDecl
}

func (t *TypeDecl) Position() int { return t.Off }

type TypeDeclGroup struct {
	Off   int // position of the opening parenthesis
	Types []*TypeDecl
}

func (t *TypeDeclGroup) Position() int { return t.Off }

type Const struct {
	Off  int // position of the identifier
	File *File
	Name string
	Type Type
	Init Expr
	Iota int
}

func (c *Const) Position() int { return c.Off }

type ConstDecl struct {
	Names  []*Const
	Type   Type
	Values []Expr
}

func (c *ConstDecl) Position() int { return c.Names[0].Position() }

type ConstDeclGroup struct {
	Off    int // position of the opening parenthesis
	Consts []*ConstDecl
}

func (c *ConstDeclGroup) Position() int { return c.Off }

type Var struct {
	Off  int // position of the identifier
	File *File
	Name string
	Type Type
	Init *AssignStmt
}

func (v *Var) Position() int { return v.Off }

type VarDecl struct {
	Names []*Var
	Type  Type
	Init  []Expr
}

func (v *VarDecl) Position() int { return v.Names[0].Position() }

type VarDeclGroup struct {
	Off  int // position of the opening parenthesis
	Vars []*VarDecl
}

func (v *VarDeclGroup) Position() int { return v.Off }

type Func struct {
	Off    int // position of the `func` keyword
	Decl   *FuncDecl
	Recv   *Param
	Sig    *FuncType
	Blk    *Block
	Up     Scope
	Labels map[string]*Label
}

func (f *Func) Position() int { return f.Off }
func (f *Func) Type() Type    { return f.Sig }

type FuncDecl struct {
	Off  int // position of the identifier
	File *File
	Name string
	Func Func
}

func (f *FuncDecl) Position() int { return f.Off }

//
// Expressions
//

// Operations; Must be the same as scanner token codes.
const (
	NOP = 0
	DCL = 1

	PLUS   = '+'
	MINUS  = '-'
	MUL    = '*'
	DIV    = '/'
	REM    = '%'
	BITAND = '&'
	BITOR  = '|'
	BITXOR = '^'
	LT     = '<'
	GT     = '>'
	NOT    = '!'

	SHL  = scanner.SHL
	SHR  = scanner.SHR
	ANDN = scanner.ANDN
	AND  = scanner.AND
	OR   = scanner.OR
	RECV = scanner.RECV
	EQ   = scanner.EQ
	NE   = scanner.NE
	LE   = scanner.LE
	GE   = scanner.GE
)

// Precedence table for binary expressions.
var opPrec = map[uint]uint{
	'*':  5,
	'/':  5,
	'%':  5,
	SHL:  5,
	SHR:  5,
	'&':  5,
	ANDN: 5,
	'+':  4,
	'-':  4,
	'|':  4,
	'^':  4,
	EQ:   3,
	NE:   3,
	LT:   3,
	LE:   3,
	GT:   3,
	GE:   3,
	AND:  2,
	OR:   1,
}

type ConstValue struct {
	Off   int // position of the leftmost subexpression, or literal of the value
	Typ   Type
	Value Value
}

func (c *ConstValue) Position() int { return c.Off }
func (c *ConstValue) Type() Type    { return c.Typ }

// Builtin values
const (
	BUILTIN_NIL = iota
	BUILTIN_IOTA
)

type BuiltinValue uint

type Bool bool

type Rune rune

type UntypedInt struct {
	*big.Int
}

type Int uint64

type UntypedFloat struct {
	*big.Float
}
type Float float64

type UntypedComplex struct {
	Re *big.Float
	Im *big.Float
}

type Complex complex128

type String string

type OperandName struct {
	Off  int
	Decl Symbol
	Typ  Type
}

func (x *OperandName) Position() int { return x.Off }
func (x *OperandName) Type() Type    { return x.Typ }

type KeyedElement struct {
	Key Expr
	Elt Expr
}

type CompLiteral struct {
	Off  int // position of the type, or the opening brace
	Typ  Type
	Elts []*KeyedElement
}

func (x *CompLiteral) Position() int { return x.Off }
func (x *CompLiteral) Type() Type    { return x.Typ }

type Call struct {
	Off  int // position of the func expression
	Func Expr
	Typ  Type
	Xs   []Expr
	Ell  bool
}

func (x *Call) Position() int { return x.Off }
func (x *Call) Type() Type    { return x.Typ }

type Conversion struct {
	Off int // position of the type
	Typ Type
	X   Expr
}

func (x *Conversion) Position() int { return x.Off }
func (x *Conversion) Type() Type    { return x.Typ }

type MethodExpr struct {
	Off  int
	RTyp Type
	Id   string
	Typ  Type
}

func (x *MethodExpr) Position() int { return x.Off }
func (x *MethodExpr) Type() Type    { return x.Typ }

type ParensExpr struct {
	Off int // position of the opening parenthesis
	X   Expr
}

func (x *ParensExpr) Position() int { return x.Off }
func (*ParensExpr) Type() Type      { panic("not reached") }

type TypeAssertion struct {
	Off int // position of the expression
	Typ Type
	X   Expr
}

func (x *TypeAssertion) Position() int { return x.Off }
func (x *TypeAssertion) Type() Type    { return x.Typ }

type Selector struct {
	Off int // position of the struct expression
	X   Expr
	Id  string
	Typ Type
}

func (x *Selector) Position() int { return x.Off }
func (x *Selector) Type() Type    { return x.Typ }

type IndexExpr struct {
	Off  int // position of the indexed expression
	X, I Expr
	Typ  Type
}

func (x *IndexExpr) Position() int { return x.Off }
func (x *IndexExpr) Type() Type    { return x.Typ }

type SliceExpr struct {
	Off         int // position of the sliced expression
	X           Expr
	Lo, Hi, Cap Expr
	Typ         Type
}

func (x *SliceExpr) Position() int { return x.Off }
func (x *SliceExpr) Type() Type    { return x.Typ }

type UnaryExpr struct {
	Off int // position of the unary operator
	Op  uint
	X   Expr
	Typ Type
}

func (x *UnaryExpr) Position() int { return x.Off }
func (x *UnaryExpr) Type() Type    { return x.Typ }

type BinaryExpr struct {
	Off  int // position of the left operand
	Op   uint
	X, Y Expr
	Typ  Type
}

func (x *BinaryExpr) Position() int { return x.Off }
func (x *BinaryExpr) Type() Type    { return x.Typ }

//
// Types
//

// The TypeVar type represents type variables, used by the semantic checker
// during type inference phase.
type TypeVar struct {
	Off  int // position of the varible identifier or the expression
	File *File
	Name string
	Type Type
}

func (t *TypeVar) Position() int { return t.Off }

type QualifiedId struct {
	Off     int
	Pkg, Id string
}

func (t *QualifiedId) Position() int { return t.Off }
func (*QualifiedId) Type() Type      { panic("not reached") }

const (
	BUILTIN_VOID_TYPE = iota
	BUILTIN_NIL_TYPE
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

func (t *BuiltinType) Position() int { return -1 }

type ArrayType struct {
	Off int // position of the opening bracket
	Dim Expr
	Elt Type
}

func (t *ArrayType) Position() int { return t.Off }

type SliceType struct {
	Off int // position of the opening bracket
	Elt Type
}

func (t *SliceType) Position() int { return t.Off }

type PtrType struct {
	Off  int // positio of the asterisk
	Base Type
}

func (t *PtrType) Position() int { return t.Off }

type MapType struct {
	Off      int // position of the `map` keyword
	Key, Elt Type
}

func (t *MapType) Position() int { return t.Off }

type ChanType struct {
	Off        int // position of the `chan` keyword or the `<-` token
	Send, Recv bool
	Elt        Type
}

func (t *ChanType) Position() int { return t.Off }

type Field struct {
	Off  int // position of the identifier, or type, for anonymous fields
	Name string
	Type Type
	Tag  string
}

func (f *Field) Position() int { return f.Off }

type StructType struct {
	Off    int // position of the `struct` keyword
	Fields []Field
}

func (t *StructType) Position() int { return t.Off }

type TupleType struct {
	Off  int
	Type []Type
}

func (t *TupleType) Position() int { return t.Off }

type Param struct {
	Off  int // position of the identifier
	Name string
	Type Type
}

func (p *Param) Position() int { return p.Off }

type FuncType struct {
	Off     int // position of the `func` keyword
	Params  []Param
	Returns []Param
	Var     bool
}

func (t *FuncType) Position() int { return t.Off }

type InterfaceType struct {
	Off      int // position of the `interface` keyword
	Embedded []Type
	Methods  []*FuncDecl
}

func (t *InterfaceType) Position() int { return t.Off }

//
// Statements
//
type EmptyStmt struct {
	// position of the semilocon, or the closing brace following an empty
	// statement
	Off int
}

func (s *EmptyStmt) Position() int { return s.Off }

type Block struct {
	blockScope
	// position of the opening brace, or the colon preceding a statement list
	// in a swicth or a select case
	Off  int
	Body []Stmt
}

func (s *Block) Position() int { return s.Off }

type Label struct {
	Off   int // position of the label
	Label string
	Stmt  Stmt
	Blk   Scope
}

func (s *Label) Position() int { return s.Off }

type GoStmt struct {
	Off int // position of the `go` keyword
	X   Expr
}

func (s *GoStmt) Position() int { return s.Off }

type ReturnStmt struct {
	Off int // position of the `return` keyword
	Xs  []Expr
}

func (s *ReturnStmt) Position() int { return s.Off }

type BreakStmt struct {
	Off   int // position of the `break` keyword
	Label string
	Dst   Stmt
}

func (s *BreakStmt) Position() int { return s.Off }

type ContinueStmt struct {
	Off   int // position of the `continue` keyword
	Label string
	Dst   Stmt
}

func (s *ContinueStmt) Position() int { return s.Off }

type GotoStmt struct {
	Off   int // position of the `goto` keyword
	Label string
	Dst   Stmt
}

func (s *GotoStmt) Position() int { return s.Off }

type FallthroughStmt struct {
	Off int // position of the `fallthrough` keyword
	Dst Stmt
}

func (s *FallthroughStmt) Position() int { return s.Off }

type SendStmt struct {
	Off int // position of the channel expression
	Ch  Expr
	X   Expr
}

func (s *SendStmt) Position() int { return s.Off }

type RecvStmt struct {
	Off  int // position of the leftmost identifier, or position of the receive expression
	Op   uint
	X, Y Expr
	Rcv  Expr
}

func (s *RecvStmt) Position() int { return s.Off }

type IncStmt struct {
	Off int // position of the expression
	X   Expr
}

func (s *IncStmt) Position() int { return s.Off }

type DecStmt struct {
	Off int // position of the expression
	X   Expr
}

func (s *DecStmt) Position() int { return s.Off }

type AssignStmt struct {
	Off int // position of the first element on the LHS
	Op  uint
	LHS []Expr
	RHS []Expr
}

func (s *AssignStmt) Position() int { return s.Off }

type ExprStmt struct {
	Off int // position of the expression
	X   Expr
}

func (s *ExprStmt) Position() int { return s.Off }

type IfStmt struct {
	blockScope
	Off  int // position of the `if` keyword
	Init Stmt
	Cond Expr
	Then *Block
	Else Stmt
}

func (s *IfStmt) Position() int { return s.Off }

type ForStmt struct {
	blockScope
	Off  int // position of the `for`  keyword
	Init Stmt
	Cond Expr
	Post Stmt
	Blk  *Block
}

func (s *ForStmt) Position() int { return s.Off }

type ForRangeStmt struct {
	blockScope
	Off   int // position of the `for` keyword
	Op    uint
	LHS   []Expr
	Range Expr
	Blk   *Block
}

func (s *ForRangeStmt) Position() int { return s.Off }

type DeferStmt struct {
	Off int // position of the `defer` keyword
	X   Expr
}

func (s *DeferStmt) Position() int { return s.Off }

type ExprCaseClause struct {
	Xs  []Expr
	Blk *Block
}

type ExprSwitchStmt struct {
	blockScope
	Off   int // position of the `switch` keyword
	Init  Stmt
	X     Expr
	Cases []ExprCaseClause
}

func (s *ExprSwitchStmt) Position() int { return s.Off }

type TypeCaseClause struct {
	Types []Type
	Blk   *Block
}

type TypeSwitchStmt struct {
	blockScope
	Off   int // position of the `switch` keyword
	Init  Stmt
	Id    string
	X     Expr
	Cases []TypeCaseClause
}

func (s *TypeSwitchStmt) Position() int { return s.Off }

type CommClause struct {
	Comm Stmt
	Blk  *Block
}

type SelectStmt struct {
	blockScope
	Off   int // position of the `select` keyword
	Comms []CommClause
}

func (s *SelectStmt) Position() int { return s.Off }
