package profile

import (
	"usoftirqueue/pkg/option"
	"usoftirqueue/pkg/policy"

	"k8s.io/klog/v2"
)

/*

ecs.ebmc7.32xlarge
	numa 0: 32 phy; 64 ht: 0-31,64-95
	numa 1: 32 phy; 64 ht: 32-63,96-127

	main eni:
		eth0: 32 queues: 64 irqs

	pod eni:
		eth1: 16 queues: 32 irqs
		...
		ethX: 16 queues: 32 irqs

TODO: other instance type (AMD EPYC: 8 more NUMA nodes)

*/

func (p *Profile) SetupEBM() (err error) {

	numCPU := p.cpus.NumCPU()

	// no standard cores: 12c / 24c / 96c / 104c / 192c ... : numCPU != numIRQ
	if numCPU&(numCPU-1) != 0 {
		return p.policies[policy.Steering](
			// option.WithCPUSelectPolicy(option.CPUSelectPolicyScatter), // default
			// option.WithCPUSelectMask(option.CPUSelectMaskAll),         // default
			option.WithFlowMode(option.FlowModeSplitMainNIC),
		).Apply()
	}

	if len(p.nics) > 2 {
		klog.InfoS("num nics > 2",
			"CPUSelectPolicy", "CPUSelectPolicyScatter",
			"CPUSelectMask", "CPUSelectMaskAll",
			"FlowMode", "FlowModeFromOneOfNIC --> FlowModeSplitMainNIC",
		)
		return p.policies[policy.Steering](
			// option.WithCPUSelectPolicy(option.CPUSelectPolicyScatter), // default
			// option.WithCPUSelectMask(option.CPUSelectMaskAll),         // default
			option.WithFlowMode(option.FlowModeSplitMainNIC),
		).Apply()
	} else {
		klog.InfoS("num nics ≤ 2",
			"CPUSelectPolicy", "CPUSelectPolicyScatter",
			"CPUSelectMask", "CPUSelectMaskLocalNUMA",
			"FlowMode", "FlowModeFromOneOfNIC --> FlowModeSplitMainNIC",
		)
		return p.policies[policy.Steering](
			// option.WithCPUSelectPolicy(option.CPUSelectPolicyScatter), // default
			option.WithCPUSelectMask(option.CPUSelectMaskLocalNUMA),
			option.WithFlowMode(option.FlowModeSplitMainNIC),
		).Apply()
	}
}
