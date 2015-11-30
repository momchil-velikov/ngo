package ok

func F(int) (int, int)

func G() int {
	x, y := F(1)
	for x > y {
	}
	for x, y = F(2); x > y; {
	}
	for u, v := F(3); u > v; {
	}
	for x, v := F(4); x > v; {
	}
}
