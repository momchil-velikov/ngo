package ok

func G() bool
func R() []int
func Ch() chan int
func Any() interface{}

func F() {
L0:
	for {
	B0:
		break
	L1:
		for {
			if G() {
			B1:
				break
			}
		}
	}

L2:
	for {
	L3:
		switch {
		default:
			if G() {
			B2:
				break L2
			} else {
			B3:
				break L3
			}
		B4:
			break
		}
	}

L4:
	for i, v := range R() {
	L5:
		switch v {
		case 0:
		B5:
			break L4
		default:
		B6:
			break
		}
	}

L6:
	switch Any().(type) {
	case []int:
	B7:
		break L6
	default:
	B8:
		break
	}

	for {
	L7:
		select {
		case <-Ch():
		B9:
			break L7
		default:
		B10:
			break
		}
	}
}
