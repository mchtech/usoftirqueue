package profile

import (
	"usoftirqueue/pkg/policy"

	"k8s.io/klog/v2"
)

func (p *Profile) SetupKVM() (err error) {

	numCPU := p.cpus.NumCPU()

	// large node: prefer to redirect softirqs to application cores
	if numCPU > p.largeNode {
		klog.Info("numCPU > threshold: use steering()")
		return p.policies[policy.Steering](
		// option.WithCPUSelectStrategy(option.CPUSelectStrategyScatter), // default
		// option.WithCPUSelectMask(option.CPUSelectMaskAll),             // default
		// option.WithTrafficSource(option.TrafficSourceFromOneOfNIC),    // default
		).Apply()
	}

	// check max irqs
	if len(p.nics) > 0 && len(p.nics[0].IRQs) < numCPU {
		klog.Info("numIRQ < numCPU: use steering()")
		return p.policies[policy.Steering](
		// option.WithCPUSelectStrategy(option.CPUSelectStrategyScatter), // default
		// option.WithCPUSelectMask(option.CPUSelectMaskAll),             // default
		// option.WithTrafficSource(option.TrafficSourceFromOneOfNIC),    // default
		).Apply()
	}

	klog.Info("use static()")
	return p.policies[policy.Static](
	// option.WithCPUSelectStrategy(option.CPUSelectStrategyScatter), // default
	// option.WithCPUSelectMask(option.CPUSelectMaskAll),             // default
	// option.WithTrafficSource(option.TrafficSourceFromOneOfNIC),    // default
	).Apply()
}
