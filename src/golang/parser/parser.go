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
	imports, ok := p.parse_import_decls()
	if !ok {
		return nil
	}

	// Parse toplevel declarations.
	decls, ok := p.parse_toplevel_decls()
	if !ok {
		return nil
	}

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
func (p *parser) parse_import_decls() (imports []ast.Import, ok bool) {
	imports = nil
	ok = true

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
				p.match(';')
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
func (p *parser) parse_toplevel_decls() (decls []ast.XDecl, ok bool) {
	ok = true
	return
}
