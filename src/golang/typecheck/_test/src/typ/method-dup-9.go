package typ

type S struct {
	F int
}

type T S

func (T) F() {}
