package typ

type A struct {
	X B
}

type B struct {
	X C
}

type C struct {
	X [3]struct{ Y A }
}
