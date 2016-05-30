package call

var (
	a struct{ X, Y int }
	b = close(a)
)
