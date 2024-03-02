package util

import (
	"testing"
	"usoftirqueue/pkg/system"
)

func TestGetCPUInfo(t *testing.T) {
	cpu, err := system.GetCPUInfo()
	if err != nil {
		t.Errorf("GetCPUInfo error: %v", err)
		return
	}
	cpus := GetCoreList(cpu, -1, false)
	if len(cpus) != int(cpu.TotalThreads) {
		t.Errorf("GetCoreList error: %v", err)
		return
	}
}

// test IterateCPU
func TestIterateCPU(t *testing.T) {

	cpus := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	expect1 := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 1, 2, 3, 4}
	expect2 := []int{5, 6, 7, 8, 9, 10, 11, 12, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	l1 := len(expect1)
	l2 := len(expect2)

	// actual
	ci := NewCircularIterator(cpus)
	actual11 := ci.NextN(l1)
	actual12 := ci.NextN(l2)

	oi := NewAccumulateOffsetIterator(cpus)
	actual21 := oi.NextN(l1)
	actual22 := oi.NextN(l2)

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

func TestIterator(t *testing.T) {

	data := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}
	it := NewAccumulateOffsetIterator(data)

	r1 := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}
	r2 := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 0}
	r3 := []int{2, 5, 8, 11}
	r4 := []int{3, 7, 11}
	r5 := []int{4, 7, 10, 1}

	c1 := it.NewInterval(1).NextN(12)
	c2 := it.NewInterval(1).NextN(12)
	c3 := it.NewInterval(3).NextN(4)
	c4 := it.NewInterval(4).NextN(3)
	c5 := it.NewInterval(3).NextN(4)

	t.Logf("expect %v, actual %v", r1, c1)
	t.Logf("expect %v, actual %v", r2, c2)
	t.Logf("expect %v, actual %v", r3, c3)
	t.Logf("expect %v, actual %v", r4, c4)
	t.Logf("expect %v, actual %v", r5, c5)

	if ok := IsSameSlice(r1, c1) && IsSameSlice(r2, c2) && IsSameSlice(r3, c3) && IsSameSlice(r4, c4) && IsSameSlice(r5, c5); !ok {
		t.Fail()
	}
}
