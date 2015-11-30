package ok

func F(int)

func G() {
	defer F(1)
}
