package binary

type T0 int
type T1 int

const (
	A = T0(1)
	B = T1(2)
	C = A ^ B
)
