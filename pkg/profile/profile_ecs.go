package profile

import (
	"usoftirqueue/pkg/option"
	"usoftirqueue/pkg/policy"
)

func (p *Profile) SetupECS() (err error) {

	numCPU := int(p.cpus.TotalThreads)

	// no standard cores: 12c / 24c / 96c / 104c / 192c ... : numCPU != numIRQ
	if numCPU&(numCPU-1) != 0 {
		return p.policies[policy.Steering](
			// option.WithCPUSelectPolicy(option.CPUSelectPolicyScatter), // default
			// option.WithCPUSelectMask(option.CPUSelectMaskAll),         // default
			option.WithFlowMode(option.FlowModeSplitMainNIC),
		).Apply()
	}

	// large node: prefer to redirect softirqs to application cores
	if numCPU > p.largeNode {
		return p.policies[policy.Steering](
			// option.WithCPUSelectPolicy(option.CPUSelectPolicyScatter), // default
			// option.WithCPUSelectMask(option.CPUSelectMaskAll),         // default
			option.WithFlowMode(option.FlowModeSplitMainNIC),
		).Apply()
	}

	// standard 2^n cores: 2c / 4c / 8c ... : numCPU == numIRQ
	return p.policies[policy.Static](
	// option.WithCPUSelectPolicy(option.CPUSelectPolicyScatter), // default
	// option.WithCPUSelectMask(option.CPUSelectMaskAll),         // default
	).Apply()
}
