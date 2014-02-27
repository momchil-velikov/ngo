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

func TestTypeName(tst *testing.T) {
    src := `
package foo
type a int
type ( b int; c bar.Z; )
type (
    d uint
    e float64
)
`
    t, e := Parse("type-name-1.go", src)
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
type ( c [...][3]bar.T )
type ( d [...][3][...]bar.T )
`
    t, e := Parse("array-type-1.go", src)
    if e != nil {
        tst.Error(e)
    }

    exp := `package foo

type a [3]int

type b [][3]bar.T

type c [...][3]bar.T

type d [...][3][...]bar.T

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

func TestChanType(tst *testing.T) {
    src := `package foo
type a chan int
type b <-chan uint
type c chan<- float32
type d chan<- chan uint
type e <-chan chan map[uint][]float64
type f chan (<-chan uint)
type g chan<- <-chan uint
`
    t, e := Parse("chan-type.go", src)
    if e != nil {
        tst.Error(e)
    }

    exp := `package foo

type a chan int

type b <-chan uint

type c chan<- float32

type d chan<- chan uint

type e <-chan chan map[uint][]float64

type f chan (<-chan uint)

type g chan<- <-chan uint

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

func TestStructType(tst *testing.T) {
    src := `package foo
    type S struct { a uint; b, c *float64 "field1" 
                    bar.Z ; S "foo"; *T "bar" ;
                    c <-chan string; e, f string
                    g []struct { x, y float64; z struct {zn, zf float32} }
                    h struct {}
    }
    type T struct { S }
`
    exp := `package foo

type S struct {
    a uint
    b *float64 "field1"
    c *float64 "field1"
    bar.Z
    S "foo"
    *T "bar"
    c <-chan string
    e string
    f string
    g []struct {
        x float64
        y float64
        z struct {
            zn float32
            zf float32
        }
    }
    h struct{}
}

type T struct {
    S
}

`
    t, e := Parse("struct-type-1.go", src)
    if e != nil {
        tst.Error(e)
    }

    f := t.Format()
    if f != exp {
        tst.Errorf("Error output:\n->|%s|<-\n", f)
    }
}

func TestStructTypeError(tst *testing.T) {
    src := `package foo
type S struct { a uint; b, *float64 "field1" 
bar.Z ; S "foo"
ch -chan string; e, f string
g []struct { x, y float64; z struct zn, zf float32} }
}
`
    exp := `package foo

type S struct {
    a uint
    b *float64 "field1"
    bar.Z
    S "foo"
    e string
    f string
    g []struct {
        x float64
        y float64
        z struct {
            zn float32
            zf float32
        }
    }
}

`
    t, e := Parse("struct-type-error.go", src)
    if e != nil {
        tst.Log(e)
    }

    f := t.Format()
    if f != exp {
        tst.Errorf("Error output:\n->|%s|<-\n", f)
    }

}

func TestFuncType(tst *testing.T) {
    src := `package foo
type T2 func()
type T3 func(x int) int
type T4 func(a, _ int, z float32) bool
type T5 func(a, b int, z float32) (bool)
type T6 func(prefix string, values ...int)
type T7 func(a, b int, z float64, opt ...struct{}) (success bool)
type T8 func(int, int, float64) (float64, *[]int)
type T9 func(n int) func(p *T)
`
    exp := `package foo

type T2 func()

type T3 func(x int) int

type T4 func(a int, _ int, z float32) bool

type T5 func(a int, b int, z float32) bool

type T6 func(prefix string, values ...int)

type T7 func(a int, b int, z float64, opt ...struct{}) (success bool)

type T8 func(int, int, float64) (float64, *[]int)

type T9 func(n int) func(p *T)

`
    t, e := Parse("func-type-1.go", src)
    if e != nil {
        tst.Error(e)
    }

    f := t.Format()
    if f != exp {
        tst.Errorf("Error output:\n->|%s|<-\n", f)
    }
}

func TestFuncTypeError(tst *testing.T) {
    src := `package foo
type T1 func(uint
type T2 func (P.t, a func), )
type T3 func (struct  a uint; b } , ... struct{})
type T4 func (a, b [4]*, c uint)
type T5 func (a, [4]*, c uint, d.)
`
    exp := `package foo

type T1 func(uint)

type T2 func(P.t, a func())

type T3 func(struct {
        a uint
        b
    }, ...struct{})

type T4 func(c uint)

type T5 func(a uint, c uint, d)

`
    t, e := Parse("func-type-error.go", src)
    if e != nil {
        tst.Log(e)
    }

    f := t.Format()
    if f != exp {
        tst.Errorf("Error output:\n->|%s|<-\n", f)
    }

}

func TestInterfaceType(tst *testing.T) {
    src := `package foo
type (
    T3 interface{}
    T4 interface {T3; Foo();  Bar(a interface{});  }
    T5 interface {
        Baz (T3, T4) (r foo.T5)
        foo.T3
        T4
    }
)

`
    exp := `package foo

type T3 interface{}

type T4 interface {
    T3
    Foo()
    Bar(a interface{})
}

type T5 interface {
    foo.T3
    T4
    Baz(T3, T4) (r foo.T5)
}

`
    t, e := Parse("interface-type-1.go", src)
    if e != nil {
        tst.Error(e)
    }

    f := t.Format()
    if f != exp {
        tst.Errorf("Error output:\n->|%s|<-\n", f)
    }
}

func TestConstDecl(tst *testing.T) {
    src := `package foo

const a = 1
const a, b = 2
const a float32 = 4; const a, b float64 = 2, 5
const c
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
const c
const (
    a = 1
    a, b = 2
    a float32 = 4
    a, b float64 = 2, 5
    c
    d
)
`
    t, e := Parse("const-decl.go", src)
    if e != nil {
        tst.Error(e)
    }

    f := t.Format()
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
const (
    a = 
    a,  = 2
    a float32 = 4; , b float64 = 2 5
    c; d
)

`
    exp := `package foo

const a
const a = 2
const a float32 = 4
const b float64 = 2
const c
const (
    a
    a = 2
    a float32 = 4
    b float64 = 2
    c
    d
)
`
    t, e := Parse("const-decl-error.go", src)
    if e != nil {
        tst.Log(e)
    }

    f := t.Format()
    if f != exp {
        tst.Errorf("Error output:\n->|%s|<-\n", f)
    }

}

func TestVarDecl(tst *testing.T) {
    src := `package foo

var a = 1
var a, b = 2
var a float32 = 4; var a, b float64 = 2, 5
var c
var (
    a = 1
    a, b = 2
    a float32 = 4; a, b float64 = 2, 5
    c; d
)
var()
`
    exp := `package foo

var a = 1
var a, b = 2
var a float32 = 4
var a, b float64 = 2, 5
var c
var (
    a = 1
    a, b = 2
    a float32 = 4
    a, b float64 = 2, 5
    c
    d
)
`
    t, e := Parse("var-decl.go", src)
    if e != nil {
        tst.Error(e)
    }

    f := t.Format()
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

var a
var a = 2
var a float32 = 4
var b float64 = 2
var c
var (
    a
    a = 2
    a float32 = 4
    b float64 = 2
    c
    d
)
`
    t, e := Parse("var-decl-error.go", src)
    if e != nil {
        tst.Log(e)
    }

    f := t.Format()
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
`
    exp := `package foo

func f()
func g() {
}
func h(a uint, b uint) bool
func i(a uint, b uint) (r0 []*X, r1 bool)
func (r T) f()
func (r *T) g() {
}
func (*T) h(a uint, b uint) bool
func (r *T) i(a uint, b uint) (r0 []*X, r1 bool) {
}
`
    t, e := Parse("func-decl.go", src)
    if e != nil {
        tst.Error(e)
    }

    f := t.Format()
    if f != exp {
        tst.Errorf("Error output:\n->|%s|<-\n", f)
    }
}

func TestFuncDeclError(tst *testing.T) {
    src := `package foo

func ()
func g() {
func h( a,  uint) bool
func i( , b uint) (r0 []*X r1 bool)

func (r T) f
func (r *T) g() {
func (*) h( , b uint) bool
func () i( a,  uint) (r0 []*X r1 bool) 
}
`
    exp := `package foo

func g() {
}
func h(a, uint) bool
func i(b uint) (r0 []*X, r1 bool)
func (r T) f()
func (r *T) g() {
}
func h(b uint) bool
func i(a, uint) (r0 []*X, r1 bool)
`
    t, e := Parse("func-decl-error.go", src)
    if e != nil {
        tst.Log(e)
    }

    f := t.Format()
    if f != exp {
        tst.Errorf("Error output:\n->|%s|<-\n", f)
    }
}
