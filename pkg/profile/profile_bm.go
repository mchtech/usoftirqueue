package profile

import (
	"usoftirqueue/pkg/option"
	"usoftirqueue/pkg/policy"

	"k8s.io/klog/v2"
)

func (p *Profile) SetupBM() (err error) {
	klog.Info("use steering(one_of_nic)")
	return p.policies[policy.Steering](
		// option.WithCPUSelectStrategy(option.CPUSelectStrategyScatter), // default
		// option.WithCPUSelectMask(option.CPUSelectMaskAll),             // default
		option.WithTrafficSource(option.TrafficSourceFromOneOfNIC),
	).Apply()
}
