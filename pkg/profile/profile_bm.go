package profile

import (
	"usoftirqueue/pkg/option"
	"usoftirqueue/pkg/policy"
)

func (p *Profile) SetupBM() (err error) {
	return p.policies[policy.Isolate](
		// option.WithCPUSelectPolicy(option.CPUSelectPolicyScatter), // default
		// option.WithCPUSelectMask(option.CPUSelectMaskAll),         // default
		option.WithFlowMode(option.FlowModeFromOneOfNIC),
	).Apply()
}
