package sel

type S struct{}

func (S) F(int)
func (*S) G(float64)

var D = S.G // error, needs pointer receiver
