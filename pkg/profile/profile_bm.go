package profile

import (
	"usoftirqueue/pkg/policy"
)

func (p *Profile) SetupBM() (err error) {
	return p.policies[policy.Isolate](
	// option.WithCPUSelectPolicy(option.CPUSelectPolicyFromIterator),
	// option.WithCPUSelectMask(option.CPUSelectMaskAuto),
	// option.WithFlowMode(option.FlowModeFromOneOfNIC),
	).Apply()
}
