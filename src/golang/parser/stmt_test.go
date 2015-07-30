package parser

import (
	"golang/ast"
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

	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
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

	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
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

	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
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

	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
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

	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
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

	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
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

	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
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

	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
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

	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
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

func foo(in, out chan int, ch chan (<-chan int), i int) {
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

	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
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

	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
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

	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
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

	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
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

	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
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
    if s := (S{1, 2}); s.x > 1 {}
    if fn := func(i int) bool {
        if s := (S{1, 2}); s.x < i {
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

	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestForStmt(tst *testing.T) {
	src := `package main

func foo() {
for {
  loop()
}
for cond {
  loop()
}
for init(); ; {
  loop()
}
for ; cond() ; {
  loop()
}
for ; ; post() {
  loop()
}
for init(); cond(); {
  loop()
}
for init() ; ; post() {
  baz()
}
for ; cond() ; post() {
  baz()
}
for init() ; cond() ; post() {
  baz()
}
for p.token != s.EOF && p.token != t1 && p.token != t2 {
  p.next()
}
for s:= (S{1, 2}); s.x > 0; tweak(&s) {
  loop(s)
}
for x, y := 1, 1; x < n; x, y = y, x + y {
  fmt.Println(x)
}
for i := range a {
  fmt.Println(a[i])
}
for i, v := range a {
  fmt.Println(i, v)
}
for _, v := range a {
  fmt.Println(v)
}
i := 0
for range a {  fmt.Println(i, a[i]);i++}
}
`
	exp := `package main

func foo() {
    for {
        loop()
    }
    for cond {
        loop()
    }
    for init(); ; {
        loop()
    }
    for cond() {
        loop()
    }
    for ; ; post() {
        loop()
    }
    for init(); cond(); {
        loop()
    }
    for init(); ; post() {
        baz()
    }
    for ; cond(); post() {
        baz()
    }
    for init(); cond(); post() {
        baz()
    }
    for p.token != s.EOF && p.token != t1 && p.token != t2 {
        p.next()
    }
    for s := (S{1, 2}); s.x > 0; tweak(&s) {
        loop(s)
    }
    for x, y := 1, 1; x < n; x, y = y, x + y {
        fmt.Println(x)
    }
    for i := range a {
        fmt.Println(a[i])
    }
    for i, v := range a {
        fmt.Println(i, v)
    }
    for _, v := range a {
        fmt.Println(v)
    }
    i := 0
    for range a {
        fmt.Println(i, a[i])
        i++
    }
}
`
	t, e := Parse("for-stmt.go", src)
	if e != nil {
		tst.Error(e)
	}

	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestDeferStmt(tst *testing.T) {
	src := `package main
func foo(i int) int {
defer bar()
}
`
	exp := `package main

func foo(i int) int {
    defer bar()
}
`
	t, e := Parse("send-stmt.go", src)
	if e != nil {
		tst.Error(e)
	}

	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestExprSwitchStmt(tst *testing.T) {
	src := `package main

func foo() {
  switch {}
  switch x + 1 {}
  switch x:= foo(); x + 1 {}
  switch x= foo(); x + 1 {}
  switch x + 1; y + 1 {}
  switch <-ch; {}
  switch <-ch {}
  switch t := <-ch; t {}
  switch { case x < 0: {return -1; ;}; case x == 0: return 0; default: return 1 }
}
`
	exp := `package main

func foo() {
    switch {}
    switch x + 1 {}
    switch x := foo(); x + 1 {}
    switch x = foo(); x + 1 {}
    switch x + 1; y + 1 {}
    switch <-ch; {}
    switch <-ch {}
    switch t := <-ch; t {}
    switch {
    case x < 0:
        {
            return -1
        }
    case x == 0:
        return 0
    default:
        return 1
    }
}
`
	t, e := Parse("expr-switch-stmt.go", src)
	if e != nil {
		tst.Error(e)
	}

	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}

}

func TestTypeSwitchStmt(tst *testing.T) {
	src := `package main

func foo() {
  switch x.(type) {}
  switch t := x.(type) {}
  switch x:= foo(); t := x.(type) {}
switch i := x.(type) {
case nil:printString("x is nil")
case int:printInt(i)
case float64:printFloat64(i)
case func(int) float64:	printFunction(i)
case bool, string:	printString("type is bool or string")
default:	printString("don't know the type")
}
}
`
	exp := `package main

func foo() {
    switch x.(type) {}
    switch t := x.(type) {}
    switch x := foo(); t := x.(type) {}
    switch i := x.(type) {
    case nil:
        printString("x is nil")
    case int:
        printInt(i)
    case float64:
        printFloat64(i)
    case func(int) float64:
        printFunction(i)
    case bool, string:
        printString("type is bool or string")
    default:
        printString("don't know the type")
    }
}
`
	t, e := Parse("type-switch-stmt.go", src)
	if e != nil {
		tst.Error(e)
	}

	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestSelectStmt(tst *testing.T) {
	src := `package main

func foo() {
  select{}
  select{
    case ch1 <-ex:
      bar()
    default:
      baz()
    case x, y = <- ch2:
      xyzzy();;
    case u, v:= <- ch3:
      quux()
    case x++:
;
 case <- ch3:
  }
}
`
	exp := `package main

func foo() {
    select {}
    select {
    case ch1 <- ex:
        bar()
    default:
        baz()
    case x, y = <-ch2:
        xyzzy()
    case u, v := <-ch3:
        quux()
    case x++:
    case <-ch3:
    }
}
`
	t, e := Parse("select-switch-stmt.go", src)
	if e != nil {
		tst.Error(e)
	}

	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestLabeledStmt(tst *testing.T) {
	src := `package main

func foo(n uint) {
L1: n++
  if n > 2 {
     L2: n--
  }
}
`

	exp := `package main

func foo(n uint) {
L1:
    n++
    if n > 2 {
    L2:
        n--
    }
}
`
	t, e := Parse("labeled-stmt.go", src)
	if e != nil {
		tst.Error(e)
	}

	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}

func TestLabeledEmptyStmt(tst *testing.T) {
	src := `package main

func f() {
L1: goto L1; L2: L3: ; goto L2
}
`

	exp := `package main

func f() {
L1:
    goto L1
L2:
L3:
    goto L2
}
`
	t, e := Parse("labeled-empty-stmt.go", src)
	if e != nil {
		tst.Error(e)
	}

	ctx := new(ast.FormatContext).Init()
	f := t.Format(ctx)
	if f != exp {
		tst.Errorf("Error output:\n->|%s|<-\n", f)
	}
}
