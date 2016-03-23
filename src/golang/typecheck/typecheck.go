package typecheck

import "golang/ast"

func CheckPackage(pkg *ast.Package) error {
	if err := verifyTypes(pkg); err != nil {
		return err
	}
	if err := verifyExprs(pkg); err != nil {
		return err
	}

	return nil
}

// Returns the first type literal in a chain of TypeNames.
func unnamedType(typ ast.Type) ast.Type {
	for {
		switch t := typ.(type) {
		case *ast.TypeDecl:
			typ = t.Type
		default:
			return t
		}
	}
}

func singleValueType(typ ast.Type) ast.Type {
	if t, ok := typ.(*ast.TupleType); ok {
		if !t.Strict {
			return t.Type[0]
		} else {
			return nil
		}
	}
	return typ
}

func builtinType(typ ast.Type) *ast.BuiltinType {
	t, _ := unnamedType(typ).(*ast.BuiltinType)
	return t
}

func defaultType(x ast.Expr) ast.Type {
	t := unnamedType(x.Type())
	if t != nil {
		return t
	}

	// Only untyped constants can have a nil type.
	c := x.(*ast.ConstValue)
	switch c.Value.(type) {
	case ast.Bool:
		return ast.BuiltinBool
	case ast.Rune:
		return ast.BuiltinInt32
	case ast.UntypedInt:
		return ast.BuiltinInt
	case ast.UntypedFloat:
		return ast.BuiltinFloat64
	case ast.UntypedComplex:
		return ast.BuiltinComplex128
	case ast.String:
		return ast.BuiltinString
	default:
		panic("not reached")
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
	if str, ok := unnamedType(dcl.Type).(*ast.StructType); ok {
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
		ifc := unnamedType(t).(*ast.InterfaceType)
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
	switch unnamedType(typ).(type) {
	case *ast.InterfaceType:
	case *ast.PtrType:
	default:
		s0, s1, pp := findFieldOrMethod(pkg, typ, name)
		return s0, s1, pp || (ptr != nil)
	}

	// "For a value x of type I where I is an interface type, x.f denotes the
	// actual method with name f of the dynamic value of x."
	if ifc, ok := unnamedType(orig).(*ast.InterfaceType); ok {
		m := findInterfaceMethod(pkg, ifc, name)
		return fieldOrMethod{M: m}, fieldOrMethod{}, false
	}

	// "As an exception, if the type of x is a named pointer type and (*x).f
	// is a valid selector expression denoting a field (but not a method), x.f
	// is shorthand for (*x).f."
	dcl, ok := orig.(*ast.TypeDecl)
	if ok {
		if ptr, ok := unnamedType(dcl.Type).(*ast.PtrType); ok {
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
	if str, ok := unnamedType(typ).(*ast.StructType); ok {
		return findPromotedFieldOrMethod(pkg, str, name)
	}
	return fieldOrMethod{}, fieldOrMethod{}, false
}

// Finds and returns the method NAME of the type TYP (which can be a typename
// or an interface type) or, if TYP is declared as a struct type, the field
// NAME.
func findImmediateFieldOrMethod(
	pkg *ast.Package, typ ast.Type, name string) fieldOrMethod {

	if ifc, ok := unnamedType(typ).(*ast.InterfaceType); ok {
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
	if str, ok := unnamedType(typ).(*ast.StructType); ok {
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
			if s, ok := unnamedType(t).(*ast.StructType); ok {
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
	typ := x.Type()
	if typ == nil {
		switch x.(*ast.ConstValue).Value.(type) {
		case ast.Bool, ast.String:
			return false
		default:
			return true
		}
	}
	t := builtinType(typ)
	if t == nil {
		return false
	}
	return t.IsArith()
}

// Returns true if the expression X as an non-arithmetic untyped constant or
// of a known, non-arithmwetic type.
func definitelyNotArith(x ast.Expr) bool {
	typ := x.Type()
	if typ == nil {
		switch x.(*ast.ConstValue).Value.(type) {
		case ast.Bool, ast.String:
			return true
		default:
			return false
		}
	}
	t := builtinType(typ)
	if t == nil {
		return false
	}
	return !t.IsArith()
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
		if _, ok := unnamedType(x.X.Type()).(*ast.SliceType); ok {
			return true
		}
		return _isAddressable(x.X)
	case *ast.Selector:
		return _isAddressable(x.X)
	default:
		return false
	}
}

// Returns the length of the array type T. It must be called only after the
// length is a constant value.
func getArrayLength(t *ast.ArrayType) int64 {
	c, ok := t.Dim.(*ast.ConstValue)
	if !ok || c.Typ != ast.BuiltinInt {
		panic("not reached")
	}
	return int64(c.Value.(ast.Int))
}

// Returns the method set of the interface type T.
func ifaceMethodSet(t *ast.InterfaceType) methodSet {
	return ifaceMethodSetRec(make(methodSet), t)
}

func ifaceMethodSetRec(set methodSet, t *ast.InterfaceType) methodSet {
	for _, e := range t.Embedded {
		set = ifaceMethodSetRec(set, unnamedType(e).(*ast.InterfaceType))
	}
	for _, m := range t.Methods {
		set[m.Name] = m
	}
	return set
}

// Returns true if types S and T are identical.
// https://golang.org/ref/spec#Type_identity
func identicalTypes(s ast.Type, t ast.Type) bool {
	if s == t {
		// "Two named types are identical if their type names originate in the
		// same TypeSpec". Also fastpath for comparing a type to itself.
		return true
	}

	switch s := s.(type) {
	case *ast.TypeDecl:
		// T is either an unnamed type, or does not originate from the same
		// TypeSpec as S.
		return false
	case *ast.BuiltinType:
		if t, ok := t.(*ast.BuiltinType); ok {
			return s.Kind == t.Kind
		}
		return false
	case *ast.ArrayType:
		if t, ok := t.(*ast.ArrayType); ok {
			return getArrayLength(s) == getArrayLength(t) && identicalTypes(s.Elt, t.Elt)
		}
		return false
	case *ast.SliceType:
		if t, ok := t.(*ast.SliceType); ok {
			return identicalTypes(s.Elt, t.Elt)
		}
		return false
	case *ast.PtrType:
		if t, ok := t.(*ast.PtrType); ok {
			return identicalTypes(s.Base, t.Base)
		}
		return false
	case *ast.MapType:
		if t, ok := t.(*ast.MapType); ok {
			return identicalTypes(s.Key, t.Key) && identicalTypes(s.Elt, t.Elt)
		}
		return false
	case *ast.ChanType:
		if t, ok := t.(*ast.ChanType); ok {
			return s.Send == t.Send && s.Recv == t.Recv && identicalTypes(s.Elt, t.Elt)
		}
		return false
	case *ast.StructType:
		if t, ok := t.(*ast.StructType); !ok {
			return false
		} else if len(s.Fields) != len(t.Fields) {
			return false
		} else {
			for i := range s.Fields {
				sf, tf := &s.Fields[i], &t.Fields[i]
				// "corresponding fields [must] have the same names, [...],
				// and identical tags. Two anonymous fields are considered to
				// have the same name."
				if sf.Name != tf.Name || sf.Tag != tf.Tag {
					return false
				}
				// "Lower-case field names from different packages are always
				// different."
				if s.File == nil || t.File == nil || s.File.Pkg != t.File.Pkg &&
					!ast.IsExported(sf.Name) {
					return false
				}
				// "corresponding fields have [...] identical types"
				if !identicalTypes(sf.Type, tf.Type) {
					return false
				}
			}
			return true
		}
	case *ast.FuncType:
		if t, ok := t.(*ast.FuncType); !ok {
			return false
		} else if len(s.Params) != len(t.Params) || len(s.Returns) != len(t.Returns) ||
			s.Var != t.Var {
			// "the same number of parameters and result values, [...], and
			// either both functions are variadic or neither is."
			return false
		} else {
			// "corresponding parameter and result types are identical"
			for i := range s.Params {
				if !identicalTypes(s.Params[i].Type, t.Params[i].Type) {
					return false
				}
			}
			for i := range s.Returns {
				if !identicalTypes(s.Returns[i].Type, t.Returns[i].Type) {
					return false
				}
			}
			return true
		}
	case *ast.InterfaceType:
		if t, ok := t.(*ast.InterfaceType); !ok {
			return false
		} else {
			// "Interface types are identical if they have the same set of methods ...
			ss, st := ifaceMethodSet(s), ifaceMethodSet(t)
			if len(ss) != len(st) {
				return false
			}
			for name, ms := range ss {
				mt := st[name]
				// "... with the same name
				if mt == nil {
					return false
				}
				// "Lower-case method names from different packages are always
				// different."
				if ms.File == nil || mt.File == nil || ms.File.Pkg != mt.File.Pkg &&
					!ast.IsExported(ms.Name) {
					return false
				}
				//  "... "nd identical function types."
				if !identicalTypes(ms.Func.Sig, mt.Func.Sig) {
					return false
				}
			}
			return true
		}
	default:
		panic("not reached")
	}
}

// Returns true if TYP implements IFC.
func implements(typ ast.Type, ifc *ast.InterfaceType) bool {
	// Get the receiver base type.
	ptr := false
	if p, ok := typ.(*ast.PtrType); ok {
		ptr = true
		typ = p.Base
	}
	dcl, ok := typ.(*ast.TypeDecl)
	if !ok {
		// If the given type is not a (pointer to a) type name, the only
		// implemented interface is `interface{}`.
		return len(ifc.Embedded) == 0 && len(ifc.Methods) == 0
	}

	if impl, ok := dcl.Type.(*ast.InterfaceType); ok {
		return !ptr && ifaceImplements(impl, ifc)
	} else {
		return typenameImplements(ptr, dcl, ifc)
	}
}

func ifaceImplements(impl *ast.InterfaceType, ifc *ast.InterfaceType) bool {
	implSet := ifaceMethodSet(impl)
	for _, fn := range ifaceMethodSet(ifc) {
		m := implSet[fn.Name]
		if m == nil || !identicalTypes(m.Func.Sig, fn.Func.Sig) {
			return false
		}
	}
	return true
}

func typenameImplements(ptr bool, dcl *ast.TypeDecl, ifc *ast.InterfaceType) bool {
	for _, fn := range ifaceMethodSet(ifc) {
		// Look for an implementation among the methods with value receivers.
		i := 0
		for ; i < len(dcl.Methods); i++ {
			m := dcl.Methods[i]
			if fn.Name == m.Name {
				if !identicalTypes(fn.Func.Sig, m.Func.Sig) {
					return false
				}
				break
			}
		}
		// Do not consider pointer receiver methods if the original type
		// wasn't a pointer type.
		if !ptr {
			if i == len(dcl.Methods) {
				// No method name matches.
				return false
			}
			continue
		}
		// Look for an implementation among the methods with pointer
		// receivers.
		for i = 0; i < len(dcl.PMethods); i++ {
			m := dcl.PMethods[i]
			if fn.Name == m.Name {
				if !identicalTypes(fn.Func.Sig, m.Func.Sig) {
					return false
				}
				break
			}
		}
		if i == len(dcl.PMethods) {
			// No method name matches.
			return false
		}
	}
	return true
}

// Checks whether the expression X is assignable to a variable of type DST. If
// it is assignable, returns the expression itself and true. If the expression
// is an untyped constant, representable by DST, returns the converted
// constant and true. Otherwise, returns false.
// https://golang.org/ref/spec#Assignability
func isAssignable(dst ast.Type, x ast.Expr) (ast.Expr, bool) {
	if src := x.Type(); src == nil {
		// "x is an untyped constant representable by a value of type T."
		if c, ok := x.(*ast.ConstValue); !ok {
			panic("not reached")
		} else if t := builtinType(dst); t == nil {
			return nil, false
		} else if v := convertConst(t, nil, c.Value); v == nil {
			return nil, false
		} else {
			return &ast.ConstValue{Off: x.Position(), Typ: dst, Value: v}, true
		}
	} else {
		return x, isAssignableType(dst, src)
	}
}

// Returns true, if a value of type SRC is assignable to a variable of type DST.
func isAssignableType(dst ast.Type, src ast.Type) bool {
	// "x's type is identical to T."
	if identicalTypes(dst, src) {
		return true
	}
	// "x's type V and T have identical underlying types and at least one of V
	// or T is not a named type."
	usrc, udst := unnamedType(src), unnamedType(dst)
	if (usrc == src || udst == dst) && identicalTypes(usrc, udst) {
		return true
	}
	// "T is an interface type and x implements T.
	if ifc, ok := udst.(*ast.InterfaceType); ok && implements(src, ifc) {
		return true
	}
	// "x is a bidirectional channel value, T is a channel type, x's type V
	// and T have identical element types, and at least one of V or T is not a
	// named type.
	if ch, ok := usrc.(*ast.ChanType); ok && ch.Send && ch.Recv {
		if dch, ok := udst.(*ast.ChanType); ok && (usrc == src || udst == dst) {
			return identicalTypes(ch.Elt, dch.Elt)
		}
	}
	// "x is the predeclared identifier nil and T is a pointer, function,
	// slice, map, channel, or interface type."
	if usrc == ast.BuiltinNilType {
		switch udst.(type) {
		case *ast.PtrType, *ast.FuncType, *ast.SliceType, *ast.MapType, *ast.ChanType,
			*ast.InterfaceType:
			return true
		}
	}
	return false
}
