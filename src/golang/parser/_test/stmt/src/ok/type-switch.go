package ok

func F(int) (interface{}, int)

func G() int {
	x, y := F(1)
	switch x.(type) {
	case int:
		x + y
	}
	switch x, y = F(2); x.(type) {
	case int:
		x + y
	}
	switch u, v := F(3); x.(type) {
	case int:
		u + v
	}
	switch x, v := F(4); x.(type) {
	case int:
		x + v
	}
	switch x, y := F(5); t := x.(type) {
	case int:
		x + t
	case float32:
		t
	default:
	}
}
