package ast

import "fmt"

// The Symbol interface denotes entities, which are put into various Scopes
// (symbol tables) and can be looked up by name
type Symbol interface {
	symbol()
	DeclaredAt() (string, int, *File) // Returns the source position of the declaration
}

// The names of the imported packages are inserted into the file scope of the
// importing source file. Their position refers to the import declaration.
func (Import) symbol() {}
func (i *Import) DeclaredAt() (string, int, *File) {
	return i.Name, i.Off, i.File
}

func (TypeDecl) symbol() {}
func (t *TypeDecl) DeclaredAt() (string, int, *File) {
	return t.Name, t.Off, t.File
}

func (Const) symbol() {}
func (c *Const) DeclaredAt() (string, int, *File) {
	return c.Name, c.Off, c.File
}

func (Var) symbol() {}
func (v *Var) DeclaredAt() (string, int, *File) {
	return v.Name, v.Off, v.File
}

func (FuncDecl) symbol() {}
func (f *FuncDecl) DeclaredAt() (string, int, *File) {
	return f.Name, f.Off, f.File
}

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
func (p *Package) Parent() Scope { return universeScope }

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
		sym = universeScope.Lookup(name)
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

var universeScope *_UniverseScope

func init() {
	u := &_UniverseScope{make(map[string]Symbol)}

	u.dcl["#nil"] = &TypeDecl{Name: "#nil", Type: &BuiltinType{BUILTIN_NIL}}
	u.dcl["bool"] = &TypeDecl{Name: "bool", Type: &BuiltinType{BUILTIN_BOOL}}
	ui8 := &BuiltinType{BUILTIN_UINT8}
	u.dcl["byte"] = &TypeDecl{Name: "byte", Type: ui8}
	u.dcl["uint8"] = &TypeDecl{Name: "uint8", Type: ui8}
	u.dcl["uint16"] = &TypeDecl{Name: "uint16", Type: &BuiltinType{BUILTIN_UINT16}}
	u.dcl["uint32"] = &TypeDecl{Name: "uint32", Type: &BuiltinType{BUILTIN_UINT32}}
	u.dcl["uint64"] = &TypeDecl{Name: "uint64", Type: &BuiltinType{BUILTIN_UINT64}}
	u.dcl["int8"] = &TypeDecl{Name: "int8", Type: &BuiltinType{BUILTIN_INT16}}
	u.dcl["int16"] = &TypeDecl{Name: "int16", Type: &BuiltinType{BUILTIN_INT16}}
	i32 := &BuiltinType{BUILTIN_INT32}
	u.dcl["rune"] = &TypeDecl{Name: "rune", Type: i32}
	u.dcl["int32"] = &TypeDecl{Name: "int32", Type: i32}
	u.dcl["int64"] = &TypeDecl{Name: "int64", Type: &BuiltinType{BUILTIN_INT64}}
	u.dcl["float32"] = &TypeDecl{Name: "float32", Type: &BuiltinType{BUILTIN_FLOAT32}}
	u.dcl["float64"] = &TypeDecl{Name: "float64", Type: &BuiltinType{BUILTIN_FLOAT64}}
	u.dcl["complex64"] = &TypeDecl{Name: "complex64", Type: &BuiltinType{BUILTIN_COMPLEX64}}
	u.dcl["complex128"] = &TypeDecl{Name: "complex128", Type: &BuiltinType{BUILTIN_COMPLEX128}}
	u.dcl["uint"] = &TypeDecl{Name: "uint", Type: &BuiltinType{BUILTIN_UINT}}
	u.dcl["int"] = &TypeDecl{Name: "int", Type: &BuiltinType{BUILTIN_INT}}
	u.dcl["uintptr"] = &TypeDecl{Name: "uintptr", Type: &BuiltinType{BUILTIN_UINTPTR}}
	universeScope = u
}
