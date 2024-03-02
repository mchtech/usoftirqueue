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

	master eni:
		eth0: 32 queues: 64 irqs

	terway eni:
		eth1: 16 queues: 32 irqs
		...
		ethX: 16 queues: 32 irqs

TODO: other instance type (AMD EPYC: 8 more NUMA nodes)

*/

func (p *Profile) SetupEBM() (err error) {
	if len(p.nics) > 2 {
		klog.InfoS("num nics > 2",
			"CPUSelectPolicy", "CPUSelectPolicyFromFirstCore --> CPUSelectPolicyFromPreviousCore",
			"CPUSelectMask", "CPUSelectMaskAuto --> CPUSelectMaskAll",
			"FlowMode", "FlowModeFromOneOfNIC --> FlowModeSplitMainNIC",
		)
		return p.policies[policy.Isolate](
			option.WithCPUSelectPolicy(option.CPUSelectPolicyFromPreviousCore),
			option.WithCPUSelectMask(option.CPUSelectMaskAll),
			option.WithFlowMode(option.FlowModeSplitMainNIC),
		).Apply()
	}

	klog.InfoS("num nics ≤ 2",
		"CPUSelectPolicy", "CPUSelectPolicyFromFirstCore",
		"CPUSelectMask", "CPUSelectMaskAuto --> CPUSelectMaskLocalNUMA",
		"FlowMode", "FlowModeFromOneOfNIC --> FlowModeSplitMainNIC",
	)
	return p.policies[policy.Isolate](
		option.WithCPUSelectPolicy(option.CPUSelectPolicyFromFirstCore),
		option.WithCPUSelectMask(option.CPUSelectMaskLocalNUMA),
		option.WithFlowMode(option.FlowModeSplitMainNIC),
	).Apply()
}
