package index

var (
	A [4]int8
	B *[5]int16
	C []int32
	D string
	E map[float64]float32

	ax = A[3]
	bx = B[4]
	i  = int8(1.0)
	cx = C[i]
	dx = D[0]
	ex = E[1.1]
)
