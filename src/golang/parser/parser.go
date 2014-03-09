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
    f := p.parse_file()
    if p.errors == nil {
        return f, nil
    } else {
        return f, ErrorList(p.errors)
    }
}

// Append an error message to the parser error messages list
func (p *parser) error(msg string) {
    e := parse_error{p.scan.Name, p.scan.TLine, p.scan.TPos, msg}
    p.errors = append(p.errors, e)
}

// Emit an expected token mismatch error.
func (p *parser) expect_error(exp, act uint) {
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
        p.expect_error(token, p.token)
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
func (p *parser) match_value(token uint) string {
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
        p.expect_error(token, p.token)
    }
    for p.token != s.EOF && p.token != token {
        p.next()
    }
    p.next()
}

// Skip tokens, until either T1 or T2 token is found. Report an error
// only if some tokens were skipped.
func (p *parser) sync2(t1, t2 uint) {
    if p.token != t1 {
        p.expect_error(t1, p.token)
    }
    for p.token != s.EOF && p.token != t1 && p.token != t2 {
        p.next()
    }
    if p.token == t1 {
        p.next()
    }
}

// Skip tokens, until beginning of a toplevel declaration found
func (p *parser) sync_decl() {
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
func (p *parser) parse_file() *ast.File {
    // Parse package name
    package_name := p.parse_package_clause()
    if len(package_name) == 0 {
        return nil
    }

    if !p.match(';') {
        return nil
    }

    // Parse import declaration(s)
    imports := p.parse_import_decls()

    // Parse toplevel declarations.
    decls := p.parse_toplevel_decls()

    p.match(s.EOF)

    return &ast.File{package_name, imports, decls}
}

// Parse a package clause. Return the package name or an empty string on error.
//
// PackageClause  = "package" PackageName .
// PackageName    = identifier .
func (p *parser) parse_package_clause() string {
    p.match(s.PACKAGE)
    return p.match_value(s.ID)
}

// Parse import declarations(s).
//
// ImportDecl       = "import" ( ImportSpec | "(" { ImportSpec ";" } ")" ) .
func (p *parser) parse_import_decls() (imports []ast.Import) {
    for p.token == s.IMPORT {
        p.match(s.IMPORT)
        if p.token == '(' {
            p.next()
            for p.token != s.EOF && p.token != ')' {
                if name, path := p.parse_import_spec(); len(path) > 0 {
                    imports = append(imports, ast.Import{name, path})
                }
                if p.token != ')' {
                    // Spec does not say a semicolon here is optional, Google Go
                    // allows it to be missing.
                    p.sync(';')
                }
            }
            p.match(')')
        } else {
            if name, path := p.parse_import_spec(); len(path) > 0 {
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
func (p *parser) parse_import_spec() (name string, path string) {
    if p.token == '.' {
        name = "."
        p.next()
    } else if p.token == s.ID {
        name = p.match_value(s.ID)
    } else {
        name = ""
    }
    path = p.match_value(s.STRING)
    return
}

// Parse toplevel declaration(s)
//
// Declaration   = ConstDecl | TypeDecl | VarDecl .
// TopLevelDecl  = Declaration | FunctionDecl | MethodDecl .
func (p *parser) parse_toplevel_decls() (dcls []ast.Decl) {
    for {
        var f func() ast.Decl
        switch p.token {
        case s.TYPE:
            f = p.parse_type_decl
        case s.CONST:
            f = p.parse_const_decl
        case s.VAR:
            f = p.parse_var_decl
        case s.FUNC:
            f = p.parse_func_decl
        case s.EOF:
            return
        default:
            p.error("expected type, const, bar or func/method declaration")
            p.sync_decl()
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
func (p *parser) parse_type_decl() ast.Decl {
    p.match(s.TYPE)
    if p.token == '(' {
        p.next()
        var ts []*ast.TypeDecl = nil
        for p.token != s.EOF && p.token != ')' {
            ts = append(ts, p.parse_type_spec())
            if p.token != ')' {
                p.sync2(';', ')')
            }
        }
        p.match(')')
        return &ast.TypeGroup{Decls: ts}
    } else {
        return p.parse_type_spec()
    }
}

// TypeSpec = identifier Type .
func (p *parser) parse_type_spec() *ast.TypeDecl {
    id := p.match_value(s.ID)
    t := p.parse_type()
    return &ast.TypeDecl{Name: id, Type: t}
}

// Determine if the given TOKEN could be a beginning of a typespec.
func is_type_lookahead(token uint) bool {
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
func (p *parser) parse_type() ast.TypeSpec {
    switch p.token {
    case s.ID:
        return p.parse_qual_id()

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
            return &ast.SliceType{p.parse_type()}
        } else if p.token == s.DOTS {
            p.next()
            p.match(']')
            return &ast.ArrayType{Dim: nil, EltType: p.parse_type()}
        } else {
            e := p.parse_expr()
            p.match(']')
            t := p.parse_type()
            return &ast.ArrayType{Dim: e, EltType: t}
        }

    // PointerType = "*" BaseType .
    // BaseType = Type .
    case '*':
        p.next()
        return &ast.PtrType{p.parse_type()}

    // MapType     = "map" "[" KeyType "]" ElementType .
    // KeyType     = Type .
    case s.MAP:
        p.next()
        p.match('[')
        k := p.parse_type()
        p.match(']')
        t := p.parse_type()
        return &ast.MapType{k, t}

    // ChannelType = ( "chan" [ "<-" ] | "<-" "chan" ) ElementType .
    case s.RECV:
        p.next()
        p.match(s.CHAN)
        t := p.parse_type()
        return &ast.ChanType{Send: false, Recv: true, EltType: t}
    case s.CHAN:
        p.next()
        send, recv := true, true
        if p.token == s.RECV {
            p.next()
            send, recv = true, false
        }
        t := p.parse_type()
        return &ast.ChanType{Send: send, Recv: recv, EltType: t}

    case s.STRUCT:
        return p.parse_struct_type()

    case s.FUNC:
        return p.parse_func_type()

    case s.INTERFACE:
        return p.parse_interface_type()

    case '(':
        p.next()
        t := p.parse_type()
        p.match(')')
        return t

    default:
        p.error("expected typespec")
        return &ast.Error{}
    }
}

// TypeName = identifier | QualifiedIdent .
func (p *parser) parse_qual_id() *ast.QualId {
    pkg := p.match_value(s.ID)
    var id string
    if p.token == '.' {
        p.next()
        id = p.match_value(s.ID)
    } else {
        pkg, id = "", pkg
    }
    return &ast.QualId{pkg, id}
}

// StructType     = "struct" "{" { FieldDecl ";" } "}" .
func (p *parser) parse_struct_type() *ast.StructType {
    var fs []*ast.FieldDecl = nil
    p.match(s.STRUCT)
    p.match('{')
    for p.token != s.EOF && p.token != '}' {
        fs = append(fs, p.parse_field_decl())
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
func (p *parser) parse_field_decl() *ast.FieldDecl {
    if p.token == '*' {
        // Anonymous field.
        p.next()
        pt := &ast.PtrType{p.parse_qual_id()}
        tag := p.parse_tag_opt()
        return &ast.FieldDecl{Names: nil, Type: pt, Tag: tag}
    } else if p.token == s.ID {
        pkg := p.match_value(s.ID)
        if p.token == '.' {
            // If the field decl begins with a qualified-id, it's parsed as an
            // anonymous field.
            p.next()
            id := p.match_value(s.ID)
            t := &ast.QualId{pkg, id}
            tag := p.parse_tag_opt()
            return &ast.FieldDecl{Names: nil, Type: t, Tag: tag}
        } else if p.token == s.STRING || p.token == ';' || p.token == '}' {
            // If it's only a single identifier, with no separate type
            // declaration, it's also an anonymous filed.
            t := &ast.QualId{"", pkg}
            tag := p.parse_tag_opt()
            return &ast.FieldDecl{Names: nil, Type: t, Tag: tag}
        } else {
            ids := p.parse_id_list(pkg)
            t := p.parse_type()
            tag := p.parse_tag_opt()
            return &ast.FieldDecl{Names: ids, Type: t, Tag: tag}
        }
    }
    p.error("Invalid field declaration")
    return &ast.FieldDecl{Names: nil, Type: &ast.Error{}}
}

func (p *parser) parse_tag_opt() (tag string) {
    if p.token == s.STRING {
        tag = p.match_value(s.STRING)
    } else {
        tag = ""
    }
    return
}

// IdentifierList = identifier { "," identifier } .
func (p *parser) parse_id_list(id string) (ids []string) {
    if len(id) == 0 {
        id = p.match_value(s.ID)
    }
    if len(id) > 0 {
        ids = append(ids, id)
    }
    for p.token == ',' {
        p.next()
        id = p.match_value(s.ID)
        if len(id) > 0 {
            ids = append(ids, id)
        }
    }
    return ids
}

// FunctionType = "func" Signature .
func (p *parser) parse_func_type() *ast.FuncType {
    p.match(s.FUNC)
    return p.parse_signature()
}

// Signature = Parameters [ Result ] .
// Result    = Parameters | Type .
func (p *parser) parse_signature() *ast.FuncType {
    ps := p.parse_parameters()
    var rs []*ast.ParamDecl = nil
    if p.token == '(' {
        rs = p.parse_parameters()
    } else if is_type_lookahead(p.token) {
        rs = []*ast.ParamDecl{&ast.ParamDecl{Type: p.parse_type()}}
    }
    return &ast.FuncType{ps, rs}
}

// Parameters    = "(" [ ParameterList [ "," ] ] ")" .
// ParameterList = ParameterDecl { "," ParameterDecl } .
func (p *parser) parse_parameters() (ds []*ast.ParamDecl) {
    p.match('(')
    if p.token == ')' {
        p.next()
        return nil
    }
    for p.token != s.EOF && p.token != ')' && p.token != ';' {
        d := p.parse_param_decl()
        ds = append(ds, d...)
        if p.token != ')' {
            p.match(',')
        }
    }
    p.match(')')
    return ds
}

// Parse an identifier or a type.  Return a param decl with filled either the
// parameter name or the parameter type, depending upon what was parsed.
func (p *parser) parse_id_or_type() *ast.ParamDecl {
    if p.token == s.ID {
        id := p.match_value(s.ID)
        if p.token == '.' {
            p.next()
            name := p.match_value(s.ID)
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
        t := p.parse_type() // FIXME: parenthesized types not allowed
        return &ast.ParamDecl{Type: t, Variadic: v}
    }
}

// Parse a list of (qualified) identifiers or types.
func (p *parser) parse_id_or_type_list() (ps []*ast.ParamDecl) {
    for {
        ps = append(ps, p.parse_id_or_type())
        if p.token != ',' {
            break
        }
        p.next()
    }
    return ps
}

// ParameterDecl = [ IdentifierList ] [ "..." ] Type .
func (p *parser) parse_param_decl() (ps []*ast.ParamDecl) {
    ps = p.parse_id_or_type_list()
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
    t := p.parse_type()
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
func (p *parser) parse_interface_type() *ast.InterfaceType {
    p.match(s.INTERFACE)
    p.match('{')
    emb := []*ast.QualId(nil)
    meth := []*ast.MethodSpec(nil)
    for p.token != s.EOF && p.token != '}' {
        id := p.match_value(s.ID)
        if p.token == '.' {
            p.next()
            name := p.match_value(s.ID)
            emb = append(emb, &ast.QualId{Pkg: id, Id: name})
        } else if p.token == '(' {
            sig := p.parse_signature()
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
func (p *parser) parse_const_decl() ast.Decl {
    p.match(s.CONST)
    if p.token == '(' {
        p.next()
        cs := []*ast.ConstDecl(nil)
        for p.token != s.EOF && p.token != ')' {
            cs = append(cs, p.parse_const_spec())
            if p.token != ')' {
                p.sync2(';', ')')
            }
        }
        p.match(')')
        return &ast.ConstGroup{Decls: cs}
    } else {
        return p.parse_const_spec()
    }
}

// ConstSpec = IdentifierList [ [ Type ] "=" ExpressionList ] .
func (p *parser) parse_const_spec() *ast.ConstDecl {
    var (
        t   ast.TypeSpec
        es  []ast.Expr
    )
    ids := p.parse_id_list("")
    if is_type_lookahead(p.token) {
        t = p.parse_type()
    }
    if p.token == '=' {
        p.next()
        es = p.parse_expr_list()
    }
    return &ast.ConstDecl{Names: ids, Type: t, Values: es}
}

// VarDecl = "var" ( VarSpec | "(" { VarSpec ";" } ")" ) .
func (p *parser) parse_var_decl() ast.Decl {
    p.match(s.VAR)
    if p.token == '(' {
        p.next()
        vs := []*ast.VarDecl(nil)
        for p.token != s.EOF && p.token != ')' {
            vs = append(vs, p.parse_var_spec())
            if p.token != ')' {
                p.sync2(';', ')')
            }
        }
        p.match(')')
        return &ast.VarGroup{Decls: vs}
    } else {
        return p.parse_var_spec()
    }
}

// VarSpec = IdentifierList ( Type [ "=" ExpressionList ] | "=" ExpressionList ) .
func (p *parser) parse_var_spec() *ast.VarDecl {
    var (
        t   ast.TypeSpec
        es  []ast.Expr
    )
    ids := p.parse_id_list("")
    if is_type_lookahead(p.token) {
        t = p.parse_type()
    }
    if p.token == '=' {
        p.next()
        es = p.parse_expr_list()
    }
    return &ast.VarDecl{Names: ids, Type: t, Init: es}
}

// MethodDecl   = "func" Receiver MethodName ( Function | Signature ) .
// FunctionDecl = "func" FunctionName ( Function | Signature ) .
// FunctionName = identifier .
// Function     = Signature FunctionBody .
// FunctionBody = Block .
func (p *parser) parse_func_decl() ast.Decl {
    var r *ast.Receiver
    p.match(s.FUNC)
    if p.token == '(' {
        r = p.parse_receiver()
    }
    name := p.match_value(s.ID)
    sig := p.parse_signature()
    var blk *ast.Block = nil
    if p.token == '{' {
        blk = p.parse_block()
    }
    return &ast.FuncDecl{Name: name, Recv: r, Sig: sig, Body: blk}
}

// Receiver     = "(" [ identifier ] [ "*" ] BaseTypeName ")" .
// BaseTypeName = identifier .
func (p *parser) parse_receiver() *ast.Receiver {
    p.match('(')
    var n string
    if p.token == s.ID {
        n = p.match_value(s.ID)
    }
    ptr := false
    if p.token == '*' {
        p.next()
        ptr = true
    }
    t := p.match_value(s.ID)
    p.sync2(')', ';')
    var tp ast.TypeSpec = &ast.QualId{Id: t}
    if ptr {
        tp = &ast.PtrType{Base: tp}
    }
    return &ast.Receiver{Name: n, Type: tp}
}

func (p *parser) parse_block() *ast.Block {
    p.match('{')
    p.match('}')
    return &ast.Block{}
}

// ExpressionList = Expression { "," Expression } .
func (p *parser) parse_expr_list() (es []ast.Expr) {
    for {
        es = append(es, p.parse_expr())
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
func (p *parser) parse_expr_or_type() (ast.Expr, ast.TypeSpec) {
    return p.parse_or_expr_or_type()
}

func (p *parser) parse_expr() ast.Expr {
    if e, _ := p.parse_or_expr_or_type(); e != nil {
        return e
    }
    p.error("Type not allowed in this context")
    return &ast.Error{}
}

// Expression = UnaryExpr | Expression binary_op UnaryExpr .
// `Expression` from the Go Specification is reaplaced by the following
// ``xxxExpr` productions.

// LogicalOrExpr = LogicalAndExpr { "||" LogicalAndExpr }
func (p *parser) parse_or_expr_or_type() (ast.Expr, ast.TypeSpec) {
    a0, t := p.parse_and_expr_or_type()
    if a0 == nil {
        return nil, t
    }
    for p.token == s.OR {
        p.next()
        a1, _ := p.parse_and_expr_or_type()
        a0 = &ast.BinaryExpr{Op: s.OR, Arg0: a0, Arg1: a1}
    }
    return a0, nil
}

// LogicalAndExpr = CompareExpr { "&&" CompareExpr }
func (p *parser) parse_and_expr_or_type() (ast.Expr, ast.TypeSpec) {
    a0, t := p.parse_compare_expr_or_type()
    if a0 == nil {
        return nil, t
    }
    for p.token == s.AND {
        p.next()
        a1, _ := p.parse_compare_expr_or_type()
        a0 = &ast.BinaryExpr{Op: s.AND, Arg0: a0, Arg1: a1}
    }
    return a0, nil
}

// CompareExpr = AddExpr { rel_op AddExpr }
func (p *parser) parse_compare_expr_or_type() (ast.Expr, ast.TypeSpec) {
    a0, t := p.parse_add_expr_or_type()
    if a0 == nil {
        return nil, t
    }
    for is_rel_op(p.token) {
        op := p.token
        p.next()
        a1, _ := p.parse_add_expr_or_type()
        a0 = &ast.BinaryExpr{Op: op, Arg0: a0, Arg1: a1}
    }
    return a0, nil
}

// AddExpr = MulExpr { add_op MulExpr}
func (p *parser) parse_add_expr_or_type() (ast.Expr, ast.TypeSpec) {
    a0, t := p.parse_mul_expr_or_type()
    if a0 == nil {
        return nil, t
    }
    for is_add_op(p.token) {
        op := p.token
        p.next()
        a1, _ := p.parse_mul_expr_or_type()
        a0 = &ast.BinaryExpr{Op: op, Arg0: a0, Arg1: a1}
    }
    return a0, nil
}

// MulExpr = UnaryExpr { mul_op UnaryExpr }
func (p *parser) parse_mul_expr_or_type() (ast.Expr, ast.TypeSpec) {
    a0, t := p.parse_unary_expr_or_type()
    if a0 == nil {
        return nil, t
    }
    for is_mul_op(p.token) {
        op := p.token
        p.next()
        a1, _ := p.parse_unary_expr_or_type()
        a0 = &ast.BinaryExpr{Op: op, Arg0: a0, Arg1: a1}
    }
    return a0, nil
}

// binary_op  = "||" | "&&" | rel_op | add_op | mul_op .

// rel_op     = "==" | "!=" | "<" | "<=" | ">" | ">=" .
func is_rel_op(t uint) bool {
    switch t {
    case s.EQ, s.NE, s.LT, s.LE, s.GT, s.GE:
        return true
    default:
        return false
    }
}

// add_op     = "+" | "-" | "|" | "^" .
func is_add_op(t uint) bool {
    switch t {
    case '+', '-', '|', '^':
        return true
    default:
        return false
    }
}

// mul_op     = "*" | "/" | "%" | "<<" | ">>" | "&" | "&^" .
func is_mul_op(t uint) bool {
    switch t {
    case '*', '/', '%', s.SHL, s.SHR, '&', s.ANDN:
        return true
    default:
        return false
    }
}

// UnaryExpr  = PrimaryExpr | unary_op UnaryExpr .
// unary_op   = "+" | "-" | "!" | "^" | "*" | "&" | "<-" .
func (p *parser) parse_unary_expr_or_type() (ast.Expr, ast.TypeSpec) {
    switch p.token {
    case '+', '-', '!', '^', '&':
        op := p.token
        p.next()
        a, _ := p.parse_unary_expr_or_type()
        return &ast.UnaryExpr{Op: op, Arg: a}, nil
    case '*':
        p.next()
        if a, t := p.parse_unary_expr_or_type(); t == nil {
            return &ast.UnaryExpr{Op: '*', Arg: a}, nil
        } else {
            return nil, &ast.PtrType{Base: t}
        }
    case s.RECV:
        p.next()
        if a, t := p.parse_unary_expr_or_type(); t == nil {
            return &ast.UnaryExpr{Op: s.RECV, Arg: a}, nil
        } else if ch, ok := t.(*ast.ChanType); ok {
            ch.Recv, ch.Send = true, false
            return nil, ch
        } else {
            p.error("invalid receive operation")
            return &ast.Error{}, nil
        }
    default:
        return p.parse_primary_expr_or_type()
    }
}

func (p *parser) need_type(ex ast.Expr) (ast.TypeSpec, bool) {
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

func (p *parser) parse_primary_expr_or_type() (ast.Expr, ast.TypeSpec) {
    var (
        ex  ast.Expr
        t   ast.TypeSpec
    )
    // Handle initial Operand, Conversion or a BuiltinCall
    switch p.token {
    case s.ID: // CompositeLit, MethodExpr, Conversion, BuiltinCall, OperandName
        id := p.match_value(s.ID)
        if p.token == '{' {
            ex = p.parse_composite_literal(&ast.QualId{Id: id})
        } else if p.token == '(' {
            ex = p.parse_call(&ast.QualId{Id: id})
        } else {
            ex = &ast.QualId{Id: id}
        }
    case '(': // Expression, MethodExpr, Conversion
        p.next()
        ex, t = p.parse_expr_or_type()
        p.match(')')
        if ex == nil {
            if p.token == '(' {
                ex = p.parse_conversion(t)
            } else {
                return nil, t
            }
        }
    case '[', s.STRUCT, s.MAP: // Conversion, CompositeLit
        t = p.parse_type()
        if p.token == '(' {
            ex = p.parse_conversion(t)
        } else if p.token == '{' {
            ex = p.parse_composite_literal(t)
        } else {
            return nil, t
        }
    case s.FUNC: // Conversion, FunctionLiteral
        t = p.parse_type()
        if p.token == '(' {
            ex = p.parse_conversion(t)
        } else if p.token == '{' {
            ex = p.parse_func_literal(t)
        } else {
            return nil, t
        }
    case '*', s.RECV:
        panic("should no reach here")
    case s.INTERFACE, s.CHAN: // Conversion
        t = p.parse_type()
        if p.token == '(' {
            ex = p.parse_conversion(t)
        } else {
            return nil, t
        }
    case s.INTEGER, s.FLOAT, s.IMAGINARY, s.RUNE, s.STRING: // BasicLiteral
        k := p.token
        v := p.match_value(k)
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
                t = p.parse_type()
                ex = &ast.TypeAssertion{Type: t, Arg: ex}
            } else {
                // PrimaryExpr Selector
                // Selector = "." identifier .
                id := p.match_value(s.ID)
                ex = &ast.Selector{Arg: ex, Id: id}
            }
        case '[':
            // PrimaryExpr Index
            // PrimaryExpr Slice
            ex = p.parse_index_or_slice(ex)
        case '(':
            // PrimaryExpr Call
            ex = p.parse_call(ex)
        case '{':
            // Composite literal
            typ, ok := p.need_type(ex)
            if !ok {
                p.error("invalid type for composite literal")
            }
            ex = p.parse_composite_literal(typ)
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
func (p *parser) parse_composite_literal(typ ast.TypeSpec) ast.Expr {
    elts := p.parse_literal_value()
    return &ast.CompLiteral{Type: typ, Elts: elts}
}

// LiteralValue  = "{" [ ElementList [ "," ] ] "}" .
// ElementList   = Element { "," Element } .
func (p *parser) parse_literal_value() (elts []*ast.Element) {
    p.match('{')
    for p.token != s.EOF && p.token != '}' {
        elts = append(elts, p.parse_element())
        if p.token != '}' {
            p.match(',')
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
func (p *parser) parse_element() *ast.Element {
    var k ast.Expr
    if p.token != '{' {
        k = p.parse_expr()
        if p.token != ':' {
            return &ast.Element{Key: nil, Value: k}
        }
        p.match(':')
    }
    if p.token == '{' {
        elts := p.parse_literal_value()
        e := &ast.CompLiteral{Type: nil, Elts: elts}
        return &ast.Element{Key: k, Value: e}
    } else {
        e := p.parse_expr()
        return &ast.Element{Key: k, Value: e}
    }
}

// Conversion = Type "(" Expression [ "," ] ")" .
func (p *parser) parse_conversion(typ ast.TypeSpec) ast.Expr {
    p.match('(')
    a := p.parse_expr()
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
func (p *parser) parse_call(f ast.Expr) ast.Expr {
    p.match('(')
    if p.token == ')' {
        p.next()
        return &ast.Call{Func: f}
    }
    var as []ast.Expr
    a, t := p.parse_expr_or_type()
    if a != nil {
        as = append(as, a)
    }
    seen_dots := false
    if p.token == s.DOTS {
        p.next()
        seen_dots = true
    }
    if p.token != ')' {
        p.match(',')
    }
    for p.token != s.EOF && p.token != ')' && !seen_dots {
        as = append(as, p.parse_expr())
        if p.token == s.DOTS {
            p.next()
            seen_dots = true
        }
        if p.token != ')' {
            p.match(',')
        }
    }
    p.match(')')
    return &ast.Call{Func: f, Type: t, Args: as, Ellipsis: seen_dots}
}

// FunctionLit = "func" Function .
func (p *parser) parse_func_literal(typ ast.TypeSpec) ast.Expr {
    sig := typ.(*ast.FuncType)
    p.match('{')
    p.match('}')
    return &ast.FuncLiteral{Sig: sig, Body: &ast.Block{}}
}

// Index          = "[" Expression "]" .
// Slice          = "[" ( [ Expression ] ":" [ Expression ] ) |
//                      ( [ Expression ] ":" Expression ":" Expression )
//                  "]" .
func (p *parser) parse_index_or_slice(e ast.Expr) ast.Expr {
    p.match('[')
    if p.token == ':' {
        p.next()
        if p.token == ']' {
            p.next()
            return &ast.SliceExpr{Array: e}
        }
        h := p.parse_expr()
        if p.token == ']' {
            p.next()
            return &ast.SliceExpr{Array: e, High: h}
        }
        p.match(':')
        c := p.parse_expr()
        p.match(']')
        return &ast.SliceExpr{Array: e, High: h, Cap: c}
    } else {
        i := p.parse_expr()
        if p.token == ']' {
            p.next()
            return &ast.IndexExpr{Array: e, Idx: i}
        }
        p.match(':')
        if p.token == ']' {
            p.next()
            return &ast.SliceExpr{Array: e, Low: i}
        }
        h := p.parse_expr()
        if p.token == ']' {
            p.next()
            return &ast.SliceExpr{Array: e, Low: i, High: h}
        }
        p.match(':')
        c := p.parse_expr()
        p.match(']')
        return &ast.SliceExpr{Array: e, Low: i, High: h, Cap: c}
    }
}
