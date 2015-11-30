package ok

var (
	A chan int
	B int
)

func F() {
	A <- B
}
