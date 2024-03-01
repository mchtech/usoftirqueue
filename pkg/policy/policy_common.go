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
	dryrun, dryrunQueue bool
	// vars
	cpus *cpu.Info
	nics []string
	// option
	option *option.Option
}

func NewCommonPolicy(dryrun, dryrunQueue bool, cpu *cpu.Info, nics []string, opts ...option.OptionFunc) Policy {
	return &CommonPolicy{
		name:   Common,
		dryrun: dryrun,
		cpus:   cpu,
		nics:   nics,
		option: option.ParseOptions(opts...),

		dryrunQueue: dryrunQueue,
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

func (p *CommonPolicy) NICs() []string {
	return p.nics
}

func (p *CommonPolicy) Apply() (err error) {

	klog.InfoS("apply policy", "name", p.Name())

	if p.dryrun && !p.dryrunQueue {
		klog.Warning("in order to get all IRQs, NIC queue needs to tuned first: ignore `-dryrun` param in this step")
	}

	for _, eth := range p.nics {

		// tune NIC queue channel to max
		if err = nic.TuneNICQueueChannelMax(p.cpus.Processors[0].NumThreads, p.cpus.TotalThreads, eth, p.dryrunQueue); err != nil {
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
