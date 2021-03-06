package binary

const (
	ab = bool(true) == false
	bb = false != bool(false)

	aub = true == false
	bub = false != false

	ci8 = int8(1) > int8(-2)
	di8 = -1 < int8(2)
	ei8 = int8(1) == 2
	fi8 = -1 != int8(2)
	gi8 = int8(1) >= 2
	hi8 = 1 <= int8(2)

	cu8 = uint8(1) > uint8(2)
	du8 = 1 < uint8(2)
	eu8 = uint8(1) == 2
	fu8 = 2 != uint8(1)
	gu8 = uint8(1) >= 2
	hu8 = 1 <= uint8(2)

	ci64 = int64(1) > int64(-0x8000000000000000)
	di64 = -1 < int64(0x7fffffffffffffff)
	ei64 = int64(1) == 2
	fi64 = -1 != int64(2)
	gi64 = int64(1) >= 2
	hi64 = 1 <= int64(2)

	af = float32(1.1) > float32(2.2)
	bf = 1.1 < float32(2.2)
	cf = float32(1.1) == 2.2
	df = 2.2 != float32(1.1)
	ef = float32(1.1) >= 2.2
	ff = 1.1 <= float32(2.2)

	as = string("abc") > string("abd")
	bs = "abc" < string("abd")
	cs = string("abc") == "abd"
	ds = "abd" != string("abc")
	es = string("abc") >= "abd"
	fs = "abc" <= string("abd")

	aus = "abc" > "abd"
	bus = "abc" < "abd"
	cus = "abc" == "abd"
	dus = "abd" != "abc"
	eus = "abc" >= "abd"
	fus = "abc" <= "abd"

	cui = 1 > -2
	dui = -1 < 2
	eui = 1 == 2
	fui = -1 != 2
	gui = 1 >= 2
	hui = 1 <= 2

	cur = 'a' > -'b'
	dur = -'a' < 'b'
	eur = 'a' == 'b'
	fur = -'a' != 'b'
	gur = 'a' >= 'b'
	hur = 'a' <= 'b'

	cuf = 1.1 > -2.1
	duf = -1.1 < 2.1
	euf = 1.1 == 2.1
	fuf = -1.1 != 2.1
	guf = 1.1 >= 2.1
	huf = 1.1 <= 2.1

	cir = 'a' > -2
	dir = -1 < 'b'
	eir = 'a' == 2
	fir = -1 != 'b'
	gir = 'a' >= 2
	hir = 1 <= 'b'

	cif = 1 > -2.1
	dif = -1.1 < 2
	eif = 1 == 2.1
	fif = -1.1 != 2
	gif = 1 >= 2.1
	hif = 1.1 <= 2

	crf = 'a' > -2.1
	drf = -1.1 < 'b'
	erf = 'a' == 2.1
	frf = -1.1 != 'b'
	grf = 'a' >= 2.1
	hrf = 1.1 <= 'b'
)

type (
	I interface {
		F()
	}
	J interface {
		F()
		G()
	}
	S struct {
		X int
	}

	Bool    bool
	Int     int16
	Float   float32
	Complex complex64
	String  string
	PInt    *Int
	Vec     [4]float32
)

var (
	A I

	x0 = A == nil
	x1 = nil == A

	B  J
	x2 = A != B

	B0, B1 bool
	x3     = B0 == B1

	P0 PInt
	P1 *Int
	x4 = P0 == P1
	x5 = P1 != P0
	x6 = P0 == nil
	x7 = nil != P1

	Ch0, Ch1 chan []int
	x8       = Ch0 == Ch1

	S0 S
	S1 struct{ X int }
	x9 = S0 == S1

	U   Vec
	V   [4]float32
	x10 = V == U
	x11 = U != V

	x12 Bool = U[0] > V[1]
)

var (
	x, y uint = 1, 2
	a         = (1 << x) < (1 << y)
)
