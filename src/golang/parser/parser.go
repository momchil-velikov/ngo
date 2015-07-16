package parser

import (
	"fmt"
	"golang/ast"
	s "golang/scanner"
)

type parser struct {
	errors []error
	scan   s.Scanner

	token uint // current token
}

func (p *parser) init(name string, src string) {
	p.scan.Init(name, src)
}

// Parse a source file
func Parse(name string, src string) (*ast.File, error) {
	var p parser
	p.init(name, src)

	p.next()
	f := p.parseFile()
	if p.errors == nil {
		return f, nil
	} else {
		return f, ErrorList(p.errors)
	}
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

// Advance to the next token iff the current one is TOKEN.
func (p *parser) matchValue(token uint) string {
	if p.expect(token) {
		value := p.scan.Value
		p.next()
		return value
	} else {
		return ""
	}
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

// Skip tokens, until either T1, T2 or T3 token is found. Report an error
// only if some tokens were skipped. Consume T1.
func (p *parser) sync3(t1, t2, t3 uint) {
	if p.token != t1 {
		p.expectError(t1, p.token)
	}
	for p.token != s.EOF && p.token != t1 && p.token != t2 && p.token != t3 {
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
	return p.matchValue(s.ID)
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
func (p *parser) parseImportSpec() (name string, path string) {
	if p.token == '.' {
		name = "."
		p.next()
	} else if p.token == s.ID {
		name = p.matchValue(s.ID)
	} else {
		name = ""
	}
	path = p.matchValue(s.STRING)
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
	id := p.matchValue(s.ID)
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
	pkg := p.matchValue(s.ID)
	var id string
	if p.token == '.' {
		p.next()
		id = p.matchValue(s.ID)
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
		pkg := p.matchValue(s.ID)
		if p.token == '.' {
			// If the field decl begins with a qualified-id, it's parsed as an
			// anonymous field.
			p.next()
			id := p.matchValue(s.ID)
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

func (p *parser) parseTagOpt() (tag string) {
	if p.token == s.STRING {
		tag = p.matchValue(s.STRING)
	} else {
		tag = ""
	}
	return
}

// IdentifierList = identifier { "," identifier } .
func (p *parser) parseIdList(id string) (ids []string) {
	if len(id) == 0 {
		id = p.matchValue(s.ID)
	}
	if len(id) > 0 {
		ids = append(ids, id)
	}
	for p.token == ',' {
		p.next()
		id = p.matchValue(s.ID)
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
func (p *parser) parseParameters() (ds []*ast.ParamDecl) {
	p.match('(')
	if p.token == ')' {
		p.next()
		return nil
	}
	for p.token != s.EOF && p.token != ')' && p.token != ';' {
		d := p.parseParamDecl()
		ds = append(ds, d...)
		if p.token != ')' {
			p.sync3(',', ')', ';')
		}
	}
	p.match(')')
	return ds
}

// Parse an identifier or a type.  Return a param decl with filled either the
// parameter name or the parameter type, depending upon what was parsed.
func (p *parser) parseIdOrType() *ast.ParamDecl {
	if p.token == s.ID {
		id := p.matchValue(s.ID)
		if p.token == '.' {
			p.next()
			name := p.matchValue(s.ID)
			return &ast.ParamDecl{Type: &ast.QualId{id, name}}
			// Fallthrough as if the dot wasn't there.
		}
		return &ast.ParamDecl{Name: id}
	} else {
		v := false
		if p.token == s.DOTS {
			p.next()
			v = true
		}
		t := p.parseType() // FIXME: parenthesized types not allowed
		return &ast.ParamDecl{Type: t, Variadic: v}
	}
}

// Parse a list of (qualified) identifiers or types.
func (p *parser) parseIdOrTypeList() (ps []*ast.ParamDecl) {
	for {
		ps = append(ps, p.parseIdOrType())
		if p.token != ',' {
			break
		}
		p.next()
	}
	return ps
}

// ParameterDecl = [ IdentifierList ] [ "..." ] Type .
func (p *parser) parseParamDecl() (ps []*ast.ParamDecl) {
	ps = p.parseIdOrTypeList()
	if p.token == ')' || /* missing closing paren */ p.token == ';' || p.token == '{' {
		// No type follows, then all the decls must be types.
		for _, d := range ps {
			if d.Type == nil {
				d.Type = &ast.QualId{"", d.Name}
				d.Name = ""
			}
		}
		return ps
	}
	// Otherwise, all the list elements must be identifiers, followed by a type.
	v := false
	if p.token == s.DOTS {
		p.next()
		v = true
	}
	t := p.parseType()
	for _, d := range ps {
		if d.Type == nil {
			d.Type = t
			d.Variadic = v
		} else {
			p.error("Invalid parameter list") // FIXME: more specific message
		}
	}
	return ps
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
		id := p.matchValue(s.ID)
		if p.token == '.' {
			p.next()
			name := p.matchValue(s.ID)
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
		es = p.parseExprList()
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
		es = p.parseExprList()
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
	name := p.matchValue(s.ID)
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
		n = p.matchValue(s.ID)
	}
	ptr := false
	if p.token == '*' {
		p.next()
		ptr = true
	}
	t := p.matchValue(s.ID)
	p.sync2(')', ';')
	var tp ast.TypeSpec = &ast.QualId{Id: t}
	if ptr {
		tp = &ast.PtrType{Base: tp}
	}
	return &ast.Receiver{Name: n, Type: tp}
}

func (p *parser) parseBlock() *ast.Block {
	p.match('{')
	p.match('}')
	return &ast.Block{}
}

// ExpressionList = Expression { "," Expression } .
func (p *parser) parseExprList() (es []ast.Expr) {
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
	return p.parseOrExprOrType()
}

func (p *parser) parseExpr() ast.Expr {
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

// UnaryExpr  = PrimaryExpr | unary_op UnaryExpr .
// unary_op   = "+" | "-" | "!" | "^" | "*" | "&" | "<-" .
func (p *parser) parseUnaryExprOrType() (ast.Expr, ast.TypeSpec) {
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
	var (
		ex ast.Expr
		t  ast.TypeSpec
	)
	// Handle initial Operand, Conversion or a BuiltinCall
	switch p.token {
	case s.ID: // CompositeLit, MethodExpr, Conversion, BuiltinCall, OperandName
		id := p.matchValue(s.ID)
		if p.token == '{' {
			ex = p.parseCompositeLiteral(&ast.QualId{Id: id})
		} else if p.token == '(' {
			ex = p.parseCall(&ast.QualId{Id: id})
		} else {
			ex = &ast.QualId{Id: id}
		}
	case '(': // Expression, MethodExpr, Conversion
		p.next()
		ex, t = p.parseExprOrType()
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
		v := p.matchValue(k)
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
				t = p.parseType()
				ex = &ast.TypeAssertion{Type: t, Arg: ex}
			} else {
				// PrimaryExpr Selector
				// Selector = "." identifier .
				id := p.matchValue(s.ID)
				ex = &ast.Selector{Arg: ex, Id: id}
			}
		case '[':
			// PrimaryExpr Index
			// PrimaryExpr Slice
			ex = p.parseIndexOrSlice(ex)
		case '(':
			// PrimaryExpr Call
			ex = p.parseCall(ex)
		case '{':
			// Composite literal
			typ, ok := p.needType(ex)
			if !ok {
				p.error("invalid type for composite literal")
			}
			ex = p.parseCompositeLiteral(typ)
		default:
			return ex, nil
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
	a := p.parseExpr()
	if p.token == ',' {
		p.next()
	}
	p.match(')')
	return &ast.Conversion{Type: typ, Arg: a}
}

// Call           = "(" [ ArgumentList [ "," ] ] ")" .
// ArgumentList   = ExpressionList [ "..." ] .
// BuiltinCall = identifier "(" [ BuiltinArgs [ "," ] ] ")" .
// BuiltinArgs = Type [ "," ArgumentList ] | ArgumentList .
func (p *parser) parseCall(f ast.Expr) ast.Expr {
	p.match('(')
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
	p.match('{')
	p.match('}')
	return &ast.FuncLiteral{Sig: sig, Body: &ast.Block{}}
}

// Index          = "[" Expression "]" .
// Slice          = "[" ( [ Expression ] ":" [ Expression ] ) |
//                      ( [ Expression ] ":" Expression ":" Expression )
//                  "]" .
func (p *parser) parseIndexOrSlice(e ast.Expr) ast.Expr {
	p.match('[')
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
