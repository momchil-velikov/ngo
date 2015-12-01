package err

func F(int) (chan int, int)

func G() int {
	x, _ := F(1)
	select {
	case x := <-y:
	}
}
