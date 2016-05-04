package index

type I16 int16
type F32 float32

var (
	A     map[I16]F32
	B     = 1.1
	C, ok = A[B]
)
