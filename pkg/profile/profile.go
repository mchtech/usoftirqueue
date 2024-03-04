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
	largeNode int
	// vars
	board *board.BoardInfo
	cpus  *cpu.Topology
	nics  []*nic.NIC
	// spec
	policies map[policy.IRQPolicy]func(opts ...option.OptionFunc) policy.Policy
	// overrides
	irqPolicy policy.IRQPolicy
	userOpts  []option.OptionFunc
}

func New(nicRegex, irqPolicy, cpuStrategy, cpuMask, flowMode string, largeNodeThresh, expectConnections int) (p *Profile, err error) {

	var (
		bi   *board.BoardInfo
		cpus *cpu.Topology
		nics []*nic.NIC
	)

	// board info
	if bi, err = board.Collect(); err != nil {
		klog.ErrorS(err, "failed to get board info")
		return
	}
	klog.InfoS("board info", "chassis", bi.ChassisVendor, "product", bi.ProductName, "bios", bi.BIOSVendor)

	// get CPU info
	if cpus, err = cpu.List(); err != nil {
		klog.ErrorS(err, "failed to get CPU info")
		return nil, err
	}
	klog.InfoS("current CPU", "numa", cpus.NumNUMA(), "total", cpus.NumCPU(), "local", cpus.NumLocalCPU(0))

	// get NICs
	if nics, err = nic.List(nicRegex); err != nil {
		klog.ErrorS(err, "failed to get NICs", "regexp", nicRegex)
		return nil, err
	}
	klog.InfoS("target NICs", "nics", nic.NICList(nics).Names(), "filter", nicRegex)

	// user-override param
	var (
		ip   policy.IRQPolicy
		opts []option.OptionFunc
	)
	if irqPolicy != "auto" {
		ip = policy.IRQPolicy(irqPolicy)
	}
	if cpuStrategy != "auto" {
		opts = append(opts, option.WithCPUSelectStrategy(option.StringToCPUSelectStrategy[cpuStrategy]))
	}
	if cpuMask != "auto" {
		opts = append(opts, option.WithCPUSelectMask(option.StringToCPUSelectMask[cpuMask]))
	}
	if flowMode != "auto" {
		opts = append(opts, option.WithFlowMode(option.StringToFlowMode[flowMode]))
	}

	// init profile
	p = &Profile{
		board:     bi,
		cpus:      cpus,
		nics:      nics,
		largeNode: largeNodeThresh,
		irqPolicy: ip,
		userOpts:  opts,
		policies: map[policy.IRQPolicy]func(opts ...option.OptionFunc) policy.Policy{
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
	if p.irqPolicy != "" {
		return p.SetupUser(p.irqPolicy, p.userOpts...)
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
	default:
		klog.InfoS("unknown chassis", "vendor", p.board)
		err = p.SetupBM()
	}
	return
}
