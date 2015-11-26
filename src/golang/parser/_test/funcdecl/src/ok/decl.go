package ok

type (
	A int
	B string
	C []byte
)

func F(a A, b B, c C, _ int, d ...C) (e A, f, g B, _ float32) {
	return 1, "", "", 1.1
}

func _() {}
