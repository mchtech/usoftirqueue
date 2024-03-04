package profile

import (
	hcpu "usoftirqueue/pkg/hardware/cpu"
	"usoftirqueue/pkg/hardware/nic"
	"usoftirqueue/pkg/option"
	"usoftirqueue/pkg/policy"

	"github.com/jaypipes/ghw/pkg/cpu"
	"k8s.io/klog/v2"
)

type Profile struct {
	// params
	largeNode int
	// vars
	cpus *cpu.Info
	nics []*nic.NIC
	// spec
	policies map[string]func(opts ...option.OptionFunc) policy.Policy
}

func New(nicRegex string, largeNode, expectConnections int) (p *Profile, err error) {

	var (
		cpus *cpu.Info
		nics []*nic.NIC
	)

	// get CPU count
	if cpus, err = hcpu.GetCPUInfo(); err != nil {
		klog.ErrorS(err, "failed to get CPU info")
		return nil, err
	}
	klog.InfoS("current CPU", "total cpus", cpus.TotalThreads, "local cpus", cpus.Processors[0].NumThreads)

	// get NICs

	if nics, err = nic.List(nicRegex); err != nil {
		klog.ErrorS(err, "failed to get NICs", "regexp", nicRegex)
		return nil, err
	}
	klog.InfoS("current NICs", "nics", nic.NICList(nics).Names(), "filter", nicRegex)

	p = &Profile{
		largeNode: largeNode,
		cpus:      cpus,
		nics:      nics,
		policies: map[string]func(opts ...option.OptionFunc) policy.Policy{
			policy.Static: func(opts ...option.OptionFunc) policy.Policy {
				return policy.NewStaticPolicy(cpus, nics, opts...)
			},
			policy.Steering: func(opts ...option.OptionFunc) policy.Policy {
				var newOpts []option.OptionFunc
				if expectConnections != 0 {
					newOpts = append(newOpts, option.WithExpectConnections(expectConnections))
				}
				newOpts = append(newOpts, opts...)
				return policy.NewSteeringPolicy(cpus, nics, newOpts...)
			},
		},
	}
	return
}
