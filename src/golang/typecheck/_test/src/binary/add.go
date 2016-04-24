package binary

type Int int

const (
	A, B int8 = 1, 2
	C         = A + B

	D uint16 = 1
	E        = D + 3

	F, G float64 = 1.125, 2.25
	H            = F + G

	I float32 = 1.0000005
	J         = I + 1.00000001

	K float64 = 1.0000005
	L         = K + 1.00000001

	S  string = "abc"
	SS        = S + "αβγ"

	M   = 1
	N   = M + 1
	D12 = S0 + S0

	U complex64 = 0.125i
	V           = U + 1.0

	X complex128 = 0.125i
	Y            = X + 1.25

	Z0 Int = 3
	Z1     = Z0 + 4
	Z2     = 4 + Z0
	Z3     = Z0 + Z1

	// int, rune, float, complex,
	S0 = 1
	S1 = 'a'
	S2 = 1.125
	S3 = 1.25i

	D0  = S0 + S1
	D1  = S1 + S0
	D13 = S1 + S1

	D2  = S0 + S2
	D3  = S2 + S0
	D14 = S2 + S2

	D4  = S0 + S3
	D5  = S3 + S0
	D15 = S3 + S3

	D6 = S1 + S2
	D7 = S2 + S1

	D8 = S1 + S3
	D9 = S3 + S1

	D10 = S2 + S3
	D11 = S3 + S2

	S4  = "aβγ"
	S5  = "xyζ"
	D16 = S4 + S5
)

var (
	a Int = 1
	b     = a + 1
	c     = 1 + a
	d     = a + a

	s0 = "aβγ"
	s1 = "xyζ"
	ss = s0 + s1
)
