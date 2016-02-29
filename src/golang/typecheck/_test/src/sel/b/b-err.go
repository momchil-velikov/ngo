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
	c C
	z = c.X // error
)
