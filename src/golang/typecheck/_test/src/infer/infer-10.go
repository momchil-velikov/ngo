package infer

var s = f()

func f() struct {
	X [len(s.Y)]int
	Y [1]int
}
