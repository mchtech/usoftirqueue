package policy

import (
	"usoftirqueue/pkg/nic"

	"github.com/jaypipes/ghw/pkg/cpu"
)

var (
	Policies []Policy
)

type Policy interface {
	Name() string
	Apply() error

	DryRun() bool
	CPUs() *cpu.Info
	NICs() []*nic.NIC
}
