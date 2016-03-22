package a

type A0 int32

type A1 int32

type A2 int

type B0 [3]A0

type B1 [3]A0

type B2 [2]A0

type B3 [3]A1

type C0 []A0

type C1 []A0

type C2 []A1

type D0 *A0

type D1 *A0

type D2 *A1

type E0 map[A0]A0

type E1 map[A0]A0

type E2 map[A1]A0

type E3 map[A0]A1

type F0 chan A0

type F1 chan A0

type F2 chan A1

type F3 <-chan A0

type F4 chan<- A0

type G0 struct {
	X A0
	Y A1
}

type G1 struct {
	X A0
	Y A1
}

type G2 struct {
	X A0
	Z A1
}

type G3 struct {
	X A0 `a:"b"`
	Y A1
}

type G4 struct {
	X A0 `a:"b"`
	Y A0
}

type G5 struct {
	X A0 `a:"b"`
	y A1
}

type G6 struct {
	X A0 `a:"b"`
	y A1
}

type G7 struct {
	X A0 `a:"b"`
}

type H0 func()

type H1 func()

type H2 func(A0, A1) (A0, A1)

type H3 func(A0, A1) (A0, A1)

type H4 func(A0) (A0, A1)

type H5 func(A0, A1) A0

type H6 func(A1, A1) (A0, A1)

type H7 func(A0, A1) (A0, A0)

type H8 func(A0, ...A1) (A0, A1)

type H9 func(A0, ...A1) (A0, A1)

type I0 interface {
	F(A0)
}

type I1 interface {
	F(A0)
}

type I2 interface {
	I0
	G(A1) A0
}

type I3 interface {
	I1
	G(A1) A0
}

type I4 interface {
	I1
	G(A1) A0
	H() A1
}

type I5 interface {
	H() A1
	I3
}

type I6 interface {
	H() A1
	gG()
	I3
}

type I7 interface {
	H() A1
	I3
	gG()
}

type I8 interface {
	H() A1
	I3
}

type I9 interface {
	H() A0
	I3
}
