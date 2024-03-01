package profile

import (
	"usoftirqueue/pkg/option"
	"usoftirqueue/pkg/policy"
)

func (p *Profile) SetupECS() (err error) {

	numCPU := int(p.cpus.TotalThreads)

	// large node: prefer to redirect softirqs to application cores
	if numCPU > p.largeNode {
		return p.policies[policy.Isolate](
			// option.WithCPUSelectPolicy(option.CPUSelectPolicyFromFirstCore), // default
			// option.WithCPUSelectMask(option.CPUSelectMaskAuto), // default
			option.WithFlowMode(option.FlowModeSplitMainNIC),
		).Apply()
	}

	// no standard cores: 6c / 12c / 24c ... : numCPU != numIRQ
	if numCPU&(numCPU-1) != 0 {
		return p.policies[policy.Isolate](
			// option.WithCPUSelectPolicy(option.CPUSelectPolicyFromFirstCore), // default
			// option.WithCPUSelectMask(option.CPUSelectMaskAuto), // default
			option.WithFlowMode(option.FlowModeSplitMainNIC),
		).Apply()
	}

	// standard cores: 2c / 4c / 8c ... : numCPU == numIRQ
	return p.policies[policy.Spread]().Apply()
}
