package binary

type Int int

const (
	A, B int8 = 17, -2
	C         = A % B

	D uint16 = 17
	E        = D % 3

	M   = 17
	N   = M % 2
	D12 = S0 % S0

	Z0 Int = 17
	Z1     = Z0 % -4
	Z2     = -4 % Z0
	Z3     = Z0 % Z1

	S0 = 17
	S1 = 'a'

	D0  = S0 % S1
	D1  = S1 % S0
	D13 = S1 % S1
)

var (
	a Int = 1
	b     = a % 1
	c     = 1 % a
	d     = a % a
)
