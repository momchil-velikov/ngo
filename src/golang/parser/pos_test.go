package parser

import (
	"golang/ast"
	"testing"
)

func TestImportPosition(tst *testing.T) {
	src := `package main

import "fmt"
import (
  "a"
  . "b"
)
`
	exp := `package main

import /* #21 */"fmt"
import /* #38 */"a"
import /* #44 */. "b"
`
	if t, e := Parse("import-group-pos.go", src); e == nil {
		ctx := new(ast.FormatContext).Init()
		ctx.EmitSourcePositions(ast.IdentPos)
		f := t.Format(ctx)
		if f != exp {
			tst.Error(f)
		}
	} else {
		tst.Error(e)
	}
}

func TestLiteralPosition(tst *testing.T) {
	src := `package main

const (
  a = 12
  b = 1.12
  c = 'χ'
  d = 1i
  e = "1"
)
`
	exp := `package main

const (
    a = /* #28 */12
    b = /* #37 */1.12
    c = /* #48 */'χ'
    d = /* #59 */1.0i
    e = /* #68 */"1"
)
`
	if t, e := Parse("literal-pos.go", src); e == nil {
		ctx := new(ast.FormatContext).Init()
		ctx.EmitSourcePositions(ast.ExprPos)
		f := t.Format(ctx)
		if f != exp {
			tst.Error(f)
		}
	} else {
		tst.Error(e)
	}
}

func TestParensExprPosition(tst *testing.T) {
	src := `package main
var a = (b+c)/d - x/((u+v)*(u-v))
`
	exp := `package main

var a = /* #21 */(b + c) / d - x / /* #33 */(/* #34 */(u + v) * /* #40 */(u - v))
`
	if t, e := Parse("parens-expr-pos.go", src); e == nil {
		ctx := new(ast.FormatContext).Init()
		ctx.EmitSourcePositions(ast.ExprPos)
		f := t.Format(ctx)
		if f != exp {
			tst.Error(f)
		}
	} else {
		tst.Error(e)
	}
}

func TestUnaryExprPosition(tst *testing.T) {
	src := `package main
var a, b, c, d = +e, -f, *g, <-h
`
	exp := `package main

var a, b, c, d = /* #30 */+e, /* #34 */-f, /* #38 */*g, /* #42 */<-h
`
	if t, e := Parse("unary-expr-pos.go", src); e == nil {
		ctx := new(ast.FormatContext).Init()
		ctx.EmitSourcePositions(ast.ExprPos)
		f := t.Format(ctx)
		if f != exp {
			tst.Error(f)
		}
	} else {
		tst.Error(e)
	}
}

func TestArrayTypePosition(tst *testing.T) {
	src := `package main
type a [N]int
type b [N][M]float64
`
	exp := `package main

type a /* #20 */[N]int
type b /* #34 */[N]/* #37 */[M]float64
`
	if t, e := Parse("array-type-pos.go", src); e == nil {
		ctx := new(ast.FormatContext).Init()
		ctx.EmitSourcePositions(ast.TypePos)
		f := t.Format(ctx)
		if f != exp {
			tst.Error(f)
		}
	} else {
		tst.Error(e)
	}
}

func TestSliceTypePosition(tst *testing.T) {
	src := `package main
type a []int
type b [][]float64
`
	exp := `package main

type a /* #20 */[]int
type b /* #33 */[]/* #35 */[]float64
`
	if t, e := Parse("slice-type-pos.go", src); e == nil {
		ctx := new(ast.FormatContext).Init()
		ctx.EmitSourcePositions(ast.TypePos)
		f := t.Format(ctx)
		if f != exp {
			tst.Error(f)
		}
	} else {
		tst.Error(e)
	}
}

func TestPtrTypePosition(tst *testing.T) {
	src := `package main
type a *int
type b **float64
`
	exp := `package main

type a /* #20 */*int
type b /* #32 */*/* #33 */*float64
`
	if t, e := Parse("slice-type-pos.go", src); e == nil {
		ctx := new(ast.FormatContext).Init()
		ctx.EmitSourcePositions(ast.TypePos)
		f := t.Format(ctx)
		if f != exp {
			tst.Error(f)
		}
	} else {
		tst.Error(e)
	}
}

func TestMapTypePosition(tst *testing.T) {
	src := `package main
type a map[int]int
type b map[int]map[int]float64
`
	exp := `package main

type a /* #20 */map[int]int
type b /* #39 */map[int]/* #47 */map[int]float64
`
	if t, e := Parse("map-type-pos.go", src); e == nil {
		ctx := new(ast.FormatContext).Init()
		ctx.EmitSourcePositions(ast.TypePos)
		f := t.Format(ctx)
		if f != exp {
			tst.Error(f)
		}
	} else {
		tst.Error(e)
	}
}

func TestChanTypePosition(tst *testing.T) {
	src := `package main
type a <-chan int
type b chan <-chan int
type c chan (<-chan int)
`
	exp := `package main

type a /* #20 */<-chan int
type b /* #38 */chan<- /* #45 */chan int
type c /* #61 */chan (/* #67 */<-chan int)
`
	if t, e := Parse("chan-type-pos.go", src); e == nil {
		ctx := new(ast.FormatContext).Init()
		ctx.EmitSourcePositions(ast.TypePos)
		f := t.Format(ctx)
		if f != exp {
			tst.Error(f)
		}
	} else {
		tst.Error(e)
	}
}

func TestStructTypePosition(tst *testing.T) {
	src := `package main
type a struct { f int }
type b struct {
  f struct { f int }
}
`
	exp := `package main

type a /* #20 */struct {
    f int
}
type b /* #44 */struct {
    f /* #57 */struct {
        f int
    }
}
`
	if t, e := Parse("struct-type-pos.go", src); e == nil {
		ctx := new(ast.FormatContext).Init()
		ctx.EmitSourcePositions(ast.TypePos)
		f := t.Format(ctx)
		if f != exp {
			tst.Error(f)
		}
	} else {
		tst.Error(e)
	}
}

func TestFuncTypePosition(tst *testing.T) {
	src := `package main
type (
  a func()
  b func() func(int) int
  c func() func(int, func(int) int) int
)
`
	exp := `package main

type (
    a /* #24 */func()
    b /* #35 */func() /* #42 */func(int) int
    c /* #60 */func() /* #67 */func(int, /* #77 */func(int) int) int
)
`
	if t, e := Parse("struct-type-pos.go", src); e == nil {
		ctx := new(ast.FormatContext).Init()
		ctx.EmitSourcePositions(ast.TypePos)
		f := t.Format(ctx)
		if f != exp {
			tst.Error(f)
		}
	} else {
		tst.Error(e)
	}
}

func TestInterfaceTypePosition(tst *testing.T) {
	src := `package foo
type (
    T3 interface{}
    T4 interface {T3; Foo();  Bar(a interface{});  }
)
`
	exp := `package foo

type (
    T3 /* #26 */interface{}
    T4 /* #45 */interface {
        T3
        Foo()
        Bar(a /* #74 */interface{})
    }
)
`
	if t, e := Parse("interface-type-pos.go", src); e == nil {
		ctx := new(ast.FormatContext).Init()
		ctx.EmitSourcePositions(ast.TypePos)
		f := t.Format(ctx)
		if f != exp {
			tst.Error(f)
		}
	} else {
		tst.Error(e)
	}
}

func TestEmptyStmtPosition(tst *testing.T) {
	src := `package foo
func f() {;}
func g() {
  switch x {
  case y:
  case z:
    x++; ; ;
  }
}
`
	exp := `package foo

func f() {
    /* #22 */
}
func g() {
    /* #38 */switch x {
    case y:
    case z:
        x++
        /* #78 */
        /* #80 */
    }
}
`
	if t, e := Parse("empty-stmt-pos.go", src); e == nil {
		ctx := new(ast.FormatContext).Init()
		ctx.EmitSourcePositions(ast.StmtPos)
		f := t.Format(ctx)
		if f != exp {
			tst.Error(f)
		}
	} else {
		tst.Error(e)
	}
}

func TestLabeledStmtPosition(tst *testing.T) {
	src := `package foo
func g() {
L1: if f() { L2: g(); goto L1 }
}
`
	exp := `package foo

func g() {
/* #23 */L1:
    /* #27 */if f() {
    /* #36 */L2:
        g()
        /* #45 */goto L1
    }
}
`
	if t, e := Parse("labeled-stmt-pos.go", src); e == nil {
		ctx := new(ast.FormatContext).Init()
		ctx.EmitSourcePositions(ast.StmtPos)
		f := t.Format(ctx)
		if f != exp {
			tst.Error(f)
		}
	} else {
		tst.Error(e)
	}
}

func TestGoStmtPosition(tst *testing.T) {
	src := `package foo
func g() {
    go f()
}
`
	exp := `package foo

func g() {
    /* #27 */go f()
}
`
	if t, e := Parse("go-stmt-pos.go", src); e == nil {
		ctx := new(ast.FormatContext).Init()
		ctx.EmitSourcePositions(ast.StmtPos)
		f := t.Format(ctx)
		if f != exp {
			tst.Error(f)
		}
	} else {
		tst.Error(e)
	}
}

func TestReturnStmtPosition(tst *testing.T) {
	src := `package foo
func g() {
    return f(), h()
}
`
	exp := `package foo

func g() {
    /* #27 */return f(), h()
}
`
	if t, e := Parse("return-stmt-pos.go", src); e == nil {
		ctx := new(ast.FormatContext).Init()
		ctx.EmitSourcePositions(ast.StmtPos)
		f := t.Format(ctx)
		if f != exp {
			tst.Error(f)
		}
	} else {
		tst.Error(e)
	}
}

func TestBreakStmtPosition(tst *testing.T) {
	src := `package foo
func g() {
  for h() {
    if p() {
      break Q
    } else {
      break
    }
  }
}
`
	exp := `package foo

func g() {
    /* #25 */for h() {
        /* #39 */if p() {
            /* #54 */break Q
        } else {
            /* #81 */break
        }
    }
}
`
	if t, e := Parse("break-stmt-pos.go", src); e == nil {
		ctx := new(ast.FormatContext).Init()
		ctx.EmitSourcePositions(ast.StmtPos)
		f := t.Format(ctx)
		if f != exp {
			tst.Error(f)
		}
	} else {
		tst.Error(e)
	}
}

func TestContinueStmtPosition(tst *testing.T) {
	src := `package foo
func g() {
  for h() {
    if p() {
      continue Q
    } else {
      continue
    }
  }
}
`
	exp := `package foo

func g() {
    /* #25 */for h() {
        /* #39 */if p() {
            /* #54 */continue Q
        } else {
            /* #84 */continue
        }
    }
}
`
	if t, e := Parse("continue-stmt-pos.go", src); e == nil {
		ctx := new(ast.FormatContext).Init()
		ctx.EmitSourcePositions(ast.StmtPos)
		f := t.Format(ctx)
		if f != exp {
			tst.Error(f)
		}
	} else {
		tst.Error(e)
	}
}

func TestGotoStmtPosition(tst *testing.T) {
	src := `package foo
func g() {
  goto L
}
`
	exp := `package foo

func g() {
    /* #25 */goto L
}
`
	if t, e := Parse("goto-stmt-pos.go", src); e == nil {
		ctx := new(ast.FormatContext).Init()
		ctx.EmitSourcePositions(ast.StmtPos)
		f := t.Format(ctx)
		if f != exp {
			tst.Error(f)
		}
	} else {
		tst.Error(e)
	}
}

func TestFallthroughStmtPosition(tst *testing.T) {
	src := `package foo
func g() {
   fallthrough
}
`
	exp := `package foo

func g() {
    /* #26 */fallthrough
}
`
	if t, e := Parse("fallthrough-stmt-pos.go", src); e == nil {
		ctx := new(ast.FormatContext).Init()
		ctx.EmitSourcePositions(ast.StmtPos)
		f := t.Format(ctx)
		if f != exp {
			tst.Error(f)
		}
	} else {
		tst.Error(e)
	}
}

func TestIfStmtPosition(tst *testing.T) {
	src := `package foo
func g() {
   if a {
      x()
   } else if b {
      y()
   } else {
      z()
   }
}
`
	exp := `package foo

func g() {
    /* #26 */if a {
        x()
    } else /* #53 */if b {
        y()
    } else {
        z()
    }
}
`
	if t, e := Parse("if-stmt-pos.go", src); e == nil {
		ctx := new(ast.FormatContext).Init()
		ctx.EmitSourcePositions(ast.StmtPos)
		f := t.Format(ctx)
		if f != exp {
			tst.Error(f)
		}
	} else {
		tst.Error(e)
	}
}

func TestForStmtPosition(tst *testing.T) {
	src := `package foo
func g() {
  for i < n {
    for j > i {
    }
  }
}
`
	exp := `package foo

func g() {
    /* #25 */for i < n {
        /* #41 */for j > i {}
    }
}
`
	if t, e := Parse("for-stmt-pos.go", src); e == nil {
		ctx := new(ast.FormatContext).Init()
		ctx.EmitSourcePositions(ast.StmtPos)
		f := t.Format(ctx)
		if f != exp {
			tst.Error(f)
		}
	} else {
		tst.Error(e)
	}
}

func TestForRangeStmtPosition(tst *testing.T) {
	src := `package foo
func g() {
  for i := range a {
  }
}
`
	exp := `package foo

func g() {
    /* #25 */for i := range a {}
}
`
	if t, e := Parse("for-range-stmt-pos.go", src); e == nil {
		ctx := new(ast.FormatContext).Init()
		ctx.EmitSourcePositions(ast.StmtPos)
		f := t.Format(ctx)
		if f != exp {
			tst.Error(f)
		}
	} else {
		tst.Error(e)
	}
}

func TestDeferStmtPosition(tst *testing.T) {
	src := `package foo
func g() {
    defer f()
}
`
	exp := `package foo

func g() {
    /* #27 */defer f()
}
`
	if t, e := Parse("defer-stmt-pos.go", src); e == nil {
		ctx := new(ast.FormatContext).Init()
		ctx.EmitSourcePositions(ast.StmtPos)
		f := t.Format(ctx)
		if f != exp {
			tst.Error(f)
		}
	} else {
		tst.Error(e)
	}
}

func TestExprSwitchStmtPosition(tst *testing.T) {
	src := `package foo
func g() {
  switch x {
  case y:
  case z:
    switch u {}
  }
}
`
	exp := `package foo

func g() {
    /* #25 */switch x {
    case y:
    case z:
        /* #60 */switch u {}
    }
}
`
	if t, e := Parse("expr-swicth-stmt-pos.go", src); e == nil {
		ctx := new(ast.FormatContext).Init()
		ctx.EmitSourcePositions(ast.StmtPos)
		f := t.Format(ctx)
		if f != exp {
			tst.Error(f)
		}
	} else {
		tst.Error(e)
	}
}

func TestTypeSwitchStmtPosition(tst *testing.T) {
	src := `package foo
func g() {
  switch x.(type) {
  case y:
  case z:
    switch u.(type) {}
  }
}
`
	exp := `package foo

func g() {
    /* #25 */switch x.(type) {
    case y:
    case z:
        /* #67 */switch u.(type) {}
    }
}
`
	if t, e := Parse("type-switch-stmt-pos.go", src); e == nil {
		ctx := new(ast.FormatContext).Init()
		ctx.EmitSourcePositions(ast.StmtPos)
		f := t.Format(ctx)
		if f != exp {
			tst.Error(f)
		}
	} else {
		tst.Error(e)
	}
}

func TestSelectStmtPosition(tst *testing.T) {
	src := `package foo
func f() {
  select {
  case ch1 <- ex:
    select {}
  default:
  }
}
`
	exp := `package foo

func f() {
    /* #25 */select {
    case ch1 <- ex:
        /* #56 */select {}
    default:
    }
}
`
	if t, e := Parse("select-stmt-pos.go", src); e == nil {
		ctx := new(ast.FormatContext).Init()
		ctx.EmitSourcePositions(ast.StmtPos)
		f := t.Format(ctx)
		if f != exp {
			tst.Error(f)
		}
	} else {
		tst.Error(e)
	}
}

func TestIdentPosition(tst *testing.T) {
	src := `package foo
const a, b,  c = 1, 2, 3
var x, y = 1, 2
func (r Tp) Fn()
type S struct { x, y int }
func f(a, b c)
type A interface { Foo () }
`
	exp := `package foo

const /* #18 */a, /* #21 */b, /* #25 */c = 1, 2, 3
var /* #41 */x, /* #44 */y = 1, 2
func (/* #59 */r /* #61 */Tp) /* #65 */Fn()
type S struct {
    /* #86 */x, /* #89 */y /* #91 */int
}
func /* #102 */f(/* #104 */a, /* #107 */b /* #109 */c)
type A interface {
    /* #131 */Foo()
}
`
	if t, e := Parse("ident-pos.go", src); e == nil {
		ctx := new(ast.FormatContext).Init()
		ctx.EmitSourcePositions(ast.IdentPos)
		f := t.Format(ctx)
		if f != exp {
			tst.Error(f)
		}
	} else {
		tst.Error(e)
	}
}
