package policy

import (
	"usoftirqueue/pkg/hardware/nic"

	"github.com/jaypipes/ghw/pkg/cpu"
)

var (
	Policies []Policy
)

type Policy interface {
	Name() string
	CPUs() *cpu.Info
	NICs() []*nic.NIC
	Apply() error
}
