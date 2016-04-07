package typ

const c0 int8 = 'a'

type (
	Bool bool
	Int  int
	Rune rune
	S    struct{ X, Y float32 }
	T    struct{ X, Y float32 }
	I    interface {
		F()
	}
	A struct{ X [2]int }
)

func (A) F()

func F() [2]int

var (
	v0 int
	v1 int  = v0
	v2      = 'a'
	v3 rune = v2
	v4 Rune = 'a'

	v7  S                      = struct{ X, Y float32 }{1.1, 2.2}
	v8  struct{ X, Y float32 } = v7
	v9  struct{ X, Y float32 } = struct{ X, Y float32 }{1.1, 2.2}
	v10 I                      = A{}

	ch0 chan A
	ch1 <-chan A = ch0
	ch2 chan<- A = ch0
	ch3 chan<- A = ch2

	z0 *A             = nil
	z1 func()         = nil
	z2 []A            = nil
	z3 map[int]string = nil
	z4 chan A         = nil
	z5 I              = nil
	z6 interface{}    = nil

	a A
	b [2]int = a.X

	u [2]int = F()

	s chan [2]int
	r chan [2]int = s

	b0 bool
	b1 bool = b0
	b2      = true
	b3 bool = b2

	ifc     interface{}
	b4, ok0 bool = ifc.(bool)
	b5, ok1      = ifc.(bool)

	m       map[int]bool
	b6, ok2 bool = m[1]
	b7, ok3      = m[2]

	ch      chan Bool
	b8, ok4 Bool = <-ch
	b9, ok5      = <-ch
)
