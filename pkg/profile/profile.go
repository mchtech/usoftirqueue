package profile

import (
	"usoftirqueue/pkg/option"
	"usoftirqueue/pkg/policy"
	"usoftirqueue/pkg/system"

	"github.com/jaypipes/ghw/pkg/cpu"
	"k8s.io/klog/v2"
)

type Profile struct {
	// params
	dryrun    bool
	largeNode int
	// vars
	cpus *cpu.Info
	nics []string
	// spec
	policies map[string]func(opts ...option.OptionFunc) policy.Policy
}

func New(dryrun bool, nicRegex string, largeNode, expectConnections int) (p *Profile, err error) {

	var (
		cpus *cpu.Info
		nics []string
	)

	// get CPU count
	if cpus, err = system.GetCPUInfo(); err != nil {
		klog.ErrorS(err, "failed to get CPU info")
		return nil, err
	}
	klog.InfoS("current CPU", "total cpus", cpus.TotalThreads, "local cpus", cpus.Processors[0].NumThreads)

	// get NICs
	if nics, err = system.GetNICs(nicRegex); err != nil {
		klog.ErrorS(err, "failed to get NICs", "regexp", nicRegex)
		return nil, err
	}
	klog.InfoS("current NICs", "nics", nics, "filter", nicRegex)

	p = &Profile{
		dryrun:    dryrun,
		largeNode: largeNode,
		cpus:      cpus,
		nics:      nics,
		policies: map[string]func(opts ...option.OptionFunc) policy.Policy{
			policy.Spread: func(opts ...option.OptionFunc) policy.Policy {
				return policy.NewSpreadPolicy(dryrun, cpus, nics, opts...)
			},
			policy.Isolate: func(opts ...option.OptionFunc) policy.Policy {
				var newOpts []option.OptionFunc
				if expectConnections != 0 {
					newOpts = append(newOpts, option.WithExpectConnections(expectConnections))
				}
				newOpts = append(newOpts, opts...)
				return policy.NewIsolatePolicy(dryrun, cpus, nics, newOpts...)
			},
		},
	}
	return
}
