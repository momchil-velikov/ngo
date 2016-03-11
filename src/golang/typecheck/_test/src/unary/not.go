package unary

const (
	A = false
	B = !A
	C = true
	D = !C
)

const (
	cA = bool(true)
	cB = !cA
	cC = bool(false)
	cD = !cC
)

var (
	vA bool
	vB = !vA
)
