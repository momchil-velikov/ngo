package parser

import (
	"errors"
	"golang/ast"
	"golang/constexpr"
	"unicode"
	"unicode/utf8"
)

func isExported(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsLetter(r) && unicode.IsUpper(r)
}

type resolver struct {
	scope ast.Scope
}

func (r *resolver) enterScope(scope ast.Scope) {
	if r.scope != scope.Parent() {
		panic("internal error: invalid scope nesting")
	}
	r.scope = scope
}

func (r *resolver) exitScope() {
	r.scope = r.scope.Parent()
}

func ResolvePackage(
	p *ast.UnresolvedPackage, loc ast.PackageLocator) (*ast.Package, error) {

	pkg := &ast.Package{
		Path:  p.Path,
		Name:  p.Name,
		Files: nil,
		Decls: make(map[string]ast.Symbol),
	}

	r := resolver{scope: ast.UniverseScope}
	r.enterScope(pkg)

	// Insert package and file scope declarations into the symbol table.
	for _, f := range p.Files {
		file, err := r.declareTopLevel(f, pkg, loc)
		if err != nil {
			return nil, err
		}
		pkg.Files = append(pkg.Files, file)
	}

	// Declare lower level names and resolve all.
	for i, f := range p.Files {
		if err := r.resolveTopLevel(pkg.Files[i], f.Decls); err != nil {
			return nil, err
		}
	}

	return pkg, nil
}

func findImport(path string, is []*ast.Import) (*ast.Import, int) {
	i := 0
	for i < len(is) && path < is[i].Path {
		i++
	}
	if i >= len(is) || path != is[i].Path {
		return nil, i
	}
	return is[i], i
}

func insertImport(is []*ast.Import, i int, imp *ast.Import) []*ast.Import {
	is = append(is, nil)
	copy(is[i+1:], is[i:])
	is[i] = imp
	return is
}

func (r *resolver) declareTopLevel(
	f *ast.UnresolvedFile, pkg *ast.Package, loc ast.PackageLocator) (*ast.File, error) {

	file := &ast.File{
		Off:     f.Off,
		Pkg:     pkg,
		Imports: f.Imports,
		Name:    f.Name,
		SrcMap:  f.SrcMap,
		Decls:   make(map[string]ast.Symbol),
	}

	// Declare imported package names.
	for _, i := range file.Imports {
		path := string(constexpr.String(i.Path))
		dep, pos := findImport(path, pkg.Deps)
		if dep == nil {
			p, err := loc.FindPackage(path)
			if err != nil {
				return nil, err
			}
			dep = &ast.Import{Path: path, Pkg: p}
			pkg.Deps = insertImport(pkg.Deps, pos, dep)
		}
		i.File = file
		i.Pkg = dep.Pkg
		if i.Name == "_" {
			// do not import package decls
			continue
		}
		if i.Name == "." {
			// Import package exported identifiers into the file block.
			for name, sym := range i.Pkg.Decls {
				if isExported(name) {
					if err := file.Declare(name, sym); err != nil {
						return nil, err
					}
				}
			}
		} else {
			// Declare the package name in the file block.
			if len(i.Name) == 0 {
				i.Name = i.Pkg.Name
			}
			if err := file.Declare(i.Name, i); err != nil {
				return nil, err
			}
		}
	}

	// Declare top-level names.
	for _, d := range f.Decls {
		var err error
		switch d := d.(type) {
		case *ast.TypeDecl:
			err = r.declareTypeName(d, file)
		case *ast.TypeDeclGroup:
			for _, d := range d.Types {
				if err := r.declareTypeName(d, file); err != nil {
					return nil, err
				}
			}
		case *ast.ConstDecl:
			err = r.declareConst(d, file)
		case *ast.ConstDeclGroup:
			for _, c := range d.Consts {
				if err := r.declareConst(c, file); err != nil {
					return nil, err
				}
			}
		case *ast.VarDecl:
			err = r.declarePkgVar(d, file)
		case *ast.VarDeclGroup:
			for _, d := range d.Vars {
				if err := r.declarePkgVar(d, file); err != nil {
					return nil, err
				}
			}
		case *ast.FuncDecl:
			err = r.declareFunc(d, file)
		default:
			panic("not reached")
		}
		if err != nil {
			return nil, err
		}
	}

	return file, nil
}

func (r *resolver) resolveTopLevel(file *ast.File, ds []ast.Decl) error {
	r.enterScope(file)
	defer r.exitScope()
	// Resolve right-hand sides of package-level variable initialization
	// statements.
	for _, st := range file.Init {
		for i := range st.RHS {
			x, err := r.resolveExpr(st.RHS[i])
			if err != nil {
				return err
			}
			st.RHS[i] = x
		}
	}
	for _, d := range ds {
		var err error
		switch d := d.(type) {
		case *ast.TypeDecl:
			err = r.resolveTypeDecl(d)
		case *ast.TypeDeclGroup:
			for _, d := range d.Types {
				if err = r.resolveTypeDecl(d); err != nil {
					break
				}
			}
		case *ast.ConstDecl:
			err = r.resolveConst(d)
		case *ast.ConstDeclGroup:
			err = r.resolveConstGroup(d, false)
		case *ast.VarDecl:
			err = r.resolvePkgVar(d)
		case *ast.VarDeclGroup:
			for _, d := range d.Vars {
				if err = r.resolvePkgVar(d); err != nil {
					break
				}
			}
		case *ast.FuncDecl:
			err = r.resolveFunc(&d.Func)
		default:
			panic("not reached")
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *resolver) declareTypeName(t *ast.TypeDecl, file *ast.File) error {
	if t.Name == "_" {
		return nil
	}
	t.File = file
	return r.scope.Declare(t.Name, t)
}

func (r *resolver) resolveTypeDecl(t *ast.TypeDecl) error {
	typ, err := r.resolveType(t.Type)
	if err != nil {
		return err
	}
	t.Type = typ
	return nil
}

func (r *resolver) declareConst(dcl *ast.ConstDecl, file *ast.File) error {
	for _, c := range dcl.Names {
		if c.Name == "_" {
			continue
		}
		c.File = file
		if err := r.scope.Declare(c.Name, c); err != nil {
			return err
		}
	}
	return nil
}

func (r *resolver) resolveConstSpec(dcl *ast.ConstDecl) error {
	typ, err := r.resolveType(dcl.Type)
	if err != nil {
		return err
	}
	dcl.Type = typ
	for i := range dcl.Values {
		x, err := r.resolveExpr(dcl.Values[i])
		if err != nil {
			return err
		}
		dcl.Values[i] = x
	}
	return nil
}

func (r *resolver) resolveConst(dcl *ast.ConstDecl) error {
	if err := r.resolveConstSpec(dcl); err != nil {
		return err
	}
	if len(dcl.Values) != len(dcl.Names) {
		return errors.New("number of idents must be equal to the number of expressions")
	}
	for i, c := range dcl.Names {
		c.Type = dcl.Type
		c.Init = dcl.Values[i]
	}
	return nil
}

func (r *resolver) resolveConstGroup(group *ast.ConstDeclGroup, blk bool) error {
	var (
		iota = 0
		typ  ast.Type
		init []ast.Expr
	)
	for _, d := range group.Consts {
		if err := r.resolveConstSpec(d); err != nil {
			return err
		}
		if blk {
			// In blocks, declare the constant names, so the they are
			// available to following ConstSpec's in the group.
			if err := r.declareConst(d, r.scope.File()); err != nil {
				return err
			}
		}
		if len(d.Values) == 0 {
			// Repeat the type and the initialization expression from the
			// preceding ConstSpec
			d.Type = typ
			d.Values = init
		}
		if len(d.Values) != len(d.Names) {
			return errors.New(
				"number of idents must be equal to the number of expressions")
		}
		for i, c := range d.Names {
			c.Iota = iota
			c.Type = d.Type
			c.Init = d.Values[i]
		}
		iota++
		typ = d.Type
		init = d.Values
	}
	return nil
}

// Declares a variable at package scope and creates the assignment statement
// for the initialization.
func (r *resolver) declarePkgVar(vr *ast.VarDecl, file *ast.File) error {
	if len(vr.Init) > 0 {
		// Create an assignment statement, which is executed, according to
		// dependency order, upon package initialization, to set initial
		// values of the package-level variables. Make every variable refer to
		// its initialization statement.
		lhs := make([]ast.Expr, len(vr.Names))
		init := &ast.AssignStmt{Op: ast.NOP, LHS: lhs, RHS: vr.Init}
		vr.Init = nil
		for i, v := range vr.Names {
			lhs[i] = v
			v.Init = init
		}
		file.Init = append(file.Init, init)
	}
	for _, v := range vr.Names {
		if v.Name == "_" {
			continue
		}
		v.File = file
		if err := file.Pkg.Declare(v.Name, v); err != nil {
			return err
		}
	}
	return nil
}

func (r *resolver) resolvePkgVar(d *ast.VarDecl) error {
	typ, err := r.resolveType(d.Type)
	if err != nil {
		return err
	}
	for _, v := range d.Names {
		v.Type = typ
	}
	return nil
}

func (r *resolver) declareFunc(fn *ast.FuncDecl, file *ast.File) error {
	fn.File = file
	if fn.Name == "_" {
		return nil
	}
	if fn.Func.Recv != nil {
		return nil
	}
	return r.scope.Declare(fn.Name, fn)
}

func (r *resolver) declareParam(p *ast.Param) error {
	if p.Name == "_" || len(p.Name) == 0 {
		return nil
	}
	v := &ast.Var{Off: p.Off, File: r.scope.File(), Name: p.Name, Type: p.Type}
	return r.scope.Declare(v.Name, v)
}

func (r *resolver) declareParams(ps []ast.Param) error {
	for i := range ps {
		if err := r.declareParam(&ps[i]); err != nil {
			return err
		}
	}
	return nil
}

// Checks if TYP is of the form `T` or `*T` where `T`, the receiver base type,
// is a (unqualified) TypeName, declared in this package. The type `T` must
// also not be a pointer or an interface type. Returns `T` or an error.
// cf. https://golang.org/ref/spec#Method_declarations
func (r *resolver) isReceiverType(typ ast.Type) (*ast.TypeDecl, error) {
	if ptr, ok := typ.(*ast.PtrType); ok {
		typ = ptr.Base
	}
	d, ok := typ.(*ast.TypeDecl)
	if !ok {
		return nil, errors.New("invalid receiver type (is an unnamed type)")
	}
	if d.File == nil || r.scope.Package() != d.File.Pkg {
		return nil, errors.New(
			"receiver type and method must be declared in the same package")
	}
	return d, nil
}

// Returns the first type literal in a possibly empty chain of TypeNames.
func unnamedType(dcl *ast.TypeDecl) ast.Type {
	typ := dcl.Type
	for t, ok := typ.(*ast.TypeDecl); ok; t, ok = typ.(*ast.TypeDecl) {
		typ = t.Type
	}
	return typ
}

// Checks that a method name is unique in the method set of DCL and, if DCL is
// a TypeName of a struct type, among the field names.
func isUniqueMethod(name string, dcl *ast.TypeDecl) error {
	for _, fn := range dcl.Methods {
		if name == fn.Name {
			return errors.New("duplicate method name")
		}
	}
	for _, fn := range dcl.PMethods {
		if name == fn.Name {
			return errors.New("duplicate method name")
		}
	}
	if s := checkStructType(dcl); s != nil {
		for i := range s.Fields {
			if name == fieldName(&s.Fields[i]) {
				return errors.New("type has both field and method named " + name)
			}
		}
	}
	return nil
}

func (r *resolver) resolveFunc(fn *ast.Func) error {
	if rcv := fn.Recv; rcv != nil {
		typ, err := r.resolveType(rcv.Type)
		if err != nil {
			return err
		}
		base, err := r.isReceiverType(typ)
		if err != nil {
			return err
		}
		switch unnamedType(base).(type) {
		case *ast.PtrType:
			return errors.New("receiver base type cannot be a pointer")
		case *ast.InterfaceType:
			return errors.New("receiver base type cannot be an interface")
		}
		if err := isUniqueMethod(fn.Decl.Name, base); err != nil {
			return err
		}
		rcv.Type = typ
		if typ == base {
			base.Methods = append(base.Methods, fn.Decl)
		} else {
			base.PMethods = append(base.PMethods, fn.Decl)
		}
	}
	typ, err := r.resolveType(fn.Sig)
	if err != nil {
		return err
	}
	fn.Sig = typ.(*ast.FuncType)
	if fn.Blk == nil {
		return nil
	}
	// The function body forms an extra scope, holding only labels and
	// transparent for ordinary name lookups.
	fn.Up = r.scope
	r.enterScope(fn)
	defer r.exitScope()
	// Declare parameter and return names in the scope of the function block.
	fn.Blk.Up = r.scope
	// FIXME: scope is not exited on error, not sure if it matters at all
	r.enterScope(fn.Blk)
	if err := r.declareParams(fn.Sig.Params); err != nil {
		return err
	}
	if err := r.declareParams(fn.Sig.Returns); err != nil {
		return err
	}
	if fn.Recv != nil {
		if err := r.declareParam(fn.Recv); err != nil {
			return err
		}
	}
	r.exitScope()
	blk, err := r.resolveBlock(fn.Blk)
	if err != nil {
		return err
	}
	fn.Blk = blk
	// Now check the goto, break and continue statement destinations
	ck := jumpResolver{fn: fn}
	_, err = ck.VisitBlock(fn.Blk)
	return err
}

func (r *resolver) lookupIdent(id *ast.QualifiedId) (ast.Symbol, error) {
	if len(id.Pkg) > 0 {
		if i := r.scope.Lookup(id.Pkg); i == nil {
			return nil, errors.New("package name " + id.Pkg + " not declared")
		} else if i, ok := i.(*ast.ImportDecl); !ok {
			return nil, errors.New(id.Pkg + " does not refer to package name")
		} else if d := i.Pkg.Lookup(id.Id); d == nil {
			return nil, errors.New(id.Pkg + "." + id.Id + " not declared")
		} else if !isExported(id.Id) {
			return nil, errors.New(id.Pkg + "." + id.Id + " is not exported")
		} else {
			return d, nil
		}
	} else {
		if d := r.scope.Lookup(id.Id); d == nil {
			return nil, errors.New(id.Id + " not declared")
		} else {
			return d, nil
		}
	}
}

func fieldName(f *ast.Field) string {
	if len(f.Name) == 0 {
		typ := f.Type
		if t, ok := f.Type.(*ast.PtrType); ok {
			typ = t.Base
		}
		t := typ.(*ast.TypeDecl)
		return t.Name
	} else {
		return f.Name
	}
}

func checkDuplicateFieldNames(s *ast.StructType) error {
	for i := range s.Fields {
		name := fieldName(&s.Fields[i])
		if name == "_" {
			continue
		}
		for j := i + 1; j < len(s.Fields); j++ {
			other := fieldName(&s.Fields[j])
			if name == other {
				return errors.New("field name " + name + " is duplicated")
			}
		}
	}
	return nil
}

// Resolves identifiers in function parameter and return values list.
func (r *resolver) resolveParams(ps []ast.Param) (err error) {
	var carry ast.Type
	count := 0
	n := len(ps)
	for i := n - 1; i >= 0; i-- {
		p := &ps[i]
		if p.Type != nil {
			carry, err = r.resolveType(p.Type)
			if err != nil {
				return
			}
		}
		p.Type = carry
		if len(p.Name) > 0 {
			count++
		}
	}
	if count != n && count > 0 {
		err = errors.New("param/return names must either all be present or all be absent")
	}
	return
}

func (r *resolver) resolveType(t ast.Type) (ast.Type, error) {
	if t == nil {
		return nil, nil
	}
	return t.TraverseType(r)
}

func (r *resolver) VisitError(*ast.Error) (*ast.Error, error) {
	panic("not reached")
}

func (r *resolver) VisitTypeName(t *ast.QualifiedId) (ast.Type, error) {
	if t.Id == "_" {
		return nil, errors.New("`_` is not a typename")
	}
	d, err := r.lookupIdent(t)
	if err != nil {
		return nil, err
	}
	td, ok := d.(*ast.TypeDecl)
	if !ok {
		return nil, errors.New(t.Id + " is not a typename")
	}
	return td, nil
}

func (*resolver) VisitTypeDeclType(*ast.TypeDecl) (ast.Type, error) {
	panic("not reached")
}

func (*resolver) VisitBuiltinType(*ast.BuiltinType) (ast.Type, error) {
	panic("not reached")
}

func (r *resolver) VisitArrayType(t *ast.ArrayType) (ast.Type, error) {
	elt, err := r.resolveType(t.Elt)
	if err != nil {
		return nil, err
	}
	t.Elt = elt
	return t, nil
}

func (r *resolver) VisitSliceType(t *ast.SliceType) (ast.Type, error) {
	elt, err := r.resolveType(t.Elt)
	if err != nil {
		return nil, err
	}
	t.Elt = elt
	return t, nil
}

func (r *resolver) VisitPtrType(t *ast.PtrType) (ast.Type, error) {
	b, err := r.resolveType(t.Base)
	if err != nil {
		return nil, err
	}
	t.Base = b
	return t, nil
}

func (r *resolver) VisitMapType(t *ast.MapType) (ast.Type, error) {
	key, err := r.resolveType(t.Key)
	if err != nil {
		return nil, err
	}
	t.Key = key
	elt, err := r.resolveType(t.Elt)
	if err != nil {
		return nil, err
	}
	t.Elt = elt
	return t, nil
}

func (r *resolver) VisitChanType(t *ast.ChanType) (ast.Type, error) {
	elt, err := r.resolveType(t.Elt)
	if err != nil {
		return nil, err
	}
	t.Elt = elt
	return t, nil
}

func (r *resolver) VisitStructType(t *ast.StructType) (ast.Type, error) {
	for i := range t.Fields {
		fd := &t.Fields[i]
		typ, err := r.resolveType(fd.Type)
		if err != nil {
			return nil, err
		}
		fd.Type = typ
	}
	return t, checkDuplicateFieldNames(t)
}

func (r *resolver) VisitFuncType(t *ast.FuncType) (ast.Type, error) {
	if err := r.resolveParams(t.Params); err != nil {
		return nil, err
	}
	return t, r.resolveParams(t.Returns)
}

func (r *resolver) VisitInterfaceType(t *ast.InterfaceType) (ast.Type, error) {
	for i := range t.Methods {
		m := &t.Methods[i]
		if m.Name == "_" {
			return nil, errors.New("blank method name is not allowed")
		}
		typ, err := r.resolveType(m.Type)
		if err != nil {
			return nil, err
		}
		m.Type = typ
	}
	return t, nil
}

// Checks if the given type refers (possibly via several typenames) to a
// struct type.
func checkStructType(typ ast.Type) *ast.StructType {
	for {
		switch t := typ.(type) {
		case *ast.TypeDecl:
			typ = t.Type
		case *ast.StructType:
			return t
		default:
			return nil
		}
	}
}

// Finds the package where the named type TYP is originally declared. Returns
// nil, if TYP does not refer to a typename. If TYP begins a chain of
// typenames, returns the one, which refers directly to the structure.
func findTypeOrigPackage(typ ast.Type) *ast.Package {
	var dcl *ast.TypeDecl
	for d, ok := typ.(*ast.TypeDecl); ok; d, ok = typ.(*ast.TypeDecl) {
		dcl = d
		typ = dcl.Type
	}
	if dcl == nil {
		return nil
	}
	return dcl.File.Pkg
}

//  Finds the field NAME within the structure type STR. Does not consider
//  promoted fields.
func findField(str *ast.StructType, name string) *ast.Field {
	for i := range str.Fields {
		f := &str.Fields[i]
		if name == f.Name {
			return f
		}
	}
	return nil
}

// Removes outermost ParensExpr from an expression.
func removeParens(x ast.Expr) ast.Expr {
	for y, ok := x.(*ast.ParensExpr); ok; y, ok = x.(*ast.ParensExpr) {
		x = y.X
	}
	return x
}

// Checks if the expression X is in the form `ID` or `*ID`, where ID is a TypeName.
func (r *resolver) isType(x ast.Expr) (ast.Type, error) {
	x = removeParens(x)
	switch x := x.(type) {
	case *ast.QualifiedId:
		if x.Id == "_" {
			return nil, errors.New("`_` is not a valid operand or a typename")
		}
		if len(x.Pkg) > 0 {
			if d := r.scope.Lookup(x.Pkg); d == nil {
				return nil, errors.New(x.Pkg + " not declared")
			} else if imp, ok := d.(*ast.ImportDecl); !ok {
				return nil, nil
			} else if d := imp.Pkg.Lookup(x.Id); d == nil {
				return nil, errors.New(x.Pkg + "." + x.Id + " not declared")
			} else if !isExported(x.Id) {
				return nil, errors.New(x.Pkg + "." + x.Id + " is not exported")
			} else if d, ok := d.(*ast.TypeDecl); ok {
				return d, nil
			}
		} else {
			if d := r.scope.Lookup(x.Id); d == nil {
				return nil, errors.New(x.Id + " not declared")
			} else if d, ok := d.(*ast.TypeDecl); ok {
				return d, nil
			}
		}
	case *ast.UnaryExpr:
		if x.Op == '*' {
			typ, err := r.isType(x.X)
			if err != nil {
				return nil, err
			}
			if typ != nil {
				return &ast.PtrType{Off: x.Off, Base: typ}, nil
			}
		}
	}
	return nil, nil
}

// Converts a symbol declaration to an expression to fill the place of an
// OperandName in the AST.
func operandName(d ast.Symbol) ast.Expr {
	switch op := d.(type) {
	case *ast.Var:
		return op
	case *ast.Const:
		return op
	case *ast.FuncDecl:
		return &op.Func
	default:
		panic("not reached")
	}
}

// Declares in SCOPE a sequence of identifiers, which are on the left-hand
// side of the ":=" token and only the identifiers, not already declared in
// SCOPE.  Checks there is at least one such identifier.
func (r *resolver) declareLHS(xs ...ast.Expr) error {
	// Check that the left-hand side contains only identifiers and at least
	// one is not declared in current scope.
	new := false
	for i := range xs {
		x, ok := xs[i].(*ast.QualifiedId)
		if !ok || len(x.Pkg) > 0 {
			return errors.New("non-name on the left side of :=")
		}
		if x.Id == "_" || r.scope.Find(x.Id) != nil {
			continue
		}
		new = true
	}
	if !new {
		return errors.New("no new variables on the left side of :=")
	}
	// Check that there are no duplicate identifiers.
	for i := range xs {
		x := xs[i].(*ast.QualifiedId)
		if x.Id == "_" {
			continue
		}
		ys := xs[i+1:]
		for j := range ys {
			y := ys[j].(*ast.QualifiedId)
			if y.Id == "_" {
				continue
			}
			if x.Id == y.Id {
				return errors.New("duplicate ident on the left side of :=")
			}
		}
	}
	// Do the declaration.
	for _, x := range xs {
		id := x.(*ast.QualifiedId)
		if id.Id == "_" || r.scope.Find(id.Id) != nil {
			continue
		}
		v := &ast.Var{Off: id.Off, File: r.scope.File(), Name: id.Id}
		r.scope.Declare(v.Name, v) // cannot fail here
	}
	return nil
}

// Resolves an expression on the left-hand size of an assignment statement or
// a short variable declaration.  Replaces occurences blank QualifiedId with
// `ast.Blank`.
func (r *resolver) resolveLHS(x ast.Expr) (ast.Expr, error) {
	if x == nil {
		return nil, nil
	}
	if id, ok := x.(*ast.QualifiedId); ok && len(id.Pkg) == 0 && id.Id == "_" {
		return ast.Blank, nil
	}
	return r.resolveExpr(x)
}

func (r *resolver) resolveExpr(x ast.Expr) (ast.Expr, error) {
	if x == nil {
		return nil, nil
	}
	return x.TraverseExpr(r)
}

func (*resolver) VisitConstValue(*ast.ConstValue) (ast.Expr, error) {
	panic("not reached")
}

func (r *resolver) VisitLiteral(x *ast.Literal) (ast.Expr, error) {
	return x, nil // FIXME
}

func (r *resolver) VisitCompLiteral(x *ast.CompLiteral) (ast.Expr, error) {
	typ, err := r.resolveType(x.Type)
	if err != nil {
		return nil, err
	}
	x.Type = typ

	for _, e := range x.Elts {
		v, err := r.resolveExpr(e.Elt)
		if err != nil {
			return nil, err
		}
		e.Elt = v
	}

	// For struct types, do not resolve key expressions: they should be field
	// names.
	if checkStructType(x.Type) != nil {
		return x, nil
	}

	for _, e := range x.Elts {
		k, err := r.resolveExpr(e.Key)
		if err != nil {
			return nil, err
		}
		e.Key = k
	}
	return x, nil
}

func (r *resolver) VisitCall(x *ast.Call) (ast.Expr, error) {
	// A Conversion, which begins with a Typename is parsed as a Call. Check
	// for this case and tranform the Call into a Conversion.
	typ, err := r.isType(x.Func)
	if err != nil {
		return nil, err
	}
	if typ == nil {
		// Not a conversion.
		fn, err := r.resolveExpr(x.Func)
		if err != nil {
			return nil, err
		}
		x.Func = fn
		typ, err := r.resolveType(x.Type)
		if err != nil {
			return nil, err
		}
		x.Type = typ
		for i := range x.Xs {
			y, err := r.resolveExpr(x.Xs[i])
			if err != nil {
				return nil, err
			}
			x.Xs[i] = y
		}
		return x, nil
	} else {
		if x.Type != nil || len(x.Xs) != 1 || x.Ell {
			return nil, errors.New("invalid conversion argument")
		}
		x, err := r.resolveExpr(x.Xs[0])
		if err != nil {
			return nil, err
		}
		return &ast.Conversion{Type: typ, X: x}, nil
	}
}

func (r *resolver) VisitConversion(x *ast.Conversion) (ast.Expr, error) {
	typ, err := r.resolveType(x.Type)
	if err != nil {
		return nil, err
	}
	x.Type = typ
	y, err := r.resolveExpr(x.X)
	if err != nil {
		return nil, err
	}
	x.X = y
	return x, nil
}

func (r *resolver) VisitParensExpr(x *ast.ParensExpr) (ast.Expr, error) {
	return r.resolveExpr(x.X)
}

func (r *resolver) VisitFunc(x *ast.Func) (ast.Expr, error) {
	return x, r.resolveFunc(x)
}

func (r *resolver) VisitTypeAssertion(x *ast.TypeAssertion) (ast.Expr, error) {
	typ, err := r.resolveType(x.Type)
	if err != nil {
		return nil, err
	}
	x.Type = typ
	y, err := r.resolveExpr(x.X)
	if err != nil {
		return nil, err
	}
	x.X = y
	return x, nil
}

func (r *resolver) VisitSelector(x *ast.Selector) (ast.Expr, error) {
	// Check for a MethodExpr parsed as a Selector
	typ, err := r.isType(x.X)
	if err != nil {
		return nil, err
	}
	if typ != nil {
		return &ast.MethodExpr{Type: typ, Id: x.Id}, nil
	}
	y, err := r.resolveExpr(x.X)
	if err != nil {
		return nil, err
	}
	x.X = y
	return x, nil
}

func (r *resolver) VisitIndexExpr(x *ast.IndexExpr) (ast.Expr, error) {
	i, err := r.resolveExpr(x.I)
	if err != nil {
		return nil, err
	}
	x.I = i
	y, err := r.resolveExpr(x.X)
	if err != nil {
		return nil, err
	}
	x.X = y
	return x, nil
}

func (r *resolver) VisitSliceExpr(x *ast.SliceExpr) (ast.Expr, error) {
	lo, err := r.resolveExpr(x.Lo)
	if err != nil {
		return nil, err
	}
	x.Lo = lo
	hi, err := r.resolveExpr(x.Hi)
	if err != nil {
		return nil, err
	}
	x.Hi = hi
	cap, err := r.resolveExpr(x.Cap)
	if err != nil {
		return nil, err
	}
	x.Cap = cap
	y, err := r.resolveExpr(x.X)
	if err != nil {
		return nil, err
	}
	x.X = y
	return x, nil
}

func (r *resolver) VisitUnaryExpr(x *ast.UnaryExpr) (ast.Expr, error) {
	y, err := r.resolveExpr(x.X)
	if err != nil {
		return nil, err
	}
	x.X = y
	return x, nil
}

func (r *resolver) VisitBinaryExpr(x *ast.BinaryExpr) (ast.Expr, error) {
	u, err := r.resolveExpr(x.X)
	if err != nil {
		return nil, err
	}
	x.X = u
	v, err := r.resolveExpr(x.Y)
	if err != nil {
		return nil, err
	}
	x.Y = v
	return x, nil
}

func (*resolver) VisitVar(*ast.Var) (ast.Expr, error) {
	panic("not reached")
}

func (*resolver) VisitConst(*ast.Const) (ast.Expr, error) {
	panic("not reached")
}

func (r *resolver) VisitOperandName(x *ast.QualifiedId) (ast.Expr, error) {
	if x.Id == "_" {
		return nil, errors.New("`_` is not a valid operand")
	}
	if len(x.Pkg) > 0 {
		// Depending on what the constituent names (hereafter referred to by
		// PKG and ID) of the QualifiedId resolve to, there are a few options:
		//  * If PKG refers to a package, then ID must refer to an exported
		//    non-type declaration, in which case the whole QualifiedId is an
		//    OperandName, otherwise it's an error.
		//  * If PKG refers to a type declaration, then the expression is a
		//    MethodExpr
		//  * If PKG refers to a non-type declaration, then the expression is
		//    a Selector
		if d := r.scope.Lookup(x.Pkg); d == nil {
			return nil, errors.New(x.Pkg + " not declared")
		} else if imp, ok := d.(*ast.ImportDecl); ok {
			if d := imp.Pkg.Lookup(x.Id); d == nil {
				return nil, errors.New(x.Pkg + "." + x.Id + " not declared")
			} else if _, ok := d.(*ast.TypeDecl); ok {
				return nil, errors.New("invalid operand " + x.Pkg + "." + x.Id)
			} else if !isExported(x.Id) {
				return nil, errors.New(x.Pkg + "." + x.Id + " is not exported")
			} else {
				return operandName(d), nil
			}
		} else if t, ok := d.(*ast.TypeDecl); ok {
			return &ast.MethodExpr{Type: t, Id: x.Id}, nil
		} else {
			return &ast.Selector{X: operandName(d), Id: x.Id}, nil
		}
	} else {
		// A single identifier must be simply a valid operand.
		if d := r.scope.Lookup(x.Id); d == nil {
			return nil, errors.New(x.Id + " not declared")
		} else if _, ok := d.(*ast.TypeDecl); ok {
			return nil, errors.New("invalid operand " + x.Id)
		} else {
			return operandName(d), nil
		}
	}
}

func (*resolver) VisitMethodExpr(*ast.MethodExpr) (ast.Expr, error) {
	panic("internal error: the parser does not generate MethodExpr")
}

func (r *resolver) resolveBlock(blk *ast.Block) (*ast.Block, error) {
	blk.Up = r.scope
	r.enterScope(blk)
	defer r.exitScope()
	ss := []ast.Stmt{}
	for i := range blk.Body {
		st, err := r.resolveStmt(blk.Body[i])
		if err != nil {
			return nil, err
		}
		if st != nil {
			ss = append(ss, st)
		}
	}
	blk.Body = ss
	return blk, nil
}

func (r *resolver) resolveStmt(s ast.Stmt) (ast.Stmt, error) {
	if s == nil {
		return nil, nil
	}
	return s.TraverseStmt(r)
}

func (r *resolver) VisitTypeDecl(s *ast.TypeDecl) (ast.Stmt, error) {
	if err := r.declareTypeName(s, r.scope.File()); err != nil {
		return nil, err
	}
	return nil, r.resolveTypeDecl(s)
}

func (r *resolver) VisitTypeDeclGroup(s *ast.TypeDeclGroup) (ast.Stmt, error) {
	for _, d := range s.Types {
		if err := r.declareTypeName(d, r.scope.File()); err != nil {
			return nil, err
		}
		if err := r.resolveTypeDecl(d); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (r *resolver) VisitConstDecl(s *ast.ConstDecl) (ast.Stmt, error) {
	if err := r.resolveConst(s); err != nil {
		return nil, err
	}
	return nil, r.declareConst(s, r.scope.File())
}

func (r *resolver) VisitConstDeclGroup(s *ast.ConstDeclGroup) (ast.Stmt, error) {
	return nil, r.resolveConstGroup(s, true)
}

func (r *resolver) VisitVarDecl(s *ast.VarDecl) (ast.Stmt, error) {
	typ, err := r.resolveType(s.Type)
	if err != nil {
		return nil, err
	}
	s.Type = typ
	for i := range s.Init {
		x, err := r.resolveExpr(s.Init[i])
		if err != nil {
			return nil, err
		}
		s.Init[i] = x
	}

	// Even if there are not initialization expressions, create the assignment
	// statement. The RHS side being empty means zero-initialization of the
	// variables.
	init := &ast.AssignStmt{Op: ast.NOP, LHS: make([]ast.Expr, len(s.Names)), RHS: s.Init}
	s.Init = nil
	for i, v := range s.Names {
		init.LHS[i] = v
		v.Init = init
	}

	for _, v := range s.Names {
		if v.Name == "_" {
			continue
		}
		v.File = r.scope.File()
		if err := r.scope.Declare(v.Name, v); err != nil {
			return nil, err
		}
	}
	return init, nil
}

func (r *resolver) VisitVarDeclGroup(s *ast.VarDeclGroup) (ast.Stmt, error) {
	ss := []ast.Stmt{}
	for _, v := range s.Vars {
		st, err := r.VisitVarDecl(v)
		if err != nil {
			return nil, err
		}
		ss = append(ss, st)
	}

	b := &ast.Block{Body: ss}
	b.Up = r.scope
	return b, nil
}

func (*resolver) VisitEmptyStmt(*ast.EmptyStmt) (ast.Stmt, error) {
	return nil, nil
}

func (r *resolver) VisitBlock(s *ast.Block) (ast.Stmt, error) {
	return r.resolveBlock(s)
}

func (r *resolver) VisitLabel(s *ast.Label) (ast.Stmt, error) {
	fn := r.scope.Func()
	s.Blk = r.scope
	if err := fn.DeclareLabel(s.Label, s); err != nil {
		return nil, err
	}
	st, err := r.resolveStmt(s.Stmt)
	if err != nil {
		return nil, err
	}
	s.Stmt = st
	return st, nil
}

func (r *resolver) VisitGoStmt(s *ast.GoStmt) (ast.Stmt, error) {
	x, err := r.resolveExpr(s.X)
	if err != nil {
		return nil, err
	}
	s.X = x
	return s, nil
}

func (r *resolver) VisitReturnStmt(s *ast.ReturnStmt) (ast.Stmt, error) {
	for i := range s.Xs {
		x, err := r.resolveExpr(s.Xs[i])
		if err != nil {
			return nil, err
		}
		s.Xs[i] = x
	}
	return s, nil
}

func (r *resolver) VisitBreakStmt(s *ast.BreakStmt) (ast.Stmt, error) {
	return s, nil
}

func (r *resolver) VisitContinueStmt(s *ast.ContinueStmt) (ast.Stmt, error) {
	return s, nil
}

func (r *resolver) VisitGotoStmt(s *ast.GotoStmt) (ast.Stmt, error) {
	return s, nil
}

func (r *resolver) VisitFallthroughStmt(s *ast.FallthroughStmt) (ast.Stmt, error) {
	return s, nil
}

func (r *resolver) VisitSendStmt(s *ast.SendStmt) (ast.Stmt, error) {
	ch, err := r.resolveExpr(s.Ch)
	if err != nil {
		return nil, err
	}
	s.Ch = ch
	x, err := r.resolveExpr(s.X)
	if err != nil {
		return nil, err
	}
	s.X = x
	return s, nil
}

func (r *resolver) VisitRecvStmt(s *ast.RecvStmt) (ast.Stmt, error) {
	x, err := r.resolveExpr(s.Rcv)
	if err != nil {
		return nil, err
	}
	s.Rcv = x
	if s.Op == ast.DCL {
		var err error
		if s.Y == nil {
			err = r.declareLHS(s.X)
		} else {
			err = r.declareLHS(s.X, s.Y)
		}
		if err != nil {
			return nil, err
		}
	}
	x, err = r.resolveLHS(s.X)
	if err != nil {
		return nil, err
	}
	s.X = x
	y, err := r.resolveLHS(s.Y)
	if err != nil {
		return nil, err
	}
	s.Y = y
	return s, nil
}

func (r *resolver) VisitIncStmt(s *ast.IncStmt) (ast.Stmt, error) {
	x, err := r.resolveExpr(s.X)
	if err != nil {
		return nil, err
	}
	s.X = x
	return s, nil
}

func (r *resolver) VisitDecStmt(s *ast.DecStmt) (ast.Stmt, error) {
	x, err := r.resolveExpr(s.X)
	if err != nil {
		return nil, err
	}
	s.X = x
	return s, nil
}

func (r *resolver) VisitAssignStmt(s *ast.AssignStmt) (ast.Stmt, error) {
	// Resolve right-hand side(s).
	for i := range s.RHS {
		x, err := r.resolveExpr(s.RHS[i])
		if err != nil {
			return nil, err
		}
		s.RHS[i] = x

	}
	if s.Op == ast.DCL {
		if err := r.declareLHS(s.LHS...); err != nil {
			return nil, err
		}
	}
	// Resolve left-hand side(s).
	for i := range s.LHS {
		x, err := r.resolveLHS(s.LHS[i])
		if err != nil {
			return nil, err
		}
		s.LHS[i] = x
	}
	// The statement becomes an ordinary assignment.
	s.Op = ast.NOP
	return s, nil
}

func (r *resolver) VisitExprStmt(s *ast.ExprStmt) (ast.Stmt, error) {
	x, err := r.resolveExpr(s.X)
	if err != nil {
		return nil, err
	}
	s.X = x
	return s, nil
}

func (r *resolver) VisitIfStmt(s *ast.IfStmt) (ast.Stmt, error) {
	s.Up = r.scope
	r.enterScope(s)
	defer r.exitScope()

	st, err := r.resolveStmt(s.Init)
	if err != nil {
		return nil, err
	}
	s.Init = st

	x, err := r.resolveExpr(s.Cond)
	if err != nil {
		return nil, err
	}
	s.Cond = x

	blk, err := r.resolveBlock(s.Then)
	if err != nil {
		return nil, err
	}
	s.Then = blk

	st, err = r.resolveStmt(s.Else)
	if err != nil {
		return nil, err
	}
	s.Else = st
	return s, nil
}

func (r *resolver) VisitForStmt(s *ast.ForStmt) (ast.Stmt, error) {
	s.Up = r.scope
	r.enterScope(s)
	defer r.exitScope()

	st, err := r.resolveStmt(s.Init)
	if err != nil {
		return nil, err
	}
	s.Init = st

	x, err := r.resolveExpr(s.Cond)
	if err != nil {
		return nil, err
	}
	s.Cond = x

	// The Post statement cannot be a declaration.
	if d, ok := s.Post.(*ast.AssignStmt); ok && d.Op == ast.DCL {
		return nil, errors.New("cannot declare in for post-statement")
	}

	st, err = r.resolveStmt(s.Post)
	if err != nil {
		return nil, err
	}
	s.Post = st

	blk, err := r.resolveBlock(s.Blk)
	if err != nil {
		return nil, err
	}
	s.Blk = blk
	return s, nil
}

func (r *resolver) VisitForRangeStmt(s *ast.ForRangeStmt) (ast.Stmt, error) {
	s.Up = r.scope
	r.enterScope(s)
	defer r.exitScope()

	// Resolve the range expression before potential LHS variables are
	// declared.
	x, err := r.resolveExpr(s.Range)
	if err != nil {
		return nil, err
	}
	s.Range = x

	// Declare the LHS variables.
	if s.Op == ast.DCL {
		if err := r.declareLHS(s.LHS...); err != nil {
			return nil, err
		}
	}

	// Resolve the LHS.
	for i := range s.LHS {
		x, err := r.resolveLHS(s.LHS[i])
		if err != nil {
			return nil, err
		}
		s.LHS[i] = x

	}

	// Resolve the loop body.
	b, err := r.resolveBlock(s.Blk)
	if err != nil {
		return nil, err
	}
	s.Blk = b
	return s, nil
}

func (r *resolver) VisitDeferStmt(s *ast.DeferStmt) (ast.Stmt, error) {
	x, err := r.resolveExpr(s.X)
	if err != nil {
		return nil, err
	}
	s.X = x
	return s, nil
}

func (r *resolver) VisitExprSwitchStmt(s *ast.ExprSwitchStmt) (ast.Stmt, error) {
	s.Up = r.scope
	r.enterScope(s)
	defer r.exitScope()

	st, err := r.resolveStmt(s.Init)
	if err != nil {
		return nil, err
	}
	s.Init = st

	x, err := r.resolveExpr(s.X)
	if err != nil {
		return nil, err
	}
	s.X = x

	for i := range s.Cases {
		c := &s.Cases[i]
		for i := range c.Xs {
			x, err := r.resolveExpr(c.Xs[i])
			if err != nil {
				return nil, err
			}
			c.Xs[i] = x

		}
		b, err := r.resolveBlock(c.Blk)
		if err != nil {
			return nil, err
		}
		c.Blk = b

	}
	return s, nil
}

func (r *resolver) VisitTypeSwitchStmt(s *ast.TypeSwitchStmt) (ast.Stmt, error) {
	s.Up = r.scope
	r.enterScope(s)
	defer r.exitScope()

	st, err := r.resolveStmt(s.Init)
	if err != nil {
		return nil, err
	}
	s.Init = st

	x, err := r.resolveExpr(s.X)
	if err != nil {
		return nil, err
	}
	s.X = x

	for i := range s.Cases {
		c := &s.Cases[i]
		for i := range c.Types {
			t, err := r.resolveType(c.Types[i])
			if err != nil {
				return nil, err
			}
			c.Types[i] = t
		}
		// Declare the variable from the type switch guard in every block,
		// except the default. Block has no declarations whatsoever at this
		// point, so the below `Declare` cannot fail.
		if len(s.Id) > 0 && len(c.Types) > 0 {
			v := &ast.Var{Off: s.Off, File: r.scope.File(), Name: s.Id}
			c.Blk.Declare(v.Name, v)
		}
		b, err := r.resolveBlock(c.Blk)
		if err != nil {
			return nil, err
		}
		c.Blk = b
	}
	return s, nil
}

func (r *resolver) VisitSelectStmt(s *ast.SelectStmt) (ast.Stmt, error) {
	s.Up = r.scope
	r.enterScope(s)
	defer r.exitScope()

	for i := range s.Comms {
		c := &s.Comms[i]
		// Resolve the CommCase statement in the scope of the clause block, so
		// eventual variable names are declared in that scope.
		c.Blk.Up = r.scope
		r.enterScope(c.Blk)
		st, err := r.resolveStmt(c.Comm)
		r.exitScope()
		if err != nil {
			return nil, err
		}
		c.Comm = st
		b, err := r.resolveBlock(c.Blk)
		if err != nil {
			return nil, err
		}
		c.Blk = b
	}
	return s, nil
}
