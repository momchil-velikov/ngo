package ast

type Decl interface {
    Formatter
    decl()
}

type File struct {
    PackageName string
    Imports     []Import
    Decls       []Decl
}

type Import struct {
    Name string
    Path string
}

// Declarations
type TypeDecl struct {
    Name string
    Type TypeSpec
}

func (d TypeDecl) decl() {}

type ConstDecl struct {
    Names  []string
    Type   TypeSpec
    Values []*Expr
}

func (d ConstDecl) decl() {}

type ConstGroup struct {
    Decls []*ConstDecl
}

func (d ConstGroup) decl() {}

type VarDecl struct {
    Names []string
    Type  TypeSpec
    Init  []*Expr
}

func (v VarDecl) decl() {}

type VarGroup struct {
    Decls []*VarDecl
}

func (d VarGroup) decl() {}

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

type Block struct {
}

// Expressions
type Expr struct {
    Const string
}

// Types
type TypeSpec interface {
    Formatter
    typeSpec()
}

type TypeName struct {
    Pkg, Id string
}

func (t TypeName) typeSpec() {}

type ArrayType struct {
    Dim     *Expr
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
    Name string
    Type TypeSpec
    Tag  string
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
    Embed   []*TypeName
    Methods []*MethodSpec
}

func (t InterfaceType) typeSpec() {}
