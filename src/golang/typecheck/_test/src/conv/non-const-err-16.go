package conv

type I interface {
	F()
}

type J interface {
	I
}

var ifc = J(1)
