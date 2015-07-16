package parser

import (
	"testing"
)

func TestTypeDeclStmt(tst *testing.T) {
	src := `package main
func foo() {
  type a uint
  type ( b string; c []string
        d float32; e struct { x, y float64 })
  type f rune
}
`
	exp := `package main

func foo() {
    type a uint
    type (
        b string
        c []string
        d float32
        e struct {
            x, y float64
        }
    )
    type f rune
}
`
	t, e := Parse("type-decl-stmt.go", src)
	if e != nil {
		tst.Error(e)
	}

	f := t.Format()
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestConstDeclStmt(tst *testing.T) {
	src := `package main
func foo() {
  const a = 0
  const ( b string = "b"; c = "c"
        d = 3.1415i; e float64 = 2.71)
  const f rune = 'α'
}
`
	exp := `package main

func foo() {
    const a = 0
    const (
        b string = "b"
        c = "c"
        d = 3.1415i
        e float64 = 2.71
    )
    const f rune = 'α'
}
`
	t, e := Parse("const-decl-stmt.go", src)
	if e != nil {
		tst.Error(e)
	}

	f := t.Format()
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestVarDeclStmt(tst *testing.T) {
	src := `package main
func foo() {
  var a uint
  var ( b string; c []int
        d,e float64)
  var f[]*float64
}
`
	exp := `package main

func foo() {
    var a uint
    var (
        b string
        c []int
        d, e float64
    )
    var f []*float64
}
`
	t, e := Parse("var-decl-stmt.go", src)
	if e != nil {
		tst.Error(e)
	}

	f := t.Format()
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestGoStmt(tst *testing.T) {
	src := `package main
func foo(fn func(int) int) {
   go fn(1)
}
`
	exp := `package main

func foo(fn func(int) int) {
    go fn(1)
}
`
	t, e := Parse("go-stmt.go", src)
	if e != nil {
		tst.Error(e)
	}

	f := t.Format()
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestReturnStmt(tst *testing.T) {
	src := `package main
func foo(fn func(int) int) int {
   return fn(1) + 1
}
func bar(s string, t string) (string, int, string) {
   return s + t,
          len( s ) + len( t ),
          s
}
func baz() {
  return
}
`
	exp := `package main

func foo(fn func(int) int) int {
    return fn(1) + 1
}
func bar(s string, t string) (string, int, string) {
    return s + t, len(s) + len(t), s
}
func baz() {
    return
}
`
	t, e := Parse("return-stmt.go", src)
	if e != nil {
		tst.Error(e)
	}

	f := t.Format()
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestBreakStmt(tst *testing.T) {
	src := `package main
func foo(fn func(int) int) int {
   break
   break L; break; break Q
}
`
	exp := `package main

func foo(fn func(int) int) int {
    break
    break L
    break
    break Q
}
`
	t, e := Parse("break-stmt.go", src)
	if e != nil {
		tst.Error(e)
	}

	f := t.Format()
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestContinueStmt(tst *testing.T) {
	src := `package main
func foo(fn func(int) int) int {
   continue
   continue L; continue; continue Q
}
`
	exp := `package main

func foo(fn func(int) int) int {
    continue
    continue L
    continue
    continue Q
}
`
	t, e := Parse("continue-stmt.go", src)
	if e != nil {
		tst.Error(e)
	}

	f := t.Format()
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestGotoStmt(tst *testing.T) {
	src := `package main
func foo(fn func(int) int) int {
   goto L
   goto L1; goto L2
}
`
	exp := `package main

func foo(fn func(int) int) int {
    goto L
    goto L1
    goto L2
}
`
	t, e := Parse("goto-stmt.go", src)
	if e != nil {
		tst.Error(e)
	}

	f := t.Format()
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestFallthroughStmt(tst *testing.T) {
	src := `package main
func foo(fn func(int) int) int {
   fallthrough
   fallthrough; fallthrough
}
`
	exp := `package main

func foo(fn func(int) int) int {
    fallthrough
    fallthrough
    fallthrough
}
`
	t, e := Parse("fallthrough-stmt.go", src)
	if e != nil {
		tst.Error(e)
	}

	f := t.Format()
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestSendStmt(tst *testing.T) {
	src := `package main
func foo(in, out chan int, ch chan (<-chan int), i int) {
  out <- i
  out <- <- in
  ch <- (<-chan int)(in)
  ch <- <-chan int (in)
}
`
	exp := `package main

func foo(in chan int, out chan int, ch chan (<-chan int), i int) {
    out <- i
    out <- <-in
    ch <- (<-chan int)(in)
    ch <- <-(chan int)(in)
}
`
	t, e := Parse("send-stmt.go", src)
	if e != nil {
		tst.Error(e)
	}

	f := t.Format()
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestIncDecStmt(tst *testing.T) {
	src := `package main
func foo(i int) int {
  i++; i--
 i--
 i++; i--
}
`
	exp := `package main

func foo(i int) int {
    i++
    i--
    i--
    i++
    i--
}
`
	t, e := Parse("send-stmt.go", src)
	if e != nil {
		tst.Error(e)
	}

	f := t.Format()
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestAssignStmt(tst *testing.T) {
	src := `package main
func foo() {
  x := 1
  y := 2
  x = y
  x = y + 1
  x, y = y,x+y
  x, y, 1 = y,x+y,z
     a, b := x+y, x - y
    a, b *= 2, 3
}
`
	exp := `package main

func foo() {
    x := 1
    y := 2
    x = y
    x = y + 1
    x, y = y, x + y
    x, y, 1 = y, x + y, z
    a, b := x + y, x - y
    a, b *= 2, 3
}
`
	t, e := Parse("assign-stmt.go", src)
	if e != nil {
		tst.Error(e)
	}

	f := t.Format()
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestAssignStmt1(tst *testing.T) {
	src := `package main
func foo() {
  fn := func(i int) bool {
      return i > 0
  }
}
`
	exp := `package main

func foo() {
    fn := func(i int) bool {
        return i > 0
    }
}
`
	t, e := Parse("assign-stmt-1.go", src)
	if e != nil {
		tst.Error(e)
	}

	f := t.Format()
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestEmptyStmt(tst *testing.T) {
	src := `package main
func foo() {
  {}
  ;;{};
  ;{;};
}
`
	exp := `package main

func foo() {
    {}
    {}
    {}
}
`
	t, e := Parse("empty-stmt.go", src)
	if e != nil {
		tst.Error(e)
	}

	f := t.Format()
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestIfStmt(tst *testing.T) {
	src := `package main
func foo() {
  if true {
 }
 if s := (S{1, 2}); s.x > 1 {
}
    if fn := func(i int) bool {	if s := (S{1, 2}); s.x < i { return true } else { return false } }; fn(s.y) {
		bar()
	}

p.next()
if p.token == ']' {
	p.next()
		return &ast.SliceType{p.parseType()}
 } else
   if p.token == s.DOTS {
		p.next()
	p.match(']')
		return &ast.ArrayType{Dim: nil,
                              EltType: p.parseType()}
	} else
 {
		e := p.parseExpr()
		p.match(']')
		t := p.parseType()
		return &ast.ArrayType{Dim: e,
                              EltType: t}
}}
`
	exp := `package main

func foo() {
    if true {}
    if s := S{1, 2}; s.x > 1 {}
    if fn := func(i int) bool {
        if s := S{1, 2}; s.x < i {
            return true
        } else {
            return false
        }
    }; fn(s.y) {
        bar()
    }
    p.next()
    if p.token == ']' {
        p.next()
        return &ast.SliceType{p.parseType()}
    } else if p.token == s.DOTS {
        p.next()
        p.match(']')
        return &ast.ArrayType{Dim: nil, EltType: p.parseType()}
    } else {
        e := p.parseExpr()
        p.match(']')
        t := p.parseType()
        return &ast.ArrayType{Dim: e, EltType: t}
    }
}
`
	t, e := Parse("if-stmt.go", src)
	if e != nil {
		tst.Error(e)
	}

	f := t.Format()
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}
