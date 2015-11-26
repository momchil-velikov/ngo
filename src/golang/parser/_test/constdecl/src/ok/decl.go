package ok

const A = 1
const B int = A
const C, D = A, B
const E, _, F string = "3", "1", "4"

func Fn() {
	const A = 1
	const B int = A
	const C, D = A, B
	const E, _, F string = "3", "1", "4"
}
