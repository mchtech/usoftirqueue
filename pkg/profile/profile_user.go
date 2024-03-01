package profile

import (
	"usoftirqueue/pkg/option"

	"k8s.io/klog/v2"
)

func (p *Profile) SetupUser(pl string, opts ...option.OptionFunc) (err error) {
	klog.InfoS("Setting up user policy", "name", pl, "options", opts)
	return p.policies[pl](opts...).Apply()
}
