package assert

type A interface {
	F()
}

type B interface {
	F()
	G() int
}

type S struct{}

func (S) F()     {}
func (S) G() int { return 1 }

type T struct{}

func (*T) F() {}

var (
	a A
	b B
	s S
	t T

	u, o0 = a.(B)
	v, o1 = b.(A)
	w, o2 = a.(S)
	x, o3 = b.(S)
	y, o4 = a.(*T)

	p interface{}
	q = p.(int) + 1
)
