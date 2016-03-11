package unary

type (
	T int
	P *T
)

var (
	A P
	B = *A
)
