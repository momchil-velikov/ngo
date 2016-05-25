package call

func f(int, bool) int

var (
	ch chan int
	a  = f(<-ch)
)
