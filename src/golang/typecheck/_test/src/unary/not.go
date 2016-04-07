package unary

type Bool bool

const (
	A = false
	B = !A
	C = true
	D = !C
)

const (
	cA      = bool(true)
	cB      = !cA
	cC      = bool(false)
	cD      = !cC
	cE Bool = false
)

var (
	vA bool
	vB = !vA
)

const (
	z0 = !cA
	z1
)

var (
	a, b int
	u    Bool = !(a > b)
	v         = !(a < b)
)
