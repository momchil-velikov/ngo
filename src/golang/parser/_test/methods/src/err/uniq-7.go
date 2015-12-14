package err

type S struct {
	F int
}

type T struct {
	*S
}

type TT T

func (TT) S()
