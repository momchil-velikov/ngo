package infer

var a, b = f()

type P [1]S

type Q [1]int

type S struct {
	X [len(a)]int
}

func f() (P, Q)
