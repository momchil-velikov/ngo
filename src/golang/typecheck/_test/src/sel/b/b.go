package b

import "sel/a"

type B struct {
	X int
	y int
}

type C struct {
	a.A
	B
}

var (
	b B
	x = b.X
	y = b.y

	c C
	// z = c.X // error
	z = c.y
)
