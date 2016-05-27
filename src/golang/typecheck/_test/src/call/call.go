package call

func f0()
func f1() int8
func f2() (int8, int8)
func f3(int8) int8
func f4(int8, int8) int8
func f5(int8, int8, int8) (int8, int8, int8)

func g0(x ...int8)
func g1(x ...int8) int8
func g2(x ...int8) (int8, int8)
func g3(x int8, y ...int8) int8
func g4(x int8, y int8, z ...int8) int8
func g5(x int8, y int8, z ...int8) (int8, int8, int8)
func g6() []int8

func h(*int) int

var (
	_b      = f0() // FIXME: should be error
	a       = f1()
	b, c    = f2()
	d       = f3(b)
	e       = f4(c, d)
	i, j, k = f5(c, d, e)

	ss             []int8
	l0, l1, l2, l3 = g1(), g1(a), g1(a, b, c), g1(ss...)

	m0, m1 = g2()
	m2, m3 = g2(a)
	m4, m5 = g2(a, b, c)
	m6, m7 = g2(ss...)

	n0, n1, n2, n3 = g3(a), g3(a, b), g3(a, b, c, d), g3(a, ss...)

	o0, o1, o2, o3 = g4(a, b), g4(a, b, c), g4(a, b, c, d), g4(a, b, ss...)

	p0, p1, p2   = g5(a, b)
	p3, p4, p5   = g5(a, b, c)
	p6, p7, p8   = g5(a, b, c, d)
	p9, p10, p11 = g5(a, b, ss...)

	q0, q1     = f3(f1()), f4(f2())
	q3, q4, q5 = f5(f5(a, b, c))

	r0, r1, r2, r3 = g1(g1()), g1(f5(a, b, c)), g3(f2()), g3(f5(a, b, c))

	s0, s1 = g4(f2()), g4(f5(a, b, c))

	t = g1(g6()...)
	u = g4(a, b, g6()...)

	s          uint
	v0, v1, v2 = f3(1.0), f3(0x80 >> 1), f3(0x7f << s)

	w0 = h(nil)

	ch chan int8
	x0 = f3(<-ch)
)

func print(...interface{}) int

const C = 1

var (
	z0 = print(C, 1.1, float32(2.2)+C)
)
