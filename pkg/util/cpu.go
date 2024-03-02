package util

import "github.com/jaypipes/ghw/pkg/cpu"

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
