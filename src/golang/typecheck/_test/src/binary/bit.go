package binary

type Int int16

const (
	A, B Int = 0x7081, 0x7180

	Z0 = A & B
	Z1 = A | B
	Z2 = A ^ B
	Z3 = A &^ B

	C = 0xa5a5
	D = 0x1111
	E = 'a'

	Z4  = C & D
	Z7  = C | D
	Z10 = C ^ D
	Z13 = C &^ D

	Z5  = C & E
	Z6  = E & C
	Z8  = C | E
	Z9  = E | C
	Z11 = C ^ E
	Z12 = E ^ C
	Z14 = C &^ E
	Z15 = E &^ C
)

var (
	a, b Int

	z0 = a & b
	z1 = a | b
	z2 = a ^ b
	z3 = a &^ b
)
