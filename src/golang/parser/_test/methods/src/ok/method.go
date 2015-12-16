package ok

type A int

func (a A) F() {}
func (_ *A) G(_ int)
