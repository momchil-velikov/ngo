package ok

type S struct {
	F int
}

type T struct {
	S
}

func (T) F()
