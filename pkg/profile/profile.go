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
	board *board.SystemInfo
	cpus  *cpu.Topology
	nics  []*nic.NIC
	// spec
	policies map[policy.IRQPolicy]func(opts ...option.OptionFunc) policy.Policy
	// overrides
	irqPolicy   *policy.IRQPolicy
	cpuStrategy *option.CPUSelectStrategy
	cpuMask     *option.CPUSelectMask
	flowMode    *option.FlowMode
}

func New(nicRegex string, irqPolicy string, cpuPolicy string, cpuMask string, flowMode string, largeNodeThresh int, expectConnections int) (p *Profile, err error) {

	var (
		si   *board.SystemInfo
		cpus *cpu.Topology
		nics []*nic.NIC
	)

	// board info
	if si, err = board.GetSystemInfo(); err != nil {
		klog.ErrorS(err, "failed to get system info")
		return
	}
	klog.InfoS("system info", "chassis", si.ChassisVendor, "product", si.ProductName, "bios", si.BIOSVendor)

	// get CPU count
	if cpus, err = cpu.List(); err != nil {
		klog.ErrorS(err, "failed to get CPU info")
		return nil, err
	}
	klog.InfoS("current CPU", "numa count", cpus.NumNUMA(), "total cpus", cpus.NumCPU(), "local cpus", cpus.NumLocalCPU(0))

	// get NICs
	if nics, err = nic.List(nicRegex); err != nil {
		klog.ErrorS(err, "failed to get NICs", "regexp", nicRegex)
		return nil, err
	}
	klog.InfoS("current NICs", "nics", nic.NICList(nics).Names(), "filter", nicRegex)

	p = &Profile{
		largeNode: largeNodeThresh,
		board:     si,
		cpus:      cpus,
		nics:      nics,
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

	// user override param
	if irqPolicy != "auto" {
		u := policy.IRQPolicy(irqPolicy)
		p.irqPolicy = &u
	}
	if cpuPolicy != "auto" {
		u := option.StringToCPUSelectStrategy[cpuPolicy]
		p.cpuStrategy = &u
	}
	if cpuMask != "auto" {
		u := option.StringToCPUSelectMask[cpuMask]
		p.cpuMask = &u
	}
	if flowMode != "auto" {
		u := option.StringToFlowMode[flowMode]
		p.flowMode = &u
	}
	return
}

func (p *Profile) Run() (err error) {

	// apply user-override policy
	if p.irqPolicy != nil {
		var opts []option.OptionFunc
		if p.cpuStrategy != nil {
			opts = append(opts, option.WithCPUSelectStrategy(*p.cpuStrategy))
		}
		if p.cpuMask != nil {
			opts = append(opts, option.WithCPUSelectMask(*p.cpuMask))
		}
		if p.flowMode != nil {
			opts = append(opts, option.WithFlowMode(*p.flowMode))
		}
		err = p.SetupUser(*p.irqPolicy, opts...)
		return
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
