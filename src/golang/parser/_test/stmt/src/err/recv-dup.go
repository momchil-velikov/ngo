package err

func F(int) (chan int, int)

func G() int {
	x, y := F(1)
	select {
	case z, z := <-x:
	}
}
