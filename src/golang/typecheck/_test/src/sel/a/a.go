package a

type A struct {
	X int
	y int
}

func (A) F()

func (A) g()

var (
	a A
	x = a.X
	y = a.y

	f = a.F
	g = a.g
)
