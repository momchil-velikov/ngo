package sort

import (
	"fmt"
	"math/rand"
	"testing"
)

const array_size = 10000
const key_range = 500000000

var (
	workArray   []StringKey = nil
	randomArray []StringKey = nil
)

func makeRandomArray(k int) StringKeySlice {
	a := make([]StringKey, array_size)
	for i := 0; i < array_size; i++ {
		a[i].Key = fmt.Sprintf("%d", rand.Intn(k))
	}
	return a
}

func randomShuffle(a []int) {
	for i := len(a); i > 1; i-- {
		j := rand.Intn(i)
		a[i-1], a[j] = a[j], a[i-1]
	}
}

func initRandomArray() {
	if randomArray == nil {
		randomArray = makeRandomArray(key_range)
	}
	if workArray == nil {
		workArray = make([]StringKey, array_size)
	}
}

func checkSorted(a []StringKey) bool {
	n := len(a)
	for i := 1; i < n; i++ {
		if a[i].Key < a[i-1].Key {
			return false
		}
	}
	return true
}

func Test_QS(t *testing.T) {
	initRandomArray()
	copy(workArray, randomArray)
	StringKeySlice(workArray).Quicksort()
	if !checkSorted(workArray) {
		t.Error("Array not sorted")
	}
}
