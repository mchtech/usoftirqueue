package policy

import (
	"math"
	"usoftirqueue/pkg/irq"
	"usoftirqueue/pkg/nic"
	"usoftirqueue/pkg/option"
	"usoftirqueue/pkg/tuner"

	"github.com/jaypipes/ghw/pkg/cpu"
	"k8s.io/klog/v2"
)

var (
	BaseConnectionsPerCore = int(math.Exp2(14))
	CoreToConnections      = map[int]int{
		1:   int(math.Exp2(16)),
		2:   int(math.Exp2(17)),
		4:   int(math.Exp2(18)),
		8:   int(math.Exp2(18)),
		16:  int(math.Exp2(18)),
		32:  int(math.Exp2(19)),
		64:  int(math.Exp2(20)),
		128: int(math.Exp2(21)),
		256: int(math.Exp2(22)),
	}
)

const Isolate = "isolate"

var _ Policy = &IsolatePolicy{}

type IsolatePolicy struct {
	Policy
	name   string
	option *option.Option
	plan   *tuner.Tuner
}

func NewIsolatePolicy(dryrun bool, cpu *cpu.Info, nics []*nic.NIC, opts ...option.OptionFunc) Policy {
	return &IsolatePolicy{
		Policy: NewCommonPolicy(dryrun, cpu, nics),
		name:   Isolate,
		option: option.ParseOptions(opts...),
		plan:   &tuner.Tuner{},
	}
}

func (p *IsolatePolicy) Name() string {
	return p.name
}

func (p *IsolatePolicy) ExpectConnections() int {
	if p.option.ExpectConnections == 0 {
		cpuCount := int(p.CPUs().TotalThreads)
		c, ok := CoreToConnections[cpuCount]
		if !ok {
			c = BaseConnectionsPerCore * cpuCount
		}
		p.option.ExpectConnections = c
	}
	return p.option.ExpectConnections
}

func (p *IsolatePolicy) Apply() (err error) {

	klog.InfoS("apply policy", "name", p.Name())

	if err = p.Policy.Apply(); err != nil {
		return
	}
	if err = p.apply(); err != nil {
		return
	}
	return
}

func (p *IsolatePolicy) apply() (err error) {

	// stop irqbalance
	if err = irq.StopIRQBalance(p.DryRun()); err != nil {
		klog.ErrorS(err, "failed to stop irqbalance")
		return
	}

	// tune IRQs
	for _, eth := range p.NICs() {
		if err = p.plan.TuneIRQ(eth, p.CPUs(), p.DryRun(), p.option); err != nil {
			klog.ErrorS(err, "failed to tune IRQs", "nic", eth.Name)
			return
		}
	}

	// enable RPS RFS aRFS XPS
	if err = p.plan.EnableRPSRFSXPS(
		p.NICs(),
		int(p.CPUs().TotalThreads),
		p.ExpectConnections(),
		p.DryRun(),
		p.option,
	); err != nil {
		klog.ErrorS(err, "failed to disable RPS RFS aRFS XPS", "nic", p.NICs())
		return
	}
	return
}
