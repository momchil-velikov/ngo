package typ

type I interface {
	F()
}

type J interface {
	I
	G()
}

type K interface {
	I
	H()
}

type L interface {
	J
	K
}
