package binary

type Bool bool

const (
	T, F = true, false

	Z0 = F && F
	Z1 = F && T
	Z2 = T && F
	Z3 = T && T

	Z4 = F || F
	Z5 = F || T
	Z6 = T || F
	Z7 = T || T

	TT, FF Bool = true, false

	Z8 = FF && FF
	Z9 = FF && TT
	ZA = TT && FF
	ZB = TT && TT

	ZC = FF || FF
	ZD = FF || TT
	ZE = TT || FF
	ZF = TT || TT
)

var (
	a, b Bool

	z0 = a && b
	z1 = a || b
)
