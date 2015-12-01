package err

func F(int) (int, int)

func G() int {
	if u, _ = F(3); u > 0 {
	}
}
