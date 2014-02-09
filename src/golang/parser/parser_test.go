package parser

import (
	// "fmt"
	// "golang/ast"
	"testing"
)

func TestPackageClause(tst *testing.T) {
	src := `
package foo`
	t, e := Parse("package-clause.go", src)

	if e != nil {
		tst.Error(e)
	}

	if t.PackageName != "foo" {
		tst.Errorf("Unexpected package name `%s`", t.PackageName)
	}
}

func TestImportSingle(tst *testing.T) {
	src := `
package foo

import "fmt"
`
	t, e := Parse("import-single.go", src)

	if e != nil {
		tst.Error(e)
	}

	if t.PackageName != "foo" {
		tst.Errorf("Unexpected package name `%s`", t.PackageName)
	}

	if len(t.Imports) != 1 {
		tst.Error("Expected exactly one import")
	}
}

func TestImportMultiple(tst *testing.T) {
	src := `
package foo

import  ( "fmt"; . "foo"
	s "bar" 
	)
`
	t, e := Parse("import-single.go", src)

	if e != nil {
		tst.Error(e)
	}

	if t.PackageName != "foo" {
		tst.Errorf("Unexpected package name `%s`", t.PackageName)
	}

	if len(t.Imports) != 3 {
		tst.Error("Expected exactly three imports")
	}

	exp := `package foo

import (
    "fmt"
    . "foo"
    s "bar"
)

`
	f := t.Format()
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}
