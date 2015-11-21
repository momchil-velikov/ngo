package ok

type S struct{ X, Y int }
type T struct{ A, B S }

var A = S{1, 2}

var B = A.X

var C = T{A: {1, 2}, B: {3, 4}}
var D = C.A.Y

var E = (*S).Y // pass resolution, fail typeheck
