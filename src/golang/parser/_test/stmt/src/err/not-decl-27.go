package err

func F(int) (int, int)

func G() int {
	x, y := F(1)
	switch x, z = F(2); {
	}
}
