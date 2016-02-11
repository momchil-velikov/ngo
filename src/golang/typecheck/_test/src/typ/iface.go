package typ

type A interface{}

type B interface {
	A
}

type C interface {
	B
	F(A, B) C
}
