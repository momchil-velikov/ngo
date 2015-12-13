package parser

import (
	"errors"
	"fmt"
	"golang/ast"
	"golang/constexpr"
	"golang/scanner"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	// "runtime"
)

type parser struct {
	errors   []error
	scan     scanner.Scanner
	comments []ast.Comment

	token    uint // current token
	level    int
	brackets int
}

func (p *parser) init(name string, src string) {
	p.scan.Init(name, []byte(src))
	p.level = 1
	p.brackets = 1
	p.next()
}

func (p *parser) beginBrackets() {
	p.brackets++
}

func (p *parser) endBrackets() {
	p.brackets--
}

func (p *parser) setBrackets(n int) int {
	o := p.brackets
	p.brackets = n
	return o
}

// Parse all the source files of a package
func ParsePackage(dir string, names []string) (*ast.UnresolvedPackage, error) {
	// We need at least one source file
	if len(names) == 0 {
		return nil, errors.New("no source files in " + dir)
	}
	// Parse sources
	var files []*ast.UnresolvedFile
	for _, name := range names {
		path := filepath.Join(dir, name)
		src, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}
		f, err := Parse(path, string(src))
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}

	// Check that all the source files declare the same package name.
	pkgname := files[0].PkgName
	for _, f := range files {
		if f.PkgName != pkgname {
			ln, col := f.SrcMap.Position(f.Off)
			return nil, parseError{f.Name, ln, col, "inconsistent package name"}
		}
	}
	pkg := &ast.UnresolvedPackage{
		Path:  dir,
		Name:  pkgname,
		Files: files,
	}
	for _, f := range pkg.Files {
		f.Pkg = pkg
	}
	return pkg, nil
}

// Parse a source file
func Parse(name string, src string) (*ast.UnresolvedFile, error) {
	p := parser{}
	p.init(name, src)

	f := p.parseFile()
	if p.errors == nil {
		return f, nil
	} else {
		return f, ErrorList(p.errors)
	}
}

func printIndent(w io.Writer, n int, s string) {
	for i := 0; i < n; i++ {
		fmt.Fprint(w, s)
	}
}
func (p *parser) traceIn(args ...interface{}) {
	printIndent(os.Stderr, p.level, ".")
	fmt.Fprintln(os.Stderr, "->", args)
	p.level++
}

func (p *parser) traceOut(args ...interface{}) {
	p.level--
	printIndent(os.Stderr, p.level, ".")
	fmt.Fprintln(os.Stderr, "<-", args)
}

func (p *parser) trace(args ...interface{}) func() {
	p.traceIn(args...)
	return func() { p.traceOut(args...) }
}

// Append an error message to the parser error messages list
func (p *parser) error(msg string) {
	ln, col := p.scan.Position(p.scan.TOff)
	e := parseError{p.scan.Name, ln, col, msg}
	p.errors = append(p.errors, e)
}

// Emit an expected token mismatch error.
func (p *parser) expectError(exp, act uint) {
	p.error(fmt.Sprintf("expected %s, got %s", scanner.TokenNames[exp], scanner.TokenNames[act]))
}

// Get the next token from the scanner.
func (p *parser) next() int {
	off := p.scan.TOff
	p.token = p.scan.Get()
	for p.token == scanner.LINE_COMMENT || p.token == scanner.BLOCK_COMMENT {
		p.comments = append(p.comments, ast.Comment{Off: p.scan.TOff, Text: p.scan.Value})
		p.token = p.scan.Get()
	}
	return off
}

// Check the next token is TOKEN. Return true if so, otherwise emit an error and
// return false.
func (p *parser) expect(token uint) bool {
	if p.token == token {
		return true
	} else {
		p.expectError(token, p.token)
		return false
	}
}

// Advance to the next token iff the current one is TOKEN. Returns the token
// offset before the advance.
func (p *parser) match(token uint) int {
	if p.expect(token) {
		return p.next()
	} else {
		return p.scan.TOff
	}
}

// Advances to the next token iff the current one is TOKEN. Returns token
// value.
func (p *parser) matchRaw(token uint) ([]byte, int) {
	if p.expect(token) {
		v := p.scan.Value
		return v, p.next()
	} else {
		return nil, p.scan.TOff
	}
}

// Advances to the next token iff the current one is TOKEN. Returns token
// value as string.
func (p *parser) matchString(token uint) (string, int) {
	b, o := p.matchRaw(token)
	return string(b), o
}

// Skip tokens, until given token found, then consume it. Report an error
// only if some tokens were skipped.
func (p *parser) sync(token uint) {
	if p.token != token {
		p.expectError(token, p.token)
	}
	for p.token != scanner.EOF && p.token != token {
		p.next()
	}
	p.next()
}

// Skip tokens, until either T1 or T2 token is found. Report an error
// only if some tokens were skipped. Consume T1.
func (p *parser) sync2(t1, t2 uint) {
	if p.token != t1 {
		p.expectError(t1, p.token)
	}
	for p.token != scanner.EOF && p.token != t1 && p.token != t2 {
		p.next()
	}
	if p.token == t1 {
		p.next()
	}
}

// Skip tokens, until beginning of a toplevel declaration found
func (p *parser) syncDecl() {
	for {
		switch p.token {
		case scanner.CONST, scanner.TYPE, scanner.VAR, scanner.FUNC, scanner.EOF:
			return
		default:
			p.next()
		}
	}
}

// SourceFile = PackageClause ";" { ImportDecl ";" } { TopLevelDecl ";" } .
func (p *parser) parseFile() *ast.UnresolvedFile {
	var off int

	// Parse package name
	if p.token == scanner.PACKAGE {
		off = p.scan.TOff
	}
	name := p.parsePackageClause()
	if len(name) == 0 {
		return nil
	}

	if !p.expect(';') {
		return nil
	}
	p.next()

	// Parse import declaration(s)
	is := p.parseImportDecls()

	// Parse toplevel declarations.
	ds := p.parseToplevelDecls()

	p.match(scanner.EOF)

	return &ast.UnresolvedFile{
		Off:      off,
		PkgName:  name,
		Imports:  is,
		Decls:    ds,
		Comments: p.comments,
		Name:     p.scan.Name,
		SrcMap:   p.scan.SrcMap,
	}
}

// Parse a package clause. Return the package name or an empty string on error.
//
// PackageClause  = "package" PackageName .
// PackageName    = identifier .
func (p *parser) parsePackageClause() string {
	p.match(scanner.PACKAGE)
	pkg, _ := p.matchString(scanner.ID)
	if pkg == "_" {
		p.error("`_` is invalid package name")
	}
	return pkg
}

// Parse import declarations(s).
//
// ImportDecl       = "import" ( ImportSpec | "(" { ImportSpec ";" } ")" ) .
func (p *parser) parseImportDecls() (is []*ast.ImportDecl) {
	for p.token == scanner.IMPORT {
		p.match(scanner.IMPORT)
		if p.token == '(' {
			p.next()
			for p.token != scanner.EOF && p.token != ')' {
				is = append(is, p.parseImportSpec())
				if p.token != ')' {
					p.sync2(';', ')')
				}
			}
			p.match(')')
		} else {
			is = append(is, p.parseImportSpec())
		}
		p.sync(';')
	}
	return
}

// Parse import spec
//
// ImportSpec       = [ "." | PackageName ] ImportPath .
// ImportPath       = string_lit .
func (p *parser) parseImportSpec() *ast.ImportDecl {
	var name string
	off := p.scan.TOff
	if p.token == '.' {
		name = "."
		p.next()
	} else if p.token == scanner.ID {
		name, _ = p.matchString(scanner.ID)
	} else {
		name = ""
	}
	path, _ := p.matchRaw(scanner.STRING)
	return &ast.ImportDecl{Off: off, Name: name, Path: path}
}

// Parse toplevel declaration(s)
//
// Declaration   = ConstDecl | TypeDecl | VarDecl .
// TopLevelDecl  = Declaration | FunctionDecl | MethodDecl .
func (p *parser) parseToplevelDecls() (dcls []ast.Decl) {
	for {
		var d ast.Decl
		switch p.token {
		case scanner.TYPE:
			p.match(scanner.TYPE)
			if p.token == '(' {
				d = p.parseTypeDeclGroup()
			} else {
				d = p.parseTypeSpec()
			}
		case scanner.CONST:
			p.match(scanner.CONST)
			if p.token == '(' {
				d = p.parseConstDeclGroup()
			} else {
				d = p.parseConstSpec(false)
			}
		case scanner.VAR:
			p.match(scanner.VAR)
			if p.token == '(' {
				d = p.parseVarDeclGroup()
			} else {
				d = p.parseVarSpec()
			}
		case scanner.FUNC:
			d = p.parseFuncDecl()
		case scanner.EOF:
			return
		default:
			p.error("expected type, const, var or func/method declaration")
			p.syncDecl()
		}
		if d != nil {
			dcls = append(dcls, d)
			p.match(';')
		}
	}
}

// Parse type declaration
//
// TypeDecl = "type" ( TypeSpec | "(" { TypeSpec ";" } ")" ) .
func (p *parser) parseTypeDeclGroup() *ast.TypeDeclGroup {
	p.match('(')
	grp := &ast.TypeDeclGroup{}
	for p.token != scanner.EOF && p.token != ')' {
		grp.Types = append(grp.Types, p.parseTypeSpec())
		if p.token != ')' {
			p.sync2(';', ')')
		}
	}
	p.match(')')
	return grp
}

// TypeSpec = identifier Type .
func (p *parser) parseTypeSpec() *ast.TypeDecl {
	id, off := p.matchString(scanner.ID)
	t := p.parseType()
	return &ast.TypeDecl{Off: off, Name: id, Type: t}
}

// Determine if the given TOKEN could be a beginning of a typespec.
func isTypeLookahead(token uint) bool {
	switch token {
	case scanner.ID, '[', scanner.STRUCT, '*', scanner.FUNC, scanner.INTERFACE,
		scanner.MAP, scanner.CHAN, scanner.RECV, '(':
		return true
	default:
		return false
	}
}

// Type     = TypeName | TypeLit | "(" Type ")" .
// TypeName = identifier | QualifiedIdent .
// TypeLit  = ArrayType | StructType | PointerType | FunctionType | InterfaceType |
//            SliceType | MapType | ChannelType .
func (p *parser) parseType() ast.Type {
	switch p.token {
	case scanner.ID:
		return p.parseQualifiedId()

	// ArrayType   = "[" ArrayLength "]" ElementType .
	// ArrayLength = Expression .
	// ElementType = Type .
	// SliceType = "[" "]" ElementType .
	// Allow here an array type of unspecified size, that can be used
	// only in composite literal expressions.
	// ArrayLength = Expression | "..." .
	case '[':
		off := p.next()
		if p.token == ']' {
			p.next()
			return &ast.SliceType{Off: off, Elt: p.parseType()}
		} else if p.token == scanner.DOTS {
			p.next()
			p.match(']')
			return &ast.ArrayType{Off: off, Elt: p.parseType()}
		} else {
			e := p.parseExpr()
			p.match(']')
			t := p.parseType()
			return &ast.ArrayType{Off: off, Dim: e, Elt: t}
		}

	// PointerType = "*" BaseType .
	// BaseType = Type .
	case '*':
		off := p.next()
		return &ast.PtrType{Off: off, Base: p.parseType()}

	// MapType     = "map" "[" KeyType "]" ElementType .
	// KeyType     = Type .
	case scanner.MAP:
		off := p.next()
		p.match('[')
		k := p.parseType()
		p.match(']')
		t := p.parseType()
		return &ast.MapType{Off: off, Key: k, Elt: t}

	// ChannelType = ( "chan" [ "<-" ] | "<-" "chan" ) ElementType .
	case scanner.RECV:
		off := p.next()
		p.match(scanner.CHAN)
		t := p.parseType()
		return &ast.ChanType{Off: off, Send: false, Recv: true, Elt: t}
	case scanner.CHAN:
		off := p.next()
		send, recv := true, true
		if p.token == scanner.RECV {
			p.next()
			send, recv = true, false
		}
		t := p.parseType()
		return &ast.ChanType{Off: off, Send: send, Recv: recv, Elt: t}

	case scanner.STRUCT:
		return p.parseStructType()

	case scanner.FUNC:
		return p.parseFuncType()

	case scanner.INTERFACE:
		return p.parseInterfaceType()

	case '(':
		p.next()
		t := p.parseType()
		p.match(')')
		return t

	default:
		p.error("expected typespec")
		return &ast.Error{p.scan.TOff}
	}
}

// TypeName = identifier | QualifiedIdent .
func (p *parser) parseQualifiedId() *ast.QualifiedId {
	pkg, off := p.matchString(scanner.ID)
	var id string
	if p.token == '.' {
		p.next()
		id, _ = p.matchString(scanner.ID)
	} else {
		pkg, id = "", pkg
	}
	return &ast.QualifiedId{Off: off, Pkg: pkg, Id: id}
}

// StructType     = "struct" "{" { FieldDecl ";" } "}" .
func (p *parser) parseStructType() *ast.StructType {
	fs := []ast.Field{}
	off := p.match(scanner.STRUCT)
	p.match('{')
	for p.token != scanner.EOF && p.token != '}' {
		fs = p.parseFieldDecls(fs)
		if p.token != '}' {
			p.sync2(';', '}')
		}
	}
	p.match('}')
	return &ast.StructType{Off: off, Fields: fs}
}

// FieldDecl      = (IdentifierList Type | AnonymousField) [ Tag ] .
// AnonymousField = [ "*" ] TypeName .
// Tag            = string_lit .
func (p *parser) parseFieldDecls(fs []ast.Field) []ast.Field {
	if p.token == '*' {
		// Anonymous field.
		off := p.next()
		pt := &ast.PtrType{Off: off, Base: p.parseQualifiedId()}
		return append(fs, ast.Field{Type: pt, Tag: p.parseTagOpt()})
	}

	if p.token == scanner.ID {
		name, off := p.matchString(scanner.ID)
		id := &ast.QualifiedId{Off: off, Id: name}

		// If the field decl begins with a qualified-id, it's parsed as an
		// anonymous field.
		if p.token == '.' {
			p.next()
			id.Pkg = id.Id
			id.Id, _ = p.matchString(scanner.ID)
			return append(fs, ast.Field{Type: id, Tag: p.parseTagOpt()})
		}

		// If it's only a single identifier, with no separate type
		// declaration, it's also an anonymous filed.
		if p.token == scanner.STRING || p.token == ';' || p.token == '}' {
			return append(fs, ast.Field{Type: id, Tag: p.parseTagOpt()})
		}

		fs = p.parseFieldIdList(off, name, fs)
		n := len(fs)
		fs[n-1].Type = p.parseType()
		fs[n-1].Tag = p.parseTagOpt()
		return fs
	}
	p.error("Invalid field declaration")
	return append(fs, ast.Field{Type: &ast.Error{Off: p.scan.TOff}})
}

func (p *parser) parseTagOpt() string {
	if p.token == scanner.STRING {
		tag, _ := p.matchRaw(scanner.STRING)
		return constexpr.String(tag)
	} else {
		return ""
	}
}

// IdentifierList = identifier { "," identifier } .
func (p *parser) parseIdList(id *ast.Ident) (ids []*ast.Ident) {
	if id == nil {
		name, off := p.matchString(scanner.ID)
		id = &ast.Ident{Off: off, Id: name}
	}
	ids = append(ids, id)
	for p.token == ',' {
		p.next()
		name, off := p.matchString(scanner.ID)
		if len(name) > 0 {
			ids = append(ids, &ast.Ident{Off: off, Id: name})
		}
	}
	return ids
}

// Parse an IdentifierList in the context of a ConstSpec.
func (p *parser) parseConstIdList() (id []*ast.Const) {
	name, off := p.matchString(scanner.ID)
	// Ensure the list containes at least one element even in the case of a
	// missing identifier.
	id = append(id, &ast.Const{Off: off, Name: name})
	for p.token == ',' {
		p.next()
		name, off = p.matchString(scanner.ID)
		if len(name) > 0 {
			id = append(id, &ast.Const{Off: off, Name: name})
		}
	}
	return id
}

// Parse an IdentifierList in the context of a VarSpec.
func (p *parser) parseVarIdList() (id []*ast.Var) {
	name, off := p.matchString(scanner.ID)
	// Ensure the list containes at least one element even in the case of a
	// missing identifier.
	id = append(id, &ast.Var{Off: off, Name: name})
	for p.token == ',' {
		p.next()
		name, off = p.matchString(scanner.ID)
		if len(name) > 0 {
			id = append(id, &ast.Var{Off: off, Name: name})
		}
	}
	return id
}

// Parse an IdentifierList in the context of a FieldDecl.
func (p *parser) parseFieldIdList(off int, name string, fs []ast.Field) []ast.Field {
	fs = append(fs, ast.Field{Off: off, Name: name})
	for p.token == ',' {
		p.next()
		name, off = p.matchString(scanner.ID)
		if len(name) > 0 {
			fs = append(fs, ast.Field{Off: off, Name: name})
		}
	}
	return fs
}

// FunctionType = "func" Signature .
func (p *parser) parseFuncType() *ast.FuncType {
	off := p.match(scanner.FUNC)
	t := p.parseSignature()
	t.Off = off
	return t
}

// Signature = Parameters [ Result ] .
// Result    = Parameters | Type .
func (p *parser) parseSignature() *ast.FuncType {
	ps, v := p.parseParameters()
	var rs []ast.Param
	if p.token == '(' {
		vr := false
		rs, vr = p.parseParameters()
		if vr {
			p.error("output parameters cannot be variadic")
		}
	} else if isTypeLookahead(p.token) {
		off := p.scan.TOff
		rs = []ast.Param{ast.Param{Off: off, Type: p.parseType()}}
	}
	return &ast.FuncType{Params: ps, Returns: rs, Var: v}
}

// Parameters    = "(" [ ParameterList [ "," ] ] ")" .
// ParameterList = ParameterDecl { "," ParameterDecl } .
// ParameterDecl = [ IdentifierList ] [ "..." ] Type .
func (p *parser) parseParameters() ([]ast.Param, bool) {
	var ns, ps []ast.Param
	variadic := false
	p.match('(')
	if p.token == ')' {
		p.next()
		return nil, false
	}
	for p.token != scanner.EOF && p.token != ')' {
		parm, v := p.parseIdOrType()
		if parm.Type != nil {
			convertToParamTypes(ns)
			ps = append(ps, ns...)
			ns = ns[:0]
			ps = append(ps, parm)
			if variadic {
				p.error("only the last parameter can be variadic")
			}
			variadic = variadic || v
		} else {
			ns = append(ns, parm)
			if p.token != ',' && p.token != ')' {
				parm, v = p.parseIdOrType()
				if parm.Type == nil {
					parm.Type = &ast.QualifiedId{Off: parm.Off, Id: parm.Name}
				}
				ps = append(ps, ns...)
				// The type is attached only to the last parameter in a
				// sequence like `x, y, z T` until after `T` is resolved. This
				// way we avoid duplicate resolution due to sharing of the
				// type.
				ps[len(ps)-1].Type = parm.Type
				if variadic || (len(ns) > 1 && v) {
					p.error("only the last parameter can be variadic")
				}
				ns = ns[:0]
				variadic = variadic || v
			}
		}
		if p.token != ')' {
			p.sync2(',', ')')
		}
	}
	if len(ns) > 0 {
		convertToParamTypes(ns)
		ps = append(ps, ns...)
		if variadic {
			p.error("only the last parameter can be variadic")
		}
	}
	p.match(')')
	return ps, variadic
}

// Parses an identifier or a type. A qualified identifier is considered a
// type.
func (p *parser) parseIdOrType() (ast.Param, bool) {
	if p.token == scanner.ID {
		id, off := p.matchString(scanner.ID)
		if p.token == '.' {
			p.next()
			pkg := id
			if p.token == scanner.ID {
				id, _ = p.matchString(scanner.ID)
			} else {
				p.error("incomplete qualified id")
				id = ""
			}
			return ast.Param{Off: off, Type: &ast.QualifiedId{Pkg: pkg, Id: id}}, false
		}
		return ast.Param{Off: off, Name: id}, false
	}
	v := false
	if p.token == scanner.DOTS {
		p.next()
		v = true
	}
	t := p.parseType() // FIXME: parenthesized types not allowed
	return ast.Param{Type: t}, v
}

// Converts an identifier list to a list of parameter declarations, taking
// each identifier to be a type name.
func convertToParamTypes(ps []ast.Param) {
	for i := range ps {
		p := &ps[i]
		p.Type = &ast.QualifiedId{Off: p.Off, Id: p.Name}
		p.Off = 0
		p.Name = ""
	}
}

// InterfaceType      = "interface" "{" { MethodSpec ";" } "}" .
// MethodSpec         = MethodName Signature | InterfaceTypeName .
// MethodName         = identifier .
// InterfaceTypeName  = TypeName .
func (p *parser) parseInterfaceType() *ast.InterfaceType {
	var ms []ast.MethodSpec
	off := p.match(scanner.INTERFACE)
	p.match('{')
	for p.token != scanner.EOF && p.token != '}' {
		var m ast.MethodSpec
		id, off := p.matchString(scanner.ID)
		if p.token == '(' {
			sig := p.parseSignature()
			m = ast.MethodSpec{Off: off, Name: id, Type: sig}
		} else {
			pkg := ""
			if p.token == '.' {
				p.next()
				pkg = id
				id, _ = p.matchString(scanner.ID)
			}
			m = ast.MethodSpec{Type: &ast.QualifiedId{Off: off, Pkg: pkg, Id: id}}
		}
		ms = append(ms, m)
		if p.token != '}' {
			p.sync2(';', '}')
		}
	}
	p.match('}')
	return &ast.InterfaceType{Off: off, Methods: ms}
}

// ConstDecl = "const" ( ConstSpec | "(" { ConstSpec ";" } ")" ) .
func (p *parser) parseConstDeclGroup() *ast.ConstDeclGroup {
	p.match('(')
	grp := &ast.ConstDeclGroup{}
	for p.token != scanner.EOF && p.token != ')' {
		grp.Consts = append(grp.Consts, p.parseConstSpec(true))
		if p.token != ')' {
			p.sync2(';', ')')
		}
	}
	p.match(')')
	return grp
}

// ConstSpec = IdentifierList [ [ Type ] "=" ExpressionList ] .
func (p *parser) parseConstSpec(grp bool) *ast.ConstDecl {
	var (
		t  ast.Type
		xs []ast.Expr
	)
	ids := p.parseConstIdList()
	if isTypeLookahead(p.token) {
		t = p.parseType()
	}
	if t == nil && grp {
		// Allow initialization expressions to be missing if the ConstSpec is
		// a part of a group declaration and there is no explicit type given.
		if p.token == '=' {
			p.next()
			xs = p.parseExprList(nil)
		}
	} else {
		p.match('=')
		xs = p.parseExprList(nil)
	}
	return &ast.ConstDecl{Names: ids, Type: t, Values: xs}
}

// VarDecl = "var" ( VarSpec | "(" { VarSpec ";" } ")" ) .
func (p *parser) parseVarDeclGroup() *ast.VarDeclGroup {
	p.match('(')
	grp := &ast.VarDeclGroup{}
	for p.token != scanner.EOF && p.token != ')' {
		grp.Vars = append(grp.Vars, p.parseVarSpec())
		if p.token != ')' {
			p.sync2(';', ')')
		}
	}
	p.match(')')
	return grp
}

// VarSpec = IdentifierList ( Type [ "=" ExpressionList ] | "=" ExpressionList ) .
func (p *parser) parseVarSpec() *ast.VarDecl {
	var (
		t  ast.Type
		es []ast.Expr
	)
	ids := p.parseVarIdList()
	if isTypeLookahead(p.token) {
		t = p.parseType()
	}
	if p.token == '=' {
		p.next()
		es = p.parseExprList(nil)
	}
	return &ast.VarDecl{Names: ids, Type: t, Init: es}
}

// MethodDecl   = "func" Receiver MethodName ( Function | Signature ) .
// FunctionDecl = "func" FunctionName ( Function | Signature ) .
// FunctionName = identifier .
// Function     = Signature FunctionBody .
// FunctionBody = Block .
func (p *parser) parseFuncDecl() ast.Decl {
	foff := p.match(scanner.FUNC)
	var r *ast.Param
	if p.token == '(' {
		r = p.parseReceiver()
	}
	name, off := p.matchString(scanner.ID)
	sig := p.parseSignature()
	var blk *ast.Block
	if p.token == '{' {
		blk = p.parseBlock()
	}
	fn := &ast.FuncDecl{
		Off:  off,
		Name: name,
		Func: ast.Func{
			Off:  foff,
			Recv: r,
			Sig:  sig,
			Blk:  blk,
		},
	}
	fn.Func.Decl = fn
	return fn
}

// Receiver = Parameters .
// Instead of the above production from the Go Language Specification
// we parse the receiver as
// Receiver  = "(" [identifier] Type [","] ")"
func (p *parser) parseReceiver() *ast.Param {
	off := p.match('(')
	var name *ast.QualifiedId
	if p.token == scanner.ID {
		name = p.parseQualifiedId()
	}
	switch p.token {
	case ',':
		p.next()
		fallthrough
	case ')':
		p.match(')')
		if name == nil {
			p.error("missing receiver type")
			return &ast.Param{Type: &ast.Error{Off: off}}
		}
		return &ast.Param{Type: name}
	}
	var id string
	if name != nil {
		if len(name.Pkg) > 0 {
			p.error("receiver name cannot be qualified")
		}
		off = name.Off
		id = name.Id
	}
	t := p.parseType()
	if p.token == ',' {
		p.next()
	}
	p.sync2(')', ';')
	return &ast.Param{Off: off, Name: id, Type: t}
}

// ExpressionList = Expression { "," Expression } .
func (p *parser) parseExprList(e ast.Expr) (es []ast.Expr) {
	//	defer p.trace("ExprList")()

	if e != nil {
		es = append(es, e)
	}
	for {
		es = append(es, p.parseExpr())
		if p.token != ',' {
			break
		}
		p.match(',')
	}
	return
}

// In certain contexts it may not be possible to disambiguate between Expression
// or Type based on a single (or even O(1)) tokens(s) of lookahead. Therefore,
// in such contexts, allow either one to happen and rely on a later typecheck
// pass to validate the AST.
func (p *parser) parseExprOrType() (ast.Expr, ast.Type) {
	//	defer p.trace("ExprOrType")()

	return p.parseOrExprOrType()
}

func (p *parser) parseExpr() ast.Expr {
	// defer p.trace("Expression", s.TokenNames[p.token], p.scan.TLine, p.scan.TPos)()

	if e, _ := p.parseOrExprOrType(); e != nil {
		return e
	}
	p.error("Type not allowed in this context")
	return &ast.Error{p.scan.TOff}
}

// Expression = UnaryExpr | Expression binary_op UnaryExpr .
// `Expression` from the Go Specification is replaced by the following
// ``xxxExpr` productions.

// LogicalOrExpr = LogicalAndExpr { "||" LogicalAndExpr }
func (p *parser) parseOrExprOrType() (ast.Expr, ast.Type) {
	//	defer p.trace("OrExprOrType")()

	x, t := p.parseAndExprOrType()
	if x == nil {
		return nil, t
	}
	for p.token == scanner.OR {
		p.next()
		y, _ := p.parseAndExprOrType()
		x = &ast.BinaryExpr{Op: scanner.OR, X: x, Y: y}
	}
	return x, nil
}

// LogicalAndExpr = CompareExpr { "&&" CompareExpr }
func (p *parser) parseAndExprOrType() (ast.Expr, ast.Type) {
	//	defer p.trace("AndExprOrType")()

	x, t := p.parseCompareExprOrType()
	if x == nil {
		return nil, t
	}
	for p.token == scanner.AND {
		p.next()
		y, _ := p.parseCompareExprOrType()
		x = &ast.BinaryExpr{Op: scanner.AND, X: x, Y: y}
	}
	return x, nil
}

// CompareExpr = AddExpr { rel_op AddExpr }
func (p *parser) parseCompareExprOrType() (ast.Expr, ast.Type) {
	//	defer p.trace("CompareExprOrType")()

	x, t := p.parseAddExprOrType()
	if x == nil {
		return nil, t
	}
	for isRelOp(p.token) {
		op := p.token
		p.next()
		y, _ := p.parseAddExprOrType()
		x = &ast.BinaryExpr{Op: op, X: x, Y: y}
	}
	return x, nil
}

// AddExpr = MulExpr { add_op MulExpr}
func (p *parser) parseAddExprOrType() (ast.Expr, ast.Type) {
	//	defer p.trace("AddExprOrType")()

	x, t := p.parseMulExprOrType()
	if x == nil {
		return nil, t
	}
	for isAddOp(p.token) {
		op := p.token
		p.next()
		y, _ := p.parseMulExprOrType()
		x = &ast.BinaryExpr{Op: op, X: x, Y: y}
	}
	return x, nil
}

// MulExpr = UnaryExpr { mul_op UnaryExpr }
func (p *parser) parseMulExprOrType() (ast.Expr, ast.Type) {
	//	defer p.trace("MulExprOrType")()

	x, t := p.parseUnaryExprOrType()
	if x == nil {
		return nil, t
	}
	for isMulOp(p.token) {
		op := p.token
		p.next()
		y, _ := p.parseUnaryExprOrType()
		x = &ast.BinaryExpr{Op: op, X: x, Y: y}
	}
	return x, nil
}

// binary_op  = "||" | "&&" | rel_op | add_op | mul_op .

// rel_op     = "==" | "!=" | "<" | "<=" | ">" | ">=" .
func isRelOp(t uint) bool {
	switch t {
	case scanner.EQ, scanner.NE, scanner.LT, scanner.LE, scanner.GT, scanner.GE:
		return true
	default:
		return false
	}
}

// add_op     = "+" | "-" | "|" | "^" .
func isAddOp(t uint) bool {
	switch t {
	case '+', '-', '|', '^':
		return true
	default:
		return false
	}
}

// mul_op     = "*" | "/" | "%" | "<<" | ">>" | "&" | "&^" .
func isMulOp(t uint) bool {
	switch t {
	case '*', '/', '%', scanner.SHL, scanner.SHR, '&', scanner.ANDN:
		return true
	default:
		return false
	}
}

// assign_op = [ add_op | mul_op ] "=" .
func isAssignOp(t uint) bool {
	switch t {
	case '=', scanner.PLUS_ASSIGN, scanner.MINUS_ASSIGN, scanner.MUL_ASSIGN,
		scanner.DIV_ASSIGN, scanner.REM_ASSIGN, scanner.AND_ASSIGN, scanner.OR_ASSIGN,
		scanner.XOR_ASSIGN, scanner.SHL_ASSIGN, scanner.SHR_ASSIGN, scanner.ANDN_ASSIGN:
		return true
	default:
		return false
	}
}

// UnaryExpr  = PrimaryExpr | unary_op UnaryExpr .
// unary_op   = "+" | "-" | "!" | "^" | "*" | "&" | "<-" .
func (p *parser) parseUnaryExprOrType() (ast.Expr, ast.Type) {
	//	defer p.trace("UnaryExprOrType")()

	switch p.token {
	case '+', '-', '!', '^', '&':
		op := p.token
		off := p.next()
		x, _ := p.parseUnaryExprOrType()
		return &ast.UnaryExpr{Off: off, Op: op, X: x}, nil
	case '*':
		off := p.next()
		if x, t := p.parseUnaryExprOrType(); t == nil {
			return &ast.UnaryExpr{Off: off, Op: '*', X: x}, nil
		} else {
			return nil, &ast.PtrType{Off: off, Base: t}
		}
	case scanner.RECV:
		off := p.next()
		if x, t := p.parseUnaryExprOrType(); t == nil {
			return &ast.UnaryExpr{Off: off, Op: scanner.RECV, X: x}, nil
		} else if ch, ok := t.(*ast.ChanType); ok {
			ch.Recv, ch.Send = true, false
			ch.Off = off
			return nil, ch
		} else {
			p.error("invalid receive operation")
			return &ast.Error{p.scan.TOff}, nil
		}
	default:
		return p.parsePrimaryExprOrType()
	}
}

func (p *parser) needType(x ast.Expr) (ast.Type, bool) {
	switch t := x.(type) {
	case *ast.QualifiedId:
		return t, true
	case *ast.Selector:
		if x, ok := t.X.(*ast.QualifiedId); ok && len(x.Pkg) == 0 {
			return &ast.QualifiedId{Off: x.Off, Pkg: x.Id, Id: t.Id}, true
		}
	}
	return nil, false
}

// TypeAssertion = "." "(" Type ")" .
// Also allow "." "(" "type" ")".
func (p *parser) parseTypeAssertion(x ast.Expr) ast.Expr {
	p.match('(')
	var t ast.Type
	if p.token == scanner.TYPE {
		p.next()
	} else {
		t = p.parseType()
	}
	p.match(')')
	return &ast.TypeAssertion{Type: t, X: x}
}

// PrimaryExpr =
//     Operand |
//     Conversion |
//     PrimaryExpr Selector |
//     PrimaryExpr Index |
//     PrimaryExpr Slice |
//     PrimaryExpr TypeAssertion |
//     PrimaryExpr Arguments .

func (p *parser) parsePrimaryExprOrType() (ast.Expr, ast.Type) {
	// defer p.trace("PrimaryExprOrType", s.TokenNames[p.token])()

	var (
		x ast.Expr
		t ast.Type
	)
	// Handle initial Operand or Conversion
	switch p.token {
	case scanner.ID: // CompositeLit, MethodExpr, Conversion, OperandName
		name, off := p.matchString(scanner.ID)
		id := &ast.QualifiedId{Off: off, Id: name}
		if p.token == '{' && p.brackets > 0 {
			x = p.parseCompositeLiteral(id)
		} else if p.token == '(' {
			x = p.parseArguments(id)
		} else if p.token == '.' {
			p.next()
			if p.token == '(' {
				// PrimaryExpr TypeAssertion
				// TypeAssertion = "." "(" Type ")" .
				x = p.parseTypeAssertion(id)
			} else {
				// QualifiedId or Selector. Parsed as QualifiedId
				name, _ = p.matchString(scanner.ID)
				id.Pkg = id.Id
				id.Id = name
				x = id
			}
		} else {
			x = id
		}
	case '(': // Expression, MethodExpr, Conversion
		off := p.next()
		p.beginBrackets()
		x, t = p.parseExprOrType()
		p.endBrackets()
		p.match(')')
		if x == nil {
			if p.token == '(' {
				x = p.parseConversion(t)
			} else {
				return nil, t
			}
		} else {
			x = &ast.ParensExpr{Off: off, X: x}
		}
	case '[', scanner.STRUCT, scanner.MAP: // Conversion, CompositeLit
		t = p.parseType()
		if p.token == '(' {
			x = p.parseConversion(t)
		} else if p.token == '{' {
			x = p.parseCompositeLiteral(t)
		} else {
			return nil, t
		}
	case scanner.FUNC: // Conversion, FunctionLiteral
		t = p.parseType()
		if p.token == '(' {
			x = p.parseConversion(t)
		} else if p.token == '{' {
			x = p.parseFuncLiteral(t)
		} else {
			return nil, t
		}
	case '*', scanner.RECV:
		panic("should not reach here")
	case scanner.INTERFACE, scanner.CHAN: // Conversion
		t = p.parseType()
		if p.token == '(' {
			x = p.parseConversion(t)
		} else {
			return nil, t
		}
	case scanner.INTEGER, scanner.FLOAT, scanner.IMAGINARY, scanner.RUNE: // BasicLiteral
		k := p.token
		v, off := p.matchRaw(k)
		return &ast.Literal{Off: off, Kind: k, Value: v}, nil
	case scanner.STRING:
		v, off := p.matchRaw(scanner.STRING)
		x = &ast.Literal{Off: off, Kind: scanner.STRING, Value: v}
	default:
		p.error("token cannot start neither expression nor type")
		x = &ast.Error{p.scan.TOff}
	}
	// Parse the left-recursive alternatives for PrimaryExpr, folding the
	// left-hand parts in the variable `X`.
	for {
		switch p.token {
		case '.': // TypeAssertion or Selector
			p.next()
			if p.token == '(' {
				// PrimaryExpr TypeAssertion
				// TypeAssertion = "." "(" Type ")" .
				x = p.parseTypeAssertion(x)
			} else {
				// PrimaryExpr Selector
				// Selector = "." identifier .
				id, _ := p.matchString(scanner.ID)
				x = &ast.Selector{X: x, Id: id}
			}
		case '[':
			// PrimaryExpr Index
			// PrimaryExpr Slice
			x = p.parseIndexOrSlice(x)
		case '(':
			// PrimaryExpr Arguments
			x = p.parseArguments(x)
		default:
			if p.token == '{' && p.brackets > 0 {
				// Composite literal
				typ, ok := p.needType(x)
				if !ok {
					p.error("invalid type for composite literal")
				}
				x = p.parseCompositeLiteral(typ)
			} else {
				return x, nil
			}
		}
	}
}

// Operand    = Literal | OperandName | MethodExpr | "(" Expression ")" .
// Literal    = BasicLit | CompositeLit | FunctionLit .
// BasicLit   = int_lit | float_lit | imaginary_lit | rune_lit | string_lit .
// OperandName = identifier | QualifiedIdent.

// MethodExpr    = ReceiverType "." MethodName .
// ReceiverType  = TypeName | "(" "*" TypeName ")" | "(" ReceiverType ")" .
// Parsed as Selector expression, to be fixed up later by the typecheck phase.

// CompositeLit  = LiteralType LiteralValue .
// LiteralType   = StructType | ArrayType | "[" "..." "]" ElementType |
//                 SliceType | MapType | TypeName .
func (p *parser) parseCompositeLiteral(typ ast.Type) ast.Expr {
	elts := p.parseLiteralValue()
	return &ast.CompLiteral{Type: typ, Elts: elts}
}

// LiteralValue  = "{" [ ElementList [ "," ] ] "}" .
// ElementList   = Element { "," Element } .
func (p *parser) parseLiteralValue() (elts []*ast.Element) {
	p.match('{')
	p.beginBrackets()
	for p.token != scanner.EOF && p.token != '}' {
		elts = append(elts, p.parseElement())
		if p.token != '}' {
			p.sync2(',', '}')
		}
	}
	p.endBrackets()
	p.match('}')
	return elts
}

// Element       = [ Key ":" ] Value .
// Key           = FieldName | Expression | LiteralValue .
// FieldName     = identifier .
// Value         = Expression | LiteralValue .
func (p *parser) parseElement() *ast.Element {
	var k ast.Expr
	if p.token == '{' {
		k = &ast.CompLiteral{Elts: p.parseLiteralValue()}
	} else {
		k = p.parseExpr()
	}
	if p.token != ':' {
		return &ast.Element{Value: k}
	}
	p.match(':')
	if p.token == '{' {
		return &ast.Element{
			Key:   k,
			Value: &ast.CompLiteral{Elts: p.parseLiteralValue()},
		}
	} else {
		return &ast.Element{Key: k, Value: p.parseExpr()}
	}
}

// Conversion = Type "(" Expression [ "," ] ")" .
func (p *parser) parseConversion(t ast.Type) ast.Expr {
	p.match('(')
	p.beginBrackets()
	x := p.parseExpr()
	if p.token == ',' {
		p.next()
	}
	p.endBrackets()
	p.match(')')
	return &ast.Conversion{Type: t, X: x}
}

// Arguments =
//    "(" [ ( ExpressionList | Type [ "," ExpressionList ] ) [ "..." ] [ "," ] ] ")" .
func (p *parser) parseArguments(f ast.Expr) ast.Expr {
	p.match('(')
	p.beginBrackets()
	defer p.endBrackets()
	if p.token == ')' {
		p.next()
		return &ast.Call{Func: f}
	}
	var xs []ast.Expr
	x, t := p.parseExprOrType()
	if x != nil {
		xs = append(xs, x)
	}
	dots := false
	if p.token == scanner.DOTS {
		p.next()
		dots = true
	}
	if p.token != ')' {
		p.match(',')
	}
	for p.token != scanner.EOF && p.token != ')' && !dots {
		xs = append(xs, p.parseExpr())
		if p.token == scanner.DOTS {
			p.next()
			dots = true
		}
		if p.token != ')' {
			p.sync2(',', ')')
		}
	}
	p.match(')')
	return &ast.Call{Func: f, Type: t, Xs: xs, Ell: dots}
}

// FunctionLit = "func" Function .
func (p *parser) parseFuncLiteral(typ ast.Type) ast.Expr {
	sig := typ.(*ast.FuncType)
	b := p.parseBlock()
	return &ast.Func{Sig: sig, Blk: b}
}

// Index          = "[" Expression "]" .
// Slice          = "[" ( [ Expression ] ":" [ Expression ] ) |
//                      ( [ Expression ] ":" Expression ":" Expression )
//                  "]" .
func (p *parser) parseIndexOrSlice(x ast.Expr) ast.Expr {
	p.match('[')
	p.beginBrackets()
	defer p.endBrackets()
	if p.token == ':' {
		p.next()
		if p.token == ']' {
			p.next()
			return &ast.SliceExpr{X: x}
		}
		h := p.parseExpr()
		if p.token == ']' {
			p.next()
			return &ast.SliceExpr{X: x, Hi: h}
		}
		p.match(':')
		c := p.parseExpr()
		p.match(']')
		return &ast.SliceExpr{X: x, Hi: h, Cap: c}
	} else {
		i := p.parseExpr()
		if p.token == ']' {
			p.next()
			return &ast.IndexExpr{X: x, I: i}
		}
		p.match(':')
		if p.token == ']' {
			p.next()
			return &ast.SliceExpr{X: x, Lo: i}
		}
		h := p.parseExpr()
		if p.token == ']' {
			p.next()
			return &ast.SliceExpr{X: x, Lo: i, Hi: h}
		}
		p.match(':')
		c := p.parseExpr()
		p.match(']')
		return &ast.SliceExpr{X: x, Lo: i, Hi: h, Cap: c}
	}
}

// Block = "{" StatementList "}" .
func (p *parser) parseBlock() *ast.Block {
	//	defer p.trace("Block")()
	p.match('{')
	blk := p.parseStatementList()
	p.match('}')
	return blk
}

// StatementList = { Statement ";" } .
func (p *parser) parseStatementList() *ast.Block {
	var st []ast.Stmt
	b := p.setBrackets(1)
	for p.token != scanner.EOF && p.token != '}' &&
		p.token != scanner.CASE && p.token != scanner.DEFAULT {
		st = append(st, p.parseStmt())
		p.syncEndStatement()
	}
	p.setBrackets(b)
	return &ast.Block{Body: st}
}

func (p *parser) syncEndStatement() {
	if p.token == '}' {
		return
	}
	if p.expect(';') {
		p.next()
		return
	}
	for p.token != scanner.EOF && p.token != ';' && p.token != '}' &&
		p.token != scanner.CASE && p.token != scanner.DEFAULT {
		p.next()
	}
	if p.token == ';' {
		p.next()
	}
}

// Statement =
//     Declaration | LabeledStmt | SimpleStmt |
//     GoStmt | ReturnStmt | BreakStmt | ContinueStmt | GotoStmt |
//     FallthroughStmt | Block | IfStmt | SwitchStmt | SelectStmt | ForStmt |
//     DeferStmt .
// SimpleStmt =
//     EmptyStmt | ExpressionStmt | SendStmt | IncDecStmt | Assignment |
//     ShortVarDecl .
func (p *parser) parseStmt() ast.Stmt {
	//	defer p.trace("Statement")()

	switch p.token {
	case scanner.CONST:
		p.match(scanner.CONST)
		if p.token == '(' {
			return p.parseConstDeclGroup()
		} else {
			return p.parseConstSpec(false)
		}
	case scanner.TYPE:
		p.match(scanner.TYPE)
		if p.token == '(' {
			return p.parseTypeDeclGroup()
		} else {
			return p.parseTypeSpec()
		}
	case scanner.VAR:
		p.match(scanner.VAR)
		if p.token == '(' {
			return p.parseVarDeclGroup()
		} else {
			return p.parseVarSpec()
		}
	case scanner.GO:
		return p.parseGoStmt()
	case scanner.RETURN:
		return p.parseReturnStmt()
	case scanner.BREAK:
		return p.parseBreakStmt()
	case scanner.CONTINUE:
		return p.parseContinueStmt()
	case scanner.GOTO:
		return p.parseGotoStmt()
	case scanner.FALLTHROUGH:
		return p.parseFallthroughStmt()
	case '{':
		return p.parseBlock()
	case scanner.IF:
		return p.parseIfStmt()
	case scanner.SWITCH:
		return p.parseSwitchStmt()
	case scanner.SELECT:
		return p.parseSelectStmt()
	case scanner.FOR:
		return p.parseForStmt()
	case scanner.DEFER:
		return p.parseDeferStmt()
	case ';', '}':
		return &ast.EmptyStmt{p.scan.TOff}
	}

	e := p.parseExpr()
	if p.token == ':' {
		return p.parseLabeledStmt(e)
	} else {
		return p.parseSimpleStmt(e)
	}
}

// LabeledStmt = Label ":" Statement .
// Label       = identifier .
func (p *parser) parseLabeledStmt(x ast.Expr) ast.Stmt {
	off := p.match(':')
	id := ""
	if q, ok := x.(*ast.QualifiedId); ok {
		if len(q.Pkg) > 0 {
			p.error("a label must consist of single identifier")
		}
		off = q.Off
		id = q.Id
	}
	return &ast.Label{Off: off, Label: id, Stmt: p.parseStmt()}
}

// GoStmt = "go" Expression .
func (p *parser) parseGoStmt() ast.Stmt {
	return &ast.GoStmt{Off: p.match(scanner.GO), X: p.parseExpr()}
}

// ReturnStmt = "return" [ ExpressionList ] .
func (p *parser) parseReturnStmt() ast.Stmt {
	off := p.match(scanner.RETURN)
	if p.token != ';' && p.token != '}' {
		return &ast.ReturnStmt{Off: off, Xs: p.parseExprList(nil)}
	} else {
		return &ast.ReturnStmt{Off: off}
	}
}

// BreakStmt = "break" [ Label ] .
func (p *parser) parseBreakStmt() ast.Stmt {
	off := p.match(scanner.BREAK)
	if p.token == scanner.ID {
		label, _ := p.matchString(scanner.ID)
		return &ast.BreakStmt{Off: off, Label: label}
	} else {
		return &ast.BreakStmt{Off: off}
	}
}

// ContinueStmt = "continue" [ Label ] .
func (p *parser) parseContinueStmt() ast.Stmt {
	off := p.match(scanner.CONTINUE)
	if p.token == scanner.ID {
		label, _ := p.matchString(scanner.ID)
		return &ast.ContinueStmt{Off: off, Label: label}
	} else {
		return &ast.ContinueStmt{Off: off}
	}
}

// GotoStmt = "goto" Label .
func (p *parser) parseGotoStmt() ast.Stmt {
	off := p.match(scanner.GOTO)
	label, _ := p.matchString(scanner.ID)
	return &ast.GotoStmt{Off: off, Label: label}
}

// FallthroughStmt = "fallthrough" .
func (p *parser) parseFallthroughStmt() ast.Stmt {
	return &ast.FallthroughStmt{Off: p.match(scanner.FALLTHROUGH)}
}

// IfStmt = "if" [ SimpleStmt ";" ] Expression Block [ "else" ( IfStmt | Block ) ] .
func (p *parser) parseIfStmt() ast.Stmt {
	off := p.match(scanner.IF)
	b := p.setBrackets(0)
	var init ast.Stmt
	x := p.parseExpr()
	if p.token != '{' {
		init = p.parseSimpleStmt(x)
		p.match(';')
		x = p.parseExpr()
	}
	p.setBrackets(b)
	then := p.parseBlock()
	if p.token == scanner.ELSE {
		var els ast.Stmt
		p.next()
		if p.token == scanner.IF {
			els = p.parseIfStmt()
		} else {
			els = p.parseBlock()
		}
		return &ast.IfStmt{Off: off, Init: init, Cond: x, Then: then, Else: els}
	} else {
		return &ast.IfStmt{Off: off, Init: init, Cond: x, Then: then}
	}
}

// Checks if an Expression is a TypeSwitchGuard.
func isTypeSwitchGuardExpr(x ast.Expr) (ast.Expr, bool) {
	if y, ok := x.(*ast.TypeAssertion); ok && y.Type == nil {
		return y.X, true
	} else {
		return nil, false
	}
}

// Checks if a SimpleStmt is a TypeSwitchGuard. If true, returns the statement
// constituent parts.
func (p *parser) isTypeSwitchGuardStmt(x ast.Stmt) (string, ast.Expr, bool) {
	if a, ok := x.(*ast.AssignStmt); ok && len(a.LHS) == 1 && len(a.RHS) == 1 {
		lhs, rhs := a.LHS[0], a.RHS[0]
		if y, ok := isTypeSwitchGuardExpr(rhs); ok {
			var id string
			if z, ok := lhs.(*ast.QualifiedId); !ok || len(z.Pkg) > 0 || z.Id == "_" {
				id = ""
				p.error("invalid identifier in type switch")
			} else {
				id = z.Id
			}
			if a.Op != scanner.DEFINE {
				p.error("type switch guard must use := instead of =")
			}
			return id, y, true
		}
	}
	return "", nil, false
}

// SwitchStmt = ExprSwitchStmt | TypeSwitchStmt .
// ExprSwitchStmt = "switch" [ SimpleStmt ";" ] [ Expression ] "{" { ExprCaseClause } "}".
// TypeSwitchStmt = "switch" [ SimpleStmt ";" ] TypeSwitchGuard "{" {TypeCaseClause} "}" .
// TypeSwitchGuard = [ identifier ":=" ] PrimaryExpr "." "(" "type" ")" .
func (p *parser) parseSwitchStmt() ast.Stmt {
	off := p.match(scanner.SWITCH)

	b := p.setBrackets(0)
	defer p.setBrackets(b)

	if p.token == '{' {
		return p.parseExprSwitchStmt(off, nil, nil)
	}

	x := p.parseExpr()
	if p.token == '{' {
		if y, ok := isTypeSwitchGuardExpr(x); ok {
			return p.parseTypeSwitchStmt(off, nil, "", y)
		} else {
			return p.parseExprSwitchStmt(off, nil, x)
		}
	}

	init := p.parseSimpleStmt(x)
	if p.token == '{' {
		if id, y, ok := p.isTypeSwitchGuardStmt(init); ok {
			return p.parseTypeSwitchStmt(off, nil, id, y)
		}
	}
	p.match(';')

	if p.token == '{' {
		return p.parseExprSwitchStmt(off, init, nil)
	}

	x = p.parseExpr()
	if p.token == '{' {
		if y, ok := isTypeSwitchGuardExpr(x); ok {
			return p.parseTypeSwitchStmt(off, init, "", y)
		} else {
			return p.parseExprSwitchStmt(off, init, x)
		}
	}

	stmt := p.parseSimpleStmt(x)
	if id, y, ok := p.isTypeSwitchGuardStmt(stmt); ok {
		return p.parseTypeSwitchStmt(off, init, id, y)
	}

	p.error("invalid switch expression")
	p.sync('{')
	return p.parseExprSwitchStmt(off, init, nil)
}

// ExprCaseClause = ExprSwitchCase ":" StatementList .
// ExprSwitchCase = "case" ExpressionList | "default" .
func (p *parser) parseExprSwitchStmt(off int, init ast.Stmt, x ast.Expr) ast.Stmt {
	var (
		cs  []ast.ExprCaseClause
		def bool
	)
	p.match('{')
	for p.token == scanner.CASE || p.token == scanner.DEFAULT {
		t := p.token
		p.next()
		var xs []ast.Expr
		if t == scanner.CASE {
			xs = p.parseExprList(nil)
		} else {
			if def {
				p.error("multiple defaults in switch")
			}
			def = true
		}
		p.match(':')
		cs = append(cs, ast.ExprCaseClause{Xs: xs, Blk: p.parseStatementList()})
	}
	p.match('}')
	return &ast.ExprSwitchStmt{Off: off, Init: init, X: x, Cases: cs}
}

// TypeCaseClause  = TypeSwitchCase ":" StatementList .
// TypeSwitchCase  = "case" TypeList | "default" .
func (p *parser) parseTypeSwitchStmt(off int, init ast.Stmt, id string, x ast.Expr) ast.Stmt {
	def := false
	var cs []ast.TypeCaseClause
	p.match('{')
	for p.token == scanner.CASE || p.token == scanner.DEFAULT {
		t := p.token
		p.next()
		var ts []ast.Type
		if t == scanner.CASE {
			ts = p.parseTypeList()
		} else {
			if def {
				p.error("multiple defaults in switch")
			}
			def = true
		}
		p.match(':')
		cs = append(cs, ast.TypeCaseClause{Types: ts, Blk: p.parseStatementList()})
	}
	p.match('}')
	return &ast.TypeSwitchStmt{Off: off, Init: init, Id: id, X: x, Cases: cs}
}

// TypeList = Type { "," Type } .
func (p *parser) parseTypeList() (ts []ast.Type) {
	for {
		ts = append(ts, p.parseType())
		if p.token != ',' {
			break
		}
		p.next()
	}
	return
}

// Checks if X is a (possibly parenthesized) receive expression.
func isReceiveExpr(x ast.Expr) bool {
	for {
		switch y := x.(type) {
		case *ast.ParensExpr:
			x = y.X
		case *ast.UnaryExpr:
			return y.Op == scanner.RECV
		default:
			return false
		}
	}
}

// SendStmt = Channel "<-" Expression .
// RecvStmt   = [ ExpressionList "=" | IdentifierList ":=" ] RecvExpr .
// RecvExpr   = Expression .
func (p *parser) parseSendOrRecv() ast.Stmt {
	x := p.parseExpr()
	if p.token == scanner.RECV {
		// Send statement.
		p.next()
		y := p.parseExpr()
		return &ast.SendStmt{Ch: x, X: y}
	} else {
		// Receive statement.
		op := uint(0)
		var y, rcv ast.Expr
		if p.token == ',' {
			p.next()
			y = p.parseExpr()
			op = p.token
			if op != '=' && op != scanner.DEFINE {
				p.error("expected = or := in a receive statement")
				op = '='
			} else {
				p.next()
			}
			rcv = p.parseExpr()
		} else if p.token == '=' || p.token == scanner.DEFINE {
			op = p.token
			p.next()
			rcv = p.parseExpr()
		} else {
			rcv = x
			x = nil
		}
		if !isReceiveExpr(rcv) {
			p.error("receive statement must contain a receive expression")
		}
		return &ast.RecvStmt{Op: op, X: x, Y: y, Rcv: rcv}
	}
}

// SelectStmt = "select" "{" { CommClause } "}" .
// CommClause = CommCase ":" StatementList .
// CommCase   = "case" ( SendStmt | RecvStmt ) | "default" .
func (p *parser) parseSelectStmt() ast.Stmt {
	def := false
	var cs []ast.CommClause
	off := p.match(scanner.SELECT)
	p.match('{')
	for p.token == scanner.CASE || p.token == scanner.DEFAULT {
		t := p.token
		p.next()
		var c ast.Stmt
		if t == scanner.CASE {
			c = p.parseSendOrRecv()
		} else {
			if def {
				p.error("multiple defaults in select")
			}
			def = true
		}
		p.match(':')
		cs = append(cs, ast.CommClause{Comm: c, Blk: p.parseStatementList()})
	}
	p.match('}')
	return &ast.SelectStmt{Off: off, Comms: cs}
}

// ForStmt = "for" [ Condition | ForClause | RangeClause ] Block .
// Condition = Expression .
// ForClause = [ InitStmt ] ";" [ Condition ] ";" [ PostStmt ] .
// InitStmt = SimpleStmt .
// PostStmt = SimpleStmt .
// RangeClause = [ ExpressionList "=" | IdentifierList ":=" ] "range" Expression .
func (p *parser) parseForStmt() ast.Stmt {
	off := p.match(scanner.FOR)
	b := p.setBrackets(0)
	defer p.setBrackets(b)

	if p.token == '{' {
		// infinite loop: "for { ..."
		return &ast.ForStmt{Off: off, Blk: p.parseBlock()}
	}

	if p.token == scanner.RANGE {
		// range for: "for range ex { ..."
		p.next()
		return &ast.ForRangeStmt{
			Off:   off,
			Op:    '=',
			Range: p.parseExpr(),
			Blk:   p.parseBlock(),
		}
	}

	if p.token == ';' {
		// ordinary for : "for ; [cond] ; [post] { ..."
		p.next()
		var (
			x    ast.Expr
			post ast.Stmt
		)
		if p.token != ';' {
			x = p.parseExpr()
		}
		p.match(';')
		if p.token != '{' {
			post = p.parseSimpleStmt(p.parseExpr())
		}
		return &ast.ForStmt{Off: off, Cond: x, Post: post, Blk: p.parseBlock()}
	}

	x := p.parseExpr()
	if p.token == '{' {
		// "while" loop "for cond { ..."
		return &ast.ForStmt{Off: off, Cond: x, Blk: p.parseBlock()}
	}

	init := p.parseSimpleStmtOrRange(x)
	x = nil
	if r, ok := init.(*ast.ForRangeStmt); ok {
		// range for
		r.Off = off
		r.Blk = p.parseBlock()
		return r
	}

	// ordinary for: "for init ; [cond] ; [post] { ..."
	p.match(';')
	if p.token != ';' {
		x = p.parseExpr()
	}
	p.match(';')
	var post ast.Stmt
	if p.token != '{' {
		post = p.parseSimpleStmt(p.parseExpr())
	}
	return &ast.ForStmt{Off: off, Init: init, Cond: x, Post: post, Blk: p.parseBlock()}
}

// DeferStmt = "defer" Expression .
func (p *parser) parseDeferStmt() ast.Stmt {
	return &ast.DeferStmt{Off: p.match(scanner.DEFER), X: p.parseExpr()}
}

// SimpleStmt =
//     EmptyStmt | ExpressionStmt | SendStmt | IncDecStmt | Assignment |
//     ShortVarDecl .
func (p *parser) parseSimpleStmt(e ast.Expr) ast.Stmt {
	switch p.token {
	case scanner.RECV:
		return p.parseSendStmt(e)
	case scanner.INC, scanner.DEC:
		return p.parseIncDecStmt(e)
	}
	if p.token == ';' || p.token == ':' || p.token == '}' || p.token == '{' {
		return &ast.ExprStmt{e}
	}
	var es []ast.Expr
	if p.token == ',' {
		p.next()
		es = p.parseExprList(e)
	} else {
		es = append(es, e)
	}
	t := p.token
	if t == scanner.DEFINE {
		p.next()
		return p.parseShortVarDecl(es)
	} else if isAssignOp(t) {
		p.next()
		return p.parseAssignment(t, es)
	}

	p.error("Invalid statement")
	return &ast.Error{p.scan.TOff}
}

// Parses a SimpleStmt or a RangeClause. Used by `parseForStmt` to
// disambiguate between a range for and an ordinary for.
func (p *parser) parseSimpleStmtOrRange(e ast.Expr) ast.Stmt {
	switch p.token {
	case scanner.RECV:
		return p.parseSendStmt(e)
	case scanner.INC, scanner.DEC:
		return p.parseIncDecStmt(e)
	}
	if p.token == ';' || p.token == '{' {
		return &ast.ExprStmt{e}
	}
	var es []ast.Expr
	if p.token == ',' {
		p.next()
		es = p.parseExprList(e)
	} else {
		es = append(es, e)
	}
	t := p.token
	if t == scanner.DEFINE || isAssignOp(t) {
		p.next()
		if p.token == scanner.RANGE {
			return p.parseRangeClause(t, es)
		}
		if t == scanner.DEFINE {
			return p.parseShortVarDecl(es)
		} else {
			return p.parseAssignment(t, es)
		}
	}
	p.error("Invalid statement")
	return &ast.Error{p.scan.TOff}
}

// SendStmt = Channel "<-" Expression .
// Channel  = Expression .
func (p *parser) parseSendStmt(ch ast.Expr) ast.Stmt {
	p.match(scanner.RECV)
	return &ast.SendStmt{Ch: ch, X: p.parseExpr()}
}

// IncDecStmt = Expression ( "++" | "--" ) .
func (p *parser) parseIncDecStmt(e ast.Expr) ast.Stmt {
	if p.token == scanner.INC {
		p.match(scanner.INC)
		return &ast.IncStmt{e}
	} else {
		p.match(scanner.DEC)
		return &ast.DecStmt{e}
	}
}

// Assignment = ExpressionList assign_op ExpressionList .
// assign_op = [ add_op | mul_op ] "=" .
func (p *parser) parseAssignment(op uint, lhs []ast.Expr) ast.Stmt {
	rhs := p.parseExprList(nil)
	return &ast.AssignStmt{op, lhs, rhs}
}

// ShortVarDecl = IdentifierList ":=" ExpressionList .
func (p *parser) parseShortVarDecl(lhs []ast.Expr) ast.Stmt {
	rhs := p.parseExprList(nil)
	return &ast.AssignStmt{scanner.DEFINE, lhs, rhs}
}

// RangeClause = [ ExpressionList "=" | IdentifierList ":=" ] "range" Expression .
func (p *parser) parseRangeClause(op uint, lhs []ast.Expr) *ast.ForRangeStmt {
	p.match(scanner.RANGE)
	return &ast.ForRangeStmt{Op: op, LHS: lhs, Range: p.parseExpr()}
}
