package err

func F(int) (chan int, int)

func G() int {
	_, y := F(1)
	select {
	case x = <-y:
	}
}
