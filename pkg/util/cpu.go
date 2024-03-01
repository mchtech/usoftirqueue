package util

import (
	"fmt"
	"sync"

	"github.com/jaypipes/ghw"
	"github.com/jaypipes/ghw/pkg/cpu"
)

func GetCoreList(cpus *cpu.Info, numa int, excludeHT bool) (r []int) {
	processors := cpus.Processors
	if numa >= 0 {
		processors = []*cpu.Processor{processors[numa]}
	}
	for _, p := range processors {
		for _, c := range p.Cores {
			if excludeHT {
				r = append(r, c.LogicalProcessors[0])
			} else {
				r = append(r, c.LogicalProcessors...)
			}
		}
	}
	return
}

var (
	once         sync.Once
	onceIterator func() int
)

func GetCoreListByIterator(cpus *cpu.Info, numa int, excludeHT bool, count int) (r []int) {
	once.Do(func() {
		onceIterator, _ = IterateCPU(cpus, numa, excludeHT)
	})
	for i := 0; i < count; i++ {
		r = append(r, onceIterator())
	}
	return
}

// generate Iterator for cpus
func IterateCPU(cpu *ghw.CPUInfo, numaID int, excludeHT bool) (iterator func() int, err error) {

	var (
		p, c, t int
	)

	if numaID >= 0 && numaID >= len(cpu.Processors) {
		err = fmt.Errorf("numa %d out of range", numaID)
		return
	}

	if numaID >= 0 {
		p = numaID
	}

	iterator = func() int {
		if t >= len(cpu.Processors[p].Cores[c].LogicalProcessors) {
			t = 0
			c++
			if c >= len(cpu.Processors[p].Cores) {
				c = 0
				if numaID < 0 {
					p++
					if p >= len(cpu.Processors) {
						p = 0
					}
				}
			}
		}
		v := cpu.Processors[p].Cores[c].LogicalProcessors[t]
		if !excludeHT {
			t++
		}
		return v
	}
	return
}
