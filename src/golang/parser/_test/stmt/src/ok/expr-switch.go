package ok

func F(int) (int, int)

func G() int {
	x, y := F(1)
	switch {
	case x > y:
	}
	switch x, y = F(2); {
	case x > y:
	}
	switch u, v := F(3); {
	case u > v:
	}
	switch x, v := F(4); {
	case x > v:
	default:
	}
}
