package ast

import (
	"fmt"

	"unicode"
	"unicode/utf8"
)

func isExported(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsLetter(r) && unicode.IsUpper(r)
}

// The Symbol interface denotes entities, which are put into various Scopes
// (symbol tables) and can be looked up by name
type Symbol interface {
	symbol()
	// Returns source position of the declaration
	DeclaredAt() (string, int, *File)
	// Returns true if the declaration begins with an upper case letter
	IsExported() bool
}

// The names of the imported packages are inserted into the file scope of the
// importing source file. Their position refers to the import declaration.
func (ImportDecl) symbol() {}
func (i *ImportDecl) DeclaredAt() (string, int, *File) {
	return i.Name, i.Off, i.File
}
func (*ImportDecl) IsExported() bool { return false }

func (TypeDecl) symbol() {}
func (t *TypeDecl) DeclaredAt() (string, int, *File) {
	return t.Name, t.Off, t.File
}
func (t *TypeDecl) IsExported() bool { return isExported(t.Name) }

func (Const) symbol() {}
func (c *Const) DeclaredAt() (string, int, *File) {
	return c.Name, c.Off, c.File
}
func (c *Const) IsExported() bool { return isExported(c.Name) }

func (Var) symbol() {}
func (v *Var) DeclaredAt() (string, int, *File) {
	return v.Name, v.Off, v.File
}
func (v *Var) IsExported() bool { return isExported(v.Name) }

func (FuncDecl) symbol() {}
func (f *FuncDecl) DeclaredAt() (string, int, *File) {
	return f.Name, f.Off, f.File
}
func (f *FuncDecl) IsExported() bool { return isExported(f.Name) }

func (Label) symbol() {}
func (l *Label) DeclaredAt() (string, int, *File) {
	return l.Label, l.Off, l.Blk.File()
}
func (l *Label) IsExported() bool { return false }

// The Scope interface determines the set of visible names at each point of a
// program
type Scope interface {
	Parent() Scope
	Package() *Package
	File() *File
	Func() *Func
	Declare(string, Symbol) error
	Lookup(string) Symbol
	Find(string) Symbol
}

type redeclarationError struct {
	old, new Symbol
}

func (e *redeclarationError) Error() string {
	name, off, file := e.new.DeclaredAt()
	ln, col := file.SrcMap.Position(off)
	s0 := fmt.Sprintf("%s:%d:%d: %s redeclared\n", file.Name, ln, col, name)
	_, off, file = e.old.DeclaredAt()
	ln, col = file.SrcMap.Position(off)
	s1 := fmt.Sprintf("\tprevious declaration at %s:%d:%d", file.Name, ln, col)
	return s0 + s1
}

// Package scope
func (p *Package) Parent() Scope { return UniverseScope }

func (p *Package) Package() *Package { return p }

func (*Package) File() *File { return nil }

func (*Package) Func() *Func { return nil }

func (p *Package) Declare(name string, sym Symbol) error {
	// When declaring an identifier at package scope, check it is not already
	// declared at the scope of file, which contain the declaration of the
	// said identifier; "no identifier may be declared in both the file and
	// package block".
	_, _, file := sym.DeclaredAt()
	old := file.Find(name)
	if old == nil {
		old = p.Decls[name]
		if old == nil {
			p.Decls[name] = sym
			return nil
		}
	}
	return &redeclarationError{old: old, new: sym}
}

func (p *Package) Lookup(name string) Symbol {
	sym := p.Find(name)
	if sym == nil {
		sym = UniverseScope.Lookup(name)
	}
	return sym
}

func (p *Package) Find(name string) Symbol {
	return p.Decls[name]
}

// File scope
func (f *File) Parent() Scope { return f.Pkg }

func (f *File) Package() *Package { return f.Pkg }

func (f *File) File() *File { return f }

func (*File) Func() *Func { return nil }

func (f *File) Declare(name string, sym Symbol) error {
	if old := f.Decls[name]; old != nil {
		return &redeclarationError{old: old, new: sym}
	}
	f.Decls[name] = sym
	return nil
}

func (f *File) Lookup(name string) Symbol {
	sym := f.Find(name)
	if sym == nil {
		sym = f.Pkg.Lookup(name)
	}
	return sym
}

func (f *File) Find(name string) Symbol {
	return f.Decls[name]
}

// Function scope.
func (fn *Func) Parent() Scope { return fn.Up }

func (fn *Func) Package() *Package { return fn.Up.Package() }

func (fn *Func) File() *File { return fn.Up.File() }

func (fn *Func) Func() *Func { return fn }

func (fn *Func) Declare(name string, sym Symbol) error {
	if fn.Labels == nil {
		fn.Labels = make(map[string]*Label)
	}
	if old := fn.Labels[name]; old != nil {
		return &redeclarationError{old: old, new: sym}
	}
	fn.Labels[name] = sym.(*Label)
	return nil
}

func (fn *Func) Lookup(name string) Symbol { return fn.Up.Lookup(name) }

func (fn *Func) Find(name string) Symbol {
	return fn.Labels[name]
}

// Block scope
func (b *Block) Parent() Scope { return b.Up }

func (b *Block) Package() *Package { return b.Up.Package() }

func (b *Block) File() *File { return b.Up.File() }

func (b *Block) Func() *Func { return b.Up.Func() }

func (b *Block) Declare(name string, sym Symbol) error {
	if old := b.Decls[name]; old != nil {
		return &redeclarationError{old: old, new: sym}
	}
	b.Decls[name] = sym
	return nil
}

func (b *Block) Lookup(name string) Symbol {
	sym := b.Find(name)
	if sym == nil {
		sym = b.Up.Lookup(name)
	}
	return sym
}

func (b *Block) Find(name string) Symbol {
	return b.Decls[name]
}

// Universe scope
type _UniverseScope struct {
	dcl map[string]Symbol
}

func (*_UniverseScope) Parent() Scope { return nil }

func (*_UniverseScope) Package() *Package { return nil }

func (*_UniverseScope) File() *File { return nil }

func (*_UniverseScope) Func() *Func { return nil }

func (*_UniverseScope) Declare(name string, sym Symbol) error {
	panic("should not try to declare names at Universe scope")
}

func (u *_UniverseScope) Lookup(name string) Symbol {
	return u.dcl[name]
}

func (u *_UniverseScope) Find(name string) Symbol {
	return u.dcl[name]
}
