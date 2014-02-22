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

// Get the next token from the scannrt.
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
        ok := true
        ds := []ast.Decl(nil)
        d := ast.Decl(nil)
        switch p.token {
        case s.TYPE:
            ds, ok = p.parse_type_decls()
            if ds != nil {
                dcls = append(dcls, ds...)
            }
        case s.CONST:
            d, ok = p.parse_const_decl()
            if d != nil {
                dcls = append(dcls, d)
            }
        case s.VAR:
            d, ok = p.parse_var_decl()
            if d != nil {
                dcls = append(dcls, d)
            }
        case s.FUNC:
            d, ok = p.parse_func_decl()
            if d != nil {
                dcls = append(dcls, d)
            }
        default:
            return
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
func (p *parser) parse_type_decls() (ts []ast.Decl, fine bool) {
    fine = true
    p.match(s.TYPE)

    if p.token == '(' {
        p.next()
        for p.token != s.EOF && p.token != ')' {
            id, t, ok := p.parse_type_spec()
            if ok {
                ts = append(ts, &ast.TypeDecl{id, t})
            } else {
                fine = false
                p.skip_until(';')
            }
            if p.token != ')' {
                p.match(';')
            }
        }
        p.match(')')
    } else {
        id, t, ok := p.parse_type_spec()
        if ok {
            ts = append(ts, &ast.TypeDecl{id, t})
        } else {
            fine = false
            p.skip_until(';')
        }
    }
    return
}

// TypeSpec = identifier Type .
func (p *parser) parse_type_spec() (id string, t ast.TypeSpec, ok bool) {
    id, ok = p.match_valued(s.ID)
    if ok {
        t, ok = p.parse_type()
    }
    return
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
        return p.parse_type_name()

    // ArrayType   = "[" ArrayLength "]" ElementType .
    // ArrayLength = Expression .
    // ElementType = Type .
    // SliceType = "[" "]" ElementType .
    case '[':
        p.next()
        if p.token == ']' {
            p.next()
            t, ok := p.parse_type()
            if ok {
                return &ast.SliceType{t}, true
            }
        } else {
            e, ok := p.parse_expr()
            if ok && p.match(']') {
                t, ok := p.parse_type()
                if ok {
                    return &ast.ArrayType{e, t}, true
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
func (p *parser) parse_type_name() (ast.TypeSpec, bool) {
    pkg, _ := p.match_valued(s.ID)
    if p.token == '.' {
        p.next()
        if id, ok := p.match_valued(s.ID); ok {
            return &ast.TypeName{pkg, id}, true
        } else {
            return nil, false
        }
    } else {
        return &ast.TypeName{"", pkg}, true
    }
}

// StructType     = "struct" "{" { FieldDecl ";" } "}" .
func (p *parser) parse_struct_type() (*ast.StructType, bool) {
    var fs []*ast.FieldDecl = nil
    p.match(s.STRUCT)
    p.match('{')
    for p.token != s.EOF && p.token != '}' {
        if f, ok := p.parse_field_decl(); ok {
            fs = append(fs, f...)
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
func (p *parser) parse_field_decl() (fs []*ast.FieldDecl, ok bool) {
    if p.token == '*' {
        // Anonymous field.
        p.next()
        if t, ok := p.parse_type_name(); ok {
            t = &ast.PtrType{t}
            tag := p.parse_tag_opt()
            fs = append(fs, &ast.FieldDecl{"", t, tag})
            return fs, ok
        }
    } else if p.token == s.ID {
        // If the field decl begins with a qualified-id, it's parsed as
        // an anonymous field.
        pkg, _ := p.match_valued(s.ID)
        if p.token == '.' {
            p.next()
            if id, ok := p.match_valued(s.ID); ok {
                t := &ast.TypeName{pkg, id}
                tag := p.parse_tag_opt()
                fs = append(fs, &ast.FieldDecl{"", t, tag})
                return fs, true
            }
        } else if p.token == s.STRING || p.token == ';' || p.token == '}' {
            // If it's only a single identifier, with no separate type
            // declaration, it's also an anonymous filed.
            t := &ast.TypeName{"", pkg}
            tag := p.parse_tag_opt()
            fs = append(fs, &ast.FieldDecl{"", t, tag})
            return fs, true
        } else if ids, ok := p.parse_id_list(pkg); ok {
            if t, ok := p.parse_type(); ok {
                tag := p.parse_tag_opt()
                for _, id := range ids {
                    fs = append(fs, &ast.FieldDecl{id, t, tag})
                }
                return fs, true
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
                return &ast.ParamDecl{Type: &ast.TypeName{id, name}}, true
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
                dcl.Type = &ast.TypeName{"", dcl.Name}
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
    emb := []*ast.TypeName(nil)
    meth := []*ast.MethodSpec(nil)
    for p.token != s.EOF && p.token != '}' {
        if id, ok := p.match_valued(s.ID); ok {
            if p.token == '.' {
                p.next()
                if name, ok := p.match_valued(s.ID); ok {
                    emb = append(emb, &ast.TypeName{Pkg: id, Id: name})
                }
            } else if p.token == '(' {
                if sig, ok := p.parse_signature(); ok {
                    meth = append(meth, &ast.MethodSpec{Name: id, Sig: sig})
                }
            } else {
                emb = append(emb, &ast.TypeName{Pkg: "", Id: id})
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
        var es []*ast.Expr = nil
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

func (p *parser) parse_expr() (*ast.Expr, bool) {
    cst, ok := p.match_valued(s.INTEGER)
    if ok {
        return &ast.Expr{cst}, true
    } else {
        return nil, false
    }
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
        var es []*ast.Expr = nil
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
    var tp ast.TypeSpec = &ast.TypeName{Id: t}
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
func (p *parser) parse_expr_list() (es []*ast.Expr) {
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
