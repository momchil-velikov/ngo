package binary

type Z struct {
	X uint
	Y int8
}

var (
	S = uint(1)

	A = 2 << S

	B uint = 1.0 << S

	C = (1.0 << S) + uint8(2.0<<S)

	D = int16(1.0<<S) + (2.0 << S)

	E int8 = ^(1.0 << S)

	F uint16 = (1.0 << S) + (2.0 << S)

	G = (1 << S) + ^(2 << S)

	H = []int64{(1.0 << S) + ^(2.0 << S)}

	I = map[int8]uint16{(1.0 << S) + ^(2.0 << S): (1.0 << S) + ^(2.0 << S)}

	J = Z{1.0 << S, 1.0 << S}

	K = Z{Y: 1.0 << S, X: 1.0 << S}

	L = I[(1.0<<S)+^(2.0<<S)]

	a []int

	M = a[(1.0<<S)+^(2.0<<S)]

	N = a[1.0<<S : 2.0<<S : 3.0<<S]
)
