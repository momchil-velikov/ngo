package index

type I16 int16
type F32 float32

var (
	A     map[I16]F32
	B, ok = A[1.1]
)
