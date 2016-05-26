package call

var (
	a0 = make([]int)
	a1 = make([]int, 3)
	a2 = make([]int, 3, 4)

	b0 = make(map[string]int)
	b1 = make(map[string]int, 2)

	c0 = make(chan int)
	c1 = make(chan int, 2)
)

type SInt []int

var (
	x = make(SInt, 2.0)
	n uint
	m int32
	y = make(SInt, n, m)
)
