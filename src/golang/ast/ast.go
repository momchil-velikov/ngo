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

type BaseType struct {
	Pkg, Id string
}

func (t BaseType) typeSpec() {}

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
