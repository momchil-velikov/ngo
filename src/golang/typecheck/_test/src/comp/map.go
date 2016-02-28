package comp

type M map[int]string

var (
	A = map[int]string{}
	B = M{1: "a", 2: "b"}
	C = map[[2]int][]float64{{1, 2}: {3.14}}
	b = bool(true)
	D = map[bool]int{true: 0, false: 1, b: false}
	i = int8(4)
	E = map[int8]int{-1: 0, i: 0, 1: 1, 8: 3}
	u = uintptr(12)
	F = map[uintptr]int{1: 0, 2: 1, u: 3}
	f = float32(1.2)
	G = map[float32]int{f: 1, 5.8: 7}
	s = string("foo")
	H = map[string]int{"foo": 1, "bar": 2, s: 3}
)
