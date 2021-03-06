package build

// The build parser is a simplified Go parser, sufficient for parsing
// package import clauses.

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type parser struct {
	scan  scanner
	token token
	err   error
	skipc bool
}

func (p *parser) init(name string, rd io.RuneReader) {
	p.scan.init(name, rd)
}

// Get the next token from the scanner.
func (p *parser) next() {
	p.token, p.err = p.scan.Get()
	//fmt.Fprintln(os.Stderr, ">>> parser.next():", p.token.String())
	for p.skipc && (p.token.Kind == tLINE_COMMENT || p.token.Kind == tBLOCK_COMMENT) {
		p.token, p.err = p.scan.Get()
		//fmt.Fprintln(os.Stderr, ">>> parser.next():", p.token.String())
	}
}

// Advance to the next token iff the current one is TOKEN.
func (p *parser) match(token uint) bool {
	if p.token.Kind == token {
		p.next()
		return true
	} else {
		return false
	}
}

type parserError struct {
	name    string
	ln, col uint
	msg     string
}

func (e *parserError) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s", e.name, e.ln, e.col, e.msg)
}

func (p *parser) error(msg string) error {
	return &parserError{name: p.scan.name, ln: p.token.Line, col: p.token.Col, msg: msg}
}

// Parse all the source files of a package
func parsePackage(path string, dir string, names []string) (*Package, error) {
	// Parse sources
	var files []*File
	for _, name := range names {
		path := filepath.Join(dir, name)
		rd, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		f, err := parseFile(path, bufio.NewReader(rd))
		rd.Close()
		if err != nil {
			return nil, err
		}
		f.Path = path
		files = append(files, f)
	}

	// We used to check for package name consistencty accross source files and
	// with package directory. We don't do this anymore, because build tag
	// based filtering is done afterwards and an ignored source file may cause
	// the check to fail here.  Let the real Go parser do it.

	pkg := &Package{
		Dir:     dir,
		Path:    path,
		Name:    filepath.Base(dir),
		Files:   files,
		Imports: make(map[string]*Package),
	}
	return pkg, nil
}

// Parse a source file
func parseFile(path string, rd io.RuneReader) (*File, error) {
	p := parser{}
	p.init(path, rd)
	return p.parseFile()
}

// SourceFile = PackageClause ";" { ImportDecl ";" }
func (p *parser) parseFile() (*File, error) {
	f := &File{}

	// Parse initial batch of comment lines.
	p.next()
	for p.token.Kind == tBLOCK_COMMENT || p.token.Kind == tLINE_COMMENT {
		if p.token.Kind == tLINE_COMMENT {
			f.Comments = append(f.Comments, p.token.Value)
		}
		p.next()
	}

	// If we've got the package keyword, consider this a Go source file.
	if p.token.Kind == tPACKAGE {
		// Ignore subsequent comments.
		p.skipc = true

		f.Package, p.err = p.parsePackageClause()
		if p.err != nil {
			return nil, p.err
		}
		if !p.match(';') {
			return nil, p.error("expected semicolon or newline to follow package clause")
		}

		// Parse import declaration(s)
		f.Imports, p.err = p.parseImportDecls()
		if p.err != nil {
			return nil, p.err
		}
	}
	return f, nil
}

// Parse a package clause. Return the package name.
//
// PackageClause  = "package" PackageName .
// PackageName    = identifier .
func (p *parser) parsePackageClause() (string, error) {
	p.match(tPACKAGE)
	name := p.token.Value
	if !p.match(tID) {
		return "", p.error("package keyword must be followed by package name")
	}
	return name, nil
}

// Parse import declarations(s).
//
// ImportDecl       = "import" ( ImportSpec | "(" { ImportSpec ";" } ")" ) .
func (p *parser) parseImportDecls() ([]string, error) {
	is := []string{}
	for p.token.Kind == tIMPORT {
		p.match(tIMPORT)
		if p.token.Kind == '(' {
			p.next()
			for p.token.Kind != tEOF && p.token.Kind != ')' {
				if i, err := p.parseImportSpec(); err != nil {
					return nil, err
				} else {
					is = append(is, i)
				}
				if p.token.Kind != ')' {
					if !p.match(';') {
						return nil, p.error("expected semicolon or newline to " +
							"follow import specification ")
					}
				}
			}
			if !p.match(')') {
				return nil, p.error("missing closing parenthesis after imports group")
			}
		} else {
			if i, err := p.parseImportSpec(); err != nil {
				return nil, err
			} else {
				is = append(is, i)
			}
		}
		if !p.match(';') {
			return nil, p.error("expected semicolon or newline to " +
				"follow import specification ")
		}
	}
	return is, nil
}

// Parse import spec
//
// ImportSpec       = [ "." | PackageName ] ImportPath .
// ImportPath       = string_lit .
func (p *parser) parseImportSpec() (string, error) {
	if p.token.Kind == '.' || p.token.Kind == tID {
		p.next()
	}
	path := p.token.Value
	if !p.match(tSTRING) {
		return "", p.error("missing import path in import specification")
	}
	return path, nil
}
