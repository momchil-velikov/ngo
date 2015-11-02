// Package sort contains sorting algorithms, operating on "generic"-like
// slices with concrete key types.
package sort

const sizeThreshold = 32

type StringKey struct {
	Key   string
	Value interface{}
}

type StringKeySlice []StringKey

func (a StringKeySlice) InsertionSort() {
	n := len(a)
	if n <= 1 {
		return
	}
	for i := 0; i < n-1; i++ {
		j := i + 1
		x := a[j]
		for ; j > 0 && x.Key < a[j-1].Key; j-- {
			a[j] = a[j-1]
		}
		a[j] = x
	}
}

func (a StringKeySlice) Partition() (int, int) {
	k := a[len(a)/2].Key
	i, j := 0, len(a)-1
	for {
		for k < a[j].Key {
			j--
		}
		for a[i].Key < k {
			i++
		}
		if i < j {
			a[i], a[j] = a[j], a[i]
			i++
			j--
		} else {
			return i, j
		}
	}
}

func (a StringKeySlice) Quicksort() {
	for len(a) > sizeThreshold {
		i, j := a.Partition()
		if i > len(a)-j-1 {
			a[j+1:].Quicksort()
			a = a[:i]
		} else {
			a[:i].Quicksort()
			a = a[j+1:]
		}
	}
	a.InsertionSort()
}
