package err

func F(int) (int, int)

func G() int {
	var x int
	for x, _ = F(3); x > 0; x-- {
		x + y
	}
}
