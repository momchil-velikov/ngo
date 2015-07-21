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

func (p *parser) setBrackets(n int) int {
	o := p.brackets
	p.brackets = n
	return o
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
	is := p.parseImportDecls()

	// Parse toplevel declarations.
	ds := p.parseToplevelDecls()

	p.match(s.EOF)

	return &ast.File{Package: name, Imports: is, Decls: ds}
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
					imports = append(imports, ast.Import{Name: name, Path: path})
				}
				if p.token != ')' {
					p.sync2(';', ')')
				}
			}
			p.match(')')
		} else {
			if name, path := p.parseImportSpec(); len(path) > 0 {
				imports = append(imports, ast.Import{Name: name, Path: path})
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
		var ts []*ast.TypeDecl
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
func (p *parser) parseType() ast.Type {
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
			return &ast.SliceType{Elt: p.parseType()}
		} else if p.token == s.DOTS {
			p.next()
			p.match(']')
			return &ast.ArrayType{Elt: p.parseType()}
		} else {
			e := p.parseExpr()
			p.match(']')
			t := p.parseType()
			return &ast.ArrayType{Dim: e, Elt: t}
		}

	// PointerType = "*" BaseType .
	// BaseType = Type .
	case '*':
		p.next()
		return &ast.PtrType{Base: p.parseType()}

	// MapType     = "map" "[" KeyType "]" ElementType .
	// KeyType     = Type .
	case s.MAP:
		p.next()
		p.match('[')
		k := p.parseType()
		p.match(']')
		t := p.parseType()
		return &ast.MapType{Key: k, Elt: t}

	// ChannelType = ( "chan" [ "<-" ] | "<-" "chan" ) ElementType .
	case s.RECV:
		p.next()
		p.match(s.CHAN)
		t := p.parseType()
		return &ast.ChanType{Send: false, Recv: true, Elt: t}
	case s.CHAN:
		p.next()
		send, recv := true, true
		if p.token == s.RECV {
			p.next()
			send, recv = true, false
		}
		t := p.parseType()
		return &ast.ChanType{Send: send, Recv: recv, Elt: t}

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
	return &ast.QualId{Pkg: pkg, Id: id}
}

// StructType     = "struct" "{" { FieldDecl ";" } "}" .
func (p *parser) parseStructType() *ast.StructType {
	var fs []*ast.FieldDecl
	p.match(s.STRUCT)
	p.match('{')
	for p.token != s.EOF && p.token != '}' {
		fs = append(fs, p.parseFieldDecl())
		if p.token != '}' {
			p.sync2(';', '}')
		}
	}
	p.match('}')
	return &ast.StructType{Fields: fs}
}

// FieldDecl      = (IdentifierList Type | AnonymousField) [ Tag ] .
// AnonymousField = [ "*" ] TypeName .
// Tag            = string_lit .
func (p *parser) parseFieldDecl() *ast.FieldDecl {
	if p.token == '*' {
		// Anonymous field.
		p.next()
		pt := &ast.PtrType{Base: p.parseQualId()}
		tag := p.parseTagOpt()
		return &ast.FieldDecl{Type: pt, Tag: tag}
	} else if p.token == s.ID {
		pkg := p.matchString(s.ID)
		if p.token == '.' {
			// If the field decl begins with a qualified-id, it's parsed as an
			// anonymous field.
			p.next()
			id := p.matchString(s.ID)
			t := &ast.QualId{Pkg: pkg, Id: id}
			tag := p.parseTagOpt()
			return &ast.FieldDecl{Type: t, Tag: tag}
		} else if p.token == s.STRING || p.token == ';' || p.token == '}' {
			// If it's only a single identifier, with no separate type
			// declaration, it's also an anonymous filed.
			t := &ast.QualId{Id: pkg}
			tag := p.parseTagOpt()
			return &ast.FieldDecl{Type: t, Tag: tag}
		} else {
			ids := p.parseIdList(pkg)
			t := p.parseType()
			tag := p.parseTagOpt()
			return &ast.FieldDecl{Names: ids, Type: t, Tag: tag}
		}
	}
	p.error("Invalid field declaration")
	return &ast.FieldDecl{Type: &ast.Error{}}
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
	var rs []*ast.ParamDecl
	if p.token == '(' {
		rs = p.parseParameters()
	} else if isTypeLookahead(p.token) {
		rs = []*ast.ParamDecl{&ast.ParamDecl{Type: p.parseType()}}
	}
	return &ast.FuncType{Params: ps, Returns: rs}
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
			ds = append(ds, &ast.ParamDecl{Type: t, Var: v})
			ids = nil
		} else {
			ids = append(ids, id)
			if p.token != ',' && p.token != ')' {
				id, t, v = p.parseIdOrType()
				if t == nil {
					t = &ast.QualId{Id: id}
				}
				ds = append(ds, &ast.ParamDecl{Names: ids, Type: t, Var: v})
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
func (p *parser) parseIdOrType() (string, ast.Type, bool) {
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
	for _, id := range ids {
		ds = append(ds, &ast.ParamDecl{Type: &ast.QualId{Id: id}})
	}
	return ds
}

// InterfaceType      = "interface" "{" { MethodSpec ";" } "}" .
// MethodSpec         = MethodName Signature | InterfaceTypeName .
// MethodName         = identifier .
// InterfaceTypeName  = TypeName .
func (p *parser) parseInterfaceType() *ast.InterfaceType {
	var (
		emb  []*ast.QualId
		meth []*ast.MethodSpec
	)
	p.match(s.INTERFACE)
	p.match('{')
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
			emb = append(emb, &ast.QualId{Id: id})
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
		var cs []*ast.ConstDecl
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
		t  ast.Type
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
		var vs []*ast.VarDecl
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
		t  ast.Type
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
	var blk *ast.Block
	if p.token == '{' {
		blk = p.parseBlock()
	}
	return &ast.FuncDecl{Name: name, Recv: r, Sig: sig, Blk: blk}
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
	var tp ast.Type = &ast.QualId{Id: t}
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
	return &ast.Error{}
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
	for p.token == s.OR {
		p.next()
		y, _ := p.parseAndExprOrType()
		x = &ast.BinaryExpr{Op: s.OR, X: x, Y: y}
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
	for p.token == s.AND {
		p.next()
		y, _ := p.parseCompareExprOrType()
		x = &ast.BinaryExpr{Op: s.AND, X: x, Y: y}
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
func (p *parser) parseUnaryExprOrType() (ast.Expr, ast.Type) {
	//	defer p.trace("UnaryExprOrType")()

	switch p.token {
	case '+', '-', '!', '^', '&':
		op := p.token
		p.next()
		x, _ := p.parseUnaryExprOrType()
		return &ast.UnaryExpr{Op: op, X: x}, nil
	case '*':
		p.next()
		if x, t := p.parseUnaryExprOrType(); t == nil {
			return &ast.UnaryExpr{Op: '*', X: x}, nil
		} else {
			return nil, &ast.PtrType{Base: t}
		}
	case s.RECV:
		p.next()
		if x, t := p.parseUnaryExprOrType(); t == nil {
			return &ast.UnaryExpr{Op: s.RECV, X: x}, nil
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

func (p *parser) needType(x ast.Expr) (ast.Type, bool) {
	switch t := x.(type) {
	case *ast.QualId:
		return t, true
	case *ast.Selector:
		if x, ok := t.X.(*ast.QualId); ok && len(x.Pkg) == 0 {
			return &ast.QualId{Pkg: x.Id, Id: t.Id}, true
		}
	}
	return nil, false
}

// TypeAssertion = "." "(" Type ")" .
// Also allow "." "(" "type" ")".
func (p *parser) parseTypeAssertion(x ast.Expr) ast.Expr {
	p.match('(')
	var t ast.Type
	if p.token == s.TYPE {
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
//     BuiltinCall |
//     PrimaryExpr Selector |
//     PrimaryExpr Index |
//     PrimaryExpr Slice |
//     PrimaryExpr TypeAssertion |
//     PrimaryExpr Call .

func (p *parser) parsePrimaryExprOrType() (ast.Expr, ast.Type) {
	// defer p.trace("PrimaryExprOrType", s.TokenNames[p.token])()

	var (
		x ast.Expr
		t ast.Type
	)
	// Handle initial Operand, Conversion or a BuiltinCall
	switch p.token {
	case s.ID: // CompositeLit, MethodExpr, Conversion, BuiltinCall, OperandName
		id := p.matchString(s.ID)
		if p.token == '{' && p.brackets > 0 {
			x = p.parseCompositeLiteral(&ast.QualId{Id: id})
		} else if p.token == '(' {
			x = p.parseCall(&ast.QualId{Id: id})
		} else {
			x = &ast.QualId{Id: id}
		}
	case '(': // Expression, MethodExpr, Conversion
		p.next()
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
			x = &ast.ParensExpr{X: x}
		}
	case '[', s.STRUCT, s.MAP: // Conversion, CompositeLit
		t = p.parseType()
		if p.token == '(' {
			x = p.parseConversion(t)
		} else if p.token == '{' {
			x = p.parseCompositeLiteral(t)
		} else {
			return nil, t
		}
	case s.FUNC: // Conversion, FunctionLiteral
		t = p.parseType()
		if p.token == '(' {
			x = p.parseConversion(t)
		} else if p.token == '{' {
			x = p.parseFuncLiteral(t)
		} else {
			return nil, t
		}
	case '*', s.RECV:
		panic("should not reach here")
	case s.INTERFACE, s.CHAN: // Conversion
		t = p.parseType()
		if p.token == '(' {
			x = p.parseConversion(t)
		} else {
			return nil, t
		}
	case s.INTEGER, s.FLOAT, s.IMAGINARY, s.RUNE, s.STRING: // BasicLiteral
		k := p.token
		v := p.matchRaw(k)
		return &ast.Literal{Kind: k, Value: v}, nil
	default:
		p.error("token cannot start neither expression nor type")
		x = &ast.Error{}
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
				id := p.matchString(s.ID)
				x = &ast.Selector{X: x, Id: id}
			}
		case '[':
			// PrimaryExpr Index
			// PrimaryExpr Slice
			x = p.parseIndexOrSlice(x)
		case '(':
			// PrimaryExpr Call
			x = p.parseCall(x)
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
			return &ast.Element{Value: k}
		}
		p.match(':')
	}
	if p.token == '{' {
		elts := p.parseLiteralValue()
		e := &ast.CompLiteral{Elts: elts}
		return &ast.Element{Key: k, Value: e}
	} else {
		e := p.parseExpr()
		return &ast.Element{Key: k, Value: e}
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
	var xs []ast.Expr
	x, t := p.parseExprOrType()
	if x != nil {
		xs = append(xs, x)
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
		xs = append(xs, p.parseExpr())
		if p.token == s.DOTS {
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
	return &ast.FuncLiteral{Sig: sig, Blk: b}
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
	st := p.parseStatementList()
	p.match('}')
	return &ast.Block{Body: st}
}

// StatementList = { Statement ";" } .
func (p *parser) parseStatementList() []ast.Stmt {
	var st []ast.Stmt
	b := p.setBrackets(1)
	for p.token != s.EOF && p.token != '}' && p.token != s.CASE && p.token != s.DEFAULT {
		st = append(st, p.parseStmt())
		p.syncEndStatement()
	}
	p.setBrackets(b)
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
	x := p.parseExpr()
	return &ast.GoStmt{X: x}
}

// ReturnStmt = "return" [ ExpressionList ] .
func (p *parser) parseReturnStmt() ast.Stmt {
	p.match(s.RETURN)
	if p.token != ';' && p.token != '}' {
		return &ast.ReturnStmt{Xs: p.parseExprList(nil)}
	} else {
		return &ast.ReturnStmt{}
	}
}

// BreakStmt = "break" [ Label ] .
func (p *parser) parseBreakStmt() ast.Stmt {
	p.match(s.BREAK)
	if p.token == s.ID {
		return &ast.BreakStmt{Label: p.matchString(s.ID)}
	} else {
		return &ast.BreakStmt{}
	}
}

// ContinueStmt = "continue" [ Label ] .
func (p *parser) parseContinueStmt() ast.Stmt {
	p.match(s.CONTINUE)
	if p.token == s.ID {
		return &ast.ContinueStmt{Label: p.matchString(s.ID)}
	} else {
		return &ast.ContinueStmt{}
	}
}

// GotoStmt = "goto" Label .
func (p *parser) parseGotoStmt() ast.Stmt {
	p.match(s.GOTO)
	return &ast.GotoStmt{Label: p.matchString(s.ID)}
}

// FallthroughStmt = "fallthrough" .
func (p *parser) parseFallthroughStmt() ast.Stmt {
	p.match(s.FALLTHROUGH)
	return &ast.FallthroughStmt{}
}

// IfStmt = "if" [ SimpleStmt ";" ] Expression Block [ "else" ( IfStmt | Block ) ] .
func (p *parser) parseIfStmt() ast.Stmt {
	p.match(s.IF)
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
	if p.token == s.ELSE {
		var els ast.Stmt
		p.next()
		if p.token == s.IF {
			els = p.parseIfStmt()
		} else {
			els = p.parseBlock()
		}
		return &ast.IfStmt{Init: init, Cond: x, Then: then, Else: els}
	} else {
		return &ast.IfStmt{Init: init, Cond: x, Then: then}
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

	b := p.setBrackets(0)
	defer p.setBrackets(b)

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
func (p *parser) parseExprSwitchStmt(init ast.Stmt, x ast.Expr) ast.Stmt {
	var (
		cs  []ast.ExprCaseClause
		def bool
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
		cs = append(cs, ast.ExprCaseClause{Xs: xs, Body: p.parseStatementList()})
	}
	p.match('}')
	return &ast.ExprSwitchStmt{Init: init, X: x, Cases: cs}
}

// TypeCaseClause  = TypeSwitchCase ":" StatementList .
// TypeSwitchCase  = "case" TypeList | "default" .
func (p *parser) parseTypeSwitchStmt(init ast.Stmt, id string, x ast.Expr) ast.Stmt {
	def := false
	var cs []ast.TypeCaseClause
	p.match('{')
	for p.token == s.CASE || p.token == s.DEFAULT {
		t := p.token
		p.next()
		var ts []ast.Type
		if t == s.CASE {
			ts = p.parseTypeList()
		} else {
			if def {
				p.error("multiple defaults in switch")
			}
			def = true
		}
		p.match(':')
		cs = append(cs, ast.TypeCaseClause{Types: ts, Body: p.parseStatementList()})
	}
	p.match('}')
	return &ast.TypeSwitchStmt{Init: init, Id: id, X: x, Cases: cs}
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
		var c ast.Stmt
		if t == s.CASE {
			c = p.parseSimpleStmt(p.parseExpr())
		} else {
			if def {
				p.error("multiple defaults in select")
			}
			def = true
		}
		p.match(':')
		cs = append(cs, ast.CommClause{Comm: c, Body: p.parseStatementList()})
	}
	p.match('}')
	return &ast.SelectStmt{Comms: cs}
}

// ForStmt = "for" [ Condition | ForClause | RangeClause ] Block .
// Condition = Expression .
// ForClause = [ InitStmt ] ";" [ Condition ] ";" [ PostStmt ] .
// InitStmt = SimpleStmt .
// PostStmt = SimpleStmt .
// RangeClause = [ ExpressionList "=" | IdentifierList ":=" ] "range" Expression .
func (p *parser) parseForStmt() ast.Stmt {
	p.match(s.FOR)
	b := p.setBrackets(0)
	defer p.setBrackets(b)

	if p.token == '{' {
		// infinite loop: "for { ..."
		return &ast.ForStmt{Blk: p.parseBlock()}
	}

	if p.token == s.RANGE {
		// range for: "for range ex { ..."
		p.next()
		return &ast.ForRangeStmt{Op: '=', Range: p.parseExpr(), Blk: p.parseBlock()}
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
		return &ast.ForStmt{Cond: x, Post: post, Blk: p.parseBlock()}
	}

	x := p.parseExpr()
	if p.token == '{' {
		// "while" loop "for cond { ..."
		return &ast.ForStmt{Cond: x, Blk: p.parseBlock()}
	}

	init := p.parseSimpleStmtOrRange(x)
	x = nil
	if r, ok := init.(*ast.ForRangeStmt); ok {
		// range for
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
	return &ast.ForStmt{Init: init, Cond: x, Post: post, Blk: p.parseBlock()}
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
