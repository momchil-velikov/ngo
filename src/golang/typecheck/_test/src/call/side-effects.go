package call

type (
	Int int
	Str struct {
		X, Y int
	}
	Arr [10]Int
)

func f() int
func fx() interface{}
func fs() *Str
func fa() []Int

var (
	A   = f()
	B   = 1
	C   = int32(f())
	D   = int32(A)
	x   interface{}
	E   = fx().(int32)
	F   = x.(int32)
	G   = (interface{})(fx()).(int)
	s   Str
	H   = s.X
	I   = fs().Y
	arr Arr
	J   = arr[B]
	K   = arr[f()]
	L   = fa()[B]
	M   = fa()[f()]
	N   = arr[A:]
	O   = arr[B:A]
	P   = arr[:B:A]
	Q   = N[A:]
	R   = O[B:A]
	S   = P[:B:A]
	T   = fa()[:B:A]
	U   = T[f():B:A]
	V   = T[:f():A]
	W   = T[:B:f()]
	X   = -f()
	Y   = -A
	ch  chan int
	Z   = <-ch

	// FIXME: Postponed until issue #71 resolved
	// Cr = int32(<-ch)
	// Dr = int32(Z)
	// Gr = (interface{})(<-ch).(int)

	// chS chan Str
	// Ir  = (<-chS).Y

	// Kr  = arr[<-ch]
	// chA chan []Int

	// Lr = (<-chA)[B]
	// Mr = (<-chA)[f()]
	// Tr = (<-chA)[:B:A]
	// Ur = T[<-ch:B:A]
	// Vr = T[:<-ch:A]
	// Wr = T[:B:<-ch]

	AA = A + A
	AB = f() + B
	AC = B + f()

	AD = [...]int{1: A, 2: B}
	AE = [...]int{1: f(), 2: B}
	AF = [...]int{1: A, 2: f()}

	m map[int]int

	AG = map[int]int{f(): 1, 2: B}
	AH = map[int]int{1: A, f(): 2}

	L0 = len([...]int{1: A, 2: B})
	L1 = len([...]int{1: f(), 2: B})

	// FIXME: Postponed until issue #72 resolved
	// C0 = cap([...]int{1: A, 2: B})
	// C1 = cap([...]int{1: f(), 2: B})
)
