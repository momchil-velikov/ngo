package typ

type A map[S]int

type S struct {
	X T
}

type T struct {
	X A
}
