package typecheck

import "golang/ast"

func CheckPackage(pkg *ast.Package) error {
	if err := checkPkgLevelTypes(pkg); err != nil {
		return err
	}
	return nil
}

// Finds the selector NAME among the fields and methods of the type TYP. If
// there are two or more identically named fields or methods at the same
// shallowest depth, returns any two of them, with the intention for the
// caller to report an ambiguous selector error at the appropriate source
// position. The final `bool` output is true whenever the original type was a
// pointer type, or the field or method was promoted via one or more anonymous
// fields of a pointer type.
type fieldOrMethod struct {
	F *ast.Field
	M *ast.FuncDecl
}

func findSelector(typ ast.Type, name string) (fieldOrMethod, fieldOrMethod, bool) {
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
		s0, s1, p := findFieldOrMethod(typ, name)
		return s0, s1, p || (ptr != nil)
	}

	// "For a value x of type I where I is an interface type, x.f denotes the
	// actual method with name f of the dynamic value of x."
	if ifc, ok := unnamedType(orig).(*ast.InterfaceType); ok {
		return fieldOrMethod{M: findInterfaceMethod(ifc, name)}, fieldOrMethod{}, false
	}

	// "As an exception, if the type of x is a named pointer type and (*x).f
	// is a valid selector expression denoting a field (but not a method), x.f
	// is shorthand for (*x).f."
	dcl, ok := orig.(*ast.TypeDecl)
	if ok {
		if ptr, ok := unnamedType(dcl.Type).(*ast.PtrType); ok {
			s0, s1, p := findFieldOrMethod(ptr.Base, name)
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

func findFieldOrMethod(typ ast.Type, name string) (fieldOrMethod, fieldOrMethod, bool) {
	m := findImmediateFieldOrMethod(typ, name)
	if m.F != nil || m.M != nil {
		return m, fieldOrMethod{}, false
	}
	if str, ok := unnamedType(typ).(*ast.StructType); ok {
		return findPromotedFieldOrMethod(str, name)
	}
	return fieldOrMethod{}, fieldOrMethod{}, false
}

// Finds and returns the method NAME of the type TYP (which can be a typename
// or an interface type) or, if TYP is declared as a struct type, the field
// NAME.
func findImmediateFieldOrMethod(typ ast.Type, name string) fieldOrMethod {
	if ifc, ok := unnamedType(typ).(*ast.InterfaceType); ok {
		return fieldOrMethod{M: findInterfaceMethod(ifc, name)}
	}
	if dcl, ok := typ.(*ast.TypeDecl); ok {
		for i := range dcl.Methods {
			if name == dcl.Methods[i].Name {
				return fieldOrMethod{M: dcl.Methods[i]}
			}
		}
		for i := range dcl.PMethods {
			if name == dcl.PMethods[i].Name {
				return fieldOrMethod{M: dcl.PMethods[i]}
			}
		}
		typ = dcl.Type
	}
	if str, ok := unnamedType(typ).(*ast.StructType); ok {
		for i := range str.Fields {
			if name == fieldName(&str.Fields[i]) {
				return fieldOrMethod{F: &str.Fields[i]}
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
	str *ast.StructType, name string) (fieldOrMethod, fieldOrMethod, bool) {

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
			m := findImmediateFieldOrMethod(t, name)
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
func findInterfaceMethod(typ *ast.InterfaceType, name string) *ast.FuncDecl {
	for _, m := range typ.Methods {
		if name == m.Name {
			return m
		}
	}
	for _, t := range typ.Embedded {
		d := t.(*ast.TypeDecl)
		if m := findInterfaceMethod(d.Type.(*ast.InterfaceType), name); m != nil {
			return m
		}
	}
	return nil
}
