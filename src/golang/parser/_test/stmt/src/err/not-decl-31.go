package err

func F(int) (interface{}, int)

func G() int {
	x, y := F(1)
	switch x, z = F(2); x.(type) {
	}
}
