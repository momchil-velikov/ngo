package call

var (
	ch <-chan int
	a  = close(ch)
)
