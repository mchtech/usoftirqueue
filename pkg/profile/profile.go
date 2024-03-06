package profile

import (
	"usoftirqueue/pkg/hardware/board"
	"usoftirqueue/pkg/hardware/cpu"
	"usoftirqueue/pkg/hardware/nic"
	"usoftirqueue/pkg/option"
	"usoftirqueue/pkg/policy"

	// "github.com/jaypipes/ghw/pkg/cpu"
	"k8s.io/klog/v2"
)

type Profile struct {
	// params
	largeNode         int
	expectConnections int
	// vars
	board *board.BoardInfo
	cpus  *cpu.Topology
	nics  []*nic.NIC
	// spec
	policies map[policy.IRQPolicy]func(opts ...option.OptionFunc) policy.Policy
	// overrides
	userPolicy     policy.IRQPolicy
	userPolicyOpts []option.OptionFunc
}

func New(nicRegex string, largeNodeThresh, expectConnections int, userPolicy policy.IRQPolicy, userPolicyOpts ...option.OptionFunc) (p *Profile, err error) {

	var (
		bi   *board.BoardInfo
		cpus *cpu.Topology
		nics nic.NICList
	)

	// get NICs
	if nics, err = nic.List(nicRegex); err != nil {
		klog.ErrorS(err, "failed to get NICs", "regexp", nicRegex)
		return nil, err
	}
	klog.InfoS("target NICs", "filter", nicRegex, "nics", nics.Names(), "irqs", nics.IRQs())

	// get CPU topology
	if cpus, err = cpu.List(); err != nil {
		klog.ErrorS(err, "failed to get CPU topology")
		return nil, err
	}
	klog.InfoS("CPU topology", "numa", cpus.NumNUMA(), "total", cpus.NumCPU(), "local", cpus.NumLocalCPU(0), "ht", cpus.NumHyperThread(0), "cores", cpus.CoreIDs(-1, false))

	// board info
	if bi, err = board.Collect(); err != nil {
		klog.ErrorS(err, "failed to get board info")
		return
	}
	klog.InfoS("board info", "chassis", bi.ChassisVendor, "product", bi.ProductName, "bios", bi.BIOSVendor)

	// init profile
	p = &Profile{
		board:             bi,
		cpus:              cpus,
		nics:              nics,
		largeNode:         largeNodeThresh,
		expectConnections: expectConnections,
		userPolicy:        userPolicy,
		userPolicyOpts:    userPolicyOpts,
		policies: map[policy.IRQPolicy]func(opts ...option.OptionFunc) policy.Policy{
			policy.Common: func(opts ...option.OptionFunc) policy.Policy {
				return policy.NewCommonPolicy(cpus, nics)
			},
			policy.Static: func(opts ...option.OptionFunc) policy.Policy {
				return policy.NewStaticPolicy(cpus, nics, opts...)
			},
			policy.Steering: func(opts ...option.OptionFunc) policy.Policy {
				var newOpts []option.OptionFunc
				if expectConnections != 0 {
					newOpts = append(newOpts, option.WithExpectConnections(expectConnections))
				}
				newOpts = append(newOpts, opts...)
				return policy.NewSteeringPolicy(cpus, nics, newOpts...)
			},
		},
	}
	return
}

func (p *Profile) Run() (err error) {

	// apply user-override policy
	if p.userPolicy != "" {
		return p.SetupUser(p.userPolicy, p.userPolicyOpts...)
	}

	// apply auto-detect policy
	switch p.board.ChassisVendor {
	case "Red Hat":
		err = p.SetupKVM()
	case "AlibabaCloud", "Alibaba Cloud":
		switch p.board.BIOSVendor {
		case "SeaBIOS":
			err = p.SetupECS()
		default:
			err = p.SetupEBM()
		}
	case "HuaweiCloud", "Huawei Cloud":
		switch p.board.ProductName {
		case "OpenStack Nova":
			err = p.SetupECS()
		default:
			err = p.SetupEBM()
		}
	case "Amazon EC2":
		err = p.SetupECS()
	default:
		klog.InfoS("unknown chassis", "vendor", p.board)
		err = p.SetupBM()
	}
	return
}
