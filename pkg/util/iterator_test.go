package util

import (
	"testing"
)

func TestIterator(t *testing.T) {

	data := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}
	it := NewBiasIterator(data)

	r1 := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}
	r2 := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 0}
	r3 := []int{2, 5, 8, 11}
	r4 := []int{3, 7, 11}
	r5 := []int{4, 7, 10, 1}

	c1 := it.Interval(1).Generate(12)
	c2 := it.Interval(1).Generate(12)
	c3 := it.Interval(3).Generate(4)
	c4 := it.Interval(4).Generate(3)
	c5 := it.Interval(3).Generate(4)

	t.Logf("expect %v, actual %v", r1, c1)
	t.Logf("expect %v, actual %v", r2, c2)
	t.Logf("expect %v, actual %v", r3, c3)
	t.Logf("expect %v, actual %v", r4, c4)
	t.Logf("expect %v, actual %v", r5, c5)

	if ok := IsSameSlice(r1, c1) && IsSameSlice(r2, c2) && IsSameSlice(r3, c3) && IsSameSlice(r4, c4) && IsSameSlice(r5, c5); !ok {
		t.Fail()
	}
	// t.Fail()
}
