package ast

import s "golang/scanner"

type Node interface {
	Format(*FormatContext, uint)
	Position() int
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

// Source file
type File struct {
	Off      int // position of the "package" keyword
	Package  string
	Imports  []Decl
	Decls    []Decl
	Comments []Comment
	Name     string
	SrcMap   s.SourceMap
}

// Declaration group
type DeclGroup struct {
	Off   int // position of the group keyword ("type", "import", etc)
	Kind  uint
	Decls []Decl
}

func (d DeclGroup) decl()          {}
func (d DeclGroup) stmt()          {}
func (d *DeclGroup) Position() int { return d.Off }

type Import struct {
	Off  int
	Name string
	Path []byte
}

func (d Import) decl()          {}
func (i *Import) Position() int { return i.Off }

// Universal error node
type Error struct {
	Off int
}

func (e Error) decl()          {}
func (e Error) typ()           {}
func (e Error) expr()          {}
func (e Error) stmt()          {}
func (e *Error) Position() int { return e.Off } // FIXME: text position

//
// Declarations
//
type TypeDecl struct {
	Off  int
	Name string
	Type Type
}

func (d TypeDecl) decl()          {}
func (d TypeDecl) stmt()          {}
func (d *TypeDecl) Position() int { return d.Off }

type ConstDecl struct {
	Names  []*Ident
	Type   Type
	Values []Expr
}

func (d ConstDecl) decl() {}
func (d ConstDecl) stmt() {}

func (d *ConstDecl) Position() int {
	return d.Names[0].Position()
}

type VarDecl struct {
	Names []*Ident
	Type  Type
	Init  []Expr
}

func (v VarDecl) decl() {}
func (v VarDecl) stmt() {}
func (v *VarDecl) Position() int {
	return v.Names[0].Position()
}

type Receiver struct {
	Name *Ident
	Type Type
}

type FuncDecl struct {
	Off  int
	Name *Ident
	Recv *Receiver
	Sig  *FuncType
	Blk  *Block
}

func (d FuncDecl) decl()          {}
func (f *FuncDecl) Position() int { return f.Off }

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

func (e Literal) expr()          {}
func (x *Literal) Position() int { return x.Off }

type Element struct {
	Key   Expr
	Value Expr
}

type CompLiteral struct {
	Type Type
	Elts []*Element
}

func (e CompLiteral) expr()          {}
func (x *CompLiteral) Position() int { return x.Type.Position() }

type Call struct {
	Func Expr
	Type Type
	Xs   []Expr
	Ell  bool
}

func (e Call) expr()          {}
func (x *Call) Position() int { return x.Func.Position() }

type Conversion struct {
	Type Type
	X    Expr
}

func (e Conversion) expr()          {}
func (x *Conversion) Position() int { return x.Type.Position() }

type MethodExpr struct {
	Type Type
	Id   string
}

func (e MethodExpr) expr()          {}
func (x *MethodExpr) Position() int { return x.Type.Position() }

type ParensExpr struct {
	Off int
	X   Expr
}

func (e ParensExpr) expr()          {}
func (x *ParensExpr) Position() int { return x.Off }

type FuncLiteral struct {
	Sig *FuncType
	Blk *Block
}

func (e FuncLiteral) expr()          {}
func (x *FuncLiteral) Position() int { return x.Sig.Position() }

type TypeAssertion struct {
	Type Type
	X    Expr
}

func (e TypeAssertion) expr()          {}
func (x *TypeAssertion) Position() int { return x.X.Position() }

type Selector struct {
	X  Expr
	Id string
}

func (e Selector) expr()          {}
func (x *Selector) Position() int { return x.X.Position() }

type IndexExpr struct {
	X, I Expr
}

func (e IndexExpr) expr()          {}
func (x *IndexExpr) Position() int { return x.X.Position() }

type SliceExpr struct {
	X           Expr
	Lo, Hi, Cap Expr
}

func (e SliceExpr) expr()          {}
func (x *SliceExpr) Position() int { return x.X.Position() }

type UnaryExpr struct {
	Off int
	Op  uint
	X   Expr
}

func (ex UnaryExpr) expr()         {}
func (x *UnaryExpr) Position() int { return x.Off }

type BinaryExpr struct {
	Op   uint
	X, Y Expr
}

func (ex BinaryExpr) expr()         {}
func (x *BinaryExpr) Position() int { return x.X.Position() }

//
// Types
//
type Ident struct {
	Off     int
	Pkg, Id string
}

func (t Ident) typ()           {}
func (t Ident) expr()          {}
func (t *Ident) Position() int { return t.Off }

type ArrayType struct {
	Off int
	Dim Expr
	Elt Type
}

func (t ArrayType) typ()           {}
func (t *ArrayType) Position() int { return t.Off }

type SliceType struct {
	Off int
	Elt Type
}

func (t SliceType) typ()           {}
func (t *SliceType) Position() int { return t.Off }

type PtrType struct {
	Off  int
	Base Type
}

func (t PtrType) typ()           {}
func (t *PtrType) Position() int { return t.Off }

type MapType struct {
	Off      int
	Key, Elt Type
}

func (t MapType) typ()           {}
func (t *MapType) Position() int { return t.Off }

type ChanType struct {
	Off        int
	Send, Recv bool
	Elt        Type
}

func (t ChanType) typ()           {}
func (t *ChanType) Position() int { return t.Off }

type FieldDecl struct {
	Names []*Ident
	Type  Type
	Tag   []byte
}

type StructType struct {
	Off    int
	Fields []*FieldDecl
}

func (t StructType) typ()           {}
func (t *StructType) Position() int { return t.Off }

type ParamDecl struct {
	Off   int
	Names []*Ident
	Type  Type
	Var   bool
}

type FuncType struct {
	Off     int
	Params  []*ParamDecl
	Returns []*ParamDecl
}

func (f FuncType) typ()           {}
func (t *FuncType) Position() int { return t.Off }

type MethodSpec struct {
	Name *Ident
	Type Type
}

type InterfaceType struct {
	Off     int
	Methods []*MethodSpec
}

func (t InterfaceType) typ()           {}
func (t *InterfaceType) Position() int { return t.Off }

//
// Statements
//
type EmptyStmt struct {
	Off int
}

func (e EmptyStmt) stmt()          {}
func (s *EmptyStmt) Position() int { return s.Off }

type Block struct {
	Begin, End int // positions of the opening and the closing brace
	Body       []Stmt
}

func (b Block) stmt()          {}
func (b *Block) Position() int { return b.Begin }

type LabeledStmt struct {
	Off   int
	Label string
	Stmt  Stmt
}

func (l LabeledStmt) stmt()          {}
func (l *LabeledStmt) Position() int { return l.Off }

type GoStmt struct {
	Off int
	X   Expr
}

func (b GoStmt) stmt()          {}
func (g *GoStmt) Position() int { return g.Off }

type ReturnStmt struct {
	Off int
	Xs  []Expr
}

func (b ReturnStmt) stmt()          {}
func (s *ReturnStmt) Position() int { return s.Off }

type BreakStmt struct {
	Off   int
	Label string
}

func (b BreakStmt) stmt()          {}
func (s *BreakStmt) Position() int { return s.Off }

type ContinueStmt struct {
	Off   int
	Label string
}

func (b ContinueStmt) stmt()          {}
func (s *ContinueStmt) Position() int { return s.Off }

type GotoStmt struct {
	Off   int
	Label string
}

func (b GotoStmt) stmt()          {}
func (s *GotoStmt) Position() int { return s.Off }

type FallthroughStmt struct {
	Off int
}

func (b FallthroughStmt) stmt()          {}
func (s *FallthroughStmt) Position() int { return s.Off }

type SendStmt struct {
	Ch Expr
	X  Expr
}

func (b SendStmt) stmt()          {}
func (s *SendStmt) Position() int { return s.Ch.Position() }

type IncStmt struct {
	X Expr
}

func (b IncStmt) stmt()          {}
func (s *IncStmt) Position() int { return s.X.Position() }

type DecStmt struct {
	X Expr
}

func (b DecStmt) stmt()          {}
func (s *DecStmt) Position() int { return s.X.Position() }

type AssignStmt struct {
	Op  uint
	LHS []Expr
	RHS []Expr
}

func (a AssignStmt) stmt()          {}
func (s *AssignStmt) Position() int { return s.LHS[0].Position() }

type ExprStmt struct {
	X Expr
}

func (e ExprStmt) stmt()          {}
func (s *ExprStmt) Position() int { return s.X.Position() }

type IfStmt struct {
	Off  int
	Init Stmt
	Cond Expr
	Then *Block
	Else Stmt
}

func (i *IfStmt) stmt()         {}
func (s *IfStmt) Position() int { return s.Off }

type ForStmt struct {
	Off  int
	Init Stmt
	Cond Expr
	Post Stmt
	Blk  *Block
}

func (f ForStmt) stmt()          {}
func (s *ForStmt) Position() int { return s.Off }

type ForRangeStmt struct {
	Off   int
	Op    uint
	LHS   []Expr
	Range Expr
	Blk   *Block
}

func (f ForRangeStmt) stmt()          {}
func (s *ForRangeStmt) Position() int { return s.Off }

type DeferStmt struct {
	Off int
	X   Expr
}

func (d DeferStmt) stmt()          {}
func (s *DeferStmt) Position() int { return s.Off }

type ExprCaseClause struct {
	Xs   []Expr
	Body []Stmt
}

type ExprSwitchStmt struct {
	Off        int // position of the `switch` keyword
	Begin, End int // positions of the opening and the closing braces
	Init       Stmt
	X          Expr
	Cases      []ExprCaseClause
}

func (d ExprSwitchStmt) stmt()          {}
func (s *ExprSwitchStmt) Position() int { return s.Off }

type TypeCaseClause struct {
	Types []Type
	Body  []Stmt
}

type TypeSwitchStmt struct {
	Off        int // position of the `switch` keyword
	Begin, End int // positions of the opening and the closing braces
	Init       Stmt
	Id         string
	X          Expr
	Cases      []TypeCaseClause
}

func (d TypeSwitchStmt) stmt()          {}
func (s *TypeSwitchStmt) Position() int { return s.Off }

type CommClause struct {
	Comm Stmt
	Body []Stmt
}

type SelectStmt struct {
	Off        int // position of the `select` keyword
	Begin, End int // positions of the opening and the closing braces
	Comms      []CommClause
}

func (s SelectStmt) stmt()          {}
func (s *SelectStmt) Position() int { return s.Off }
