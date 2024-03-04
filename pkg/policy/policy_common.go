package policy

import (
	"usoftirqueue/pkg/hardware/cpu"
	"usoftirqueue/pkg/hardware/nic"
	"usoftirqueue/pkg/tuner"

	// "github.com/jaypipes/ghw/pkg/cpu"
	"k8s.io/klog/v2"
)

const Common IRQPolicy = "common"

var _ Policy = &CommonPolicy{}

type CommonPolicy struct {
	// interface
	name string
	// vars
	cpus *cpu.Topology
	nics []*nic.NIC
	// exec
	tuner *tuner.Tuner
}

func NewCommonPolicy(cpu *cpu.Topology, nics []*nic.NIC) *CommonPolicy {
	return &CommonPolicy{
		name: string(Common),
		cpus: cpu,
		nics: nics,
		// option: option.ParseOptions(opts...),
		tuner: tuner.New(),
	}
}

func (p *CommonPolicy) Name() string {
	return p.name
}

func (p *CommonPolicy) Apply() (err error) {

	klog.InfoS("apply policy", "name", p.name)

	for _, eth := range p.nics {
		// tune NIC queue channel to max
		if err = p.tuner.TuneNICQueueChannelMax(eth, p.cpus.TotalThreads); err != nil {
			klog.ErrorS(err, "failed to tune NIC queue channel", "nic", eth.Name)
			break
		}
		klog.InfoS("finish tune NIC queue channel", "nic", eth.Name)
		// tune NIC ring buffer to max
		if err = p.tuner.TuneNICRingBufferMax(eth); err != nil {
			klog.ErrorS(err, "failed to tune NIC ring buffer", "nic", eth.Name)
			break
		}
		klog.InfoS("finish tune NIC ring buffer", "nic", eth.Name)
	}
	return
}
