package typ

type A interface {
	F() A
	B
}

type B interface {
	F() B
	C
}

type C interface {
	F() C
	A
}
