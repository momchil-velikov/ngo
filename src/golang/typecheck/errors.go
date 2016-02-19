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

// The BadMultiValueAssign error is returned for assignments and
// initialization statements where the number of values on the right-hand side
// does not equal the number of locations on the left-hand side.
type BadMultiValueAssign struct {
	Off  int
	File *ast.File
}

func (e *BadMultiValueAssign) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: assignment count mismatch", e.File.Name, ln, col)
}

// The SingleValueContext error is returned whenever a multi-valued expression
// is used in a sigle-valur context statements.
type SingleValueContext struct {
	Off  int
	File *ast.File
}

func (e *SingleValueContext) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: multiple value expression in single-value context",
		e.File.Name, ln, col)
}

// The NotFunc error is returned on attemt to call a non-function.
type NotFunc struct {
	Off  int
	File *ast.File
	X    ast.Expr
}

func (e *NotFunc) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	// FIXME: display the failing expression(?)
	return fmt.Sprintf("%s:%d:%d: called object is not a function", e.File.Name, ln, col)
}

// The AmbiguousSelector error is returned whenever a selector is not unique
// at the shallowest deopth in a type.
type AmbiguousSelector struct {
	Off  int
	File *ast.File
	Name string
}

func (e *AmbiguousSelector) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: ambiguous selector %s\n", e.File.Name, ln, col, e.Name)
}

// The BadReceiverType error is returned whenever a type does not have the
// form `T` or `*T`, where `T` is a typename.
type BadReceiverType struct {
	Off  int
	File *ast.File
}

func (e *BadReceiverType) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: receiver type must have the form `T` or `*T`",
		e.File.Name, ln, col)
}

// The NotFound error is returned whenever a selector expression refers to a
// field or method name not present in the type.
type NotFound struct {
	Off        int
	File       *ast.File
	What, Name string
}

func (e *NotFound) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: type does not have a %s named %s",
		e.File.Name, ln, col, e.What, e.Name)
}

// The BadMethodExpr error is returned whenever the method is declared with a
// pointer receiver, but the type is not a pointer type neither the method was
// promoted through an anonymous pointer member.
type BadMethodExpr struct {
	Off  int
	File *ast.File
}

func (e *BadMethodExpr) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf(
		"%s:%d:%d: in the method expression the method must have a pointer receiver",
		e.File.Name, ln, col)
}

// The BadTypeAssertion error is returned for use of `.(type)` outside as
// switch statement. The construct `x.(type)` is parsed as a TypeAssertion
// expression with `nil` type. Such an expression is deconstructed when
// parsing a type switch statement, thus all remaining occurances are invalid.
type BadTypeAssertion struct {
	Off  int
	File *ast.File
}

func (e *BadTypeAssertion) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: invalid use if .(type) outside type switch",
		e.File.Name, ln, col)
}

// The NotConst error is returned whenever an expression does not evaluate to
// a constant value, requiored by context.
type NotConst struct {
	Off  int
	File *ast.File
	What string
}

func (e *NotConst) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: %s is not a constant", e.File.Name, ln, col, e.What)
}

// The BadTypeArg error is returned for call expressions, which do not allow
// "type" argument.
type BadTypeArg struct {
	Off  int
	File *ast.File
}

func (e *BadTypeArg) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: type argument not allowed", e.File.Name, ln, col)
}

// The BadArgNumber error is returned for call expressions, where the number
// of arguments does not match the number of parameteres.
type BadArgNumber struct {
	Off  int
	File *ast.File
}

func (e *BadArgNumber) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: argument count mismatch allowed", e.File.Name, ln, col)
}

// The TypeInferLoop error is returned for when an expression type depends
// upon iself.
type TypeInferLoop struct {
	Off  int
	File *ast.File
}

func (e *TypeInferLoop) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: type inference loop", e.File.Name, ln, col)
}

// The ExprLoop error is returned for expressions, which depend on itself.
type ExprLoop struct {
	Off  int
	File *ast.File
}

func (e *ExprLoop) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: expression evaluation loop", e.File.Name, ln, col)
}
