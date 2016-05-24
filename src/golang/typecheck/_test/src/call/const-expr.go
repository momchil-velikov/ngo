package call

const Z = 1

// Not a const initializer, but the test does not perforrm all the semantic
// checks
const Ch chan int = make(chan int)

func f() int

type (
	Int int
	Ptr *[4]Int
)

var (
	A      = 1
	B      = Z
	C      = A
	D *int = nil
	E      = f
	F      = f()
	G      = uint(A)
	H      = uint(Z)
	I      = uint(f())
	J      = (*int)(nil)
	K      = *D
	L      = &F
	M      = ^A
	N      = ^Z

	ch chan int
	O  = <-ch
	P  = <-Ch

	Q = A + B
	R = Z - Z
	V = 1 << A
	W = 1 << Z

	S = len("aaa")

	a [4]Int
	s []int
	T = len(a)
	U = len(1, 2)   // error, not considered a constant
	X = len(1 << A) // not array
	Y = len(s)

	AA = [...]int{1, 2, 3}
	AB = len([...]int{1, 2, 3})
	p  = Ptr(&a)
	AD = len(p)
	AE = cap(L)
	AF = len([...]int{f(), 2, 3})
	AG = len([...]int{1, 2: <-ch, 3})
	AH = len([...]int{1, len([...]int{1, 2}), 3})
	AI = len([...]int{1, len([...]int{1, f()}), 3})
	AJ = len([...]int{1, len([...]int{<-ch, 2}), 3})

	C0 = complex(Z, 1.2)
	C1 = complex(Z, A)
	C2 = complex(1, 2, 3)

	IM0 = imag(C0)
	IM1 = imag(complex(Z, 1.2))
	IM2 = imag(1, 2)
	RE0 = real(C0)
	RE1 = real(complex(Z, 1.2))
	RE2 = real(1, 2)

	S0 = len("aa")
	S1 = len("aa" + Z)
	S2 = len("aa"[0])
	S3 = cap("aa")

	ss string
	S4 = len(ss)

	m  map[int]func() int
	F0 = m[1]()
	F1 = A()
)
