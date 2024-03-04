package profile

import (
	"usoftirqueue/pkg/option"
	"usoftirqueue/pkg/policy"
)

func (p *Profile) SetupECS() (err error) {

	numCPU := p.cpus.NumCPU()

	// no standard cores: 12c / 24c / 96c / 104c / 192c ... : numCPU != numIRQ
	if numCPU&(numCPU-1) != 0 {
		return p.policies[policy.Steering](
			// option.WithCPUSelectStrategy(option.CPUSelectStrategyScatter), // default
			// option.WithCPUSelectMask(option.CPUSelectMaskAll),         // default
			option.WithFlowMode(option.FlowModeCloudK8sENI),
		).Apply()
	}

	// large node: prefer to redirect softirqs to application cores
	if numCPU > p.largeNode {
		return p.policies[policy.Steering](
			// option.WithCPUSelectStrategy(option.CPUSelectStrategyScatter), // default
			// option.WithCPUSelectMask(option.CPUSelectMaskAll),         // default
			option.WithFlowMode(option.FlowModeCloudK8sENI),
		).Apply()
	}

	// standard 2^n cores: 2c / 4c / 8c ... : numCPU == numIRQ
	return p.policies[policy.Static](
	// option.WithCPUSelectStrategy(option.CPUSelectStrategyScatter), // default
	// option.WithCPUSelectMask(option.CPUSelectMaskAll),         // default
	).Apply()
}
