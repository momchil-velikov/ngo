package typecheck

import (
	"bytes"
	"fmt"
	"golang/ast"
)

// The ErrorPos is used to attach a source position to another error.
type ErrorPos struct {
	Off  int
	File *ast.File
	Err  error
}

func (e *ErrorPos) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: %s", e.File.Name, ln, col, e.Err.Error())
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
	Type ast.Type
}

func (e *BadAnonType) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: `%s` is not a valid anonymous field type",
		e.File.Name, ln, col, e.Type)
}

type TypeCheckLoop struct {
	Off  int
	File *ast.File
	Loop []*ast.TypeDecl
}

func (e *TypeCheckLoop) Error() string {
	b := bytes.Buffer{}
	ln, col := e.File.SrcMap.Position(e.Off)
	fmt.Fprintf(&b, "%s:%d:%d: invalid recursive type\n", e.File.Name, ln, col)
	i := 0
	for ; i < len(e.Loop)-1; i++ {
		fmt.Fprintf(&b, "\t `%s` depends on `%s`\n", e.Loop[i].Name, e.Loop[i+1].Name)
	}
	fmt.Fprintf(&b, "\t `%s` depends on `%s`", e.Loop[i].Name, e.Loop[0].Name)
	return b.String()
}

type BadEmbed struct {
	Type *ast.TypeDecl
}

func (e *BadEmbed) Error() string {
	file := e.Type.File
	ln, col := file.SrcMap.Position(e.Type.Off)
	return fmt.Sprintf("%s:%d:%d: interace embeds non-interface type `%s`",
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
	return fmt.Sprintf("%s:%d:%d: `%s` is not a valid constant type",
		e.File.Name, ln, col, e.Type)
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
	fmt.Fprintf(&txt, "%s:%d:%d: duplicate method name `%s`\n",
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
	fmt.Fprintf(&txt, "%s:%d:%d: method name `%s` conflicts with field name\n",
		f.Name, ln, col, e.M.Name)
	off, f = e.S.DeclaredAt()
	ln, col = f.SrcMap.Position(off)
	fmt.Fprintf(&txt, "%s:%d:%d: in the declaration of type `%s`",
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
	fmt.Fprintf(&txt, "%s:%d:%d: in declaration of `%s`: duplicate method name `%s`\n",
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

// The SingleValueContext error is returned whenever a multi-value expression
// is used in a single-value context.
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
	return fmt.Sprintf("%s:%d:%d: ambiguous selector `%s`\n", e.File.Name, ln, col, e.Name)
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
	return fmt.Sprintf("%s:%d:%d: type does not have a %s named `%s`",
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
	return fmt.Sprintf("%s:%d:%d: invalid use of `.(type)` outside type switch",
		e.File.Name, ln, col)
}

// The NotConst error is returned whenever an expression does not evaluate to
// a constant value, required by context.
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
	return fmt.Sprintf("%s:%d:%d: argument count mismatch", e.File.Name, ln, col)
}

// The BadVariadicCall error is returned for call expressions using `...` in
// argument to call a non-variadic function.
type BadVariadicCall struct {
	Off  int
	File *ast.File
}

func (e *BadVariadicCall) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: invalid use of `...` to call a non-variadic function",
		e.File.Name, ln, col)
}

// The BadVariadicArg error is returned for call expressions using `...` with
// an argument og of non-slice type.
type BadVariadicArg struct {
	Off  int
	File *ast.File
}

func (e *BadVariadicArg) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: invalid use of `...` with a non-slice argument",
		e.File.Name, ln, col)
}

// The TypeInferLoop error is returned for when an expression type depends
// upon iself.
type TypeInferLoop struct {
	Loop []ast.Symbol
}

func (e *TypeInferLoop) Error() string {
	txt := bytes.Buffer{}
	fmt.Fprintln(&txt, "type inference loop:")
	n := len(e.Loop)
	prev := e.Loop[0]
	for i := 1; i < n; i++ {
		if e.Loop[i] == nil {
			continue
		}
		next := e.Loop[i]
		off, file := prev.DeclaredAt()
		ln, col := file.SrcMap.Position(off)
		fmt.Fprintf(&txt, "\t%s:%d:%d: `%s`uses `%s`\n",
			file.Name, ln, col, prev.Id(), next.Id())
		prev = next
	}

	next := e.Loop[0]
	off, file := prev.DeclaredAt()
	ln, col := file.SrcMap.Position(off)
	fmt.Fprintf(&txt, "\t%s:%d:%d: `%s` uses `%s`\n",
		file.Name, ln, col, prev.Id(), next.Id())
	return txt.String()

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

// The EvalLoop error is returned when the value of a constant or a variable
// eventually depends on itself.
type EvalLoop struct {
	Loop []ast.Symbol
}

func (e *EvalLoop) Error() string {
	txt := bytes.Buffer{}
	fmt.Fprintln(&txt, "evaluation loop:")
	n := len(e.Loop)
	prev := e.Loop[0]
	for i := 1; i < n; i++ {
		if e.Loop[i] == nil {
			continue
		}
		next := e.Loop[i]
		off, file := prev.DeclaredAt()
		ln, col := file.SrcMap.Position(off)
		fmt.Fprintf(&txt, "\t%s:%d:%d: `%s` uses `%s`\n",
			file.Name, ln, col, prev.Id(), next.Id())
		prev = next
	}

	next := e.Loop[0]
	off, file := prev.DeclaredAt()
	ln, col := file.SrcMap.Position(off)
	fmt.Fprintf(&txt, "\t%s:%d:%d: `%s` uses `%s`\n",
		file.Name, ln, col, prev.Id(), next.Id())
	return txt.String()
}

// The BadOperand error is returned whan an operation is not applicable to the
// type of an operand.
type BadOperand struct {
	Off      int
	File     *ast.File
	Op       ast.Operation
	Expected string
	Type     ast.Type
}

func (e *BadOperand) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	if e.Type == nil {
		return fmt.Sprintf("%s:%d:%d: invalid operand to `%s`: operand must be %s",
			e.File.Name, ln, col, e.Op, e.Expected)
	} else {
		return fmt.Sprintf(
			"%s:%d:%d: invalid operand to `%s`: operand must have %s (`%s` given)",
			e.File.Name, ln, col, e.Op, e.Expected, e.Type)
	}
}

// The NotAddressable error is returned for expressions, required by the
// context to have addressable values.
type NotAddressable struct {
	Off  int
	File *ast.File
}

func (e *NotAddressable) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: the expression's value must be addresable",
		e.File.Name, ln, col)
}

// The NegArrayLen error is returned when an array is declared of negetive
// length.
type NegArrayLen struct {
	Off  int
	File *ast.File
}

func (e *NegArrayLen) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: array length must be non-negative",
		e.File.Name, ln, col)
}

// The BadLiteralType error is returned when the type given for a composite
// literal is not an array, slice, struct or map type.
type BadLiteralType struct {
	Off  int
	File *ast.File
	Type ast.Type
}

func (e *BadLiteralType) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: `%s` is not a valid type for composite literals",
		e.File.Name, ln, col, e.Type)
}

// The MissingLiteralType error is returned when a composite literal elides
// type and type elision ios not allowed by the context
type MissingLiteralType struct {
	Off  int
	File *ast.File
}

func (e *MissingLiteralType) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: missing type for composite literal",
		e.File.Name, ln, col)
}

// The MissingMapKey error is returned for map literals, containing elements
// without a key.
type MissingMapKey struct {
	Off  int
	File *ast.File
}

func (e *MissingMapKey) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf(
		"%s:%d:%d: all elements in a map composite literal must have a key",
		e.File.Name, ln, col)
}

// The NotField error is returned for keys in struct composite literals, which
// aren't field names.
type NotField struct {
	Off  int
	File *ast.File
}

func (e *NotField) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: key is not a field name", e.File.Name, ln, col)
}

// The BadUnspecArrayLen error is returned when an array type has unspecified
// length outside of a top-level composite literal context.
type BadUnspecArrayLen struct {
	Off  int
	File *ast.File
}

func (e *BadUnspecArrayLen) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: unspecified array length not allowed",
		e.File.Name, ln, col)
}

// The BadArraySize error is returned when an array index in a composite
// literal or array dimension in array type declaration is not a non-negative
// integer constant, which fits in `int`.
type BadArraySize struct {
	Off  int
	File *ast.File
	What string
}

func (e *BadArraySize) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: array %s must be a non-negative `int` constant",
		e.File.Name, ln, col, e.What)
}

// The IndexOutOfBounds bounds error is returned for an array index, which is
// out of array bounds.
type IndexOutOfBounds struct {
	Off  int
	File *ast.File
	Idx  int64
}

func (e *IndexOutOfBounds) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: index `%d` out of bounds", e.File.Name, ln, col, e.Idx)
}

// The MixedStructLiteral is returned when a struct literal mixes keyed and
// non-keyed initializers.
type MixedStructLiteral struct {
	Off  int
	File *ast.File
}

func (e *MixedStructLiteral) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: struct literal mixes field:value and value initializers",
		e.File.Name, ln, col)
}

// The FieldEltMismatch error is returned when a struct composite literal with
// no keys does not contain exactly one element for each field in the struct
// type.
type FieldEltMismatch struct {
	Off  int
	File *ast.File
}

func (e *FieldEltMismatch) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf(
		"%s:%d:%d: the literal must contain exactly one element for each struct field",
		e.File.Name, ln, col)
}

// The DupLitField is returned for struct literals, which mention the same
// field more than once.
type DupLitField struct {
	Off  int
	File *ast.File
	Name string
}

func (e *DupLitField) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: duplicate field name `%s` in struct literal",
		e.File.Name, ln, col, e.Name)
}

// The DupLitIndex is returned for array or slice literals, which mention the
// same index more than once.
type DupLitIndex struct {
	Off  int
	File *ast.File
	Idx  int64
}

func (e *DupLitIndex) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: duplicate index `%d` in array/slice literal",
		e.File.Name, ln, col, e.Idx)
}

// The DupLitKey errors are returned for map literals, which mention the same
// constant key more than once.
type DupLitKey struct {
	Off  int
	File *ast.File
	Key  interface{}
}

func (e *DupLitKey) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: duplicate key `%v` in map literal",
		e.File.Name, ln, col, e.Key)
}

// The BadIndexedType error is returned for an index expression, where the
// indexed object is not one of array, pointer to array, slice, string, or
// map type.
type BadIndexedType struct {
	Off  int
	File *ast.File
	Type ast.Type
}

func (e *BadIndexedType) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: type `%s` does not support indexing or slicing",
		e.File.Name, ln, col, e.Type)
}

// The BadSliceExpr error is returned for a slice expression with capacity
// operand, where the sliced expression is of a string type.
type BadSliceExpr struct {
	Off  int
	File *ast.File
}

func (e *BadSliceExpr) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: `string` type does not support 3-index slicing",
		e.File.Name, ln, col)
}

// The NotInteger error is returned for index expression, where the index is
// not of an integral type.
type NotInteger struct {
	Off  int
	File *ast.File
	Type ast.Type
}

func (e *NotInteger) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: index must be of integer type (given `%s`)",
		e.File.Name, ln, col, e.Type)
}

// The BadShiftCount error is returned when the right operand of a shift
// expression is neither of an unsigned integer type nor an untyped constat,
// which can be converted to an unsined integer type.
type BadShiftCount struct {
	Off  int
	File *ast.File
}

func (e *BadShiftCount) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: shift count must be unsigned and integer",
		e.File.Name, ln, col)
}

// The BadIota error is returned whan the predeclard `iota` was used outside a
// const declaration
type BadIota struct {
	Off  int
	File *ast.File
}

func (e *BadIota) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: `iota` used outside const declaration",
		e.File.Name, ln, col)
}

// The NotAssignable error is returned when an expression is not asignable to
// a given type.
type NotAssignable struct {
	Off          int
	File         *ast.File
	SType, DType ast.Type
}

func (e *NotAssignable) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: `%s` is not assignable to `%s`",
		e.File.Name, ln, col, e.SType, e.DType)
}

// The NilUse error is returned on attemt to use the value `nil`
type NilUse struct {
	Off  int
	File *ast.File
}

func (e *NilUse) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: use of builtin `nil`", e.File.Name, ln, col)
}

// The BadCompareOperands error is returned for comparison expressions, where
// neither operand is assignable to the type of the other one.
type BadCompareOperands struct {
	Off          int
	File         *ast.File
	XType, YType ast.Type
}

func (e *BadCompareOperands) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: neither `%s` nor `%s` is assignable to the other",
		e.File.Name, ln, col, e.XType, e.YType)
}

// The NotNilComparable error is returned for comparison expressions, where
// one of the operands if` nil` and the other operand is not of a pointer,
// slice, map, channel type, function, or interface type.
type NotNilComparable struct {
	Off  int
	File *ast.File
	Type ast.Type
}

func (e *NotNilComparable) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: `%s` is not comparable to `nil`",
		e.File.Name, ln, col, e.Type)
}

// The NotComparable error is returned when a type is required by the context
// to be equality comparable and it isn't.
type NotComparable struct {
	Off  int
	File *ast.File
	Type ast.Type
}

func (e *NotComparable) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: values of type `%s` cannot be compared for equality",
		e.File.Name, ln, col, e.Type)
}

// The NotOrderede error is returned when a value is used as an operand to onr
// of the oderding operators `<`, `>`, `<=` or `>=` and the type is not
// "ordered".
type NotOrdered struct {
	Off  int
	File *ast.File
	Type ast.Type
}

func (e *NotOrdered) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: values of type `%s` are not ordered",
		e.File.Name, ln, col, e.Type)
}

// The NotSupportedOperation error is returned when an operator is not
// applicable to an operand of the given type.
type NotSupportedOperation struct {
	Off  int
	File *ast.File
	Op   ast.Operation
	Type ast.Type
}

func (e *NotSupportedOperation) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: operation `%s` not supported for `%s`",
		e.File.Name, ln, col, e.Op, e.Type)
}

// The BadBinaryOperands error is returned when the types of a binary
// expression (other than comparison ) are not identical.
type BadBinaryOperands struct {
	Off          int
	File         *ast.File
	Op           ast.Operation
	XType, YType ast.Type
}

func (e *BadBinaryOperands) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: invalid operation `%s`: mismatched types `%s` and `%s`",
		e.File.Name, ln, col, e.Op, e.XType, e.YType)
}

// The NotConvertible error is returned when an expression is not convertible
// to a given type.
type NotConvertible struct {
	Off          int
	File         *ast.File
	SType, DType ast.Type
}

func (e *NotConvertible) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: `%s` is not convertible to `%s`",
		e.File.Name, ln, col, e.SType, e.DType)
}

// The NotInterface error is returned for type assertions where the expression
// is not of an interface type.
type NotInterface struct {
	Off  int
	File *ast.File
	Type ast.Type
}

func (e *NotInterface) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: invalid type asserion: `%s` is not an interface type",
		e.File.Name, ln, col, e.Type)
}

// The DoesNotImplement error is returned when a concrete type is required to
// implement an interface type and it doesn't.
type DoesNotImplement struct {
	Off       int
	File      *ast.File
	Type, Ifc ast.Type
}

func (e *DoesNotImplement) Error() string {
	ln, col := e.File.SrcMap.Position(e.Off)
	return fmt.Sprintf("%s:%d:%d: type `%s` does not implement interface `%s`",
		e.File.Name, ln, col, e.Type, e.Ifc)
}
