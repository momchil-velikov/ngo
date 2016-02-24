package typecheck

import "golang/ast"

type typeckVisitor interface {
	checkTypeDecl(*ast.TypeDecl) error
	checkConstDecl(*ast.Const) error
	checkVarDecl(*ast.Var) error
	checkFuncDecl(*ast.FuncDecl) error
}

func CheckPackage(pkg *ast.Package) error {
	ck0 := &typeckPhase0{
		typeckCommon: typeckCommon{
			Pkg:  pkg,
			Done: make(map[ast.Symbol]struct{}),
		},
	}
	if err := checkDecls(pkg, ck0); err != nil {
		return err
	}

	ck1 := &typeckPhase1{
		typeckCommon: typeckCommon{
			Pkg:  pkg,
			Done: make(map[ast.Symbol]struct{}),
		},
	}
	if err := checkDecls(pkg, ck1); err != nil {
		return err
	}
	return nil
}

type typeckCommon struct {
	Pkg   *ast.Package
	File  *ast.File
	Files []*ast.File
	Done  map[ast.Symbol]struct{}
}

type typeckPhase0 struct {
	typeckCommon
	Syms []ast.Symbol
}

func checkDecls(pkg *ast.Package, ck typeckVisitor) error {
	for _, s := range pkg.Syms {
		var err error
		switch d := s.(type) {
		case *ast.TypeDecl:
			err = ck.checkTypeDecl(d)
		case *ast.Const:
			err = ck.checkConstDecl(d)
		case *ast.Var:
			err = ck.checkVarDecl(d)
		case *ast.FuncDecl:
			err = ck.checkFuncDecl(d)
		default:
			panic("not reached")
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (ck *typeckPhase0) checkTypeDecl(d *ast.TypeDecl) error {
	// No need to check things in a different package
	if d.File == nil || d.File.Pkg != ck.Pkg {
		return nil
	}
	// Check for a typechecking loop
	if l := ck.checkTypeLoop(d); l != nil {
		return &TypeCheckLoop{Off: d.Off, File: d.File, Loop: l}
	}
	// Check if we have already processed this type declaration.
	if _, ok := ck.Done[d]; ok {
		return nil
	}
	// Descend into the type literal
	ck.Done[d] = struct{}{}
	ck.beginCheck(d, d.File)
	t, err := ck.checkType(d.Type)
	ck.endCheck()
	if err != nil {
		return err
	}
	if err := checkMethodUniqueness(d); err != nil {
		return err
	}
	d.Type = t
	return nil
}

func (ck *typeckPhase0) checkConstDecl(c *ast.Const) error {
	// No need to check things in a different package
	if c.File == nil || c.File.Pkg != ck.Pkg {
		return nil
	}
	if _, ok := ck.Done[c]; ok {
		return nil
	}
	ck.Done[c] = struct{}{}

	if c.Type == nil {
		c.Type = &ast.TypeVar{Off: c.Off, File: c.File, Name: c.Name}
	} else {
		t, ok := unnamedType(c.Type).(*ast.BuiltinType)
		if !ok || t.Kind == ast.BUILTIN_NIL_TYPE {
			return &BadConstType{Off: c.Off, File: c.File, Type: c.Type}
		}
	}
	ck.beginCheck(nil, c.File)
	x, err := ck.checkExpr(c.Init)
	ck.endCheck()
	if err != nil {
		return err
	}
	c.Init = x
	return nil
}

func (ck *typeckPhase0) checkVarDecl(v *ast.Var) error {
	// No need to check things in a different package
	if v.File == nil || v.File.Pkg != ck.Pkg {
		return nil
	}
	if _, ok := ck.Done[v]; ok {
		return nil
	}
	ck.Done[v] = struct{}{}

	if v.Init == nil {
		// Check the declared type of the variable
		if v.Type == nil {
			panic("this is not happening")
		}
		ck.beginCheck(nil, v.File)
		t, err := ck.checkType(v.Type)
		ck.endCheck()
		if err != nil {
			return err
		}
		v.Type = t
		return nil
	}

	// Assign a new type variable for each variable declaration, if the
	// variable does not have a declared type. Check the type, if the variable
	// has a declared type.
	for i := range v.Init.LHS {
		op := v.Init.LHS[i].(*ast.OperandName)
		v := op.Decl.(*ast.Var)
		ck.Done[v] = struct{}{}
		if v.Type == nil {
			v.Type = &ast.TypeVar{Off: v.Off, File: v.File, Name: v.Name}
		} else {
			ck.beginCheck(nil, v.File)
			t, err := ck.checkType(v.Type)
			ck.endCheck()
			if err != nil {
				return err
			}
			v.Type = t
		}
	}

	// Infer/check the type of each expression on the right-hand side.
	for i := range v.Init.RHS {
		x, err := ck.checkExpr(v.Init.RHS[i])
		if err != nil {
			return err
		}
		v.Init.RHS[i] = x
	}

	// Bind the type variables to the corresponding expression's type.
	if len(v.Init.RHS) == 1 {
		// Case 1: one multi-valued expression.
		if tp, ok := v.Init.RHS[0].Type().(*ast.TupleType); ok && tp.Strict {
			if len(tp.Type) != len(v.Init.LHS) {
				return &BadMultiValueAssign{Off: v.Init.Off, File: v.File}
			}
			for i := range v.Init.LHS {
				op := v.Init.LHS[i].(*ast.OperandName)
				v := op.Decl.(*ast.Var)
				if t, ok := v.Type.(*ast.TypeVar); ok {
					t.Type = tp.Type[i]
				}
			}
			return nil
		}
	}

	// Case 2: one or more single valued expressions.
	if len(v.Init.LHS) != len(v.Init.RHS) {
		return &BadMultiValueAssign{Off: v.Init.Off, File: v.File}
	}
	for i := range v.Init.RHS {
		op := v.Init.LHS[i].(*ast.OperandName)
		v := op.Decl.(*ast.Var)
		t := singleValueType(v.Init.RHS[i].Type())
		if t == nil {
			return &SingleValueContext{Off: v.Init.RHS[i].Position(), File: v.File}
		}
		if tv, ok := v.Type.(*ast.TypeVar); ok {
			tv.Type = t
		}
	}

	return nil
}

func (ck *typeckPhase0) checkFuncDecl(fn *ast.FuncDecl) error {
	// No need to check things in a different package
	if fn.File == nil || fn.File.Pkg != ck.Pkg {
		return nil
	}
	// Check the receiver type.
	if fn.Func.Recv != nil {
		ck.beginCheck(nil, fn.File)
		t, err := ck.checkType(fn.Func.Recv.Type)
		ck.endCheck()
		if err != nil {
			return err
		}
		fn.Func.Recv.Type = t
	}
	// Check the signature.
	ck.beginCheck(nil, fn.File)
	t, err := ck.checkType(fn.Func.Sig)
	ck.endCheck()
	if err != nil {
		return err
	}
	fn.Func.Sig = t.(*ast.FuncType)
	return nil
}

func (ck *typeckPhase0) checkType(t ast.Type) (ast.Type, error) {
	return t.TraverseType(ck)
}

func (ck *typeckPhase0) checkExpr(x ast.Expr) (ast.Expr, error) {
	return x.TraverseExpr(ck)
}

func (ck *typeckCommon) beginFile(f *ast.File) {
	ck.Files = append(ck.Files, ck.File)
	ck.File = f
}

func (ck *typeckCommon) endFile() {
	n := len(ck.Files)
	ck.File = ck.Files[n-1]
	ck.Files = ck.Files[:n-1]
}

func (ck *typeckPhase0) beginCheck(s ast.Symbol, f *ast.File) {
	ck.beginFile(f)
	ck.Syms = append(ck.Syms, s)
}

func (ck *typeckPhase0) endCheck() {
	ck.Syms = ck.Syms[:len(ck.Syms)-1]
	ck.endFile()
}

func (ck *typeckPhase0) breakEmbedChain() {
	ck.Syms = append(ck.Syms, nil)
}

func (ck *typeckPhase0) restoreEmbedChain() {
	ck.Syms = ck.Syms[:len(ck.Syms)-1]
}

func (ck *typeckPhase0) checkTypeLoop(sym ast.Symbol) []ast.Symbol {
	for i := len(ck.Syms); i > 0; i-- {
		s := ck.Syms[i-1]
		if s == nil {
			return nil
		}
		if s == sym {
			return ck.Syms[i-1:]
		}
	}
	return nil
}

type typeckPhase1 struct {
	typeckCommon
	Const []*ast.Const
	TVars []*ast.TypeVar
	X     ast.Expr
}

func (ck *typeckPhase1) checkTypeDecl(d *ast.TypeDecl) error {
	// No need to check things in a different package
	if d.File == nil || d.File.Pkg != ck.Pkg {
		return nil
	}
	// Check if we have already processed this type declaration.
	if _, ok := ck.Done[d]; ok {
		return nil
	}
	// Descend into the type literal
	ck.Done[d] = struct{}{}
	ck.beginFile(d.File)
	t, err := ck.checkType(d.Type)
	ck.endFile()
	if err != nil {
		return err
	}
	d.Type = t
	return nil
}

func (ck *typeckPhase1) checkConstDecl(c *ast.Const) error {
	// No need to check things in a different package
	if c.File == nil || c.File.Pkg != ck.Pkg {
		return nil
	}
	// Check for a constant definition loop.
	if l := ck.checkConstLoop(c); l != nil {
		return &ConstInitLoop{Loop: l}
	}
	// Check if we have already processed this const declaration.
	if _, ok := ck.Done[c]; ok {
		return nil
	}
	ck.Done[c] = struct{}{}

	ck.beginFile(c.File)
	ck.beginEvalConst(c)
	x, err := ck.checkExpr(c.Init)
	ck.endEvalConst()
	ck.endFile()
	if err != nil {
		return err
	}
	if v, ok := x.(*ast.ConstValue); !ok || v == ast.BuiltinNil {
		return &NotConst{Off: c.Init.Position(), File: c.File, What: "const initializer"}
	}
	c.Init = x
	// FIXME: check compatibility with declared type.
	return nil
}

func (ck *typeckPhase1) checkVarDecl(v *ast.Var) error {
	// No need to check things in a different package
	if v.File == nil || v.File.Pkg != ck.Pkg {
		return nil
	}
	// Check if we have already processed this variable declaration.
	if _, ok := ck.Done[v]; ok {
		return nil
	}
	ck.Done[v] = struct{}{}

	ck.beginFile(v.File)
	defer ck.endFile()

	// If there is no initializer, just check the type.
	if v.Init == nil {
		t, err := ck.checkType(v.Type)
		if err != nil {
			return err
		}
		v.Type = t
		return nil
	}

	// Check the types of the initialization expressions.
	for i := range v.Init.RHS {
		x, err := ck.checkExpr(v.Init.RHS[i])
		if err != nil {
			return err
		}
		v.Init.RHS[i] = x
	}

	// Check the type of the variables.
	for i := range v.Init.LHS {
		op := v.Init.LHS[i].(*ast.OperandName)
		v := op.Decl.(*ast.Var)
		t, err := ck.checkType(v.Type)
		if err != nil {
			return err
		}
		v.Type = t
	}

	return nil
}

func (ck *typeckPhase1) checkFuncDecl(fn *ast.FuncDecl) error {
	// No need to check things in a different package
	if fn.File == nil || fn.File.Pkg != ck.Pkg {
		return nil
	}
	// Check if we have already processed this function/method declaration.
	if _, ok := ck.Done[fn]; ok {
		return nil
	}
	ck.Done[fn] = struct{}{}
	// Check the receiver type.
	if fn.Func.Recv != nil {
		ck.beginFile(fn.File)
		t, err := ck.checkType(fn.Func.Recv.Type)
		ck.endFile()
		if err != nil {
			return err
		}
		fn.Func.Recv.Type = t
	}
	// Check the signature.
	ck.beginFile(fn.File)
	t, err := ck.checkType(fn.Func.Sig)
	ck.endFile()
	if err != nil {
		return err
	}
	fn.Func.Sig = t.(*ast.FuncType)
	return nil
}

func (ck *typeckPhase1) checkType(t ast.Type) (ast.Type, error) {
	return t.TraverseType(ck)
}

func (ck *typeckPhase1) checkExpr(x ast.Expr) (ast.Expr, error) {
	if x == ck.X {
		return nil, &ExprLoop{Off: x.Position(), File: ck.File}
	}
	if ck.X == nil {
		ck.X = x
		defer func() { ck.X = nil }()
	}
	return x.TraverseExpr(ck)
}

func (ck *typeckPhase1) checkTypeVarLoop(tv *ast.TypeVar) []*ast.TypeVar {
	for i := len(ck.TVars); i > 0; i-- {
		t := ck.TVars[i-1]
		if t == tv {
			return ck.TVars[i-1:]
		}
	}
	return nil
}

func (ck *typeckPhase1) beginEvalConst(c *ast.Const) {
	ck.Const = append(ck.Const, c)
}

func (ck *typeckPhase1) endEvalConst() {
	ck.Const = ck.Const[:len(ck.Const)-1]
}

func (ck *typeckPhase1) checkConstLoop(c *ast.Const) []*ast.Const {
	for i := len(ck.Const) - 1; i >= 0; i-- {
		if c == ck.Const[i] {
			return ck.Const[i:]
		}
	}
	return nil
}

// Returns the first type literal in a chain of TypeNames.
func unnamedType(typ ast.Type) ast.Type {
	for {
		switch t := typ.(type) {
		case *ast.TypeDecl:
			typ = t.Type
		case *ast.TypeVar:
			if t.Type == nil {
				return t
			}
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
