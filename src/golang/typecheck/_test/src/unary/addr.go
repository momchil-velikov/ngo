package unary

type (
	T int
	P *T
	S struct{ X, T int }
)

var (
	A T
	B = &A
	C = &*B
	D = []S{{1, 2}, {3, 4}}
	E = &D[1]
	F = &D[0].X
	G = [...]T{1, 2, 3}
	H = &G[2]
	I = &[]int{1, 2, 3}[0]
	X = &S{1, 2}
	Y = &[...]T{1, 2}
	Z = &[]T{1, 2}
)
