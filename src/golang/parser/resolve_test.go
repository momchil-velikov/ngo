package parser

import (
	"golang/ast"
	"path/filepath"
	"testing"
)

func TestDeclareResolveTypeUniverse(t *testing.T) {
	const TESTDIR = "_test/02/src/a"

	up, err := ParsePackage(TESTDIR, []string{"a.go", "b.go"})
	if err != nil {
		t.Error(err)
	}
	p, err := ResolvePackage(up, nil)
	if err != nil {
		t.Error(err)
	}

	cases := []struct{ decl, ref, file string }{
		{"a", "bool", "a.go"},
		{"b", "byte", "a.go"},
		{"c", "uint8", "a.go"},
		{"d", "uint16", "a.go"},
		{"e", "uint32", "a.go"},
		{"f", "uint64", "a.go"},
		{"g", "int8", "a.go"},
		{"h", "int16", "a.go"},
		{"i", "rune", "a.go"},
		{"j", "int32", "a.go"},
		{"k", "int64", "a.go"},
		{"l", "float32", "b.go"},
		{"m", "float64", "b.go"},
		{"n", "complex64", "b.go"},
		{"o", "complex128", "b.go"},
		{"p", "uint", "b.go"},
		{"q", "int", "b.go"},
		{"r", "uintptr", "b.go"},
		{"s", "string", "b.go"},
	}
	for _, cs := range cases {
		// Test type is declared at package scope.
		sym := p.Decls[cs.decl]
		if sym == nil {
			t.Fatalf("name `%s` not found at package scope\n", cs.decl)
		}
		dcl, ok := sym.(*ast.TypeDecl)
		if !ok {
			t.Fatalf("`%s` must be a TypeDecl\n", cs.decl)
		}
		// Test type is registered under its name.
		name, _, file := sym.DeclaredAt()
		if name != cs.decl {
			t.Errorf("type `%s` declared as `%s`\n", cs.decl, name)
		}
		// Test type is declared in the respective file.
		if file == nil || filepath.Base(file.Name) != cs.file {
			t.Errorf("type '%s' must be declared in file `%s`", cs.decl, cs.file)
		}
		// Test type declaration refers to the respective typename at universe
		// scope.
		tn, ok := dcl.Type.(*ast.Typename)
		if !ok {
			t.Fatalf("declaration of `%s` does not refer to a typename\n", cs.decl)
		}
		if tn.Decl != ast.UniverseScope.Find(cs.ref) {
			if tn.Decl == nil {
				t.Errorf("`%s` refers to an invalid Typename\n", cs.decl)
			} else {
				t.Errorf(
					"`%s` expected to refer to `%s`, refers to `%s` instead\n",
					cs.decl, cs.ref, tn.Decl.Name,
				)
			}
		}
	}
}
