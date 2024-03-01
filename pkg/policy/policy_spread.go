package policy

import (
	"usoftirqueue/pkg/nic"
	"usoftirqueue/pkg/option"

	"github.com/jaypipes/ghw/pkg/cpu"
	"k8s.io/klog/v2"
)

const Spread = "spread"

var _ Policy = &SpreadPolicy{}

type SpreadPolicy struct {
	Policy
	name   string
	option *option.Option
}

func NewSpreadPolicy(dryrun, dryrunQueue bool, cpu *cpu.Info, nics []string, opts ...option.OptionFunc) Policy {
	return &SpreadPolicy{
		Policy: NewCommonPolicy(dryrun, dryrunQueue, cpu, nics),
		name:   Spread,
		option: option.ParseOptions(opts...),
	}
}

func (p *SpreadPolicy) Name() string {
	return p.name
}

func (p *SpreadPolicy) Apply() (err error) {

	klog.InfoS("apply policy", "name", p.Name())

	if err = p.Policy.Apply(); err != nil {
		return
	}
	if err = p.apply(); err != nil {
		return
	}
	return
}

func (p *SpreadPolicy) apply() (err error) {

	// stop irqbalance
	if err = nic.StopIRQBalance(p.DryRun()); err != nil {
		klog.ErrorS(err, "failed to stop irqbalance")
		return
	}

	// tune IRQs
	for _, eth := range p.NICs() {
		if err = nic.TuneNICIRQ(eth, p.CPUs(), p.DryRun(), p.option); err != nil {
			klog.ErrorS(err, "failed to tune IRQs", "nic", eth)
			return
		}
	}

	// disable RPS RFS aRFS XPS
	if err = nic.DisableRPSRFSXPS(p.NICs(), p.DryRun()); err != nil {
		klog.ErrorS(err, "failed to disable RPS RFS aRFS XPS", "nic", p.NICs())
		return
	}
	return
}
