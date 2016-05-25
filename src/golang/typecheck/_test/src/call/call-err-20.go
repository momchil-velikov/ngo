package call

func f(int, ...float64) int

var (
	x float32
	a = f(1, x...)
)
