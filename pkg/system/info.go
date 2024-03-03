package system

import (
	"sort"

	"github.com/jaypipes/ghw"
	"github.com/jaypipes/ghw/pkg/cpu"
	"github.com/jaypipes/ghw/pkg/option"
	"k8s.io/klog/v2"
)

func GetCPUInfo() (cpus *cpu.Info, err error) {
	if cpus, err = ghw.CPU(option.WithNullAlerter(), option.WithDisableTools()); err != nil {
		return
	}
	for p := range cpus.Processors {
		cpus.Processors[p].Capabilities = nil
		sort.Slice(cpus.Processors[p].Cores, func(j, k int) bool {
			return cpus.Processors[p].Cores[j].ID < cpus.Processors[p].Cores[k].ID
		})
		for c := range cpus.Processors[p].Cores {
			sort.Slice(cpus.Processors[p].Cores[c].LogicalProcessors, func(j, k int) bool {
				return cpus.Processors[p].Cores[c].LogicalProcessors[j] < cpus.Processors[p].Cores[c].LogicalProcessors[k]
			})
		}
	}
	klog.InfoS("get cpus", "cpus", cpus.YAMLString())
	return
}
