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

	O = uint16((1.0 << S) + (2.0 << S) + (3.0 << S))

	c0 = uint(1) << 1

	c1 = int(-2) >> 1

	c2 = uint(2) >> 1

	c3 = 1 << uint(1)

	c4 = -2 >> uint(1)

	c5 = 1.0 << 1

	c6 = 6.0 >> 1

	c7 = 0.0i << 1

	c8 = 0.0i >> 2

	c9 = 'a' << 1

	c10 = 'b' >> 1

	l0 = []uint8{1 << 7}

	l1 = map[uint16]uint8{1 << 15: 1 << 7}
)
