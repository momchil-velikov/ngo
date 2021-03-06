package parser

import (
	"golang/ast"
	"strings"
	"testing"
)

func TestPackageClause(tst *testing.T) {
	src := `
package foo`
	t, e := Parse("package-clause.go", src)

	if e != nil {
		tst.Error(e)
	}

	if t.PkgName != "foo" {
		tst.Errorf("Unexpected package name `%s`", t.PkgName)
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

func TestPackageError3(tst *testing.T) {
	src := `
package _`
	_, e := Parse("package-blank-errror.go", src)

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

	if t.PkgName != "foo" {
		tst.Errorf("Unexpected package name `%s`", t.PkgName)
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

	if t.PkgName != "foo" {
		tst.Errorf("Unexpected package name `%s`", t.PkgName)
	}

	exp := `package foo

import "fmt"
import . "foo"
import . "bar"
import s "baz"
`
	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
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

import "bar"
import "baz"
import "xyzzy"
import "qux"
`
	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
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
	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
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
	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
	tst.Log(f)
	if e != nil {
		tst.Log(e)
	}
}

func TestConstDecl(tst *testing.T) {
	src := `package foo

const a = 1
const a, b = 2
const a float32 = 4; const a, b float64 = 2, 5
const c = "aaa"
const (
    a = 1
    a, b = 2
    a float32 = 4; a, b float64 = 2, 5
    c; d
)
const()
`
	exp := `package foo

const a = 1
const a, b = 2
const a float32 = 4
const a, b float64 = 2, 5
const c = "aaa"
const (
    a = 1
    a, b = 2
    a float32 = 4
    a, b float64 = 2, 5
    c
    d
)
const (
)
`
	t, e := Parse("const-decl.go", src)
	if e != nil {
		tst.Error(e)
	}

	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestConstDeclError(tst *testing.T) {
	src := `package foo
const a = 
const a,  = 2
const a float32 = 4; const , b float64 = 2 5
const c
const d int
const (
    u = 
    v,  = 2
    u float32 = 4; , v float64 = 2 5
    w; x
)
`
	exp := `package foo

const a = <error>
const a = 2
const a float32 = 4
const , b float64 = 2
const c = <error>
const d int = <error>
const (
    u = v, <error>
    u float32 = 4
    , v float64 = 2
    w
    x
)
`
	t, e := Parse("const-decl-error.go", src)
	if e != nil {
		tst.Log(e)
	}

	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}

}

func TestVarDecl(tst *testing.T) {
	src := `package foo

var a = 1
var a, b = 2
var a float32 = 4; var a, b float64 = 2, 5
var c = 1
var (
    a = 1
    a, b = 2
    a float32 = 4; a, b float64 = 2, 5
    c int; d = 1
)
var()
`
	exp := `package foo

var a = 1
var a, b = 2
var a float32 = 4
var a, b float64 = 2, 5
var c = 1
var (
    a = 1
    a, b = 2
    a float32 = 4
    a, b float64 = 2, 5
    c int
    d = 1
)
var (
)
`
	t, e := Parse("var-decl.go", src)
	if e != nil {
		tst.Error(e)
	}

	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestVarDeclError(tst *testing.T) {
	src := `package foo
var a = 
var a,  = 2
var a float32 = 4; var , b float64 = 2 5
var c
var (
    a = 
    a,  = 2
    a float32 = 4; , b float64 = 2 5
    c; d
)
`
	exp := `package foo

var a = <error>
var a = 2
var a float32 = 4
var , b float64 = 2
var c <error>
var (
    a = a, <error>
    a float32 = 4
    , b float64 = 2
    c <error>
    d <error>
)
`
	t, e := Parse("var-decl-error.go", src)
	if e != nil {
		tst.Log(e)
	}

	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestFuncDecl(tst *testing.T) {
	src := `package foo

func f()
func g() {}
func h( a, b uint) bool
func i( a, b uint) (r0 []*X, r1 bool)

func (r T) f()
func (r *T) g() {}
func (*T) h( a, b uint) bool
func (r *T) i( a, b uint) (r0 []*X, r1 bool) {
}
func (r (*(T)),) f()
`
	exp := `package foo

func f()
func g() {}
func h(a, b uint) bool
func i(a, b uint) (r0 []*X, r1 bool)
func (r T) f()
func (r *T) g() {}
func (*T) h(a, b uint) bool
func (r *T) i(a, b uint) (r0 []*X, r1 bool) {}
func (r *T) f()
`
	t, e := Parse("func-decl.go", src)
	if e != nil {
		tst.Error(e)
	}

	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestFuncDeclError(tst *testing.T) {
	src := `package foo

func func ()
func h( a,  uint) bool
func i( , b uint) (r0 []*X r1 bool)
func (r T) f
func (*) h( , b uint) bool
func () i( a,  uint) (r0 []*X r1 bool)
}
`

	exp := `package foo

func (func(), uint) bool
func i(<error>, b uint) (r0 []*X)
func (r T) f(<error>) h
func (<error>) i(a, uint) (r0 []*X)
`

	t, e := Parse("func-decl-error.go", src)
	if e != nil {
		tst.Log(e)
	}

	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestBug20150730T124314(tst *testing.T) {
	src := `package p
var (
  a = func() { var b = func() { var c = 1} }
)`

	exp := `package p

var (
    a = func() {
            var b = func() {
                    var c = 1
                }
        }
)
`
	t, e := Parse("bug-2015-07-30T12:43:14.go", src)
	if e != nil {
		tst.Log(e)
	}

	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestBug20150929T234411(tst *testing.T) {
	src := `package p
func (id) name() {}
`
	exp := `package p

func (id) name() {}
`

	t, e := Parse("bug-2015-09-29T23:44:11.go", src)
	if e != nil {
		tst.Log(e)
	}

	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestParsePackage(t *testing.T) {
	const TESTDIR = "_test/pkgname/hedgp"

	// Test error if no source files.
	up, err := ParsePackage(TESTDIR, nil)
	if err == nil || !strings.Contains(err.Error(), "no source files") {
		t.Error("should return `no soutce files`` error")
	}

	// Test consistent package naming.
	up, err = ParsePackage(TESTDIR, []string{"a.go", "b.go"})
	if err != nil {
		t.Error(err)
	}
	if up.Name != "hedgp" {
		t.Error("incorrect package name: expected 'hedgp'")
	}

	// Test consistent package name 'main'.
	up, err = ParsePackage(TESTDIR, []string{"c.go", "d.go"})
	if err != nil {
		t.Error(err)
	}
	if up.Name != "main" {
		t.Error("incorrect package name: expected 'main'")
	}

	// Test inconsistent 'main'
	up, err = ParsePackage(TESTDIR, []string{"a.go", "b.go", "c.go"})
	if err == nil || !strings.Contains(err.Error(), "inconsistent package name") {
		if err != nil {
			t.Log(err)
		}
		t.Error("expected 'inconsistent package name' error")
	}

	// Test inconsistent other name.
	up, err = ParsePackage(TESTDIR, []string{"a.go", "b.go", "e.go"})
	if err == nil || !strings.Contains(err.Error(), "inconsistent package name") {
		if err != nil {
			t.Log(err)
		}
		t.Error("expected 'inconsistent package name' error")
	}

	up, err = ParsePackage(TESTDIR, []string{"c.go", "d.go", "e.go"})
	if err == nil || !strings.Contains(err.Error(), "inconsistent package name") {
		if err != nil {
			t.Log(err)
		}
		t.Error("expected 'inconsistent package name' error")
	}

	// Test source file not found.
	up, err = ParsePackage(TESTDIR, []string{"a.go", "ff.go"})
	if err == nil || !strings.Contains(err.Error(), "no such file") {
		if err != nil {
			t.Log(err)
		}
		t.Error("expected 'no such file' error")
	}

	// Test source file contains errors.
	up, err = ParsePackage(TESTDIR, []string{"a.go", "f.go"})
	if err == nil || !strings.Contains(err.Error(), "expected package") {
		if err != nil {
			t.Log(err)
		}
		t.Error("expected syntax error")
	}
}

func TestBug20160202T173058(t *testing.T) {
	src := `package p
var a
`
	_, e := Parse("bug-2016-02-02T17:30:58.go", src)
	if e == nil || !strings.Contains(e.Error(), "expected typespec") {
		t.Error("expected error `expected typespec`")
		if e == nil {
			t.Error("actual: no error")
		} else {
			t.Errorf("actual: %s\n", e.Error())
		}
	}
}
