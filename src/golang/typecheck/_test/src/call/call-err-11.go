package call

func f(int, int, ...int) int

var (
	ss []int
	a  = f(1, 2, 3, ss...)
)
