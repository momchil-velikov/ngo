package infer

var A [2]B

type B struct {
	X [len(A)]int
}
