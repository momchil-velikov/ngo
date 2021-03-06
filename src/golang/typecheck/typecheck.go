package typecheck

import "golang/ast"

func CheckPackage(pkg *ast.Package) error {
	if err := verifyTypes(pkg); err != nil {
		return err
	}
	if err := inferTypes(pkg); err != nil {
		return err
	}
	if err := verifyExprs(pkg); err != nil {
		return err
	}

	return nil
}

// Returns the first type literal in a chain of TypeNames.
func underlyingType(typ ast.Type) ast.Type {
	for {
		switch t := typ.(type) {
		case *ast.TypeDecl:
			typ = t.Type
		default:
			return t
		}
	}
}

// Returns true if TYP is the empty interface type (`interface{}`).
func isEmptyInterfaceType(typ ast.Type) bool {
	t, ok := underlyingType(typ).(*ast.InterfaceType)
	if !ok || len(t.Methods) != 0 {
		return false
	}
	for i := range t.Embedded {
		if !isEmptyInterfaceType(t.Embedded[i]) {
			return false
		}
	}
	return true
}

// Returns true if `TYP` is a TypeName.
func isNamed(typ ast.Type) bool {
	switch t := typ.(type) {
	case *ast.TypeDecl:
		return true
	case *ast.BuiltinType:
		return t.IsNamed()
	default:
		return false
	}
}

func builtinType(typ ast.Type) *ast.BuiltinType {
	t, _ := underlyingType(typ).(*ast.BuiltinType)
	return t
}

// Returns nil if TYP is "untyped" type
func isUntyped(typ ast.Type) bool {
	if typ == nil {
		return true
	}
	if t := builtinType(typ); t != nil {
		return t.IsUntyped()
	}
	return false
}

func defaultType(typ ast.Type) ast.Type {
	if typ == nil {
		panic("not reached")
	}

	u := underlyingType(typ)
	if t, ok := u.(*ast.BuiltinType); !ok || !t.IsUntyped() {
		return typ
	} else {
		switch t.Kind {
		case ast.BUILTIN_UNTYPED_BOOL:
			return ast.BuiltinBool
		case ast.BUILTIN_UNTYPED_RUNE:
			return ast.BuiltinInt32
		case ast.BUILTIN_UNTYPED_INT:
			return ast.BuiltinInt
		case ast.BUILTIN_UNTYPED_FLOAT:
			return ast.BuiltinFloat64
		case ast.BUILTIN_UNTYPED_COMPLEX:
			return ast.BuiltinComplex128
		case ast.BUILTIN_UNTYPED_STRING:
			return ast.BuiltinString
		default:
			panic("not reached")
		}
	}
}

func fieldName(f *ast.Field) string {
	if len(f.Name) > 0 {
		return f.Name
	}
	typ := f.Type
	if t, ok := f.Type.(*ast.PtrType); ok {
		typ = t.Base
	}
	// Parser guarantees that this type assertion will not fail.
	t := typ.(*ast.TypeDecl)
	return t.Name
}

// Check if a type can be compared with `==` and `!=`.
//
// IMPORTANT: Callers must have ensured that type is not an invalid recursive
// type.
func isEqualityComparable(t ast.Type) bool {
	switch t := t.(type) {
	case *ast.BuiltinType, *ast.PtrType, *ast.ChanType, *ast.InterfaceType:
		return true
	case *ast.SliceType, *ast.MapType, *ast.FuncType:
		return false
	case *ast.ArrayType:
		return isEqualityComparable(t.Elt)
	case *ast.StructType:
		for i := range t.Fields {
			if !isEqualityComparable(t.Fields[i].Type) {
				return false
			}
		}
		return true
	case *ast.TypeDecl:
		return isEqualityComparable(t.Type)
	default:
		panic("not reached")
	}
}

// Returns true if values of the type can be compared with operators `<`, `>`,
// `<=`, or `>=`
// https://golang.org/ref/spec#Comparison_operators
func isOrdered(typ ast.Type) bool {
	t := builtinType(typ)
	if t == nil {
		return false
	}
	return t.IsOrdered()
}

// Return true if the type TYP is equality comparable to `nil`.
func isNilComparable(typ ast.Type) bool {
	switch underlyingType(typ).(type) {
	case *ast.PtrType, *ast.SliceType, *ast.MapType, *ast.ChanType,
		*ast.FuncType, *ast.InterfaceType:
		return true
	default:
		return false
	}
}

// Checks for uniqueness of method names in the method set of the typename
// DCL.
// IMPORTANT: Callers must have ensured that type is not an invalid recursive
// type.
type methodSet map[string]*ast.FuncDecl

func checkMethodUniqueness(dcl *ast.TypeDecl) error {
	// Check for uniqueness of the method names in an interface type,
	// including embedded interfaces.
	if iface, ok := dcl.Type.(*ast.InterfaceType); ok {
		return checkIfaceMethodUniqueness(make(methodSet), dcl, iface)
	}

	// For non-interface types, check methods, declared on this typename.
	set := make(methodSet)
	for i := range dcl.Methods {
		f := dcl.Methods[i]
		g := set[f.Name]
		if g != nil {
			return &DupMethodName{M0: f, M1: g}
		}
		set[f.Name] = f
	}
	for i := range dcl.PMethods {
		f := dcl.PMethods[i]
		g := set[f.Name]
		if g != nil {
			return &DupMethodName{M0: f, M1: g}
		}
		set[f.Name] = f
	}

	// For struct types, additionally check for conflict with field names.
	if str, ok := underlyingType(dcl.Type).(*ast.StructType); ok {
		for i := range str.Fields {
			name := fieldName(&str.Fields[i])
			g := set[name]
			if g != nil {
				return &DupFieldMethodName{M: g, S: dcl}
			}
		}
	}
	return nil
}

// Checks for uniqueness of method names in the method set of the typename
// TOP. The parameter's DCL and IFACE denote either the top interface type
// declaration (same as TOP), or an embedded interface type.

// IMPORTANT: Callers must have ensured that type is not an invalid recursive
// type.
func checkIfaceMethodUniqueness(
	set methodSet, dcl *ast.TypeDecl, iface *ast.InterfaceType) error {

	for _, t := range iface.Embedded {
		ifc := underlyingType(t).(*ast.InterfaceType)
		if err := checkIfaceMethodUniqueness(set, dcl, ifc); err != nil {
			return err
		}
	}
	for _, m := range iface.Methods {
		if d, ok := set[m.Name]; ok {
			return &DupIfaceMethodName{
				Decl:  dcl,
				Off0:  m.Off,
				File0: m.File,
				Off1:  d.Off,
				File1: d.File,
				Name:  m.Name,
			}
		}
		set[m.Name] = m
	}
	return nil
}

// Finds the selector NAME among the fields and methods of the type TYP. If
// there are two or more identically named fields or methods at the same
// shallowest depth, returns any two of them, with the intention for the
// caller to report an ambiguous selector error at the appropriate source
// position. The final `bool` output is true whenever the original type was a
// pointer type, or the field or method was promoted via one or more anonymous
// fields of a pointer type. The selector lookup is done in the context of the
// package PKGS, such non-exported fields and methods, declared in a different
// package than PKG are invisible for the lookup: they are not found and they
// do not cause ambiguous selector errors.
type fieldOrMethod struct {
	F *ast.Field
	M *ast.FuncDecl
}

func findSelector(
	pkg *ast.Package, typ ast.Type, name string) (fieldOrMethod, fieldOrMethod, bool) {

	orig := typ
	// "For a value x of type T or *T where T is not a pointer or interface
	// type, x.f denotes the field or method at the shallowest depth in T."
	ptr, ok := typ.(*ast.PtrType)
	if ok {
		typ = ptr.Base
	}
	switch underlyingType(typ).(type) {
	case *ast.InterfaceType:
	case *ast.PtrType:
	default:
		s0, s1, pp := findFieldOrMethod(pkg, typ, name)
		return s0, s1, pp || (ptr != nil)
	}

	// "For a value x of type I where I is an interface type, x.f denotes the
	// actual method with name f of the dynamic value of x."
	if ifc, ok := underlyingType(orig).(*ast.InterfaceType); ok {
		m := findInterfaceMethod(pkg, ifc, name)
		return fieldOrMethod{M: m}, fieldOrMethod{}, false
	}

	// "As an exception, if the type of x is a named pointer type and (*x).f
	// is a valid selector expression denoting a field (but not a method), x.f
	// is shorthand for (*x).f."
	dcl, ok := orig.(*ast.TypeDecl)
	if ok {
		if ptr, ok := underlyingType(dcl.Type).(*ast.PtrType); ok {
			s0, s1, p := findFieldOrMethod(pkg, ptr.Base, name)
			if s1.F == nil && s1.M == nil {
				// We have found an unambiguous selector. Do not return
				// methods, only fields.
				if s0.F == nil {
					return fieldOrMethod{}, fieldOrMethod{}, false
				}
				return s0, fieldOrMethod{}, p
			}
			// Return ambiguous selector.
			return s0, s1, false
		}
	}

	// "In all other cases, x.f is illegal."
	return fieldOrMethod{}, fieldOrMethod{}, false
}

func findFieldOrMethod(
	pkg *ast.Package, typ ast.Type, name string) (fieldOrMethod, fieldOrMethod, bool) {

	m := findImmediateFieldOrMethod(pkg, typ, name)
	if m.F != nil || m.M != nil {
		return m, fieldOrMethod{}, false
	}
	if str, ok := underlyingType(typ).(*ast.StructType); ok {
		return findPromotedFieldOrMethod(pkg, str, name)
	}
	return fieldOrMethod{}, fieldOrMethod{}, false
}

// Finds and returns the method NAME of the type TYP (which can be a typename
// or an interface type) or, if TYP is declared as a struct type, the field
// NAME.
func findImmediateFieldOrMethod(
	pkg *ast.Package, typ ast.Type, name string) fieldOrMethod {

	if ifc, ok := underlyingType(typ).(*ast.InterfaceType); ok {
		return fieldOrMethod{M: findInterfaceMethod(pkg, ifc, name)}
	}
	if dcl, ok := typ.(*ast.TypeDecl); ok {
		for _, m := range dcl.Methods {
			if name == m.Name && isAccessibleMethod(pkg, m) {
				return fieldOrMethod{M: m}
			}
		}
		for _, m := range dcl.PMethods {
			if name == m.Name && isAccessibleMethod(pkg, m) {
				return fieldOrMethod{M: m}
			}
		}
		typ = dcl.Type
	}
	if str, ok := underlyingType(typ).(*ast.StructType); ok {
		for i := range str.Fields {
			f := &str.Fields[i]
			if name == fieldName(f) && isAccessibleField(pkg, str, name) {
				return fieldOrMethod{F: f}
			}
		}
	}
	return fieldOrMethod{}
}

// Appends the types of the anonymous fields in the structure type DCL to the
// ANON slice.
type anonField struct {
	Ptr  bool
	Type ast.Type
}

func appendAnonFields(str *ast.StructType, ptr bool, anon []anonField) []anonField {
	for i := range str.Fields {
		if len(str.Fields[i].Name) > 0 {
			continue
		}
		anon = append(anon, anonField{Ptr: ptr, Type: str.Fields[i].Type})
	}
	return anon
}

// Finds and returns a promoted field or method NAME in the struct type STR.
func findPromotedFieldOrMethod(
	pkg *ast.Package, str *ast.StructType, name string) (fieldOrMethod, fieldOrMethod, bool) {

	var anon, next []anonField
	anon = appendAnonFields(str, false, nil)
	found, fptr := fieldOrMethod{}, false
	for n := len(anon); n > 0; n = len(anon) {
		for i := 0; i < n; i++ {
			ptr := anon[i].Ptr
			t := anon[i].Type
			if p, ok := anon[i].Type.(*ast.PtrType); ok {
				t = p.Base
				ptr = true
			}
			m := findImmediateFieldOrMethod(pkg, t, name)
			if m.F != nil || m.M != nil {
				if found.F != nil || found.M != nil {
					return found, m, false
				}
				found = m
				fptr = ptr
			}
			if s, ok := underlyingType(t).(*ast.StructType); ok {
				next = appendAnonFields(s, ptr, next)
			}
		}
		if found.F != nil || found.M != nil {
			return found, fieldOrMethod{}, fptr
		}
		anon, next = next, anon[:0]
	}
	return fieldOrMethod{}, fieldOrMethod{}, false
}

// Finds method NAME in the interface type TYP.
func findInterfaceMethod(
	pkg *ast.Package, typ *ast.InterfaceType, name string) *ast.FuncDecl {

	for _, m := range typ.Methods {
		if name == m.Name {
			return m
		}
	}
	for _, t := range typ.Embedded {
		d := t.(*ast.TypeDecl)
		if m := findInterfaceMethod(pkg, d.Type.(*ast.InterfaceType), name); m != nil {
			return m
		}
	}
	return nil
}

// Checks if the method declaration M is accessible from package P.
func isAccessibleMethod(p *ast.Package, m *ast.FuncDecl) bool {
	return ast.IsExported(m.Name) || (m.File != nil && p == m.File.Pkg)
}

// Checks if the field NAME in structure S is accessible from package P.
func isAccessibleField(p *ast.Package, s *ast.StructType, name string) bool {
	return ast.IsExported(name) || (s.File != nil && p == s.File.Pkg)
}

// Finds the field NAME in the structure type STR.
func findField(str *ast.StructType, name string) *ast.Field {
	for i := range str.Fields {
		f := &str.Fields[i]
		if name == fieldName(f) {
			return f
		}
	}
	return nil
}

// Returns true if the expression X is of an arithmetic type, or an untyped
// integral or floating point constsnt.
func isArith(x ast.Expr) bool {
	if t := builtinType(x.Type()); t == nil {
		return false
	} else {
		return t.IsArith()
	}
}

// Returns true of the expression X is addressable.
// "... that is, either a variable, pointer indirection, or slice indexing
// operation; or a field selector of an addressable struct operand; or an
// array indexing operation of an addressable array. As an exception to the
// addressability requirement, x may also be a (possibly parenthesized)
// composite literal."
// https://golang.org/ref/spec#Address_operators
func isAddressable(x ast.Expr) bool {
	_, ok := x.(*ast.CompLiteral)
	return ok || _isAddressable(x)
}

func _isAddressable(x ast.Expr) bool {
	switch x := x.(type) {
	case *ast.OperandName:
		_, ok := x.Decl.(*ast.Var)
		return ok
	case *ast.UnaryExpr:
		return x.Op == '*'
	case *ast.IndexExpr:
		if _, ok := underlyingType(x.X.Type()).(*ast.SliceType); ok {
			return true
		}
		return _isAddressable(x.X)
	case *ast.Selector:
		return _isAddressable(x.X)
	default:
		return false
	}
}

// Returns the method set of the interface type T.
func ifaceMethodSet(t *ast.InterfaceType) methodSet {
	return ifaceMethodSetRec(make(methodSet), t)
}

func ifaceMethodSetRec(set methodSet, t *ast.InterfaceType) methodSet {
	for _, e := range t.Embedded {
		set = ifaceMethodSetRec(set, underlyingType(e).(*ast.InterfaceType))
	}
	for _, m := range t.Methods {
		set[m.Name] = m
	}
	return set
}

// Returns the type of on operation, performed on untyped constants of a
// different kind. The parameters U and V must be untyped, and integer, rune,
// floating point, or complex.
// https://golang.org/ref/spec#Constant_expressions
// " ... if the operands of a binary operation are different kinds of untyped
// constants, the operation and, for non-boolean operations, the result use
// the kind that appears later in this list: integer, rune, floating-point,
// complex."
func promoteUntyped(u *ast.BuiltinType, v *ast.BuiltinType) *ast.BuiltinType {
	if u == v {
		return u
	}

	rank := func(t *ast.BuiltinType) int {
		switch t.Kind {
		case ast.BUILTIN_UNTYPED_INT:
			return 0
		case ast.BUILTIN_UNTYPED_RUNE:
			return 1
		case ast.BUILTIN_UNTYPED_FLOAT:
			return 2
		case ast.BUILTIN_UNTYPED_COMPLEX:
			return 3
		default:
			panic("not reached")
		}
	}

	i, j := rank(u), rank(v)
	if i < j {
		i = j
	}

	return [...]*ast.BuiltinType{
		ast.BuiltinUntypedInt,
		ast.BuiltinUntypedRune,
		ast.BuiltinUntypedFloat,
		ast.BuiltinUntypedComplex,
	}[i]
}

// Returns true if the expression X is definitely a constant expression. If
// any errors occur, the function returns false.
// NOTE: must be called after type inference is complete.
func isConst(x ast.Expr) bool {
	return (*typeInferer).isConst(nil, x)
}

// Returns true if the expression X contains non-constant function calls or
// channel receive expressions. If any errors occur, the function returns
// false.
// NOTE: must be called after type inference is complete.
func hasSideEffects(x ast.Expr) bool {
	return (*typeInferer).hasSideEffects(nil, x)
}
