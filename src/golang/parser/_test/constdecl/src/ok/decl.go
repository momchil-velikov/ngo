package ok

const A = 1
const B int = A
const C, D = A, B
const E, F string = "3", "4"

func Fn() {
	const A = 1
	const B int = A
	const C, D = A, B
	const E, F string = "3", "4"
}
