package binary

type Int int

const (
	A, B int8 = 10, 2
	C         = A / B

	D uint16 = 10
	E        = D / 3

	F, G float64 = 1.125, 2.25
	H            = F / G

	I float32 = 1.0000001
	J         = I / 0.5

	K float64 = 1.0000001
	L         = K / 0.5

	M   = 10
	N   = M / 2
	D12 = S0 / S0

	U complex64 = 0.125i
	V           = U / 0.25

	X complex128 = 0.125i
	Y            = X / 0.25

	Z0 Int = 12
	Z1     = Z0 / 4
	Z2     = 4 / Z0
	Z3     = Z0 / Z1

	// int, rune, float, complex,
	S0 = 10
	S1 = 'a'
	S2 = 1.25
	S3 = 1.25i

	D0  = S0 / S1
	D1  = S1 / S0
	D13 = S1 / S1

	D2  = S0 / S2
	D3  = S2 / S0
	D14 = S2 / S2

	D4  = S0 / S3
	D5  = S3 / S0
	D15 = S3 / S3

	D6 = S1 / S2
	D7 = S2 / S1

	D8 = S1 / S3
	D9 = S3 / S1

	D10 = S2 / S3
	D11 = S3 / S2
)

var (
	a Int = 1
	b     = a / 1
	c     = 1 / a
	d     = a / a
)
