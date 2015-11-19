package not_typename

var X int

type A interface {
	F(int)
}

type B interface {
	X
}
