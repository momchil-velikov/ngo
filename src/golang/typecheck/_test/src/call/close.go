package call

var (
	ch  chan int
	out chan<- int

	a = close(ch)
	b = close(out)
)
