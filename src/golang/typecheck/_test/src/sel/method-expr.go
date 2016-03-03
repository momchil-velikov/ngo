package sel

type I interface {
	F(int)
}

var A = I.F

type S struct{}

func (S) F(int)      {}
func (*S) G(float64) {}

var B = S.F

var C = (*S).F

var D = (*S).G

type J struct {
	I
}

var A0 = J.F

var A1 = (*J).F

type T0 struct {
	S
}

var B0 = T0.F

var C0 = (*T0).F

var D0 = (*T0).G

type T1 struct {
	*S
}

var B1 = T1.F

var C1 = (*T1).F

var D10 = T1.G

var D11 = (*T1).G
