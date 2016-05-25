package call

func f(int, int) int
func g() (int, int)

var a = f(1, g())
