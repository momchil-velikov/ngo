package b

import "sel/a"

type B struct {
	X int
	y int
}

func (B) F()

func (B) g()

type C struct {
	a.A
	B
}

var (
	b B
	x = b.X
	y = b.y

	c C
	z = c.y

	f0 = B.F
	g0 = B.g

	g1 = C.g
)
