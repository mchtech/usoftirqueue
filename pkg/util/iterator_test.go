package util

import (
	"testing"
)

func TestRoundRobinIterator(t *testing.T) {

	cpus := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	expect1 := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 1, 2, 3, 4}
	expect2 := []int{5, 6, 7, 8, 9, 10, 11, 12, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	l1 := len(expect1)
	l2 := len(expect2)

	// actual
	rr := NewRoundRobinIterator(cpus)
	actual11 := rr.NextN(l1)
	actual12 := rr.NextN(l2)

	srr := NewScatterRoundRobinIterator(cpus)
	actual21 := srr.NextN(l1)
	actual22 := srr.NextN(l2)

	// diff
	if !IsSameSlice(expect1, actual11) || !IsSameSlice(expect2, actual12) {
		t.Errorf("CircularIterateCPU want: %v,%v, actual: %v,%v", expect1, expect2, actual11, actual12)
		return
	}
	t.Logf("CircularIterateCPU result: %v,%v", actual11, actual12)

	if !IsSameSlice(expect1, actual21) || !IsSameSlice(expect2, actual22) {
		t.Errorf("OffsetIterateCPU want: %v,%v, actual: %v,%v", expect1, expect2, actual21, actual22)
		return
	}
	t.Logf("OffsetIterateCPU result: %v,%v", actual11, actual12)
}

func TestScatterRoundRobinIterator(t *testing.T) {

	data := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}
	srr := NewScatterRoundRobinIterator(data)

	expects := [][]int{
		// 1
		{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 0},
		{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 0, 1},
		// 3
		{3, 6, 9, 0},
		{4, 7, 10, 1},
		{5, 8, 11, 2},
		{6, 9, 0, 3},
		// 4
		{7, 11, 3},
		{8, 0, 4, 8},
	}

	actuals := [][]int{
		// 1
		srr.NewInterval(1).NextN(12),
		srr.NewInterval(1).NextN(12),
		srr.NewInterval(1).NextN(12),
		// 3
		srr.NewInterval(3).NextN(4),
		srr.NewInterval(3).NextN(4),
		srr.NewInterval(3).NextN(4),
		srr.NewInterval(3).NextN(4),
		// 4
		srr.NewInterval(4).NextN(3),
		srr.NewInterval(4).NextN(4),
	}

	for i := 0; i < len(expects); i++ {
		t.Logf("expect %v, actual %v", expects[i], actuals[i])
		if ok := IsSameSlice(expects[i], actuals[i]); !ok {
			t.Fail()
		}
	}
}
