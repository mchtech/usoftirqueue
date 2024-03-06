package profile

import (
	"usoftirqueue/pkg/option"
	"usoftirqueue/pkg/policy"
)

func (p *Profile) SetupBM() (err error) {
	return p.policies[policy.Steering](
		// option.WithCPUSelectStrategy(option.CPUSelectStrategyScatter), // default
		// option.WithCPUSelectMask(option.CPUSelectMaskAll),             // default
		option.WithTrafficSource(option.TrafficSourceFromOneOfNIC),
	).Apply()
}
