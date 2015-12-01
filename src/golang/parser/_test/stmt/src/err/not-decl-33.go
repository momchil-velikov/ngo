package err

func F(int) (interface{}, int)

func G() int {
	switch x, y := F(2); x.(type) {
	case Int:
	}
}
