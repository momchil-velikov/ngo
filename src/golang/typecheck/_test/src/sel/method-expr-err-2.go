package sel

type S struct{}

func (S) F(int)
func (*S) G(float64)

type T0 struct {
	S
}

var D0 = T0.G // error, needs pointer receiver
