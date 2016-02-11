package typ

type A struct {
	X B
}
type B struct {
	X func()
}

type C func(map[A]int)
