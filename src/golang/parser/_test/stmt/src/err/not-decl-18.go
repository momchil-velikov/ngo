package err

func F(int) (int, int)

func G() int {
	x, y := F(1)
	if x > y {
		x, y = F(x)
	} else {
		x, y = F(z)
	}
}
