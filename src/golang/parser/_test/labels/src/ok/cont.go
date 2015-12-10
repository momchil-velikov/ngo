package ok

func G() bool
func R() []int
func Ch() chan int
func Any() interface{}

func F() {
L0:
	for {
	C0:
		continue
	L1:
		for {
			if G() {
			C1:
				continue
			}
		}
	}

L2:
	for {
	L3:
		for i, v := range R() {
			if G() {
			C2:
				continue L2
			} else {
			C3:
				continue L3
			}
		}
	}
}
