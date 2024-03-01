package profile

import (
	"usoftirqueue/pkg/option"
	"usoftirqueue/pkg/policy"
)

func (p *Profile) SetupBM() (err error) {
	return p.policies[policy.Isolate](
		option.WithCPUSelectPolicy(option.CPUSelectPolicyFromFirstCore),
		option.WithCPUSelectMask(option.CPUSelectMaskAll),
		option.WithFlowMode(option.FlowModeFromOneOfNIC),
	).Apply()
}
