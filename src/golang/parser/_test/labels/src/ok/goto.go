package ok

func G() bool

func F() {

B0:
G0:
	goto E0
	{
	B1:
	G1:
		goto E1
	G2:
		goto E0
		{
		G3:
			goto E0
		G4:
			goto B0
		}
	E1:
	G5:
		goto B1
	G6:
		goto B0
	}
	type A int
E0:
G7:
	goto B0
}
