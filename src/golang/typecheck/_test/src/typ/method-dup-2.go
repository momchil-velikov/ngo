package typ

type I interface {
	F()
}

type J interface {
	F()
}

type K interface {
	I
	J
}
