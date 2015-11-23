package parser

import (
	"errors"
	"golang/ast"
	"golang/constexpr"
	"golang/scanner"
	"path/filepath"
	"unicode"
	"unicode/utf8"
)

func isExported(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsLetter(r) && unicode.IsUpper(r)
}

func isValidPackageName(name string) bool {
	ch, n := utf8.DecodeRuneInString(name)
	if !isLetter(ch) {
		return false
	}
	name = name[n:]
	ch, n = utf8.DecodeRuneInString(name)
	for n > 0 {
		if !isLetter(ch) && !isDigit(ch) {
			return false
		}
		ch, n = utf8.DecodeRuneInString(name)
	}
	return true
}

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' ||
		'A' <= ch && ch <= 'Z' ||
		ch == '_' || ch >= 0x80 && unicode.IsLetter(ch)
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9' || ch >= 0x80 && unicode.IsDigit(ch)
}

func ResolvePackage(
	p *ast.UnresolvedPackage, loc ast.PackageLocator) (*ast.Package, error) {

	pkg := &ast.Package{
		Path:  p.Path,
		Name:  p.Name,
		Files: nil,
		Decls: make(map[string]ast.Symbol),
		Deps:  make(map[string]*ast.Import),
	}

	// Insert package and file scope declarations into the symbol table.
	for _, f := range p.Files {
		file, err := declareTopLevel(f, pkg, loc)
		if err != nil {
			return nil, err
		}
		pkg.Files = append(pkg.Files, file)
	}

	// Resolve right-hand sides of package-level variable initialization
	// statements.
	for _, st := range pkg.Init {
		for i := range st.RHS {
			x, err := resolveExpr(st.RHS[i], pkg)
			if err != nil {
				return nil, err
			}
			st.RHS[i] = x
		}
	}

	// Declare lower level names and resolve all.
	for i, f := range p.Files {
		err := resolveTopLevel(pkg.Files[i], f.Decls)
		if err != nil {
			return nil, err
		}
	}

	return pkg, nil
}

func declareTopLevel(
	f *ast.UnresolvedFile, pkg *ast.Package, loc ast.PackageLocator) (*ast.File, error) {

	file := &ast.File{
		Off:     f.Off,
		Pkg:     pkg,
		PkgName: f.PkgName,
		Imports: f.Imports,
		Name:    f.Name,
		SrcMap:  f.SrcMap,
		Decls:   make(map[string]ast.Symbol),
	}

	// Declare imported package names.
	for _, i := range file.Imports {
		path := constexpr.String(i.Path)
		dep, ok := pkg.Deps[path]
		if !ok {
			p, err := loc.FindPackage(path)
			if err != nil {
				return nil, err
			}
			dep = &ast.Import{Pkg: p}
			pkg.Deps[path] = dep
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
				i.Name = filepath.Base(path)
			}
			if !isValidPackageName(i.Name) {
				return nil, errors.New(i.Name + " is not a valid package name") // FIXME
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
			err = declareType(d, file, pkg)
		case *ast.TypeDeclGroup:
			for _, d := range d.Types {
				if err := declareType(d, file, pkg); err != nil {
					return nil, err
				}
			}
		case *ast.ConstDecl:
			err = declareConst(d, file, pkg, ast.InvalidIota)
		case *ast.ConstDeclGroup:
			iota := 0
			for _, d := range d.Consts {
				if err := declareConst(d, file, pkg, iota); err != nil {
					return nil, err
				}
				iota++
			}
		case *ast.VarDecl:
			err = declarePkgVar(d, file)
		case *ast.VarDeclGroup:
			for _, d := range d.Vars {
				if err := declarePkgVar(d, file); err != nil {
					return nil, err
				}
			}
		case *ast.FuncDecl:
			err = declareFunc(d, file, pkg)
		default:
			panic("not reached")
		}
		if err != nil {
			return nil, err
		}
	}

	return file, nil
}

func resolveTopLevel(f *ast.File, ds []ast.Decl) error {
	for _, d := range ds {
		var err error
		switch d := d.(type) {
		case *ast.TypeDecl:
			err = resolveTypeDecl(d, f)
		case *ast.TypeDeclGroup:
			for _, d := range d.Types {
				if err = resolveTypeDecl(d, f); err != nil {
					break
				}
			}
		case *ast.ConstDecl:
			err = resolveConstDecl(d, f)
		case *ast.ConstDeclGroup:
			for _, d := range d.Consts {
				if err = resolveConstDecl(d, f); err != nil {
					break
				}
			}
		case *ast.VarDecl:
			err = resolvePkgVar(d, f)
		case *ast.VarDeclGroup:
			for _, d := range d.Vars {
				if err = resolvePkgVar(d, f); err != nil {
					break
				}
			}
		case *ast.FuncDecl:
			err = resolveFunc(&d.Func, f)
		default:
			panic("not reached")
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func declareType(t *ast.TypeDecl, file *ast.File, scope ast.Scope) error {
	t.File = file
	return scope.Declare(t.Name, t)
}

func resolveTypeDecl(t *ast.TypeDecl, scope ast.Scope) error {
	if typ, err := resolveType(t.Type, scope); err != nil {
		return err
	} else {
		t.Type = typ
		return nil
	}
}

func declareConst(cst *ast.ConstDecl, file *ast.File, scope ast.Scope, iota int) error {
	if len(cst.Values) > 0 {
		if len(cst.Names) != len(cst.Values) {
			return errors.New("number of names does not match the number of values")
		}
	}
	for i, c := range cst.Names {
		c.File = file
		if i < len(cst.Values) {
			c.Init = cst.Values[i]
		}
		c.Iota = iota
		if err := scope.Declare(c.Name, c); err != nil {
			return err
		}
	}
	return nil
}

func resolveConstDecl(cst *ast.ConstDecl, scope ast.Scope) error {
	if typ, err := resolveType(cst.Type, scope); err != nil {
		return err
	} else {
		cst.Type = typ
	}
	for i := range cst.Values {
		if x, err := resolveExpr(cst.Values[i], scope); err != nil {
			return err
		} else {
			cst.Values[i] = x
		}
	}
	return nil
}

func resolveConst(cst *ast.Const, scope ast.Scope) error {
	if typ, err := resolveType(cst.Type, scope); err != nil {
		return err
	} else {
		cst.Type = typ
	}
	if x, err := resolveExpr(cst.Init, scope); err != nil {
		return err
	} else {
		cst.Init = x
	}
	return nil
}

// Declares a variable at package scope and creates the assignment statement
// for the initialization.
func declarePkgVar(vr *ast.VarDecl, file *ast.File) error {
	if len(vr.Init) > 0 {
		// Create an assignment statement, which is executed, according to
		// dependency order, upon package initialization, to set initial
		// values of the package-level variables. Make every variable refer to
		// its initialization statement.
		lhs := make([]ast.Expr, len(vr.Names))
		init := &ast.AssignStmt{Op: '=', LHS: lhs, RHS: vr.Init}
		vr.Init = nil
		for i, v := range vr.Names {
			lhs[i] = v
			v.Init = init
		}
		file.Pkg.Init = append(file.Pkg.Init, init)
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

func resolvePkgVar(d *ast.VarDecl, scope ast.Scope) error {
	typ, err := resolveType(d.Type, scope)
	if err != nil {
		return err
	}
	for _, v := range d.Names {
		v.Type = typ
	}
	return nil
}

func declareBlockVar(vr *ast.VarDecl, file *ast.File, scope ast.Scope) (ast.Stmt, error) {
	rhs := vr.Init
	vr.Init = nil
	lhs := make([]ast.Expr, len(vr.Names))
	for i, v := range vr.Names {
		v.File = file
		if err := scope.Declare(v.Name, v); err != nil {
			return nil, err
		}
		lhs[i] = v
	}
	return &ast.AssignStmt{Op: '=', LHS: lhs, RHS: rhs}, nil
}

func resolveBlockVar(v *ast.VarDecl, scope ast.Scope) error {
	if typ, err := resolveType(v.Type, scope); err != nil {
		return err
	} else {
		v.Type = typ
	}
	for i := range v.Init {
		if x, err := resolveExpr(v.Init[i], scope); err != nil {
			return err
		} else {
			v.Init[i] = x
		}
	}
	return nil
}

func declareFunc(fn *ast.FuncDecl, file *ast.File, scope ast.Scope) error {
	if fn.Func.Recv != nil {
		return nil
	}
	fn.File = file
	return scope.Declare(fn.Name, fn)
}

func resolveFunc(fn *ast.Func, scope ast.Scope) error {
	if r := fn.Recv; r != nil {
		if typ, err := resolveType(r.Type, scope); err != nil {
			return err
		} else {
			r.Type = typ
		}
	}
	if typ, err := resolveType(fn.Sig, scope); err != nil {
		return err
	} else {
		fn.Sig = typ.(*ast.FuncType)
	}
	// Declare parameter names in the scope of the function block.
	if fn.Blk != nil {
		for i := range fn.Sig.Params {
			p := &fn.Sig.Params[i]
			if len(p.Name) == 0 {
				return errors.New("missing formal parameter name")
			}
			v := &ast.Var{Off: p.Off, File: scope.File(), Name: p.Name, Type: p.Type}
			if err := fn.Blk.Declare(v.Name, v); err != nil {
				return err
			}
		}
		if blk, err := resolveBlock(fn.Blk, scope); err != nil {
			return err
		} else {
			fn.Blk = blk
		}
	}
	return nil
}

func lookupIdent(id *ast.QualifiedId, scope ast.Scope) (ast.Symbol, error) {
	if len(id.Pkg) > 0 {
		if d := scope.Lookup(id.Pkg); d == nil {
			return nil, errors.New("package name " + id.Pkg + " not declared")
		} else if imp, ok := d.(*ast.ImportDecl); !ok {
			return nil, errors.New(id.Pkg + " does not refer to package name")
		} else if d := imp.Pkg.Lookup(id.Id); d == nil {
			return nil, errors.New(id.Pkg + "." + id.Id + " not declared")
		} else if !isExported(id.Id) {
			return nil, errors.New(id.Pkg + "." + id.Id + " is not exported")
		} else {
			return d, nil
		}
	} else {
		if d := scope.Lookup(id.Id); d == nil {
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
func resolveParams(ps []ast.Param, scope ast.Scope) (err error) {
	var carry ast.Type
	n := len(ps)
	for i := n - 1; i >= 0; i-- {
		p := &ps[i]
		if p.Type != nil {
			carry, err = resolveType(p.Type, scope)
			if err != nil {
				return
			}
		}
		p.Type = carry
	}
	return
}

func resolveType(t ast.Type, scope ast.Scope) (ast.Type, error) {
	if t == nil {
		return nil, nil
	}
	switch t := t.(type) {
	case *ast.QualifiedId:
		if d, err := lookupIdent(t, scope); err != nil {
			return nil, err
		} else if d, ok := d.(*ast.TypeDecl); !ok {
			return nil, errors.New(t.Id + " is not a typename")
		} else {
			return d, nil
		}
	case *ast.ArrayType:
		if elt, err := resolveType(t.Elt, scope); err != nil {
			return nil, err
		} else {
			t.Elt = elt
			return t, nil
		}
	case *ast.SliceType:
		if elt, err := resolveType(t.Elt, scope); err != nil {
			return nil, err
		} else {
			t.Elt = elt
			return t, nil
		}
	case *ast.PtrType:
		if b, err := resolveType(t.Base, scope); err != nil {
			return nil, err
		} else {
			t.Base = b
			return t, nil
		}
	case *ast.MapType:
		if key, err := resolveType(t.Key, scope); err != nil {
			return nil, err
		} else if elt, err := resolveType(t.Elt, scope); err != nil {
			return nil, err
		} else {
			t.Key = key
			t.Elt = elt
			return t, nil
		}
	case *ast.ChanType:
		if elt, err := resolveType(t.Elt, scope); err != nil {
			return nil, err
		} else {
			t.Elt = elt
			return t, nil
		}
	case *ast.StructType:
		for i := range t.Fields {
			fd := &t.Fields[i]
			if typ, err := resolveType(fd.Type, scope); err != nil {
				return nil, err
			} else {
				fd.Type = typ
			}
		}
		if err := checkDuplicateFieldNames(t); err != nil {
			return nil, err
		}
		return t, nil
	case *ast.FuncType:
		if err := resolveParams(t.Params, scope); err != nil {
			return nil, err
		}
		if err := resolveParams(t.Returns, scope); err != nil {
			return nil, err
		}
		return t, nil
	case *ast.InterfaceType:
		for i := range t.Methods {
			m := &t.Methods[i]
			if typ, err := resolveType(m.Type, scope); err != nil {
				return nil, err
			} else {
				m.Type = typ
			}
		}
		return t, nil
	default:
		panic("not reached")
	}
}

// Checks if the given type refers (possibly via several typenames) to a
// struct type and returns it. Otherwise, returns nil.
func asStructType(typ ast.Type) *ast.StructType {
L:
	switch t := typ.(type) {
	case *ast.StructType:
		return t
	case *ast.TypeDecl:
		typ = t.Type
		goto L
	default:
		return nil
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
func isType(x ast.Expr, scope ast.Scope) (ast.Type, error) {
	x = removeParens(x)
	switch x := x.(type) {
	case *ast.QualifiedId:
		if len(x.Pkg) > 0 {
			if d := scope.Lookup(x.Pkg); d == nil {
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
			if d := scope.Lookup(x.Id); d == nil {
				return nil, errors.New(x.Id + " not declared")
			} else if d, ok := d.(*ast.TypeDecl); ok {
				return d, nil
			}
		}
	case *ast.UnaryExpr:
		if x.Op == '*' {
			typ, err := isType(x.X, scope)
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
	if v, ok := d.(*ast.Var); ok {
		return v
	} else if c, ok := d.(*ast.Const); ok {
		return c
	} else {
		return d.(*ast.FuncDecl)
	}
}

// Resolves the identifiers in an Expression. Removes occurances of
// ParensExpr.
func resolveExpr(x ast.Expr, scope ast.Scope) (ast.Expr, error) {
	if x == nil {
		return nil, nil
	}
	switch x := x.(type) {
	case *ast.Literal:
		return x, nil // FIXME
	case *ast.CompLiteral:
		if typ, err := resolveType(x.Type, scope); err != nil {
			return nil, err
		} else {
			x.Type = typ
		}
		// Resolve values.
		for _, e := range x.Elts {
			if v, err := resolveExpr(e.Value, scope); err != nil {
				return nil, err
			} else {
				e.Value = v
			}
		}
		// For struct types, check the keys are fields in the structure,
		// otherwise resolve the key expressions as usual.
		str := asStructType(x.Type)
		if str == nil {
			for _, e := range x.Elts {
				if k, err := resolveExpr(e.Key, scope); err != nil {
					return nil, err
				} else {
					e.Key = k
				}
			}
			return x, nil
		}
		for _, e := range x.Elts {
			if e.Key == nil {
				continue
			}
			id, ok := e.Key.(*ast.QualifiedId)
			if !ok || len(id.Pkg) > 0 || id.Id == "_" {
				return nil, errors.New("key is not a valid field name")
			}
			if findField(str, id.Id) == nil {
				return nil, errors.New(id.Id + " field name not found")
			}
			if !isExported(id.Id) {
				// If the field name is not exported, the structure must be
				// decared in the same package that contains this composite
				// literal.
				if findTypeOrigPackage(x.Type) != scope.Package() {
					return nil, errors.New("field " + id.Id + " is not accessible")
				}
			}
		}
	case *ast.Call:
		// A Conversion, which begins with a Typename is parsed as a
		// Call. Check for this case and tranform the Call into a Conversion.
		if typ, err := isType(x.Func, scope); err != nil {
			return nil, err
		} else if typ != nil {
			if x.Type != nil || len(x.Xs) != 1 || x.Ell {
				return nil, errors.New("invalid conversion argument")
			}
			if x, err := resolveExpr(x.Xs[0], scope); err != nil {
				return nil, err
			} else {
				return &ast.Conversion{Type: typ, X: x}, nil
			}
		}
		// Not a conversion.
		if fn, err := resolveExpr(x.Func, scope); err != nil {
			return nil, err
		} else {
			x.Func = fn
		}
		if typ, err := resolveType(x.Type, scope); err != nil {
			return nil, err
		} else {
			x.Type = typ
		}
		for i := range x.Xs {
			if y, err := resolveExpr(x.Xs[i], scope); err != nil {
				return nil, err
			} else {
				x.Xs[i] = y
			}
		}
		return x, nil
	case *ast.Conversion:
		if typ, err := resolveType(x.Type, scope); err != nil {
			return nil, err
		} else {
			x.Type = typ
		}
		if y, err := resolveExpr(x.X, scope); err != nil {
			return nil, err
		} else {
			x.X = y
		}
		return x, nil
	case *ast.ParensExpr:
		return resolveExpr(x.X, scope)
	case *ast.Func:
		if sig, err := resolveType(x.Sig, scope); err != nil {
			return nil, err
		} else {
			x.Sig = sig.(*ast.FuncType)
		}
		if blk, err := resolveBlock(x.Blk, scope); err != nil {
			return nil, err
		} else {
			x.Blk = blk
		}
		return x, nil
	case *ast.TypeAssertion:
		if typ, err := resolveType(x.Type, scope); err != nil {
			return nil, err
		} else {
			x.Type = typ
		}
		if y, err := resolveExpr(x.X, scope); err != nil {
			return nil, err
		} else {
			x.X = y
		}
		return x, nil
	case *ast.Selector:
		// Check for a MethodExpr parsed as a Selector
		if typ, err := isType(x.X, scope); err != nil {
			return nil, err
		} else if typ != nil {
			return &ast.MethodExpr{Type: typ, Id: x.Id}, nil
		}
		if y, err := resolveExpr(x.X, scope); err != nil {
			return nil, err
		} else {
			x.X = y
		}
		return x, nil
	case *ast.IndexExpr:
		if i, err := resolveExpr(x.I, scope); err != nil {
			return nil, err
		} else {
			x.I = i
		}
		if y, err := resolveExpr(x.X, scope); err != nil {
			return nil, err
		} else {
			x.X = y
		}
		return x, nil
	case *ast.SliceExpr:
		if lo, err := resolveExpr(x.Lo, scope); err != nil {
			return nil, err
		} else {
			x.Lo = lo
		}
		if hi, err := resolveExpr(x.Hi, scope); err != nil {
			return nil, err
		} else {
			x.Hi = hi
		}
		if cap, err := resolveExpr(x.Cap, scope); err != nil {
			return nil, err
		} else {
			x.Cap = cap
		}
		if y, err := resolveExpr(x.X, scope); err != nil {
			return nil, err
		} else {
			x.X = y
		}
		return x, nil
	case *ast.UnaryExpr:
		if y, err := resolveExpr(x.X, scope); err != nil {
			return nil, err
		} else {
			x.X = y
		}
		return x, nil
	case *ast.BinaryExpr:
		if u, err := resolveExpr(x.X, scope); err != nil {
			return nil, err
		} else {
			x.X = u
		}
		if v, err := resolveExpr(x.Y, scope); err != nil {
			return nil, err
		} else {
			x.Y = v
		}
		return x, nil
	case *ast.QualifiedId:
		if len(x.Pkg) > 0 {
			// Depending on what the constituent names (hereafter referred to
			// by PKG and ID) of the QualifiedId resolve to, there are a few
			// options:
			//  * If PKG refers to a package, then ID must refer to an
			//    exported non-type declaration, in which case the whole
			//    QualifiedId is an OperandName, otherwise it's an error.
			//  * If PKG refers to a type declaration, then the expression is
			//    a MethodExpr
			//  * If PKG refers to a non-type declaration, then the expression
			//    is a Selector
			if d := scope.Lookup(x.Pkg); d == nil {
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
			if d := scope.Lookup(x.Id); d == nil {
				return nil, errors.New(x.Id + " not declared")
			} else if _, ok := d.(*ast.TypeDecl); ok {
				return nil, errors.New("invalid operand " + x.Id)
			} else {
				return operandName(d), nil
			}
		}
	case *ast.MethodExpr:
		panic("internal error: the parser does not generate MethodExpr")
	default:
		panic("not reached")
	}
	return x, nil
}

func resolveBlock(blk *ast.Block, scope ast.Scope) (*ast.Block, error) {
	blk.Up = scope
	ss := []ast.Stmt{}
	for i := range blk.Body {
		if st, err := resolveStmt(blk.Body[i], blk); err != nil {
			return nil, err
		} else if st != nil {
			ss = append(ss, st)
		}
	}
	blk.Body = ss
	return blk, nil
}

func resolveStmt(stmt ast.Stmt, scope ast.Scope) (ast.Stmt, error) {
	switch s := stmt.(type) {
	case *ast.TypeDecl:
		if err := declareType(s, scope.File(), scope); err != nil {
			return nil, err
		}
		if err := resolveTypeDecl(s, scope); err != nil {
			return nil, err
		}
		return nil, nil
	case *ast.TypeDeclGroup:
		for _, d := range s.Types {
			if err := declareType(d, scope.File(), scope); err != nil {
				return nil, err
			}
			if err := resolveTypeDecl(d, scope); err != nil {
				return nil, err
			}
		}
		return nil, nil
	case *ast.ConstDecl:
		if err := resolveConstDecl(s, scope); err != nil {
			return nil, err
		}
		if err := declareConst(s, scope.File(), scope, ast.InvalidIota); err != nil {
			return nil, err
		}
		return nil, nil
	case *ast.ConstDeclGroup:
		iota := 0
		for _, d := range s.Consts {
			if err := resolveConstDecl(d, scope); err != nil {
				return nil, err
			}
			if err := declareConst(d, scope.File(), scope, iota); err != nil {
				return nil, err
			}
		}
		return nil, nil
	case *ast.VarDecl:
		if err := resolveBlockVar(s, scope); err != nil {
			return nil, err
		}
		if st, err := declareBlockVar(s, scope.File(), scope); err != nil {
			return nil, err
		} else {
			return st, nil
		}
	case *ast.VarDeclGroup:
		ss := []ast.Stmt{}
		for _, v := range s.Vars {
			if err := resolveBlockVar(v, scope); err != nil {
				return nil, err
			}
			if st, err := declareBlockVar(v, scope.File(), scope); err != nil {
				return nil, err
			} else {
				ss = append(ss, st)
			}
		}
		return &ast.Block{Up: scope, Body: ss, Decls: make(map[string]ast.Symbol)}, nil
	case *ast.EmptyStmt:
		return nil, nil
	case *ast.Block:
		if blk, err := resolveBlock(s, scope); err != nil {
			return nil, err
		} else {
			return blk, nil
		}
	case *ast.LabeledStmt:
		// FIXME: label at file scope
		if st, err := resolveStmt(s.Stmt, scope); err != nil {
			return nil, err
		} else {
			return st, nil
		}
	case *ast.GoStmt:
		if x, err := resolveExpr(s.X, scope); err != nil {
			return nil, err
		} else {
			s.X = x
		}
		return s, nil
	case *ast.ReturnStmt:
		for i := range s.Xs {
			if x, err := resolveExpr(s.Xs[i], scope); err != nil {
				return nil, err
			} else {
				s.Xs[i] = x
			}
		}
		return s, nil
	case *ast.BreakStmt:
		return s, nil
	case *ast.ContinueStmt:
		return s, nil
	case *ast.GotoStmt:
		return s, nil
	case *ast.FallthroughStmt:
		return s, nil
	case *ast.SendStmt:
		if ch, err := resolveExpr(s.Ch, scope); err != nil {
			return nil, err
		} else {
			s.Ch = ch
		}
		if x, err := resolveExpr(s.X, scope); err != nil {
			return nil, err
		} else {
			s.X = x
		}
		return s, nil
	case *ast.RecvStmt:
		if x, err := resolveExpr(s.Rcv, scope); err != nil {
			return nil, err
		} else {
			s.Rcv = x
		}
		if s.Op == scanner.DEFINE {
			if id, ok := s.X.(*ast.QualifiedId); !ok || len(id.Pkg) > 0 {
				return nil, errors.New("non-name on the left size of :=")
			} else {
				v := &ast.Var{Off: id.Off, File: scope.File(), Name: id.Id}
				if err := scope.Declare(v.Name, v); err != nil {
					return nil, err
				}
				s.X = v
			}
			if id, ok := s.Y.(*ast.QualifiedId); !ok || len(id.Pkg) > 0 {
				return nil, errors.New("non-name on the left size of :=")
			} else {
				v := &ast.Var{Off: id.Off, File: scope.File(), Name: id.Id}
				if err := scope.Declare(v.Name, v); err != nil {
					return nil, err
				}
				s.Y = v
			}
		} else {
			if s.X != nil {
				if x, err := resolveExpr(s.X, scope); err != nil {
					return nil, err
				} else {
					s.X = x
				}
			}
			if s.Y != nil {
				if y, err := resolveExpr(s.Y, scope); err != nil {
					return nil, err
				} else {
					s.Y = y
				}
			}
		}
		return s, nil
	case *ast.IncStmt:
		if x, err := resolveExpr(s.X, scope); err != nil {
			return nil, err
		} else {
			s.X = x
		}
		return s, nil
	case *ast.DecStmt:
		if x, err := resolveExpr(s.X, scope); err != nil {
			return nil, err
		} else {
			s.X = x
		}
		return s, nil
	case *ast.AssignStmt:
		if len(s.LHS) != len(s.RHS) {
			return nil, errors.New("assignment count mismatch")
		}
		// Resolve right-hand side(s).
		for i := range s.RHS {
			if x, err := resolveExpr(s.RHS[i], scope); err != nil {
				return nil, err
			} else {
				s.RHS[i] = x
			}
		}
		// For short variable declarations, declare only the identifiers, not
		// already declared in the current scope. Check there is at least one
		// such identifier.
		if s.Op == scanner.DEFINE {
			newvar := false
			for i := range s.LHS {
				if id, ok := s.LHS[i].(*ast.QualifiedId); ok && len(id.Pkg) == 0 {
					if d := scope.Find(id.Id); d == nil {
						v := &ast.Var{
							Off:  id.Off,
							File: scope.File(),
							Name: id.Id,
						}
						if err := scope.Declare(v.Name, v); err != nil {
							return nil, err
						}
						newvar = true
					}
				}
			}
			if !newvar {
				return nil, errors.New("no new variables on the left side of :=")
			}
		}
		// Resolve left-hand side(s).
		for i := range s.LHS {
			if x, err := resolveExpr(s.LHS[i], scope); err != nil {
				return nil, err
			} else {
				s.LHS[i] = x
			}
		}
		// The statement becomes an ordinary assignment.
		s.Op = '='
		return s, nil
	case *ast.ExprStmt:
		if x, err := resolveExpr(s.X, scope); err != nil {
			return nil, err
		} else {
			s.X = x
		}
		return s, nil
	case *ast.IfStmt:
		// If the initial statement is a short variable declaration, put an
		// extra block around the if.
		if d, ok := s.Init.(*ast.AssignStmt); ok && d.Op == scanner.DEFINE {
			blk := &ast.Block{
				Body:  make([]ast.Stmt, 2),
				Decls: make(map[string]ast.Symbol),
			}
			blk.Body[0] = s.Init
			blk.Body[1] = s
			s.Init = nil
			if b, err := resolveBlock(blk, scope); err != nil {
				return nil, err
			} else {
				return b, nil
			}
		} else {
			if x, err := resolveExpr(s.Cond, scope); err != nil {
				return nil, err
			} else {
				s.Cond = x
			}
			if blk, err := resolveBlock(s.Then, scope); err != nil {
				return nil, err
			} else {
				s.Then = blk
			}
			if st, err := resolveStmt(s.Else, scope); err != nil {
				return nil, err
			} else {
				s.Else = st
			}
			return s, nil
		}
	case *ast.ForStmt:
		// If the initial statement is a short variable declaration, put an
		// extra block around the for.
		if d, ok := s.Init.(*ast.AssignStmt); ok && d.Op == scanner.DEFINE {
			blk := &ast.Block{
				Body:  make([]ast.Stmt, 2),
				Decls: make(map[string]ast.Symbol),
			}
			blk.Body[0] = s.Init
			blk.Body[1] = s
			s.Init = nil
			if b, err := resolveBlock(blk, scope); err != nil {
				return nil, err
			} else {
				return b, nil
			}
		} else {
			if x, err := resolveExpr(s.Cond, scope); err != nil {
				return nil, err
			} else {
				s.Cond = x
			}
			// The Post statement cannot be a declaration.
			if d, ok := s.Post.(*ast.AssignStmt); ok && d.Op == scanner.DEFINE {
				return nil, errors.New("cannot declare in for post-statement")
			}
			if st, err := resolveStmt(s.Post, scope); err != nil {
				return nil, err
			} else {
				s.Post = st
			}
			if blk, err := resolveBlock(s.Blk, scope); err != nil {
				return nil, err
			} else {
				s.Blk = blk
			}
			return s, nil
		}
	case *ast.ForRangeStmt:
		// If the range-for declares new variables, put a block around the for
		// and declare the variables in this block.
		if s.Op == scanner.DEFINE {
			blk := &ast.Block{
				Up:    scope,
				Body:  make([]ast.Stmt, 2),
				Decls: make(map[string]ast.Symbol),
			}
			for i, x := range s.LHS {
				if id, ok := x.(*ast.QualifiedId); ok && len(id.Pkg) == 0 {
					v := &ast.Var{Off: id.Off, File: scope.File(), Name: id.Id}
					if err := blk.Declare(v.Name, v); err != nil {
						return nil, err
					}
					s.LHS[i] = v
				} else {
					return nil, errors.New("non-name on the left side of :=")
				}
			}
			// Resolve the range expression in the current scope.
			if x, err := resolveExpr(s.Range, scope); err != nil {
				return nil, err
			} else {
				s.Range = x
			}
			// Resolve the loop body in the scope of the new block.
			if b, err := resolveBlock(s.Blk, blk); err != nil {
				return nil, err
			} else {
				s.Blk = b
			}
			return blk, nil
		} else {
			// Resolve the LHS.
			for i := range s.LHS {
				if x, err := resolveExpr(s.LHS[i], scope); err != nil {
					return nil, err
				} else {
					s.LHS[i] = x
				}
			}
			// Resolve the range expression.
			if x, err := resolveExpr(s.Range, scope); err != nil {
				return nil, err
			} else {
				s.Range = x
			}
			// Resolve the loop body.
			if blk, err := resolveBlock(s.Blk, scope); err != nil {
				return nil, err
			} else {
				s.Blk = blk
			}
			return s, nil
		}
	case *ast.DeferStmt:
		if x, err := resolveExpr(s.X, scope); err != nil {
			return nil, err
		} else {
			s.X = x
		}
		return s, nil
	case *ast.ExprSwitchStmt:
		// If the initial statement is a short variable declaration, put an
		// extra block around the switch.
		if d, ok := s.Init.(*ast.AssignStmt); ok && d.Op == scanner.DEFINE {
			blk := &ast.Block{
				Body:  make([]ast.Stmt, 2),
				Decls: make(map[string]ast.Symbol),
			}
			blk.Body[0] = s.Init
			blk.Body[1] = s
			s.Init = nil
			if b, err := resolveBlock(blk, scope); err != nil {
				return nil, err
			} else {
				return b, nil
			}
		} else {
			if x, err := resolveExpr(s.X, scope); err != nil {
				return nil, err
			} else {
				s.X = x
			}
			for i := range s.Cases {
				c := &s.Cases[i]
				for i := range c.Xs {
					if x, err := resolveExpr(c.Xs[i], scope); err != nil {
						return nil, err
					} else {
						c.Xs[i] = x
					}
				}
				if b, err := resolveBlock(c.Blk, scope); err != nil {
					return nil, err
				} else {
					c.Blk = b
				}
			}
			return s, nil
		}
	case *ast.TypeSwitchStmt:
		// If the initial statement is a short variable declaration, put an
		// extra block around the type switch.
		if d, ok := s.Init.(*ast.AssignStmt); ok && d.Op == scanner.DEFINE {
			blk := &ast.Block{
				Body:  make([]ast.Stmt, 2),
				Decls: make(map[string]ast.Symbol),
			}
			blk.Body[0] = s.Init
			blk.Body[1] = s
			s.Init = nil
			if b, err := resolveBlock(blk, scope); err != nil {
				return nil, err
			} else {
				return b, nil
			}
		} else {
			if x, err := resolveExpr(s.X, scope); err != nil {
				return nil, err
			} else {
				s.X = x
			}
			for i := range s.Cases {
				c := &s.Cases[i]
				for i := range c.Types {
					if t, err := resolveType(c.Types[i], scope); err != nil {
						return nil, err
					} else {
						c.Types[i] = t
					}
				}
				// Declare the variable of the TypeSwitchGuard in each
				// clause. If the clause contains only one type, the variable
				// gets this type, otherwise it gets the type of the
				// PrimaryExpr.
				if len(s.Id) > 0 {
					var t ast.Type
					if len(c.Types) == 1 {
						t = c.Types[0]
					}
					v := &ast.Var{Off: s.Off, File: scope.File(), Name: s.Id, Type: t}
					if err := c.Blk.Declare(v.Name, v); err != nil {
						return nil, err
					}
				}
				if b, err := resolveBlock(c.Blk, scope); err != nil {
					return nil, err
				} else {
					c.Blk = b
				}
			}
			return s, nil
		}
	case *ast.SelectStmt:
		for i := range s.Comms {
			c := &s.Comms[i]
			// Resolve the CommCase statement in the scope of the clause
			// block, so eventual variable names are declared in that scope.
			c.Blk.Up = scope
			if st, err := resolveStmt(c.Comm, c.Blk); err != nil {
				return nil, err
			} else {
				c.Comm = st
			}
			if b, err := resolveBlock(c.Blk, scope); err != nil {
				return nil, err
			} else {
				c.Blk = b
			}
		}
		return s, nil
	default:
		panic("not reached")
	}
}
