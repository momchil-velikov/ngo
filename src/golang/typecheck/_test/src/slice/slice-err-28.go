package slice

func f() [5]int

var (
	a = f()[1:]
)
