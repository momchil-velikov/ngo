package ast

type XDecl interface {
	Formatter
}

type File struct {
	PackageName string
	Imports     []Import
	Decls       []XDecl
}

type Import struct {
	Name string
	Path string
}
