package ok

func F(int) (int, int)

func G() int {
	x, y := F(1)
	if x > y {
	}
	if x, y = F(2); x > y {
	}
	if u, v := F(3); u > v {
	}
	if x, v := F(4); x > v {
	}
	if x, y := F(5); x > y {
	} else if x < y {
	}
	if x, y := F(6); x > y {
	} else if x, y = F(66); x < y {
	}
	if x, y := F(7); x > y {
	} else if x, y := F(77); x < y {
	} else {
	}
}
