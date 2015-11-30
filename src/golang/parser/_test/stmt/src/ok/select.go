package ok

func F(int) (chan int, int)

func G() int {
	x, y := F(1)
	select {
	case x <- y:
	case y = <-x:
	case x, _ := <-x:
		x
	default:
	}
}
