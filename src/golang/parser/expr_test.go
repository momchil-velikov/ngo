package parser

import (
	"golang/ast"
	"testing"
)

func TestExpr1(tst *testing.T) {
	src := `package foo
var a = 1 + -1 * (-3 - -4 * 5)
var b = (1 + -1) * (-3 - -4 * 5)
var c = 1 + -1 + (-3 - -4 * 5)
`

	exp := `package foo

var a = 1 + -1 * (-3 - -4 * 5)
var b = (1 + -1) * (-3 - -4 * 5)
var c = 1 + -1 + (-3 - -4 * 5)
`
	t, e := Parse("expr-1.go", src)
	if e != nil {
		tst.Error(e)
	}

	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestExpr2(tst *testing.T) {
	src := `package foo

var (
    a = b > b1 && c || d && e <= e1 && f || g
)
`
	t, e := Parse("expr-1.go", src)
	if e != nil {
		tst.Error(e)
	}

	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
	if f != src {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestPrimaryExpr1(tst *testing.T) {
	src := `package foo

var (
    a = 1
    b = a
    c = foo.b
    d = S{}
    e = foo.F()
)
`
	t, e := Parse("primary-expr-1.go", src)
	if e != nil {
		tst.Error(e)
	}

	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
	if f != src {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestPrimaryExpr2(tst *testing.T) {
	src := `package foo

var (
    a = *uint (1)
    b = (*uint)(a)
    c = ([]uint)(b)
    cc = []uint(b)
    d = (*[]uint)(c)
    e = []uint{}
    f = struct { x, y uint}{}
    g = struct { x, y uint}(f)
    gg = (struct { x, y uint})(gg)
    h = map[uint]struct{ x, y float64}{}
    i = map[uint]struct{ x, y float64}(h)
    j = (map[uint]struct{ x, y float64})(i)
    k = func (uint, uint) (float64, bool){}
    l = func (uint, uint) (float64, bool)(k)
    m  = func (uint) {}
    n = (func (uint)) (m)
    n = (func (uint)) (m,)
    o = interface{ foo() uint; bar(uint) (uint, bool)} (n)
    p = <-chan uint (o)
    q = (<-chan uint)(p)
    r = chan uint (q)
    s = (chan uint)(r)
    t = a{}
    u = a.b{}
    v = s.(chan uint)
    w = a.b.([4]struct{x, y uint})
    x = (*pkg.T).M
    y = s.(type)
)
`
	exp := `package foo

var (
    a = *uint(1)
    b = (*uint)(a)
    c = []uint(b)
    cc = []uint(b)
    d = (*[]uint)(c)
    e = []uint{}
    f = struct {
            x, y uint
        }{}
    g = struct {
            x, y uint
        }(f)
    gg = struct {
            x, y uint
        }(gg)
    h = map[uint]struct {
            x, y float64
        }{}
    i = map[uint]struct {
            x, y float64
        }(h)
    j = map[uint]struct {
            x, y float64
        }(i)
    k = func(uint, uint) (float64, bool) {}
    l = func(uint, uint) (float64, bool)(k)
    m = func(uint) {}
    n = (func(uint))(m)
    n = (func(uint))(m)
    o = interface {
            foo() uint
            bar(uint) (uint, bool)
        }(n)
    p = <-(chan uint)(o)
    q = (<-chan uint)(p)
    r = (chan uint)(q)
    s = (chan uint)(r)
    t = a{}
    u = a.b{}
    v = s.(chan uint)
    w = a.b.([4]struct {
            x, y uint
        })
    x = (*pkg.T).M
    y = s.(type)
)
`
	t, e := Parse("primary-expr-2.go", src)
	if e != nil {
		tst.Error(e)
	}

	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestPrimaryExpr3(tst *testing.T) {
	src := `package foo

var (
    a = foo()
    b = foo(a)
    b = foo(a,)
    b = foo(a...,)
    c = bar([]uint)
    c = bar([]uint...)
    c = bar([]uint...,)
    d = baz(a, b, c)
    d = baz(a, b, c,)
    d = baz(a, b, c...)
    d = baz(a, b, c...,)
    d = xyzzy([]uint, b, c)
    d = xyzzy([]uint, b, c,)
    d = xyzzy([]uint, b, c...)
    d = xyzzy([]uint, b, c...,)
    e = d[a + b]
    f = e[:]
    f = e[:h]
    f = e[:h:c+c2]
    f = e[i + b - b]
    f = e[l:]
    f = e[l+i%2:h]
    f = e[l:h:c/2]
)
`
	exp := `package foo

var (
    a = foo()
    b = foo(a)
    b = foo(a)
    b = foo(a...)
    c = bar([]uint)
    c = bar([]uint...)
    c = bar([]uint...)
    d = baz(a, b, c)
    d = baz(a, b, c)
    d = baz(a, b, c...)
    d = baz(a, b, c...)
    d = xyzzy([]uint, b, c)
    d = xyzzy([]uint, b, c)
    d = xyzzy([]uint, b, c...)
    d = xyzzy([]uint, b, c...)
    e = d[a + b]
    f = e[:]
    f = e[: h]
    f = e[: h : c + c2]
    f = e[i + b - b]
    f = e[l :]
    f = e[l + i % 2 : h]
    f = e[l : h : c / 2]
)
`
	t, e := Parse("primary-expr-3.go", src)
	if e != nil {
		tst.Error(e)
	}

	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestCompLiteral(tst *testing.T) {
	src := `package foo

type Point3D struct { x, y, z float64 }
type Line struct { p, q Point3D }

var (
    origin = Point3D{}
    line = Line{origin, Point3D{y: -4, z: 12.3}}
    line = Line{origin, {y: -4, z: 12.3}, }
    pointer *Point3D = &Point3D{y: 1000}
    buffer = [10]string{}
    intSet = [6]int{1, 2, 3, 5}
    days = [...]string{"Sat", "Sun"}
    u = [...]Point{{1.5, -3.5}, {0, 0}}
    v1 = [][]int{{1, 2, 3}, {4, 5}}
    v2 = [][]int{[]int{1, 2, 3}, []int{4, 5}}
    w1 = [...]*Point{{1.5, -3.5}, {0, 0}}
    w1 = [...]*Point{&Point{1.5, -3.5}, &Point{0, 0}}
)
`
	exp := `package foo

type Point3D struct {
    x, y, z float64
}
type Line struct {
    p, q Point3D
}
var (
    origin = Point3D{}
    line = Line{origin, Point3D{y: -4, z: 12.3}}
    line = Line{origin, {y: -4, z: 12.3}}
    pointer *Point3D = &Point3D{y: 1000}
    buffer = [10]string{}
    intSet = [6]int{1, 2, 3, 5}
    days = [...]string{"Sat", "Sun"}
    u = [...]Point{{1.5, -3.5}, {0, 0}}
    v1 = [][]int{{1, 2, 3}, {4, 5}}
    v2 = [][]int{[]int{1, 2, 3}, []int{4, 5}}
    w1 = [...]*Point{{1.5, -3.5}, {0, 0}}
    w1 = [...]*Point{&Point{1.5, -3.5}, &Point{0, 0}}
)
`
	t, e := Parse("comp-literal.go", src)
	if e != nil {
		tst.Error(e)
	}

	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestBug20150929T183739(tst *testing.T) {
	src := `package p
var a = "123"[1]
`
	exp := `package p

var a = "123"[1]
`
	t, e := Parse("bug-2015-09-29T18:37:39.go", src)
	if e != nil {
		tst.Log(e)
	}

	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}
