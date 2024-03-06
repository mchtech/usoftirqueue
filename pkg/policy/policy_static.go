package policy

import (
	"usoftirqueue/pkg/hardware/cpu"
	"usoftirqueue/pkg/hardware/nic"
	"usoftirqueue/pkg/option"
	"usoftirqueue/pkg/tuner"

	"k8s.io/klog/v2"
)

const Static IRQPolicy = "static"

var _ Policy = &StaticPolicy{}

type StaticPolicy struct {
	*CommonPolicy
	name   string
	option *option.Option
	tuner  *tuner.Tuner
}

func NewStaticPolicy(cpu *cpu.Topology, nics []*nic.NIC, opts ...option.OptionFunc) *StaticPolicy {
	return &StaticPolicy{
		CommonPolicy: NewCommonPolicy(cpu, nics),
		name:         string(Static),
		option:       option.ParseOptions(opts...),
		tuner:        tuner.New(),
	}
}

func (p *StaticPolicy) Name() string {
	return p.name
}

func (p *StaticPolicy) Apply() (err error) {

	klog.InfoS("apply policy", "name", p.Name())

	if err = p.CommonPolicy.Apply(); err != nil {
		return
	}
	if err = p.apply(); err != nil {
		return
	}
	return
}

func (p *StaticPolicy) apply() (err error) {

	// stop irqbalance
	if err = p.tuner.StopIRQBalance(); err != nil {
		klog.ErrorS(err, "failed to stop irqbalance")
		return
	}

	// tune IRQs
	for _, eth := range p.nics {
		if err = p.tuner.TuneIRQ(eth, p.cpus, p.option); err != nil {
			klog.ErrorS(err, "failed to tune IRQs", "nic", eth)
			return
		}
	}

	// disable RPS RFS aRFS XPS
	if err = p.tuner.DisableRPSRFSXPS(p.nics); err != nil {
		klog.ErrorS(err, "failed to disable RPS RFS aRFS XPS", "nic", p.nics)
		return
	}
	return
}
