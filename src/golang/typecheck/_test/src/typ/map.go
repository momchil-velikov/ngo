package typ

type A map[int]int

type B map[int][]int

type C map[[3]int]A

type D map[*A]A

type E map[chan E]E

type F map[struct{ X *D }]D
