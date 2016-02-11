package typ

var x A

type A [1]B

type B [1]struct {
	X A
}
