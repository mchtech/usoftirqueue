package util

import (
	"runtime"
	"testing"
	"usoftirqueue/pkg/system"
)

// test IterateCPU
func TestIterateCPU(t *testing.T) {

	cpus, err := system.GetCPUInfo()
	if err != nil {
		t.Errorf("GetCPUInfo error: %v", err)
		return
	}

	var result, result2 []int
	count := 10
	for i := 0; i < count; i++ {
		result = append(result, GetCoreList(cpus, -1, false)...)
	}

	iterator, err := IterateCPU(cpus, -1, false)
	if err != nil {
		t.Errorf("IterateCPU error: %v", err)
		return
	}

	for i := 0; i < count*runtime.NumCPU(); i++ {
		result2 = append(result2, iterator())
	}

	if !IsSameSlice(result, result2) {
		t.Errorf("IterateCPU result: %v, want: %v", result2, result)
		return

	}
	t.Logf("IterateCPU result: %v", result2)
}
