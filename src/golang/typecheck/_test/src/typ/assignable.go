package typ

const c0 int8 = 'a'

type (
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
)
