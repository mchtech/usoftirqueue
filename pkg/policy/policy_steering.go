package policy

import (
	"math"
	"usoftirqueue/pkg/hardware/nic"
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

const Steering = "steering"

var _ Policy = &SteeringPolicy{}

type SteeringPolicy struct {
	*CommonPolicy
	name   string
	option *option.Option
	tuner  *tuner.Tuner
}

func NewSteeringPolicy(cpu *cpu.Info, nics []*nic.NIC, opts ...option.OptionFunc) *SteeringPolicy {
	return &SteeringPolicy{
		CommonPolicy: NewCommonPolicy(cpu, nics),
		name:         Steering,
		option:       option.ParseOptions(opts...),
		tuner:        tuner.New(),
	}
}

func (p *SteeringPolicy) Name() string {
	return p.name
}

func (p *SteeringPolicy) ExpectConnections() int {
	if p.option.ExpectConnections == 0 {
		cpuCount := int(p.cpus.TotalThreads)
		c, ok := CoreToConnections[cpuCount]
		if !ok {
			c = BaseConnectionsPerCore * cpuCount
		}
		p.option.ExpectConnections = c
	}
	return p.option.ExpectConnections
}

func (p *SteeringPolicy) Apply() (err error) {

	klog.InfoS("apply policy", "name", p.Name())

	if err = p.CommonPolicy.Apply(); err != nil {
		return
	}
	if err = p.apply(); err != nil {
		return
	}
	return
}

func (p *SteeringPolicy) apply() (err error) {

	// stop irqbalance
	if err = p.tuner.StopIRQBalance(); err != nil {
		klog.ErrorS(err, "failed to stop irqbalance")
		return
	}

	// tune IRQs
	for _, eth := range p.nics {
		if err = p.tuner.TuneIRQ(eth, p.cpus, p.option); err != nil {
			klog.ErrorS(err, "failed to tune IRQs", "nic", eth.Name)
			return
		}
	}

	// enable RPS RFS aRFS XPS
	if err = p.tuner.EnableRPSRFSXPS(
		p.nics,
		int(p.cpus.TotalThreads),
		p.ExpectConnections(),
		p.option,
	); err != nil {
		klog.ErrorS(err, "failed to disable RPS RFS aRFS XPS", "nic", p.nics)
		return
	}
	return
}
