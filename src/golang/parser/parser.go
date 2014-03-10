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
func (p *parser) match_valued(token uint) (value string, ok bool) {
    if p.expect(token) {
        ok = true
        value = p.scan.Value
        p.next()
    } else {
        ok = false
        value = ""
    }
    return
}

// Skip tokens, until given TOKEN found.
func (p *parser) skip_until(token uint) {
    for p.token != s.EOF && p.token != token {
        p.next()
    }
}

// Skip tokens, until either T0 or T1 token is found.
func (p *parser) skip_until2(t0, t1 uint) {
    for p.token != s.EOF && p.token != t0 && p.token != t1 {
        p.next()
    }
}

// Skip tokens, until either T0, T1 ot T2 token is found.
func (p *parser) skip_until3(t0, t1, t2 uint) {
    for p.token != s.EOF && p.token != t0 && p.token != t1 && p.token != t2 {
        p.next()
    }
}

// SourceFile = PackageClause ";" { ImportDecl ";" } { TopLevelDecl ";" } .
func (p *parser) parse_file() *ast.File {
    // Parse package name
    package_name, ok := p.parse_package_clause()
    if !ok {
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

// Parse a package clause. Return the package name or nil on error.
//
// PackageClause  = "package" PackageName .
// PackageName    = identifier .
func (p *parser) parse_package_clause() (id string, ok bool) {
    if p.match(s.PACKAGE) {
        id, ok = p.match_valued(s.ID)
    } else {
        id, ok = "", false
    }
    return
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
                name, path, ok := p.parse_import_spec()
                if ok {
                    imports = append(imports, ast.Import{name, path})
                } else {
                    p.skip_until(';')
                }
                if p.token != ')' {
                    // Spec does not say semicolon heres optional, Google Go
                    // allows it to be missing.
                    p.match(';')
                }
            }
            p.match(')')
        } else {
            name, path, ok := p.parse_import_spec()
            if ok {
                imports = append(imports, ast.Import{name, path})
            } else {
                p.skip_until(';')
            }
        }
        p.match(';')
    }
    return
}

// Parse import spec
//
// ImportSpec       = [ "." | PackageName ] ImportPath .
// ImportPath       = string_lit .
func (p *parser) parse_import_spec() (name string, path string, ok bool) {
    if p.token == '.' {
        name = "."
        p.next()
    } else if p.token == s.ID {
        name, _ = p.match_valued(s.ID)
    } else {
        name = ""
    }
    path, ok = p.match_valued(s.STRING)
    return
}

// Parse toplevel declaration(s)
//
// Declaration   = ConstDecl | TypeDecl | VarDecl .
// TopLevelDecl  = Declaration | FunctionDecl | MethodDecl .
func (p *parser) parse_toplevel_decls() (dcls []ast.Decl) {
    for {
        var f func() (ast.Decl, bool)
        switch p.token {
        case s.TYPE:
            f = p.parse_type_decl
        case s.CONST:
            f = p.parse_const_decl
        case s.VAR:
            f = p.parse_var_decl
        case s.FUNC:
            f = p.parse_func_decl
        default:
            return
        }
        d, ok := f()
        if d != nil {
            dcls = append(dcls, d)
        }
        if !(ok && p.match(';')) {
            p.skip_until_decl()
        }
    }
}

// Skip tokens, until beginning of a toplevel declaration found
func (p *parser) skip_until_decl() {
    for {
        switch p.token {
        case s.CONST, s.TYPE, s.VAR, s.FUNC, s.EOF:
            return
        default:
            p.next()
        }
    }
}

// Parse type declaration
//
// TypeDecl = "type" ( TypeSpec | "(" { TypeSpec ";" } ")" ) .
func (p *parser) parse_type_decl() (ast.Decl, bool) {
    p.match(s.TYPE)
    if p.token == '(' {
        p.next()
        var ts []*ast.TypeDecl = nil
        for p.token != s.EOF && p.token != ')' {
            if t, ok := p.parse_type_spec(); ok {
                ts = append(ts, t)
            } else {
                p.skip_until(';')
            }
            if p.token != ')' {
                p.match(';')
            }
        }
        p.match(')')
        return &ast.TypeGroup{Decls: ts}, true
    } else {
        if t, ok := p.parse_type_spec(); ok {
            return t, ok
        } else {
            return nil, false
        }
    }
}

// TypeSpec = identifier Type .
func (p *parser) parse_type_spec() (*ast.TypeDecl, bool) {
    if id, ok := p.match_valued(s.ID); ok {
        if t, ok := p.parse_type(); ok {
            return &ast.TypeDecl{Name: id, Type: t}, true
        }
    }
    return nil, false
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
func (p *parser) parse_type() (typ ast.TypeSpec, ok bool) {
    switch p.token {
    case s.ID:
        if t, ok := p.parse_qual_id(); ok {
            return t, ok // Avoid returning a non-nil interface value
        }

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
            t, ok := p.parse_type()
            if ok {
                return &ast.SliceType{t}, true
            }
        } else if p.token == s.DOTS {
            p.next()
            p.match(']')
            if t, ok := p.parse_type(); ok {
                return &ast.ArrayType{Dim: nil, EltType: t}, true
            }
        } else {
            e, ok := p.parse_expr()
            if ok && p.match(']') {
                t, ok := p.parse_type()
                if ok {
                    return &ast.ArrayType{Dim: e, EltType: t}, true
                }
            }
        }

    // PointerType = "*" BaseType .
    // BaseType = Type .
    case '*':
        p.next()
        if t, ok := p.parse_type(); ok {
            return &ast.PtrType{t}, true
        }

    // MapType     = "map" "[" KeyType "]" ElementType .
    // KeyType     = Type .
    case s.MAP:
        p.next()
        if p.match('[') {
            if k, ok := p.parse_type(); ok && p.match(']') {
                if t, ok := p.parse_type(); ok {
                    return &ast.MapType{k, t}, true
                }
            }
        }

    // ChannelType = ( "chan" [ "<-" ] | "<-" "chan" ) ElementType .
    case s.RECV:
        p.next()
        if p.match(s.CHAN) {
            if t, ok := p.parse_type(); ok {
                return &ast.ChanType{Send: false, Recv: true, EltType: t}, true
            }
        }
    case s.CHAN:
        p.next()
        send, recv := true, true
        if p.token == s.RECV {
            p.next()
            send, recv = true, false
        }
        if t, ok := p.parse_type(); ok {
            return &ast.ChanType{Send: send, Recv: recv, EltType: t}, true
        }

    case s.STRUCT:
        return p.parse_struct_type()

    case s.FUNC:
        return p.parse_func_type()

    case s.INTERFACE:
        return p.parse_interface_type()

    case '(':
        p.next()
        t, ok := p.parse_type()
        p.match(')')
        return t, ok

    default:
        p.error("expected typespec")
        return nil, false
    }

    return nil, false
}

// TypeName = identifier | QualifiedIdent .
func (p *parser) parse_qual_id() (*ast.QualId, bool) {
    pkg, _ := p.match_valued(s.ID)
    if p.token == '.' {
        p.next()
        if id, ok := p.match_valued(s.ID); ok {
            return &ast.QualId{pkg, id}, true
        } else {
            return nil, false
        }
    } else {
        return &ast.QualId{"", pkg}, true
    }
}

// StructType     = "struct" "{" { FieldDecl ";" } "}" .
func (p *parser) parse_struct_type() (*ast.StructType, bool) {
    var fs []*ast.FieldDecl = nil
    p.match(s.STRUCT)
    p.match('{')
    for p.token != s.EOF && p.token != '}' {
        if f, ok := p.parse_field_decl(); ok {
            fs = append(fs, f)
        } else {
            p.skip_until2(';', '}')
        }
        if p.token != '}' {
            p.match(';')
        }
    }
    p.match('}')

    return &ast.StructType{fs}, true
}

// FieldDecl      = (IdentifierList Type | AnonymousField) [ Tag ] .
// AnonymousField = [ "*" ] TypeName .
// Tag            = string_lit .
func (p *parser) parse_field_decl() (*ast.FieldDecl, bool) {
    if p.token == '*' {
        // Anonymous field.
        p.next()
        if t, ok := p.parse_qual_id(); ok {
            pt := &ast.PtrType{t}
            tag := p.parse_tag_opt()
            return &ast.FieldDecl{Names: nil, Type: pt, Tag: tag}, true
        }
    } else if p.token == s.ID {
        // If the field decl begins with a qualified-id, it's parsed as
        // an anonymous field.
        pkg, _ := p.match_valued(s.ID)
        if p.token == '.' {
            p.next()
            if id, ok := p.match_valued(s.ID); ok {
                t := &ast.QualId{pkg, id}
                tag := p.parse_tag_opt()
                return &ast.FieldDecl{Names: nil, Type: t, Tag: tag}, true
            }
        } else if p.token == s.STRING || p.token == ';' || p.token == '}' {
            // If it's only a single identifier, with no separate type
            // declaration, it's also an anonymous filed.
            t := &ast.QualId{"", pkg}
            tag := p.parse_tag_opt()
            return &ast.FieldDecl{Names: nil, Type: t, Tag: tag}, true
        } else if ids, ok := p.parse_id_list(pkg); ok {
            if t, ok := p.parse_type(); ok {
                tag := p.parse_tag_opt()
                return &ast.FieldDecl{Names: ids, Type: t, Tag: tag}, true
            }
        }
    }
    return nil, false
}

func (p *parser) parse_tag_opt() (tag string) {
    if p.token == s.STRING {
        tag, _ = p.match_valued(s.STRING)
    } else {
        tag = ""
    }
    return
}

// IdentifierList = identifier { "," identifier } .
func (p *parser) parse_id_list(id string) (ids []string, ok bool) {
    if len(id) == 0 {
        id, _ = p.match_valued(s.ID)
    }
    if len(id) > 0 {
        ids = append(ids, id)
    }
    for p.token == ',' {
        p.next()
        if id, ok := p.match_valued(s.ID); ok {
            ids = append(ids, id)
        }
    }
    if len(ids) > 0 {
        return ids, true
    } else {
        return nil, false
    }
}

// FunctionType = "func" Signature .
func (p *parser) parse_func_type() (*ast.FuncType, bool) {
    p.match(s.FUNC)
    return p.parse_signature()
}

// Signature = Parameters [ Result ] .
// Result    = Parameters | Type .
func (p *parser) parse_signature() (*ast.FuncType, bool) {
    ps, _ := p.parse_parameters()
    var rs []*ast.ParamDecl = nil
    if p.token == '(' {
        rs, _ = p.parse_parameters()
    } else if is_type_lookahead(p.token) {
        if t, ok := p.parse_type(); ok {
            rs = []*ast.ParamDecl{&ast.ParamDecl{Type: t}}
        }
    }
    return &ast.FuncType{ps, rs}, true
}

// Parameters    = "(" [ ParameterList [ "," ] ] ")" .
// ParameterList = ParameterDecl { "," ParameterDecl } .
func (p *parser) parse_parameters() (ds []*ast.ParamDecl, ok bool) {
    p.match('(')
    if p.token == ')' {
        p.next()
        return nil, true
    }
    for p.token != s.EOF && p.token != ')' && p.token != ';' {
        if d, ok := p.parse_param_decl(); ok {
            ds = append(ds, d...)
        } else {
            p.skip_until3(',', ')', ';')
        }
        if p.token != ')' {
            p.match(',')
        }
    }
    p.match(')')
    return ds, true
}

// Parse an identifier or a type.  Return a param decl with filled either the
// parameter name or the parameter type, depending upon what was parsed.
func (p *parser) parse_id_or_type() (*ast.ParamDecl, bool) {
    if p.token == s.ID {
        id, _ := p.match_valued(s.ID)
        if p.token == '.' {
            p.next()
            if name, ok := p.match_valued(s.ID); ok {
                return &ast.ParamDecl{Type: &ast.QualId{id, name}}, true
            }
            // Fallthrough as if the dot wasn't there.
        }
        return &ast.ParamDecl{Name: id}, true
    } else {
        v := false
        if p.token == s.DOTS {
            p.next()
            v = true
        }
        if t, ok := p.parse_type(); ok { // FIXME: parenthesized types not allowed
            return &ast.ParamDecl{Type: t, Variadic: v}, true
        } else {
            return nil, false
        }
    }
}

// Parse a list of (qualified) identifiers or types.
func (p *parser) parse_id_or_type_list() (ps []*ast.ParamDecl, ok bool) {
    for {
        if dcl, ok := p.parse_id_or_type(); ok {
            ps = append(ps, dcl)
        } else {
            p.skip_until2(',', ')')
        }
        if p.token != ',' {
            break
        }
        p.next()
    }
    return ps, true
}

// ParameterDecl = [ IdentifierList ] [ "..." ] Type .
func (p *parser) parse_param_decl() (ps []*ast.ParamDecl, ok bool) {
    ps, ok = p.parse_id_or_type_list()
    if p.token == ')' || /* missing closing paren */ p.token == ';' || p.token == '{' {
        // No type follows, then all the decls must be types.
        for _, dcl := range ps {
            if dcl.Type == nil {
                dcl.Type = &ast.QualId{"", dcl.Name}
                dcl.Name = ""
            }
        }
        return ps, true
    }
    // Otherwise, all the list elements must be identifiers, followed by a type.
    v := false
    if p.token == s.DOTS {
        p.next()
        v = true
    }
    if t, ok := p.parse_type(); ok {
        for _, dcl := range ps {
            if dcl.Type == nil {
                dcl.Type = t
                dcl.Variadic = v
            } else {
                p.error("Invalid parameter list") // FIXME: more specific message
            }
        }
        return ps, true
    } else {
        return nil, false
    }
}

// InterfaceType      = "interface" "{" { MethodSpec ";" } "}" .
// MethodSpec         = MethodName Signature | InterfaceTypeName .
// MethodName         = identifier .
// InterfaceTypeName  = TypeName .
func (p *parser) parse_interface_type() (*ast.InterfaceType, bool) {
    p.match(s.INTERFACE)
    p.match('{')
    emb := []*ast.QualId(nil)
    meth := []*ast.MethodSpec(nil)
    for p.token != s.EOF && p.token != '}' {
        if id, ok := p.match_valued(s.ID); ok {
            if p.token == '.' {
                p.next()
                if name, ok := p.match_valued(s.ID); ok {
                    emb = append(emb, &ast.QualId{Pkg: id, Id: name})
                }
            } else if p.token == '(' {
                if sig, ok := p.parse_signature(); ok {
                    meth = append(meth, &ast.MethodSpec{Name: id, Sig: sig})
                }
            } else {
                emb = append(emb, &ast.QualId{Pkg: "", Id: id})
            }
        }
        if p.token != '}' {
            p.match(';')
        }
    }
    p.match('}')
    return &ast.InterfaceType{Embed: emb, Methods: meth}, true
}

// ConstDecl = "const" ( ConstSpec | "(" { ConstSpec ";" } ")" ) .
func (p *parser) parse_const_decl() (ast.Decl, bool) {
    p.match(s.CONST)
    if p.token == '(' {
        p.next()
        cs := []*ast.ConstDecl(nil)
        for p.token != s.EOF && p.token != ')' {
            if c, ok := p.parse_const_spec(); ok {
                cs = append(cs, c)
            } else {
                p.skip_until2(';', ')')
            }
            if p.token != ')' {
                p.match(';')
            }
        }
        p.match(')')
        if len(cs) > 0 {
            return &ast.ConstGroup{Decls: cs}, true
        } else {
            return nil, false
        }
    } else {
        return p.parse_const_spec()
    }
}

// ConstSpec = IdentifierList [ [ Type ] "=" ExpressionList ] .
func (p *parser) parse_const_spec() (*ast.ConstDecl, bool) {
    if ids, ok := p.parse_id_list(""); ok {
        var t ast.TypeSpec = nil
        var es []ast.Expr = nil
        if is_type_lookahead(p.token) {
            t, _ = p.parse_type()
        }
        if p.token == '=' {
            p.next()
            es = p.parse_expr_list()
        }
        return &ast.ConstDecl{Names: ids, Type: t, Values: es}, true
    }
    return nil, false
}

// VarDecl = "var" ( VarSpec | "(" { VarSpec ";" } ")" ) .
func (p *parser) parse_var_decl() (ast.Decl, bool) {
    p.match(s.VAR)
    if p.token == '(' {
        p.next()
        vs := []*ast.VarDecl(nil)
        for p.token != s.EOF && p.token != ')' {
            if v, ok := p.parse_var_spec(); ok {
                vs = append(vs, v)
            } else {
                p.skip_until2(';', ')')
            }
            if p.token != ')' {
                p.match(';')
            }
        }
        p.match(')')
        if len(vs) > 0 {
            return &ast.VarGroup{Decls: vs}, true
        } else {
            return nil, false
        }
    } else {
        return p.parse_var_spec()
    }
}

// VarSpec = IdentifierList ( Type [ "=" ExpressionList ] | "=" ExpressionList ) .
func (p *parser) parse_var_spec() (*ast.VarDecl, bool) {
    if ids, ok := p.parse_id_list(""); ok {
        var t ast.TypeSpec = nil
        var es []ast.Expr = nil
        if is_type_lookahead(p.token) {
            if t, ok = p.parse_type(); !ok {
                p.skip_until2('=', ';')
            }
        }
        if p.token == '=' {
            p.next()
            es = p.parse_expr_list()
        }
        return &ast.VarDecl{Names: ids, Type: t, Init: es}, true
    }
    return nil, false
}

// MethodDecl   = "func" Receiver MethodName ( Function | Signature ) .
// FunctionDecl = "func" FunctionName ( Function | Signature ) .
// FunctionName = identifier .
// Function     = Signature FunctionBody .
// FunctionBody = Block .
func (p *parser) parse_func_decl() (ast.Decl, bool) {
    p.match(s.FUNC)
    var r *ast.Receiver = nil
    if p.token == '(' {
        r = p.parse_receiver()
    }
    name, _ := p.match_valued(s.ID)
    sig, ok := p.parse_signature()
    if !ok {
        p.skip_until2('{', ';')
    }
    var blk *ast.Block = nil
    if p.token == '{' {
        blk = p.parse_block()
    }
    if len(name) > 0 {
        return &ast.FuncDecl{Name: name, Recv: r, Sig: sig, Body: blk}, true
    } else {
        return nil, false
    }
}

// Receiver     = "(" [ identifier ] [ "*" ] BaseTypeName ")" .
// BaseTypeName = identifier .
func (p *parser) parse_receiver() *ast.Receiver {
    p.match('(')
    var n string
    if p.token == s.ID {
        n, _ = p.match_valued(s.ID)
    }
    ptr := false
    if p.token == '*' {
        p.next()
        ptr = true
    }
    t, ok := p.match_valued(s.ID)
    if !ok {
        p.skip_until2(')', ';')
        if p.token == ')' {
            p.next()
        }
        return nil
    }
    p.match(')')
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
        if e, ok := p.parse_expr(); ok {
            es = append(es, e)
        }
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
func (p *parser) parse_expr_or_type() (ast.Expr, ast.TypeSpec, bool) {
    return p.parse_or_expr_or_type()
}

func (p *parser) parse_expr() (ast.Expr, bool) {
    if e, _, ok := p.parse_or_expr_or_type(); ok {
        if e != nil {
            return e, true
        }
        p.error("Type not allowed in this context")
    }
    return nil, false
}

// Expression = UnaryExpr | Expression binary_op UnaryExpr .
// `Expression` from the Go Specification is reaplaced by the following
// ``xxxExpr` productions.

// LogicalOrExpr = LogicalAndExpr { "||" LogicalAndExpr }
func (p *parser) parse_or_expr_or_type() (ast.Expr, ast.TypeSpec, bool) {
    a0, t, ok := p.parse_and_expr_or_type()
    if !ok {
        return nil, nil, false
    } else if a0 == nil {
        return nil, t, true
    }
    for p.token == s.OR {
        p.next()
        a1, _, ok := p.parse_and_expr_or_type()
        if !ok || a1 == nil {
            return nil, nil, false
        }
        a0 = &ast.BinaryExpr{Op: s.OR, Arg0: a0, Arg1: a1}
    }
    return a0, nil, true
}

// LogicalAndExpr = CompareExpr { "&&" CompareExpr }
func (p *parser) parse_and_expr_or_type() (ast.Expr, ast.TypeSpec, bool) {
    a0, t, ok := p.parse_compare_expr_or_type()
    if !ok {
        return nil, nil, false
    } else if a0 == nil {
        return nil, t, true
    }
    for p.token == s.AND {
        p.next()
        a1, _, ok := p.parse_compare_expr_or_type()
        if !ok || a1 == nil {
            return nil, nil, false
        }
        a0 = &ast.BinaryExpr{Op: s.AND, Arg0: a0, Arg1: a1}
    }
    return a0, nil, true
}

// CompareExpr = AddExpr { rel_op AddExpr }
func (p *parser) parse_compare_expr_or_type() (ast.Expr, ast.TypeSpec, bool) {
    a0, t, ok := p.parse_add_expr_or_type()
    if !ok {
        return nil, nil, false
    } else if a0 == nil {
        return nil, t, true
    }
    for is_rel_op(p.token) {
        op := p.token
        p.next()
        a1, _, ok := p.parse_add_expr_or_type()
        if !ok || a1 == nil {
            return nil, nil, false
        }
        a0 = &ast.BinaryExpr{Op: op, Arg0: a0, Arg1: a1}
    }
    return a0, nil, true
}

// AddExpr = MulExpr { add_op MulExpr}
func (p *parser) parse_add_expr_or_type() (ast.Expr, ast.TypeSpec, bool) {
    a0, t, ok := p.parse_mul_expr_or_type()
    if !ok {
        return nil, nil, false
    } else if a0 == nil {
        return nil, t, true
    }
    for is_add_op(p.token) {
        op := p.token
        p.next()
        a1, _, ok := p.parse_mul_expr_or_type()
        if !ok || a1 == nil {
            return nil, nil, false
        }
        a0 = &ast.BinaryExpr{Op: op, Arg0: a0, Arg1: a1}
    }
    return a0, nil, true
}

// MulExpr = UnaryExpr { mul_op UnaryExpr }
func (p *parser) parse_mul_expr_or_type() (ast.Expr, ast.TypeSpec, bool) {
    a0, t, ok := p.parse_unary_expr_or_type()
    if !ok {
        return nil, nil, false
    } else if a0 == nil {
        return nil, t, ok
    }
    for is_mul_op(p.token) {
        op := p.token
        p.next()
        a1, _, ok := p.parse_unary_expr_or_type()
        if !ok || a1 == nil {
            return nil, nil, false
        }
        a0 = &ast.BinaryExpr{Op: op, Arg0: a0, Arg1: a1}
    }
    return a0, nil, true
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
func (p *parser) parse_unary_expr_or_type() (ast.Expr, ast.TypeSpec, bool) {
    switch p.token {
    case '+', '-', '!', '^', '&':
        op := p.token
        p.next()
        a, _, ok := p.parse_unary_expr_or_type()
        if !ok || a == nil {
            return nil, nil, false
        }
        return &ast.UnaryExpr{Op: op, Arg: a}, nil, true
    case '*':
        p.next()
        if a, t, ok := p.parse_unary_expr_or_type(); !ok {
            return nil, nil, false
        } else if t == nil {
            return &ast.UnaryExpr{Op: '*', Arg: a}, nil, true
        } else {
            return nil, &ast.PtrType{Base: t}, true
        }
    case s.RECV:
        p.next()
        if a, t, ok := p.parse_unary_expr_or_type(); !ok {
            return nil, nil, false
        } else if t == nil {
            return &ast.UnaryExpr{Op: s.RECV, Arg: a}, nil, true
        } else if ch, ok := t.(*ast.ChanType); ok {
            ch.Recv, ch.Send = true, false
            return nil, ch, true
        } else {
            return nil, nil, false
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

func (p *parser) parse_primary_expr_or_type() (ast.Expr, ast.TypeSpec, bool) {
    var (
        ex  ast.Expr
        t   ast.TypeSpec
        ok  bool
    )
    // Handle initial Operand, Conversion or a BuiltinCall
    switch p.token {
    case s.ID: // CompositeLit, MethodExpr, Conversion, BuiltinCall, OperandName
        id, _ := p.match_valued(s.ID)
        if p.token == '{' {
            ex, ok = p.parse_composite_literal(&ast.QualId{Id: id})
        } else if p.token == '(' {
            ex, ok = p.parse_call(&ast.QualId{Id: id})
        } else {
            ex, ok = &ast.QualId{Id: id}, true
        }
    case '(': // Expression, MethodExpr, Conversion
        p.next()
        if ex, t, ok = p.parse_expr_or_type(); ok {
            p.match(')')
            if ex == nil {
                if p.token == '(' {
                    ex, ok = p.parse_conversion(t)
                } else {
                    return nil, t, true
                }
            }
        }
    case '[', s.STRUCT, s.MAP: // Conversion, CompositeLit
        if t, ok = p.parse_type(); ok {
            if p.token == '(' {
                ex, ok = p.parse_conversion(t)
            } else if p.token == '{' {
                ex, ok = p.parse_composite_literal(t)
            } else {
                return nil, t, true
            }
        }
    case s.FUNC: // Conversion, FunctionLiteral
        if t, ok = p.parse_type(); ok {
            if p.token == '(' {
                ex, ok = p.parse_conversion(t)
            } else if p.token == '{' {
                ex, ok = p.parse_func_literal(t)
            } else {
                return nil, t, true
            }
        }
    case '*', s.RECV:
        panic("should no reach here")
    case s.INTERFACE, s.CHAN: // Conversion
        if t, ok = p.parse_type(); ok {
            if p.token == '(' {
                ex, ok = p.parse_conversion(t)
            } else {
                return nil, t, true
            }
        }
    case s.INTEGER, s.FLOAT, s.IMAGINARY, s.RUNE, s.STRING: // BasicLiteral
        k := p.token
        v, _ := p.match_valued(k)
        return &ast.Literal{Kind: k, Value: v}, nil, true
    default:
        p.error("token cannot start neither expression nor type")
    }
    // Parse the left-recursive alternatives for PrimaryExpr, folding the left-
    // hand parts in the variable `EX`.
    for {
        // If we haven't reached here with a parsed Expression, return an error.
        if !ok {
            return nil, nil, false
        }
        switch p.token {
        case '.': // TypeAssertion or Selector
            p.next()
            if p.token == '(' {
                // PrimaryExpr TypeAssertion
                // TypeAssertion = "." "(" Type ")" .
                if t, ok = p.parse_type(); ok {
                    ex = &ast.TypeAssertion{Type: t, Arg: ex}
                }
            } else {
                // PrimaryExpr Selector
                // Selector = "." identifier .
                if id, ok := p.match_valued(s.ID); ok {
                    ex = &ast.Selector{Arg: ex, Id: id}
                }
            }
        case '[':
            // PrimaryExpr Index
            // PrimaryExpr Slice
            ex, ok = p.parse_index_or_slice(ex)
        case '(':
            // PrimaryExpr Call
            ex, ok = p.parse_call(ex)
        case '{':
            // Composite literal
            var typ ast.TypeSpec = nil
            if typ, ok = p.need_type(ex); ok {
                ex, ok = p.parse_composite_literal(typ)
            } else {
                p.error("invalid type for composite literal")
            }
        default:
            return ex, nil, true
        }
    }
}

// Operand    = Literal | OperandName | MethodExpr | "(" Expression ")" .
// Literal    = BasicLit | CompositeLit | FunctionLit .
// BasicLit   = int_lit | float_lit | imaginary_lit | rune_lit | string_lit .
// OperandName = identifier | QualifiedIdent.

// Parsed as Selector expression, to be fixed up later by the typecheck phase.
// MethodExpr    = ReceiverType "." MethodName .
// ReceiverType  = TypeName | "(" "*" TypeName ")" | "(" ReceiverType ")" .
// func (p *parser) parse_method_expr(typ ast.TypeSpec) (ast.Expr, bool) {
//     p.match('.')
//     if id, ok := p.match_valued(s.ID); ok {
//         return &ast.MethodExpr{Type: typ, Id: id}, true
//     }
//     return nil, false
// }

// CompositeLit  = LiteralType LiteralValue .
// LiteralType   = StructType | ArrayType | "[" "..." "]" ElementType |
//                 SliceType | MapType | TypeName .
func (p *parser) parse_composite_literal(typ ast.TypeSpec) (ast.Expr, bool) {
    if elts, ok := p.parse_literal_value(); ok {
        return &ast.CompLiteral{Type: typ, Elts: elts}, true
    } else {
        return nil, false
    }
}

// LiteralValue  = "{" [ ElementList [ "," ] ] "}" .
// ElementList   = Element { "," Element } .
func (p *parser) parse_literal_value() (elts []*ast.Element, ok bool) {
    p.match('{')
    for p.token != s.EOF && p.token != '}' {
        if e, ok := p.parse_element(); ok {
            elts = append(elts, e)
        } else {
            p.skip_until2(',', '}')
        }
        if p.token != '}' {
            p.match(',')
        }
    }
    p.match('}')
    return elts, true
}

// Element       = [ Key ":" ] Value .
// Key           = FieldName | ElementIndex .
// FieldName     = identifier .
// ElementIndex  = Expression .
// Value         = Expression | LiteralValue .
func (p *parser) parse_element() (*ast.Element, bool) {
    var k ast.Expr
    var ok bool
    if p.token != '{' {
        if k, ok = p.parse_expr(); !ok {
            return nil, false
        }
        if p.token != ':' {
            return &ast.Element{Key: nil, Value: k}, true
        }
        p.match(':')
    }
    if p.token == '{' {
        if elts, ok := p.parse_literal_value(); ok {
            e := &ast.CompLiteral{Type: nil, Elts: elts}
            return &ast.Element{Key: k, Value: e}, true
        }
    } else {
        if e, ok := p.parse_expr(); ok {
            return &ast.Element{Key: k, Value: e}, true
        }
    }
    return nil, false
}

// Conversion = Type "(" Expression [ "," ] ")" .
func (p *parser) parse_conversion(typ ast.TypeSpec) (ast.Expr, bool) {
    p.match('(')
    if a, ok := p.parse_expr(); ok {
        if p.token == ',' {
            p.next()
        }
        p.match(')')
        return &ast.Conversion{Type: typ, Arg: a}, true
    }
    return nil, false
}

// Call           = "(" [ ArgumentList [ "," ] ] ")" .
// ArgumentList   = ExpressionList [ "..." ] .
// BuiltinCall = identifier "(" [ BuiltinArgs [ "," ] ] ")" .
// BuiltinArgs = Type [ "," ArgumentList ] | ArgumentList .
func (p *parser) parse_call(f ast.Expr) (ast.Expr, bool) {
    p.match('(')
    if p.token == ')' {
        p.next()
        return &ast.Call{Func: f}, true
    }
    var as []ast.Expr
    a, t, ok := p.parse_expr_or_type()
    if !ok {
        p.skip_until2(',', ')')
    } else if a != nil {
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
        if a, ok = p.parse_expr(); ok {
            as = append(as, a)
        } else {
            p.skip_until2(',', ')')
        }
        if p.token == s.DOTS {
            p.next()
            seen_dots = true
        }
        if p.token != ')' {
            p.match(',')
        }
    }
    p.match(')')
    return &ast.Call{Func: f, Type: t, Args: as, Ellipsis: seen_dots}, true
}

// FunctionLit = "func" Function .
func (p *parser) parse_func_literal(typ ast.TypeSpec) (ast.Expr, bool) {
    sig := typ.(*ast.FuncType)
    p.match('{')
    p.match('}')
    return &ast.FuncLiteral{Sig: sig, Body: &ast.Block{}}, true
}

// Index          = "[" Expression "]" .
// Slice          = "[" ( [ Expression ] ":" [ Expression ] ) |
//                      ( [ Expression ] ":" Expression ":" Expression )
//                  "]" .
func (p *parser) parse_index_or_slice(e ast.Expr) (ast.Expr, bool) {
    p.match('[')
    if p.token == ':' {
        p.next()
        if p.token == ']' {
            p.next()
            return &ast.SliceExpr{Array: e}, true
        }
        if h, ok := p.parse_expr(); ok {
            if p.token == ']' {
                p.next()
                return &ast.SliceExpr{Array: e, High: h}, true
            }
            p.match(':')
            if c, ok := p.parse_expr(); ok {
                p.match(']')
                return &ast.SliceExpr{Array: e, High: h, Cap: c}, true
            }
        }
    } else {
        if i, ok := p.parse_expr(); ok {
            if p.token == ']' {
                p.next()
                return &ast.IndexExpr{Array: e, Idx: i}, true
            }
            p.match(':')
            if p.token == ']' {
                p.next()
                return &ast.SliceExpr{Array: e, Low: i}, true
            }
            if h, ok := p.parse_expr(); ok {
                if p.token == ']' {
                    p.next()
                    return &ast.SliceExpr{Array: e, Low: i, High: h}, true
                }
                p.match(':')
                if c, ok := p.parse_expr(); ok {
                    p.match(']')
                    return &ast.SliceExpr{Array: e, Low: i, High: h, Cap: c}, true
                }
            }
        }
    }
    return nil, false
}
