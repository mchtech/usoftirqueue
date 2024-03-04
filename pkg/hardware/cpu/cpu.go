package cpu

import (
	"sort"

	"github.com/jaypipes/ghw"
	"github.com/jaypipes/ghw/pkg/cpu"
	"github.com/jaypipes/ghw/pkg/option"
	"k8s.io/klog/v2"
)

type Topology struct {
	*ghw.CPUInfo
}

func List() (*Topology, error) {
	c, err := getCPUInfo()
	if err != nil {
		return nil, err
	}
	return &Topology{c}, nil
}

func getCPUInfo() (cpus *cpu.Info, err error) {
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

func (c *Topology) NumCPU() int {
	return int(c.TotalThreads)
}

func (c *Topology) NumLocalCPU(numaID int) int {
	if numaID < 0 {
		return c.NumCPU()
	}
	return int(c.Processors[numaID].NumThreads)
}

func (c *Topology) NumNUMA() int {
	return len(c.Processors)
}

// CoreIDs returns the core IDs of the given NUMA node.
// numaID < 0 means all NUMA nodes.
func (c *Topology) CoreIDs(numaID int, excludeHyperThread bool) (r []int) {
	processors := c.Processors
	if numaID >= 0 {
		processors = []*cpu.Processor{processors[numaID]}
	}
	for _, p := range processors {
		for _, c := range p.Cores {
			if excludeHyperThread {
				r = append(r, c.LogicalProcessors[0])
			} else {
				r = append(r, c.LogicalProcessors...)
			}
		}
	}
	return
}
