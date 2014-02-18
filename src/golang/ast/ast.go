package ast

type Decl interface {
    Formatter
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

type TypeSpec interface {
    Formatter
    typeSpec()
}

type TypeDecl struct {
    Name string
    Type TypeSpec
}

type Expr struct {
    Const string
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
