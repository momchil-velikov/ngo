package unary

const (
	A = 1
	B = +A
	C = 'a'
	D = +C
	E = 1.1
	F = +E
)

const (
	cA = int(1)
	cB = +cA
	cC = rune('a')
	cD = +cC
	cE = float32(1.125)
	cF = +cE
)

var (
	vA int
	vB = +vA
	vC rune
	vD = +vC
	vE float64
	vF = +vE
)

const (
	z0 = +cA
	z1
)
