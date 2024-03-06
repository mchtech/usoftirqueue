package profile

import (
	"usoftirqueue/pkg/policy"
)

func (p *Profile) SetupKVM() (err error) {

	numCPU := p.cpus.NumCPU()

	// large node: prefer to redirect softirqs to application cores
	if numCPU > p.largeNode {
		return p.policies[policy.Steering](
		// option.WithCPUSelectStrategy(option.CPUSelectStrategyScatter), // default
		// option.WithCPUSelectMask(option.CPUSelectMaskAll),             // default
		// option.WithTrafficSource(option.TrafficSourceFromOneOfNIC),    // default
		).Apply()
	}

	// check max queue
	if len(p.nics) > 0 && len(p.nics[0].IRQs) < numCPU {
		return p.policies[policy.Steering](
		// option.WithCPUSelectStrategy(option.CPUSelectStrategyScatter), // default
		// option.WithCPUSelectMask(option.CPUSelectMaskAll),             // default
		// option.WithTrafficSource(option.TrafficSourceFromOneOfNIC),    // default
		).Apply()
	}

	return p.policies[policy.Static](
	// option.WithCPUSelectStrategy(option.CPUSelectStrategyScatter), // default
	// option.WithCPUSelectMask(option.CPUSelectMaskAll),             // default
	// option.WithTrafficSource(option.TrafficSourceFromOneOfNIC),    // default
	).Apply()
}
