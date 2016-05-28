package call

var (
	a  [3]int
	p  *[3]int
	s  []float32
	ch chan int

	c0 = cap(a)
	c1 = cap(p)
	c2 = cap(s)
	c3 = cap(ch)
	c4 = cap([...]int{1, 2, 4})
	c5 = cap(&[...]int{1, 2, 4})
	c6 = cap([...]int{1, <-ch, 2})
	c7 = cap([...]int{1, cap(s), 2})
)
