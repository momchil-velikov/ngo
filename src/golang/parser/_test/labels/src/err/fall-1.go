package err

func F(x interface{}) {
	switch x.(type) {
	case []int:
		fallthrough
	default:
	}
}
