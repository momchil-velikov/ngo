package call

type I interface {
	F()
}

type J interface {
	I
}

func f(J) int

const C = 1.1i

var a = f(C)
