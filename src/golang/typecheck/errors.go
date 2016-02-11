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
	Off  int
	File *ast.File
	Type *ast.TypeDecl
}

func (e *BadEmbed) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: interace embeds non-interface type %s",
		e.File.Name, ln, col, e.Type.Name)
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
