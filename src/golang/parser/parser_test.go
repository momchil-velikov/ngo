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

func TestPackageError1(tst *testing.T) {
	src := `
foo`
	_, e := Parse("package-error1.go", src)

	if e == nil {
		tst.Error("Unexpected lack of error")
	} else {
		tst.Log(e)
	}
}

func TestPackageError2(tst *testing.T) {
	src := `
package`
	_, e := Parse("package-error2.go", src)

	if e == nil {
		tst.Error("Unexpected lack of error")
	} else {
		tst.Log(e)
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
    .
    "bar"
	s "baz"
	)
`
	t, e := Parse("import-single.go", src)

	if e != nil {
		tst.Error(e)
	}

	if t.PackageName != "foo" {
		tst.Errorf("Unexpected package name `%s`", t.PackageName)
	}

	if len(t.Imports) != 4 {
		tst.Error("Expected exactly three imports")
	}

	exp := `package foo

import (
    "fmt"
    . "foo"
    . "bar"
    s "baz"
)

`
	f := t.Format()
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestImportSeq(tst *testing.T) {
	src := `
package foo
import "bar"
import ( "baz"; "xyzzy" ; ) ; import "qux"
`
	t, e := Parse("import-seq.go", src)

	if e != nil {
		tst.Error(e)
	}

	exp := `package foo

import (
    "bar"
    "baz"
    "xyzzy"
    "qux"
)

`
	f := t.Format()
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestImportSeq1(tst *testing.T) {
	src := `
package foo
import "bar"
import ( "baz"; "xyzzy" ) ; import "qux"
`
	t, e := Parse("import-seq.go", src)

	if e != nil {
		tst.Error(e)
	}

	exp := `package foo

import (
    "bar"
    "baz"
    "xyzzy"
    "qux"
)

`
	f := t.Format()
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestImportError(tst *testing.T) {
	src := `package foo
import ( "foo"
    .
    s "bar"
    t
`
	t, e := Parse("import-error.go", src)
	f := t.Format()
	tst.Log(f)
	if e != nil {
		tst.Log(e)
	}
}

func TestImportError1(tst *testing.T) {
	src := `
package foo
import ( "baz"; s ) 
`
	t, e := Parse("import-error1.go", src)
	f := t.Format()
	tst.Log(f)
	if e != nil {
		tst.Log(e)
	}
}

func TestBaseType1(tst *testing.T) {
	src := `
package foo
type a int
type ( b int; c bar.Z; )
type (
    d uint
    e float64
)
`
	t, e := Parse("base-type-1.go", src)
	if e != nil {
		tst.Error(e)
	}

	exp := `package foo

type a int

type b int

type c bar.Z

type d uint

type e float64

`
	f := t.Format()
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestArrayType1(tst *testing.T) {
	src := `
package foo
type a [3]int
type ( b [][3]bar.T )

`
	t, e := Parse("array-type-1.go", src)
	if e != nil {
		tst.Error(e)
	}

	exp := `package foo

type a [3]int

type b [][3]bar.T

`
	f := t.Format()
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestPtrType(tst *testing.T) {
	src := `
package foo
type a *int
type b [3]*int
type ( b []*[3]bar.T )

`
	t, e := Parse("ptr-type.go", src)
	if e != nil {
		tst.Error(e)
	}

	exp := `package foo

type a *int

type b [3]*int

type b []*[3]bar.T

`
	f := t.Format()
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestMapType(tst *testing.T) {
	src := `
package foo
type a map[int]int
type b map[string][3]*int
type ( b []*[3]map[*int]bar.T )

`
	t, e := Parse("map-type.go", src)
	if e != nil {
		tst.Error(e)
	}

	exp := `package foo

type a map[int]int

type b map[string][3]*int

type b []*[3]map[*int]bar.T

`
	f := t.Format()
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestTypeError1(tst *testing.T) {
	src := `package foo
type a uint
type b
type c [string

`
	t, e := Parse("type-error-1.go", src)
	f := t.Format()
	tst.Log(f)
	if e == nil {
		tst.Error("Unexpected lack of error")
	} else {
		tst.Log(e)
	}
}

func TestTypeError2(tst *testing.T) {
	src := `package foo
type ( a uint; b[] ; c float64 )

`
	t, e := Parse("type-error-2.go", src)
	f := t.Format()
	tst.Log(f)
	if e == nil {
		tst.Error("Unexpected lack of error")
	} else {
		tst.Log(e)
	}
}
