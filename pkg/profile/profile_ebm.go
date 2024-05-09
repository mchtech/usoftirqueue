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

	if len(p.nics) > 2 {

		// no 2^n cores: 12c / 24c / 96c / 104c / 152c / 192c / 384c ... : numCPU != numIRQ
		if numCPU&(numCPU-1) != 0 {
			klog.Info("num nic > 2 and num cpu != 2^n, so numIRQ != numCPU: use steering(cpu_roundrobin, cloud_k8s_eni)")
			return p.policies[policy.Steering](
				option.WithCPUSelectStrategy(option.CPUSelectStrategyRoundRobin),
				// option.WithCPUSelectMask(option.CPUSelectMaskAll),             // default
				option.WithTrafficSource(option.TrafficSourceCloudK8sENI),
			).Apply()
		}

		klog.Info("num nic > 2, use steering(cloud_k8s_eni)")
		return p.policies[policy.Steering](
			// option.WithCPUSelectStrategy(option.CPUSelectStrategyScatter), // default
			// option.WithCPUSelectMask(option.CPUSelectMaskAll),             // default
			option.WithTrafficSource(option.TrafficSourceCloudK8sENI),
		).Apply()
	} else {

		// no 2^n cores: 12c / 24c / 96c / 104c / 152c / 192c / 384c ... : numCPU != numIRQ
		if numCPU&(numCPU-1) != 0 {
			klog.Info("num nic ≤ 2 and num cpu != 2^n, so numIRQ != numCPU: use steering(cpu_roundrobin, cpu_local, cloud_k8s_eni)")
			return p.policies[policy.Steering](
				option.WithCPUSelectStrategy(option.CPUSelectStrategyRoundRobin),
				option.WithCPUSelectMask(option.CPUSelectMaskLocalNUMA),
				option.WithTrafficSource(option.TrafficSourceCloudK8sENI),
			).Apply()
		}

		klog.Info("num nic ≤ 2, use steering(cpu_local, cloud_k8s_eni)")
		return p.policies[policy.Steering](
			// option.WithCPUSelectStrategy(option.CPUSelectStrategyScatter), // default
			option.WithCPUSelectMask(option.CPUSelectMaskLocalNUMA),
			option.WithTrafficSource(option.TrafficSourceCloudK8sENI),
		).Apply()
	}
}
