package index

type I16 int16
type I32 int32
type F32 float32

var (
	A [4]int8
	B *[5]I16
	C []I32
	D string
	E map[float64]F32

	ax = A[3]
	bx = B[4]
	i  = int8(1.0)
	cx = C[i]
	dx = D[0]
	ex = E[1.1]
	fx = E[1.1] + 2.2
)
