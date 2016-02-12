package typ

type I interface {
	F()
}

type II I

type J interface {
	II
	F()
}
