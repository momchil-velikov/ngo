package ok

func F(int) map[int]int

func G() int {
	var x, y int
	for x, y = range F(1) {
	}
	for u, v := range F(2) {
	}
	for x, v := range F(4) {
	}
}
