package policy

import (
	"usoftirqueue/pkg/hardware/nic"
	"usoftirqueue/pkg/option"
	"usoftirqueue/pkg/tuner"

	"github.com/jaypipes/ghw/pkg/cpu"
	"k8s.io/klog/v2"
)

const Spread = "spread"

var _ Policy = &SpreadPolicy{}

type SpreadPolicy struct {
	Policy
	name   string
	option *option.Option
	tuner  *tuner.Tuner
}

func NewSpreadPolicy(cpu *cpu.Info, nics []*nic.NIC, opts ...option.OptionFunc) Policy {
	return &SpreadPolicy{
		Policy: NewCommonPolicy(cpu, nics),
		name:   Spread,
		option: option.ParseOptions(opts...),
		tuner:  tuner.New(),
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
	if err = p.tuner.StopIRQBalance(); err != nil {
		klog.ErrorS(err, "failed to stop irqbalance")
		return
	}

	// tune IRQs
	for _, eth := range p.NICs() {
		if err = p.tuner.TuneIRQ(eth, p.CPUs(), p.option); err != nil {
			klog.ErrorS(err, "failed to tune IRQs", "nic", eth)
			return
		}
	}

	// disable RPS RFS aRFS XPS
	if err = p.tuner.DisableRPSRFSXPS(p.NICs()); err != nil {
		klog.ErrorS(err, "failed to disable RPS RFS aRFS XPS", "nic", p.NICs())
		return
	}
	return
}
