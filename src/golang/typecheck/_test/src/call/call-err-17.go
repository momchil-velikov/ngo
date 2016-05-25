package call

func f(int, ...float64) int

var (
	x float32
	a = f(1, 1, 2.1, x, 1.2)
)
