package sel

type A struct {
	X int
}

func (A) F()

type B struct {
	X int
}

func (*B) F()

type C struct {
	A
	B
}

type P *C

type D int

func (D) X()

type E struct {
	*A
	*D
}
