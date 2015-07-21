package parser

import (
	"fmt"
	"golang/ast"
	s "golang/scanner"
	"io"
	"os"
	// "runtime"
)

type parser struct {
	errors []error
	scan   s.Scanner

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

// Parse a source file
func Parse(name string, src string) (*ast.File, error) {
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
	e := parseError{p.scan.Name, p.scan.TLine, p.scan.TPos, msg}
	p.errors = append(p.errors, e)
}

// Emit an expected token mismatch error.
func (p *parser) expectError(exp, act uint) {
	p.error(fmt.Sprintf("expected %s, got %s", s.TokenNames[exp], s.TokenNames[act]))
}

// Get the next token from the scanner.
func (p *parser) next() {
	p.token = p.scan.Get()
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

// Advance to the next token iff the current one is TOKEN.
func (p *parser) match(token uint) bool {
	if p.expect(token) {
		p.next()
		return true
	} else {
		return false
	}
}

// Advances to the next token iff the current one is TOKEN. Returns token
// value.
func (p *parser) matchRaw(token uint) []byte {
	if p.expect(token) {
		value := p.scan.Value
		p.next()
		return value
	} else {
		return nil
	}
}

// Advances to the next token iff the current one is TOKEN. Returns token
// value as string.
func (p *parser) matchString(token uint) string {
	return string(p.matchRaw(token))
}

// Skip tokens, until given token found, then consume it. Report an error
// only if some tokens were skipped.
func (p *parser) sync(token uint) {
	if p.token != token {
		p.expectError(token, p.token)
	}
	for p.token != s.EOF && p.token != token {
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
	for p.token != s.EOF && p.token != t1 && p.token != t2 {
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
		case s.CONST, s.TYPE, s.VAR, s.FUNC, s.EOF:
			return
		default:
			p.next()
		}
	}
}

// SourceFile = PackageClause ";" { ImportDecl ";" } { TopLevelDecl ";" } .
func (p *parser) parseFile() *ast.File {
	// Parse package name
	name := p.parsePackageClause()
	if len(name) == 0 {
		return nil
	}

	if !p.match(';') {
		return nil
	}

	// Parse import declaration(s)
	imports := p.parseImportDecls()

	// Parse toplevel declarations.
	decls := p.parseToplevelDecls()

	p.match(s.EOF)

	return &ast.File{name, imports, decls}
}

// Parse a package clause. Return the package name or an empty string on error.
//
// PackageClause  = "package" PackageName .
// PackageName    = identifier .
func (p *parser) parsePackageClause() string {
	p.match(s.PACKAGE)
	return p.matchString(s.ID)
}

// Parse import declarations(s).
//
// ImportDecl       = "import" ( ImportSpec | "(" { ImportSpec ";" } ")" ) .
func (p *parser) parseImportDecls() (imports []ast.Import) {
	for p.token == s.IMPORT {
		p.match(s.IMPORT)
		if p.token == '(' {
			p.next()
			for p.token != s.EOF && p.token != ')' {
				if name, path := p.parseImportSpec(); len(path) > 0 {
					imports = append(imports, ast.Import{name, path})
				}
				if p.token != ')' {
					p.sync2(';', ')')
				}
			}
			p.match(')')
		} else {
			if name, path := p.parseImportSpec(); len(path) > 0 {
				imports = append(imports, ast.Import{name, path})
			}
		}
		p.sync(';')
	}
	return
}

// Parse import spec
//
// ImportSpec       = [ "." | PackageName ] ImportPath .
// ImportPath       = string_lit .
func (p *parser) parseImportSpec() (name string, path []byte) {
	if p.token == '.' {
		name = "."
		p.next()
	} else if p.token == s.ID {
		name = p.matchString(s.ID)
	} else {
		name = ""
	}
	path = p.matchRaw(s.STRING)
	return
}

// Parse toplevel declaration(s)
//
// Declaration   = ConstDecl | TypeDecl | VarDecl .
// TopLevelDecl  = Declaration | FunctionDecl | MethodDecl .
func (p *parser) parseToplevelDecls() (dcls []ast.Decl) {
	for {
		var f func() ast.Decl
		switch p.token {
		case s.TYPE:
			f = p.parseTypeDecl
		case s.CONST:
			f = p.parseConstDecl
		case s.VAR:
			f = p.parseVarDecl
		case s.FUNC:
			f = p.parseFuncDecl
		case s.EOF:
			return
		default:
			p.error("expected type, const, var or func/method declaration")
			p.syncDecl()
			f = nil
		}
		if f != nil {
			dcls = append(dcls, f())
			p.match(';')
		}
	}
}

// Parse type declaration
//
// TypeDecl = "type" ( TypeSpec | "(" { TypeSpec ";" } ")" ) .
func (p *parser) parseTypeDecl() ast.Decl {
	p.match(s.TYPE)
	if p.token == '(' {
		p.next()
		var ts []*ast.TypeDecl = nil
		for p.token != s.EOF && p.token != ')' {
			ts = append(ts, p.parseTypeSpec())
			if p.token != ')' {
				p.sync2(';', ')')
			}
		}
		p.match(')')
		return &ast.TypeGroup{Decls: ts}
	} else {
		return p.parseTypeSpec()
	}
}

// TypeSpec = identifier Type .
func (p *parser) parseTypeSpec() *ast.TypeDecl {
	id := p.matchString(s.ID)
	t := p.parseType()
	return &ast.TypeDecl{Name: id, Type: t}
}

// Determine if the given TOKEN could be a beginning of a typespec.
func isTypeLookahead(token uint) bool {
	switch token {
	case s.ID, '[', s.STRUCT, '*', s.FUNC, s.INTERFACE, s.MAP, s.CHAN, s.RECV, '(':
		return true
	default:
		return false
	}
}

// Type     = TypeName | TypeLit | "(" Type ")" .
// TypeName = identifier | QualifiedIdent .
// TypeLit  = ArrayType | StructType | PointerType | FunctionType | InterfaceType |
//            SliceType | MapType | ChannelType .
func (p *parser) parseType() ast.TypeSpec {
	switch p.token {
	case s.ID:
		return p.parseQualId()

	// ArrayType   = "[" ArrayLength "]" ElementType .
	// ArrayLength = Expression .
	// ElementType = Type .
	// SliceType = "[" "]" ElementType .
	// Allow here an array type of unspecified size, that can be used
	// only in composite literal expressions.
	// ArrayLength = Expression | "..." .
	case '[':
		p.next()
		if p.token == ']' {
			p.next()
			return &ast.SliceType{p.parseType()}
		} else if p.token == s.DOTS {
			p.next()
			p.match(']')
			return &ast.ArrayType{Dim: nil, EltType: p.parseType()}
		} else {
			e := p.parseExpr()
			p.match(']')
			t := p.parseType()
			return &ast.ArrayType{Dim: e, EltType: t}
		}

	// PointerType = "*" BaseType .
	// BaseType = Type .
	case '*':
		p.next()
		return &ast.PtrType{p.parseType()}

	// MapType     = "map" "[" KeyType "]" ElementType .
	// KeyType     = Type .
	case s.MAP:
		p.next()
		p.match('[')
		k := p.parseType()
		p.match(']')
		t := p.parseType()
		return &ast.MapType{k, t}

	// ChannelType = ( "chan" [ "<-" ] | "<-" "chan" ) ElementType .
	case s.RECV:
		p.next()
		p.match(s.CHAN)
		t := p.parseType()
		return &ast.ChanType{Send: false, Recv: true, EltType: t}
	case s.CHAN:
		p.next()
		send, recv := true, true
		if p.token == s.RECV {
			p.next()
			send, recv = true, false
		}
		t := p.parseType()
		return &ast.ChanType{Send: send, Recv: recv, EltType: t}

	case s.STRUCT:
		return p.parseStructType()

	case s.FUNC:
		return p.parseFuncType()

	case s.INTERFACE:
		return p.parseInterfaceType()

	case '(':
		p.next()
		t := p.parseType()
		p.match(')')
		return t

	default:
		p.error("expected typespec")
		return &ast.Error{}
	}
}

// TypeName = identifier | QualifiedIdent .
func (p *parser) parseQualId() *ast.QualId {
	pkg := p.matchString(s.ID)
	var id string
	if p.token == '.' {
		p.next()
		id = p.matchString(s.ID)
	} else {
		pkg, id = "", pkg
	}
	return &ast.QualId{pkg, id}
}

// StructType     = "struct" "{" { FieldDecl ";" } "}" .
func (p *parser) parseStructType() *ast.StructType {
	var fs []*ast.FieldDecl = nil
	p.match(s.STRUCT)
	p.match('{')
	for p.token != s.EOF && p.token != '}' {
		fs = append(fs, p.parseFieldDecl())
		if p.token != '}' {
			p.sync2(';', '}')
		}
	}
	p.match('}')
	return &ast.StructType{fs}
}

// FieldDecl      = (IdentifierList Type | AnonymousField) [ Tag ] .
// AnonymousField = [ "*" ] TypeName .
// Tag            = string_lit .
func (p *parser) parseFieldDecl() *ast.FieldDecl {
	if p.token == '*' {
		// Anonymous field.
		p.next()
		pt := &ast.PtrType{p.parseQualId()}
		tag := p.parseTagOpt()
		return &ast.FieldDecl{Names: nil, Type: pt, Tag: tag}
	} else if p.token == s.ID {
		pkg := p.matchString(s.ID)
		if p.token == '.' {
			// If the field decl begins with a qualified-id, it's parsed as an
			// anonymous field.
			p.next()
			id := p.matchString(s.ID)
			t := &ast.QualId{pkg, id}
			tag := p.parseTagOpt()
			return &ast.FieldDecl{Names: nil, Type: t, Tag: tag}
		} else if p.token == s.STRING || p.token == ';' || p.token == '}' {
			// If it's only a single identifier, with no separate type
			// declaration, it's also an anonymous filed.
			t := &ast.QualId{"", pkg}
			tag := p.parseTagOpt()
			return &ast.FieldDecl{Names: nil, Type: t, Tag: tag}
		} else {
			ids := p.parseIdList(pkg)
			t := p.parseType()
			tag := p.parseTagOpt()
			return &ast.FieldDecl{Names: ids, Type: t, Tag: tag}
		}
	}
	p.error("Invalid field declaration")
	return &ast.FieldDecl{Names: nil, Type: &ast.Error{}}
}

func (p *parser) parseTagOpt() (tag []byte) {
	if p.token == s.STRING {
		tag = p.matchRaw(s.STRING)
	} else {
		tag = nil
	}
	return
}

// IdentifierList = identifier { "," identifier } .
func (p *parser) parseIdList(id string) (ids []string) {
	if len(id) == 0 {
		id = p.matchString(s.ID)
	}
	if len(id) > 0 {
		ids = append(ids, id)
	}
	for p.token == ',' {
		p.next()
		id = p.matchString(s.ID)
		if len(id) > 0 {
			ids = append(ids, id)
		}
	}
	return ids
}

// FunctionType = "func" Signature .
func (p *parser) parseFuncType() *ast.FuncType {
	p.match(s.FUNC)
	return p.parseSignature()
}

// Signature = Parameters [ Result ] .
// Result    = Parameters | Type .
func (p *parser) parseSignature() *ast.FuncType {
	ps := p.parseParameters()
	var rs []*ast.ParamDecl = nil
	if p.token == '(' {
		rs = p.parseParameters()
	} else if isTypeLookahead(p.token) {
		rs = []*ast.ParamDecl{&ast.ParamDecl{Type: p.parseType()}}
	}
	return &ast.FuncType{ps, rs}
}

// Parameters    = "(" [ ParameterList [ "," ] ] ")" .
// ParameterList = ParameterDecl { "," ParameterDecl } .
func (p *parser) parseParameters() []*ast.ParamDecl {
	var (
		ids []string
		ds  []*ast.ParamDecl
	)
	p.match('(')
	if p.token == ')' {
		p.next()
		return nil
	}
	for p.token != s.EOF && p.token != ')' {
		id, t, v := p.parseIdOrType()
		if t != nil {
			ds = appendParamTypes(ds, ids)
			ds = append(ds, &ast.ParamDecl{Names: nil, Type: t, Variadic: v})
			ids = nil
		} else {
			ids = append(ids, id)
			if p.token != ',' && p.token != ')' {
				id, t, v = p.parseIdOrType()
				if t == nil {
					t = &ast.QualId{Pkg: "", Id: id}
				}
				ds = append(ds, &ast.ParamDecl{Names: ids, Type: t, Variadic: v})
				ids = nil
			}
		}
		if p.token != ')' {
			p.sync2(',', ')')
		}
	}
	ds = appendParamTypes(ds, ids)
	p.match(')')
	return ds
}

// Parses an identifier or a type. A qualified identifier is considered a
// type.
func (p *parser) parseIdOrType() (string, ast.TypeSpec, bool) {
	if p.token == s.ID {
		id := p.matchString(s.ID)
		if p.token == '.' {
			p.next()
			pkg := id
			if p.token == s.ID {
				id = p.matchString(s.ID)
			} else {
				p.error("incomplete qualified id")
				id = ""
			}
			return "", &ast.QualId{Pkg: pkg, Id: id}, false
		}
		return id, nil, false
	}
	v := false
	if p.token == s.DOTS {
		p.next()
		v = true
	}
	t := p.parseType() // FIXME: parenthesized types not allowed
	return "", t, v
}

// Converts an identifier list to a list of parameter declarations, taking
// each identifier to be a type name.
func appendParamTypes(ds []*ast.ParamDecl, ids []string) []*ast.ParamDecl {
	for i := range ids {
		t := &ast.QualId{Id: ids[i]}
		ds = append(ds, &ast.ParamDecl{Names: nil, Type: t, Variadic: false})
	}
	return ds
}

// InterfaceType      = "interface" "{" { MethodSpec ";" } "}" .
// MethodSpec         = MethodName Signature | InterfaceTypeName .
// MethodName         = identifier .
// InterfaceTypeName  = TypeName .
func (p *parser) parseInterfaceType() *ast.InterfaceType {
	p.match(s.INTERFACE)
	p.match('{')
	emb := []*ast.QualId(nil)
	meth := []*ast.MethodSpec(nil)
	for p.token != s.EOF && p.token != '}' {
		id := p.matchString(s.ID)
		if p.token == '.' {
			p.next()
			name := p.matchString(s.ID)
			emb = append(emb, &ast.QualId{Pkg: id, Id: name})
		} else if p.token == '(' {
			sig := p.parseSignature()
			meth = append(meth, &ast.MethodSpec{Name: id, Sig: sig})
		} else {
			emb = append(emb, &ast.QualId{Pkg: "", Id: id})
		}
		if p.token != '}' {
			p.sync2(';', '}')
		}
	}
	p.match('}')
	return &ast.InterfaceType{Embed: emb, Methods: meth}
}

// ConstDecl = "const" ( ConstSpec | "(" { ConstSpec ";" } ")" ) .
func (p *parser) parseConstDecl() ast.Decl {
	p.match(s.CONST)
	if p.token == '(' {
		p.next()
		cs := []*ast.ConstDecl(nil)
		for p.token != s.EOF && p.token != ')' {
			cs = append(cs, p.parseConstSpec())
			if p.token != ')' {
				p.sync2(';', ')')
			}
		}
		p.match(')')
		return &ast.ConstGroup{Decls: cs}
	} else {
		return p.parseConstSpec()
	}
}

// ConstSpec = IdentifierList [ [ Type ] "=" ExpressionList ] .
func (p *parser) parseConstSpec() *ast.ConstDecl {
	var (
		t  ast.TypeSpec
		es []ast.Expr
	)
	ids := p.parseIdList("")
	if isTypeLookahead(p.token) {
		t = p.parseType()
	}
	if p.token == '=' {
		p.next()
		es = p.parseExprList(nil)
	}
	return &ast.ConstDecl{Names: ids, Type: t, Values: es}
}

// VarDecl = "var" ( VarSpec | "(" { VarSpec ";" } ")" ) .
func (p *parser) parseVarDecl() ast.Decl {
	p.match(s.VAR)
	if p.token == '(' {
		p.next()
		vs := []*ast.VarDecl(nil)
		for p.token != s.EOF && p.token != ')' {
			vs = append(vs, p.parseVarSpec())
			if p.token != ')' {
				p.sync2(';', ')')
			}
		}
		p.match(')')
		return &ast.VarGroup{Decls: vs}
	} else {
		return p.parseVarSpec()
	}
}

// VarSpec = IdentifierList ( Type [ "=" ExpressionList ] | "=" ExpressionList ) .
func (p *parser) parseVarSpec() *ast.VarDecl {
	var (
		t  ast.TypeSpec
		es []ast.Expr
	)
	ids := p.parseIdList("")
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
	var r *ast.Receiver
	p.match(s.FUNC)
	if p.token == '(' {
		r = p.parseReceiver()
	}
	name := p.matchString(s.ID)
	sig := p.parseSignature()
	var blk *ast.Block = nil
	if p.token == '{' {
		blk = p.parseBlock()
	}
	return &ast.FuncDecl{Name: name, Recv: r, Sig: sig, Body: blk}
}

// Receiver     = "(" [ identifier ] [ "*" ] BaseTypeName ")" .
// BaseTypeName = identifier .
func (p *parser) parseReceiver() *ast.Receiver {
	p.match('(')
	var n string
	if p.token == s.ID {
		n = p.matchString(s.ID)
	}
	ptr := false
	if p.token == '*' {
		p.next()
		ptr = true
	}
	t := p.matchString(s.ID)
	p.sync2(')', ';')
	var tp ast.TypeSpec = &ast.QualId{Id: t}
	if ptr {
		tp = &ast.PtrType{Base: tp}
	}
	return &ast.Receiver{Name: n, Type: tp}
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
func (p *parser) parseExprOrType() (ast.Expr, ast.TypeSpec) {
	//	defer p.trace("ExprOrType")()

	return p.parseOrExprOrType()
}

func (p *parser) parseExpr() ast.Expr {
	// defer p.trace("Expression", s.TokenNames[p.token], p.scan.TLine, p.scan.TPos)()

	if e, _ := p.parseOrExprOrType(); e != nil {
		return e
	}
	p.error("Type not allowed in this context")
	return &ast.Error{}
}

// Expression = UnaryExpr | Expression binary_op UnaryExpr .
// `Expression` from the Go Specification is replaced by the following
// ``xxxExpr` productions.

// LogicalOrExpr = LogicalAndExpr { "||" LogicalAndExpr }
func (p *parser) parseOrExprOrType() (ast.Expr, ast.TypeSpec) {
	//	defer p.trace("OrExprOrType")()

	a0, t := p.parseAndExprOrType()
	if a0 == nil {
		return nil, t
	}
	for p.token == s.OR {
		p.next()
		a1, _ := p.parseAndExprOrType()
		a0 = &ast.BinaryExpr{Op: s.OR, Arg0: a0, Arg1: a1}
	}
	return a0, nil
}

// LogicalAndExpr = CompareExpr { "&&" CompareExpr }
func (p *parser) parseAndExprOrType() (ast.Expr, ast.TypeSpec) {
	//	defer p.trace("AndExprOrType")()

	a0, t := p.parseCompareExprOrType()
	if a0 == nil {
		return nil, t
	}
	for p.token == s.AND {
		p.next()
		a1, _ := p.parseCompareExprOrType()
		a0 = &ast.BinaryExpr{Op: s.AND, Arg0: a0, Arg1: a1}
	}
	return a0, nil
}

// CompareExpr = AddExpr { rel_op AddExpr }
func (p *parser) parseCompareExprOrType() (ast.Expr, ast.TypeSpec) {
	//	defer p.trace("CompareExprOrType")()

	a0, t := p.parseAddExprOrType()
	if a0 == nil {
		return nil, t
	}
	for isRelOp(p.token) {
		op := p.token
		p.next()
		a1, _ := p.parseAddExprOrType()
		a0 = &ast.BinaryExpr{Op: op, Arg0: a0, Arg1: a1}
	}
	return a0, nil
}

// AddExpr = MulExpr { add_op MulExpr}
func (p *parser) parseAddExprOrType() (ast.Expr, ast.TypeSpec) {
	//	defer p.trace("AddExprOrType")()

	a0, t := p.parseMulExprOrType()
	if a0 == nil {
		return nil, t
	}
	for isAddOp(p.token) {
		op := p.token
		p.next()
		a1, _ := p.parseMulExprOrType()
		a0 = &ast.BinaryExpr{Op: op, Arg0: a0, Arg1: a1}
	}
	return a0, nil
}

// MulExpr = UnaryExpr { mul_op UnaryExpr }
func (p *parser) parseMulExprOrType() (ast.Expr, ast.TypeSpec) {
	//	defer p.trace("MulExprOrType")()

	a0, t := p.parseUnaryExprOrType()
	if a0 == nil {
		return nil, t
	}
	for isMulOp(p.token) {
		op := p.token
		p.next()
		a1, _ := p.parseUnaryExprOrType()
		a0 = &ast.BinaryExpr{Op: op, Arg0: a0, Arg1: a1}
	}
	return a0, nil
}

// binary_op  = "||" | "&&" | rel_op | add_op | mul_op .

// rel_op     = "==" | "!=" | "<" | "<=" | ">" | ">=" .
func isRelOp(t uint) bool {
	switch t {
	case s.EQ, s.NE, s.LT, s.LE, s.GT, s.GE:
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
	case '*', '/', '%', s.SHL, s.SHR, '&', s.ANDN:
		return true
	default:
		return false
	}
}

// assign_op = [ add_op | mul_op ] "=" .
func isAssignOp(t uint) bool {
	switch t {
	case '=', s.PLUS_ASSIGN, s.MINUS_ASSIGN, s.MUL_ASSIGN, s.DIV_ASSIGN,
		s.REM_ASSIGN, s.AND_ASSIGN, s.OR_ASSIGN, s.XOR_ASSIGN, s.SHL_ASSIGN,
		s.SHR_ASSIGN, s.ANDN_ASSIGN:
		return true
	default:
		return false
	}
}

// UnaryExpr  = PrimaryExpr | unary_op UnaryExpr .
// unary_op   = "+" | "-" | "!" | "^" | "*" | "&" | "<-" .
func (p *parser) parseUnaryExprOrType() (ast.Expr, ast.TypeSpec) {
	//	defer p.trace("UnaryExprOrType")()

	switch p.token {
	case '+', '-', '!', '^', '&':
		op := p.token
		p.next()
		a, _ := p.parseUnaryExprOrType()
		return &ast.UnaryExpr{Op: op, Arg: a}, nil
	case '*':
		p.next()
		if a, t := p.parseUnaryExprOrType(); t == nil {
			return &ast.UnaryExpr{Op: '*', Arg: a}, nil
		} else {
			return nil, &ast.PtrType{Base: t}
		}
	case s.RECV:
		p.next()
		if a, t := p.parseUnaryExprOrType(); t == nil {
			return &ast.UnaryExpr{Op: s.RECV, Arg: a}, nil
		} else if ch, ok := t.(*ast.ChanType); ok {
			ch.Recv, ch.Send = true, false
			return nil, ch
		} else {
			p.error("invalid receive operation")
			return &ast.Error{}, nil
		}
	default:
		return p.parsePrimaryExprOrType()
	}
}

func (p *parser) needType(ex ast.Expr) (ast.TypeSpec, bool) {
	switch t := ex.(type) {
	case *ast.QualId:
		return t, true
	case *ast.Selector:
		if a, ok := t.Arg.(*ast.QualId); ok && len(a.Pkg) == 0 {
			return &ast.QualId{Pkg: a.Id, Id: t.Id}, true
		}
	}
	return nil, false
}

// TypeAssertion = "." "(" Type ")" .
// Also allow "." "(" "type" ")".
func (p *parser) parseTypeAssertion(ex ast.Expr) ast.Expr {
	p.match('(')
	var t ast.TypeSpec
	if p.token == s.TYPE {
		p.next()
	} else {
		t = p.parseType()
	}
	p.match(')')
	return &ast.TypeAssertion{t, ex}
}

// PrimaryExpr =
//     Operand |
//     Conversion |
//     BuiltinCall |
//     PrimaryExpr Selector |
//     PrimaryExpr Index |
//     PrimaryExpr Slice |
//     PrimaryExpr TypeAssertion |
//     PrimaryExpr Call .

func (p *parser) parsePrimaryExprOrType() (ast.Expr, ast.TypeSpec) {
	// defer p.trace("PrimaryExprOrType", s.TokenNames[p.token])()

	var (
		ex ast.Expr
		t  ast.TypeSpec
	)
	// Handle initial Operand, Conversion or a BuiltinCall
	switch p.token {
	case s.ID: // CompositeLit, MethodExpr, Conversion, BuiltinCall, OperandName
		id := p.matchString(s.ID)
		if p.token == '{' && p.brackets > 0 {
			ex = p.parseCompositeLiteral(&ast.QualId{Id: id})
		} else if p.token == '(' {
			ex = p.parseCall(&ast.QualId{Id: id})
		} else {
			ex = &ast.QualId{Id: id}
		}
	case '(': // Expression, MethodExpr, Conversion
		p.next()
		p.beginBrackets()
		ex, t = p.parseExprOrType()
		p.endBrackets()
		p.match(')')
		if ex == nil {
			if p.token == '(' {
				ex = p.parseConversion(t)
			} else {
				return nil, t
			}
		}
	case '[', s.STRUCT, s.MAP: // Conversion, CompositeLit
		t = p.parseType()
		if p.token == '(' {
			ex = p.parseConversion(t)
		} else if p.token == '{' {
			ex = p.parseCompositeLiteral(t)
		} else {
			return nil, t
		}
	case s.FUNC: // Conversion, FunctionLiteral
		t = p.parseType()
		if p.token == '(' {
			ex = p.parseConversion(t)
		} else if p.token == '{' {
			ex = p.parseFuncLiteral(t)
		} else {
			return nil, t
		}
	case '*', s.RECV:
		panic("should not reach here")
	case s.INTERFACE, s.CHAN: // Conversion
		t = p.parseType()
		if p.token == '(' {
			ex = p.parseConversion(t)
		} else {
			return nil, t
		}
	case s.INTEGER, s.FLOAT, s.IMAGINARY, s.RUNE, s.STRING: // BasicLiteral
		k := p.token
		v := p.matchRaw(k)
		return &ast.Literal{Kind: k, Value: v}, nil
	default:
		p.error("token cannot start neither expression nor type")
		ex = &ast.Error{}
	}
	// Parse the left-recursive alternatives for PrimaryExpr, folding the left-
	// hand parts in the variable `EX`.
	for {
		switch p.token {
		case '.': // TypeAssertion or Selector
			p.next()
			if p.token == '(' {
				// PrimaryExpr TypeAssertion
				// TypeAssertion = "." "(" Type ")" .
				ex = p.parseTypeAssertion(ex)
			} else {
				// PrimaryExpr Selector
				// Selector = "." identifier .
				id := p.matchString(s.ID)
				ex = &ast.Selector{Arg: ex, Id: id}
			}
		case '[':
			// PrimaryExpr Index
			// PrimaryExpr Slice
			ex = p.parseIndexOrSlice(ex)
		case '(':
			// PrimaryExpr Call
			ex = p.parseCall(ex)
		default:
			if p.token == '{' && p.brackets > 0 {
				// Composite literal
				typ, ok := p.needType(ex)
				if !ok {
					p.error("invalid type for composite literal")
				}
				ex = p.parseCompositeLiteral(typ)
			} else {
				return ex, nil
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
func (p *parser) parseCompositeLiteral(typ ast.TypeSpec) ast.Expr {
	elts := p.parseLiteralValue()
	return &ast.CompLiteral{Type: typ, Elts: elts}
}

// LiteralValue  = "{" [ ElementList [ "," ] ] "}" .
// ElementList   = Element { "," Element } .
func (p *parser) parseLiteralValue() (elts []*ast.Element) {
	p.match('{')
	for p.token != s.EOF && p.token != '}' {
		elts = append(elts, p.parseElement())
		if p.token != '}' {
			p.sync2(',', '}')
		}
	}
	p.match('}')
	return elts
}

// Element       = [ Key ":" ] Value .
// Key           = FieldName | ElementIndex .
// FieldName     = identifier .
// ElementIndex  = Expression .
// Value         = Expression | LiteralValue .
func (p *parser) parseElement() *ast.Element {
	var k ast.Expr
	if p.token != '{' {
		k = p.parseExpr()
		if p.token != ':' {
			return &ast.Element{Key: nil, Value: k}
		}
		p.match(':')
	}
	if p.token == '{' {
		elts := p.parseLiteralValue()
		e := &ast.CompLiteral{Type: nil, Elts: elts}
		return &ast.Element{Key: k, Value: e}
	} else {
		e := p.parseExpr()
		return &ast.Element{Key: k, Value: e}
	}
}

// Conversion = Type "(" Expression [ "," ] ")" .
func (p *parser) parseConversion(typ ast.TypeSpec) ast.Expr {
	p.match('(')
	p.beginBrackets()
	a := p.parseExpr()
	if p.token == ',' {
		p.next()
	}
	p.endBrackets()
	p.match(')')
	return &ast.Conversion{Type: typ, Arg: a}
}

// Call           = "(" [ ArgumentList [ "," ] ] ")" .
// ArgumentList   = ExpressionList [ "..." ] .
// BuiltinCall = identifier "(" [ BuiltinArgs [ "," ] ] ")" .
// BuiltinArgs = Type [ "," ArgumentList ] | ArgumentList .
func (p *parser) parseCall(f ast.Expr) ast.Expr {
	p.match('(')
	p.beginBrackets()
	defer p.endBrackets()
	if p.token == ')' {
		p.next()
		return &ast.Call{Func: f}
	}
	var as []ast.Expr
	a, t := p.parseExprOrType()
	if a != nil {
		as = append(as, a)
	}
	dots := false
	if p.token == s.DOTS {
		p.next()
		dots = true
	}
	if p.token != ')' {
		p.match(',')
	}
	for p.token != s.EOF && p.token != ')' && !dots {
		as = append(as, p.parseExpr())
		if p.token == s.DOTS {
			p.next()
			dots = true
		}
		if p.token != ')' {
			p.sync2(',', ')')
		}
	}
	p.match(')')
	return &ast.Call{Func: f, Type: t, Args: as, Ellipsis: dots}
}

// FunctionLit = "func" Function .
func (p *parser) parseFuncLiteral(typ ast.TypeSpec) ast.Expr {
	sig := typ.(*ast.FuncType)
	b := p.parseBlock()
	return &ast.FuncLiteral{Sig: sig, Body: b}
}

// Index          = "[" Expression "]" .
// Slice          = "[" ( [ Expression ] ":" [ Expression ] ) |
//                      ( [ Expression ] ":" Expression ":" Expression )
//                  "]" .
func (p *parser) parseIndexOrSlice(e ast.Expr) ast.Expr {
	p.match('[')
	p.beginBrackets()
	defer p.endBrackets()
	if p.token == ':' {
		p.next()
		if p.token == ']' {
			p.next()
			return &ast.SliceExpr{Array: e}
		}
		h := p.parseExpr()
		if p.token == ']' {
			p.next()
			return &ast.SliceExpr{Array: e, High: h}
		}
		p.match(':')
		c := p.parseExpr()
		p.match(']')
		return &ast.SliceExpr{Array: e, High: h, Cap: c}
	} else {
		i := p.parseExpr()
		if p.token == ']' {
			p.next()
			return &ast.IndexExpr{Array: e, Idx: i}
		}
		p.match(':')
		if p.token == ']' {
			p.next()
			return &ast.SliceExpr{Array: e, Low: i}
		}
		h := p.parseExpr()
		if p.token == ']' {
			p.next()
			return &ast.SliceExpr{Array: e, Low: i, High: h}
		}
		p.match(':')
		c := p.parseExpr()
		p.match(']')
		return &ast.SliceExpr{Array: e, Low: i, High: h, Cap: c}
	}
}

// Block = "{" StatementList "}" .
func (p *parser) parseBlock() *ast.Block {
	//	defer p.trace("Block")()
	p.match('{')
	st := p.parseStatementList()
	p.match('}')
	return &ast.Block{Stmts: st}
}

// StatementList = { Statement ";" } .
func (p *parser) parseStatementList() []ast.Stmt {
	var st []ast.Stmt
	b := p.brackets
	p.brackets = 1
	for p.token != s.EOF && p.token != '}' && p.token != s.CASE && p.token != s.DEFAULT {
		st = append(st, p.parseStmt())
		p.syncEndStatement()
	}
	p.brackets = b
	return st
}

func (p *parser) syncEndStatement() {
	if p.token == '}' {
		return
	}
	if p.match(';') {
		return
	}
	for p.token != s.EOF && p.token != ';' && p.token != '}' &&
		p.token != s.CASE && p.token != s.DEFAULT {
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
	case s.CONST:
		// FIXME: these type assertions are rather dubious
		return p.parseConstDecl().(ast.Stmt)
	case s.TYPE:
		return p.parseTypeDecl().(ast.Stmt)
	case s.VAR:
		return p.parseVarDecl().(ast.Stmt)
	case s.GO:
		return p.parseGoStmt()
	case s.RETURN:
		return p.parseReturnStmt()
	case s.BREAK:
		return p.parseBreakStmt()
	case s.CONTINUE:
		return p.parseContinueStmt()
	case s.GOTO:
		return p.parseGotoStmt()
	case s.FALLTHROUGH:
		return p.parseFallthroughStmt()
	case '{':
		return p.parseBlock()
	case s.IF:
		return p.parseIfStmt()
	case s.SWITCH:
		return p.parseSwitchStmt()
	case s.SELECT:
		return p.parseSelectStmt()
	case s.FOR:
		return p.parseForStmt()
	case s.DEFER:
		return p.parseDeferStmt()
	case ';':
		return &ast.EmptyStmt{}
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
	id := ""
	q, ok := x.(*ast.QualId)
	if !ok || len(q.Pkg) > 0 {
		p.error("a label must consiste of single identifier")
	} else {
		id = q.Id
	}
	p.match(':')
	return &ast.LabeledStmt{Label: id, Stmt: p.parseStmt()}
}

// GoStmt = "go" Expression .
func (p *parser) parseGoStmt() ast.Stmt {
	p.match(s.GO)
	e := p.parseExpr()
	return &ast.GoStmt{e}
}

// ReturnStmt = "return" [ ExpressionList ] .
func (p *parser) parseReturnStmt() ast.Stmt {
	p.match(s.RETURN)
	if p.token != ';' && p.token != '}' {
		e := p.parseExprList(nil)
		return &ast.ReturnStmt{e}
	} else {
		return &ast.ReturnStmt{nil}
	}
}

// BreakStmt = "break" [ Label ] .
func (p *parser) parseBreakStmt() ast.Stmt {
	p.match(s.BREAK)
	if p.token == s.ID {
		return &ast.BreakStmt{p.matchString(s.ID)}
	} else {
		return &ast.BreakStmt{}
	}
}

// ContinueStmt = "continue" [ Label ] .
func (p *parser) parseContinueStmt() ast.Stmt {
	p.match(s.CONTINUE)
	if p.token == s.ID {
		return &ast.ContinueStmt{p.matchString(s.ID)}
	} else {
		return &ast.ContinueStmt{}
	}
}

// GotoStmt = "goto" Label .
func (p *parser) parseGotoStmt() ast.Stmt {
	p.match(s.GOTO)
	return &ast.GotoStmt{p.matchString(s.ID)}
}

// FallthroughStmt = "fallthrough" .
func (p *parser) parseFallthroughStmt() ast.Stmt {
	p.match(s.FALLTHROUGH)
	return &ast.FallthroughStmt{}
}

// IfStmt = "if" [ SimpleStmt ";" ] Expression Block [ "else" ( IfStmt | Block ) ] .
func (p *parser) parseIfStmt() ast.Stmt {
	p.match(s.IF)
	b := p.brackets
	p.brackets = 0
	var ss ast.Stmt
	e := p.parseExpr()
	if p.token != '{' {
		ss = p.parseSimpleStmt(e)
		p.match(';')
		e = p.parseExpr()
	}
	p.brackets = b
	then := p.parseBlock()
	if p.token == s.ELSE {
		var els ast.Stmt
		p.next()
		if p.token == s.IF {
			els = p.parseIfStmt()
		} else {
			els = p.parseBlock()
		}
		return &ast.IfStmt{ss, e, then, els}
	} else {
		return &ast.IfStmt{ss, e, then, nil}
	}
}

// Checks if an Expression is a TypeSwitchGuard.
func isTypeSwitchGuardExpr(x ast.Expr) (ast.Expr, bool) {
	if y, ok := x.(*ast.TypeAssertion); ok && y.Type == nil {
		return y.Arg, true
	} else {
		return nil, false
	}

}

// Checks if a SimpleStmt is a TypeSwitchGuard. If true, returns the statement
// constituent parts.
func (p *parser) isTypeSwitchGuardStmt(x ast.Stmt) (string, ast.Expr, bool) {
	if a, ok := x.(*ast.AssignStmt); ok {
		if len(a.LHS) == 1 && len(a.RHS) == 1 {
			lhs, rhs := a.LHS[0], a.RHS[0]
			if y, ok := isTypeSwitchGuardExpr(rhs); ok {
				id := ""
				if z, ok := lhs.(*ast.QualId); !ok || len(z.Pkg) > 0 {
					id = ""
					p.error("invalid identifier in type switch")
				} else {
					id = z.Id
				}
				if a.Op != s.DEFINE {
					p.error("type switch guard must use := instead of =")
				}
				return id, y, true
			}
		}
	}
	return "", nil, false
}

// SwitchStmt = ExprSwitchStmt | TypeSwitchStmt .
// ExprSwitchStmt = "switch" [ SimpleStmt ";" ] [ Expression ] "{" { ExprCaseClause } "}".
// TypeSwitchStmt = "switch" [ SimpleStmt ";" ] TypeSwitchGuard "{" {TypeCaseClause} "}" .
// TypeSwitchGuard = [ identifier ":=" ] PrimaryExpr "." "(" "type" ")" .
func (p *parser) parseSwitchStmt() ast.Stmt {
	p.match(s.SWITCH)

	b := p.brackets
	p.brackets = 0
	defer func() { p.brackets = b }()

	if p.token == '{' {
		return p.parseExprSwitchStmt(nil, nil)
	}

	x := p.parseExpr()
	if p.token == '{' {
		if y, ok := isTypeSwitchGuardExpr(x); ok {
			return p.parseTypeSwitchStmt(nil, "", y)
		} else {
			return p.parseExprSwitchStmt(nil, x)
		}
	}

	init := p.parseSimpleStmt(x)
	if p.token == '{' {
		if id, y, ok := p.isTypeSwitchGuardStmt(init); ok {
			return p.parseTypeSwitchStmt(nil, id, y)
		}
	}
	p.match(';')

	if p.token == '{' {
		return p.parseExprSwitchStmt(init, nil)
	}

	x = p.parseExpr()
	if p.token == '{' {
		return p.parseExprSwitchStmt(init, x)
	}

	stmt := p.parseSimpleStmt(x)
	if id, y, ok := p.isTypeSwitchGuardStmt(stmt); ok {
		return p.parseTypeSwitchStmt(init, id, y)
	}

	p.error("invalid switch expression")
	p.sync('{')
	return p.parseExprSwitchStmt(init, nil)
}

// ExprCaseClause = ExprSwitchCase ":" StatementList .
// ExprSwitchCase = "case" ExpressionList | "default" .
func (p *parser) parseExprSwitchStmt(init ast.Stmt, ex ast.Expr) ast.Stmt {
	var (
		cs  []ast.ExprCaseClause
		def bool = false
	)
	p.match('{')
	for p.token == s.CASE || p.token == s.DEFAULT {
		t := p.token
		p.next()
		var xs []ast.Expr
		if t == s.CASE {
			xs = p.parseExprList(nil)
		} else {
			if def {
				p.error("multiple defaults in switch")
			}
			def = true
		}
		p.match(':')
		ss := p.parseStatementList()
		cs = append(cs, ast.ExprCaseClause{xs, ss})
	}
	p.match('}')
	return &ast.ExprSwitchStmt{init, ex, cs}
}

// TypeCaseClause  = TypeSwitchCase ":" StatementList .
// TypeSwitchCase  = "case" TypeList | "default" .
func (p *parser) parseTypeSwitchStmt(init ast.Stmt, id string, ex ast.Expr) ast.Stmt {
	def := false
	var cs []ast.TypeCaseClause
	p.match('{')
	for p.token == s.CASE || p.token == s.DEFAULT {
		t := p.token
		p.next()
		var ts []ast.TypeSpec
		if t == s.CASE {
			ts = p.parseTypeList()
		} else {
			if def {
				p.error("multiple defaults in switch")
			}
			def = true
		}
		p.match(':')
		ss := p.parseStatementList()
		cs = append(cs, ast.TypeCaseClause{ts, ss})
	}
	p.match('}')
	return &ast.TypeSwitchStmt{init, id, ex, cs}
}

// TypeList = Type { "," Type } .
func (p *parser) parseTypeList() (ts []ast.TypeSpec) {
	for {
		ts = append(ts, p.parseType())
		if p.token != ',' {
			break
		}
		p.next()
	}
	return
}

// SelectStmt = "select" "{" { CommClause } "}" .
// CommClause = CommCase ":" StatementList .
// CommCase   = "case" ( SendStmt | RecvStmt ) | "default" .
// SendStmt = Channel "<-" Expression .
// RecvStmt   = [ ExpressionList "=" | IdentifierList ":=" ] RecvExpr .
// RecvExpr   = Expression .
func (p *parser) parseSelectStmt() ast.Stmt {
	def := false
	var cs []ast.CommClause
	p.match(s.SELECT)
	p.match('{')
	for p.token == s.CASE || p.token == s.DEFAULT {
		t := p.token
		p.next()
		var comm ast.Stmt
		if t == s.CASE {
			comm = p.parseSimpleStmt(p.parseExpr())
		} else {
			if def {
				p.error("multiple defaults in select")
			}
			def = true
		}
		p.match(':')
		ss := p.parseStatementList()
		cs = append(cs, ast.CommClause{comm, ss})
	}
	p.match('}')
	return &ast.SelectStmt{cs}
}

// ForStmt = "for" [ Condition | ForClause | RangeClause ] Block .
// Condition = Expression .
// ForClause = [ InitStmt ] ";" [ Condition ] ";" [ PostStmt ] .
// InitStmt = SimpleStmt .
// PostStmt = SimpleStmt .
// RangeClause = [ ExpressionList "=" | IdentifierList ":=" ] "range" Expression .
func (p *parser) parseForStmt() ast.Stmt {
	var (
		init ast.Stmt
		cond ast.Expr
		post ast.Stmt
		body *ast.Block
	)
	p.match(s.FOR)
	b := p.brackets
	p.brackets = 0
	defer func() { p.brackets = b }()

	if p.token == '{' {
		// infinite loop: "for { ..."
		return &ast.ForStmt{nil, nil, nil, p.parseBlock()}
	}

	if p.token == s.RANGE {
		// range for: "for range ex { ..."
		p.next()
		e := p.parseExpr()
		body = p.parseBlock()
		return &ast.ForRangeStmt{'=', nil, e, body}
	}

	if p.token == ';' {
		// ordinary for : "for ; [cond] ; [post] { ..."
		p.next()
		if p.token != ';' {
			cond = p.parseExpr()
		}
		p.match(';')
		if p.token != '{' {
			e := p.parseExpr()
			post = p.parseSimpleStmt(e)
		}
		body = p.parseBlock()
		return &ast.ForStmt{nil, cond, post, body}
	}

	e := p.parseExpr()
	if p.token == '{' {
		// "while" loop "for cond { ..."
		cond = e
		body = p.parseBlock()
		return &ast.ForStmt{nil, cond, nil, body}
	}

	init = p.parseSimpleStmtOrRange(e)
	if r, ok := init.(*ast.ForRangeStmt); ok {
		// range for
		r.Body = p.parseBlock()
		return r
	}

	// ordinary for: "for init ; [cond] ; [post] { ..."
	p.match(';')
	if p.token != ';' {
		cond = p.parseExpr()
	}
	p.match(';')
	if p.token != '{' {
		e := p.parseExpr()
		post = p.parseSimpleStmt(e)
	}
	body = p.parseBlock()
	return &ast.ForStmt{init, cond, post, body}
}

// DeferStmt = "defer" Expression .
func (p *parser) parseDeferStmt() ast.Stmt {
	p.match(s.DEFER)
	return &ast.DeferStmt{p.parseExpr()}
}

// SimpleStmt =
//     EmptyStmt | ExpressionStmt | SendStmt | IncDecStmt | Assignment |
//     ShortVarDecl .
func (p *parser) parseSimpleStmt(e ast.Expr) ast.Stmt {
	switch p.token {
	case s.RECV:
		return p.parseSendStmt(e)
	case s.INC, s.DEC:
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
	if t == s.DEFINE {
		p.next()
		return p.parseShortVarDecl(es)
	} else if isAssignOp(t) {
		p.next()
		return p.parseAssignment(t, es)
	}

	p.error("Invalid statement")
	return &ast.Error{}
}

// Parses a SimpleStmt or a RangeClause. Used by `parseForStmt` to
// disambiguate between a range for and an ordinary for.
func (p *parser) parseSimpleStmtOrRange(e ast.Expr) ast.Stmt {
	switch p.token {
	case s.RECV:
		return p.parseSendStmt(e)
	case s.INC, s.DEC:
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
	if t == s.DEFINE || isAssignOp(t) {
		p.next()
		if p.token == s.RANGE {
			return p.parseRangeClause(t, es)
		}
		if t == s.DEFINE {
			return p.parseShortVarDecl(es)
		} else {
			return p.parseAssignment(t, es)
		}
	}
	p.error("Invalid statement")
	return &ast.Error{}
}

// SendStmt = Channel "<-" Expression .
// Channel  = Expression .
func (p *parser) parseSendStmt(ch ast.Expr) ast.Stmt {
	p.match(s.RECV)
	return &ast.SendStmt{ch, p.parseExpr()}
}

// IncDecStmt = Expression ( "++" | "--" ) .
func (p *parser) parseIncDecStmt(e ast.Expr) ast.Stmt {
	if p.token == s.INC {
		p.match(s.INC)
		return &ast.IncStmt{e}
	} else {
		p.match(s.DEC)
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
	return &ast.AssignStmt{s.DEFINE, lhs, rhs}
}

// RangeClause = [ ExpressionList "=" | IdentifierList ":=" ] "range" Expression .
func (p *parser) parseRangeClause(op uint, lhs []ast.Expr) *ast.ForRangeStmt {
	p.match(s.RANGE)
	return &ast.ForRangeStmt{op, lhs, p.parseExpr(), nil}
}
