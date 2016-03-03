package b

import "sel/a"

type B struct{}

func (B) F()
func (B) g()

type C struct {
	a.A
	B
}

var (
	c = C.F // error, ambiguous
)
