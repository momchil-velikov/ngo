package err

var (
	A chan int
)

func F() {
	A <- B
}
