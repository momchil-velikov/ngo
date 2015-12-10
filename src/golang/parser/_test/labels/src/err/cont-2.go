package err

var v int

func F() {
	if v > 0 {
		{
			continue
		}
	}
}
