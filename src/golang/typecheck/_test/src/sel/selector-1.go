package sel

type I interface {
	F()
}

type II I

type J interface {
	I
	G()
}

type A struct {
	X int
	Y int
	J
}

func (A) MX()

func (A) MY()

type AA A

type B struct {
	A
	Y int
}

func (*B) MY()

type C struct {
	*A
	Y int
}

func (C) MY()

type P *C
