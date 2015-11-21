package not_typename

var S int

type T []int

var A interface{}
var B = A.(S).(T)
