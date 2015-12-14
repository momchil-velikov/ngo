package err

type S struct {
	F int
}

type T struct {
	S
}

func (T) S()
