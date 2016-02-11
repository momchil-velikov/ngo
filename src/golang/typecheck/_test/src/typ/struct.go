package typ

type A struct{}

type B struct {
	_, _ int
	A
	_, _ int
	X    int
}

type C struct {
	X int
	Y *C
	Z []C
}

type D struct {
	X *E
}

type E struct {
	*F
}

type F struct {
	X *D
}
