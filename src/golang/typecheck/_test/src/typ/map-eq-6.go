package typ

type A struct {
	X B
}
type B struct {
	X []int
}

type C func() map[A]int
