package typ

type I interface {
	F()
}

type J interface {
	I
	F()
}
