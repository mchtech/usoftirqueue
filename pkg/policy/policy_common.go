package policy

import (
	"usoftirqueue/pkg/nic"
	"usoftirqueue/pkg/option"

	"github.com/jaypipes/ghw/pkg/cpu"
	"k8s.io/klog/v2"
)

const Common = "common"

var _ Policy = &CommonPolicy{}

type CommonPolicy struct {
	// interface
	name string
	// params
	dryrun bool
	// vars
	cpus *cpu.Info
	nics []*nic.NIC
	// option
	option *option.Option
}

func NewCommonPolicy(dryrun bool, cpu *cpu.Info, nics []*nic.NIC, opts ...option.OptionFunc) Policy {
	return &CommonPolicy{
		name:   Common,
		dryrun: dryrun,
		cpus:   cpu,
		nics:   nics,
		option: option.ParseOptions(opts...),
	}
}

func (p *CommonPolicy) Name() string {
	return p.name
}

func (p *CommonPolicy) DryRun() bool {
	return p.dryrun
}

func (p *CommonPolicy) CPUs() *cpu.Info {
	return p.cpus
}

func (p *CommonPolicy) NICs() []*nic.NIC {
	return p.nics
}

func (p *CommonPolicy) Apply() (err error) {

	klog.InfoS("apply policy", "name", p.Name())

	for _, eth := range p.nics {

		// tune NIC queue channel to max
		if err = nic.TuneNICQueueChannelMax(eth, p.cpus.Processors[0].NumThreads, p.cpus.TotalThreads, p.dryrun); err != nil {
			klog.ErrorS(err, "failed to tune NIC queue channel", "nic", eth)
			break
		}
		klog.InfoS("finish tune NIC queue channel", "nic", eth)

		// tune NIC ring buffer to max
		if err = nic.TuneNICRingBufferMax(eth, p.dryrun); err != nil {
			klog.ErrorS(err, "failed to tune NIC ring buffer", "nic", eth)
			break
		}
		klog.InfoS("finish tune NIC ring buffer", "nic", eth)
	}
	return
}
