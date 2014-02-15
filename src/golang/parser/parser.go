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
    // fmt.Println("***" + s.TokenNames[p.token])
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
        err := false
        switch p.token {
        case s.TYPE:
            ds, ok := p.parse_type_decls()
            if ds != nil {
                dcls = append(dcls, ds...)
            }
            if !ok {
                err = true
                p.skip_until_decl()
            }
        default:
            return
        }
        if !err {
            p.match(';')
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

// Type     = TypeName | TypeLit | "(" Type ")" .
// TypeName = identifier | QualifiedIdent .
// TypeLit  = ArrayType | StructType | PointerType | FunctionType | InterfaceType |
//            SliceType | MapType | ChannelType .
func (p *parser) parse_type() (typ ast.TypeSpec, ok bool) {
    switch p.token {

    // TypeName = identifier | QualifiedIdent .
    case s.ID:
        pkg, _ := p.match_valued(s.ID)
        if p.token == '.' {
            p.next()
            id, ok := p.match_valued(s.ID)
            if ok {
                return &ast.BaseType{pkg, id}, true
            }
        } else {
            return &ast.BaseType{"", pkg}, true
        }

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
    default:
        p.error("expected typespec")
        return nil, false
    }

    return nil, false
}

// StructType     = "struct" "{" { FieldDecl ";" } "}" .
// FieldDecl      = (IdentifierList Type | AnonymousField) [ Tag ] .
// AnonymousField = [ "*" ] TypeName .
// Tag            = string_lit .

// FunctionType   = "func" Signature .
// Signature      = Parameters [ Result ] .
// Result         = Parameters | Type .
// Parameters     = "(" [ ParameterList [ "," ] ] ")" .
// ParameterList  = ParameterDecl { "," ParameterDecl } .
// ParameterDecl  = [ IdentifierList ] [ "..." ] Type .

// InterfaceType      = "interface" "{" { MethodSpec ";" } "}" .
// MethodSpec         = MethodName Signature | InterfaceTypeName .
// MethodName         = identifier .
// InterfaceTypeName  = TypeName .

// ChannelType = ( "chan" [ "<-" ] | "<-" "chan" ) ElementType .

// ConstDecl = "const" ( ConstSpec | "(" { ConstSpec ";" } ")" ) .
// ConstSpec = IdentifierList [ [ Type ] "=" ExpressionList ] .
// IdentifierList = identifier { "," identifier } .
// ExpressionList = Expression { "," Expression } .

func (p *parser) parse_expr() (*ast.Expr, bool) {
    cst, ok := p.match_valued(s.INTEGER)
    if ok {
        return &ast.Expr{cst}, true
    } else {
        return nil, false
    }
}
