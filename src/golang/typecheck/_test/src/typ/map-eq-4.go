package typ

type A struct {
	X B
}
type B struct {
	X map[int]int
}

type C chan map[A]int
