package typ

type I interface {
	F()
}

type A struct{}

func (A) F()

type B struct{}

func (*B) F()

type C *B

type D struct{}

func (D) F() int

type J interface {
	I
	G()
}

type J1 interface {
	I
	G() int
}

type E struct{}

func (*E) F()

func (*E) G(int) int

type F struct{}

func (*F) F()
