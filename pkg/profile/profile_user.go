package profile

import (
	"usoftirqueue/pkg/option"
	"usoftirqueue/pkg/policy"

	"k8s.io/klog/v2"
)

func (p *Profile) SetupUser(pl policy.IRQPolicy, opts ...option.OptionFunc) (err error) {
	klog.InfoS("Setting up user policy", "name", pl, "options", opts)
	return p.policies[pl](opts...).Apply()
}
