package typ

type A func()

type B func() A

type C func(A)

type D func(A, B, C, D) (A, B, C, D)

func f()

func g(A, B, C, D) (A, B, C, D)
