package typ

type A chan int

type B chan A

type C chan C

type D chan E

type E chan D
