package cpu

import (
	"testing"
)

func TestGetCPUInfo(t *testing.T) {
	cpu, err := List()
	if err != nil {
		t.Errorf("GetCPUInfo error: %v", err)
		return
	}
	cpus := cpu.CoreIDs(-1, false)
	if len(cpus) != int(cpu.TotalThreads) {
		t.Errorf("GetCoreList error: %v", err)
		return
	}
}
