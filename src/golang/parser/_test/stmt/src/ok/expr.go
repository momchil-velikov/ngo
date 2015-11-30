package ok

func F(ch chan int) {
	<-ch
}
