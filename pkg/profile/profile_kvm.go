package profile

import (
	"usoftirqueue/pkg/policy"
)

func (p *Profile) SetupKVM() (err error) {

	numCPU := int(p.cpus.TotalThreads)

	// large node: prefer to redirect softirqs to application cores
	if numCPU > p.largeNode {
		return p.policies[policy.Isolate](
		// option.WithCPUSelectPolicy(option.CPUSelectPolicyScatter), // default
		// option.WithCPUSelectMask(option.CPUSelectMaskAll),         // default
		// option.WithFlowMode(option.FlowModeFromOneOfNIC),          // default
		).Apply()
	}
	// // TODO: check virtio queue
	// if isVirtIO, virtMaxNumQueue := true, 1; isVirtIO && virtMaxNumQueue == 1 && numCPU > virtMaxNumQueue {
	// 	return p.policies[policy.Isolate]().Apply()
	// }
	return p.policies[policy.Spread](
	// option.WithCPUSelectPolicy(option.CPUSelectPolicyScatter), // default
	// option.WithCPUSelectMask(option.CPUSelectMaskAll),         // default
	// option.WithFlowMode(option.FlowModeFromOneOfNIC),          // default
	).Apply()
}
