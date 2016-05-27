package call

type I interface {
	F()
}

type J interface {
	I
}

func f(J) int

const C float32 = 1.1

var a = f(C)
