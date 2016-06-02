package unary

const (
	A = 1
	B = -A
	C = 'a'
	D = -C
	E = 1.1
	F = -E
	G = complex(1.1, 2.2)
	H = -G
)

const (
	cA = int(1)
	cB = -cA
	cC = rune('a')
	cD = -cC
	cE = float32(1.125)
	cF = -cE
	cG = complex(float32(1.125), 2.25)
	cH = -cG
)

var (
	vA int
	vB = -vA
	vC rune
	vD = -vC
	vE float64
	vF = -vE

	u uint
	v = -u
)

const (
	z0 = -cA
	z1
)
