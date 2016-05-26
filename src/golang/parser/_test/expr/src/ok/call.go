package ok

func F(int, int) int

var X, Y int

type T []int

var (
	A = F(X, Y)
	B = make(int, 1)
	C = make(T)
	D = make([]int, 1, 2)
)
