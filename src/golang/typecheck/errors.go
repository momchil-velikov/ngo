package typecheck

import (
	"bytes"
	"fmt"
	"golang/ast"
)

type BadMapKey struct {
	Off  int
	File *ast.File
}

func (e *BadMapKey) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: invalid type for map key", e.File.Name, ln, col)
}

type DupFieldName struct {
	Field *ast.Field
	File  *ast.File
}

func (e *DupFieldName) Error() string {
	ln, col := e.File.SrcMap.Position(e.Field.Off)
	return fmt.Sprintf("%s:%d:%d: non-unique field name `%s`",
		e.File.Name, ln, col, fieldName(e.Field))
}

// The BadAnonType is returned whenever a type is not allowed for an
// anonymous field.
type BadAnonType struct {
	Off  int
	File *ast.File
	What string
}

func (e *BadAnonType) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: embedded type cannot be a %s",
		e.File.Name, ln, col, e.What)
}

type TypeCheckLoop struct {
	Off  int
	File *ast.File
	Loop []ast.Symbol
}

func (e *TypeCheckLoop) Error() string {
	b := bytes.Buffer{}
	ln, col := e.File.SrcMap.Position(e.Off)
	fmt.Fprintf(&b, "%s:%d:%d: invalid recursive type\n", e.File.Name, ln, col)
	i := 0
	for ; i < len(e.Loop)-1; i++ {
		fmt.Fprintf(&b, "\t %s depends on %s\n", e.Loop[i].Id(), e.Loop[i+1].Id())
	}
	fmt.Fprintf(&b, "\t %s depends on %s", e.Loop[i].Id(), e.Loop[0].Id())
	return b.String()
}

type BadEmbed struct {
	Type *ast.TypeDecl
}

func (e *BadEmbed) Error() string {
	file := e.Type.File
	ln, col := file.SrcMap.Position(e.Type.Off)
	return fmt.Sprintf("%s:%d:%d: interace embeds non-interface type %s",
		file.Name, ln, col, e.Type.Name)
}

// The BadConstType error is returned when the declared type of a constant is
// not one of the allowed builtin types.
type BadConstType struct {
	Off  int
	File *ast.File
	Type ast.Type
}

func (e *BadConstType) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: type is invalid for constant declaration",
		e.File.Name, ln, col) // FIXME: describe the type
}

// The DupMethodName error is returned whenever a method name is not unique
// in the method set of a type.
type DupMethodName struct {
	M0, M1 *ast.FuncDecl
}

func (e *DupMethodName) Error() string {
	off, f := e.M0.DeclaredAt()
	ln, col := f.SrcMap.Position(off)
	txt := bytes.Buffer{}
	fmt.Fprintf(&txt, "%s:%d:%d: duplicate method name %s\n",
		f.Name, ln, col, e.M0.Name)
	off, f = e.M1.DeclaredAt()
	ln, col = f.SrcMap.Position(off)
	fmt.Fprintf(&txt, "%s:%d:%d: location of previous declaration", f.Name, ln, col)
	return txt.String()
}

// The DupFieldMethodName error is returned whenever a method has the same
// name as a field.
type DupFieldMethodName struct {
	M *ast.FuncDecl
	S *ast.TypeDecl
}

func (e *DupFieldMethodName) Error() string {
	off, f := e.M.DeclaredAt()
	ln, col := f.SrcMap.Position(off)
	txt := bytes.Buffer{}
	fmt.Fprintf(&txt, "%s:%d:%d: method name %s conflicts with field name\n",
		f.Name, ln, col, e.M.Name)
	off, f = e.S.DeclaredAt()
	ln, col = f.SrcMap.Position(off)
	fmt.Fprintf(&txt, "%s:%d:%d: in the declaration of type %s",
		f.Name, ln, col, e.S.Name)
	return txt.String()
}

// The DupIfacedMethodName error is returned whenever a method name is not
// unique amonth the method set of an interface type.
type DupIfaceMethodName struct {
	Decl         *ast.TypeDecl
	Off0, Off1   int
	File0, File1 *ast.File
	Name         string
}

func (e *DupIfaceMethodName) Error() string {
	ln, col := e.Decl.File.SrcMap.Position(e.Decl.Off)
	txt := bytes.Buffer{}
	fmt.Fprintf(&txt, "%s:%d:%d: in declaration of %s: duplicate method name %s\n",
		e.Decl.File.Name, ln, col, e.Decl.Name, e.Name)
	ln, col = e.File0.SrcMap.Position(e.Off0)
	fmt.Fprintf(&txt, "%s:%d:%d: declared here\n", e.File0.Name, ln, col)
	ln, col = e.File1.SrcMap.Position(e.Off1)
	fmt.Fprintf(&txt, "%s:%d:%d: and here", e.File1.Name, ln, col)
	return txt.String()
}
