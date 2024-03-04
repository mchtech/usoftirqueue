package option

var DryRun bool

// CPUSelectStrategy
type CPUSelectStrategy int

const (
	CPUSelectStrategyScatter    CPUSelectStrategy = iota // select cpus: scatter from cores with round offset (default)
	CPUSelectStrategyRoundRobin                          // select cpus: continue from previous core
)

var StringToCPUSelectStrategy = map[string]CPUSelectStrategy{
	"scatter":    CPUSelectStrategyScatter,
	"roundrobin": CPUSelectStrategyRoundRobin,
}

// CPUSelectMask
type CPUSelectMask int

const (
	CPUSelectMaskAll       CPUSelectMask = iota // select all cores
	CPUSelectMaskLocalNUMA                      // select local numa cores
)

var StringToCPUSelectMask = map[string]CPUSelectMask{
	"all":        CPUSelectMaskAll,
	"local_numa": CPUSelectMaskLocalNUMA,
}

// FlowMode
type FlowMode int

const (
	FlowModeFromOneOfNIC  FlowMode = iota // rx from one of all NICs
	FlowModeAvgFromAllNIC                 // rx from all NICs avg (e.g. bond)
	FlowModeCloudK8sENI                   // rx from main NIC and/or other NICs (k8s ecs/ebm: hostNetwork: eth0, CNI Network: eth1..ethx)
)

var StringToFlowMode = map[string]FlowMode{
	"one_of_all":    FlowModeFromOneOfNIC,
	"avg_from_all":  FlowModeAvgFromAllNIC,
	"cloud_k8s_eni": FlowModeCloudK8sENI,
}

type Option struct {
	CPUSelectStrategy CPUSelectStrategy
	CPUSelectMask     CPUSelectMask
	FlowMode          FlowMode // for RFS
	ExpectConnections int      // for RFS
}

type OptionFunc func(*Option)

func ParseOptions(opts ...OptionFunc) *Option {
	opt := &Option{}
	for _, o := range opts {
		o(opt)
	}
	return opt
}

func WithCPUSelectStrategy(p CPUSelectStrategy) OptionFunc {
	return func(opt *Option) {
		opt.CPUSelectStrategy = p
	}
}

func WithCPUSelectMask(m CPUSelectMask) OptionFunc {
	return func(opt *Option) {
		opt.CPUSelectMask = m
	}
}

func WithFlowMode(m FlowMode) OptionFunc {
	return func(opt *Option) {
		opt.FlowMode = m
	}
}

func WithExpectConnections(n int) OptionFunc {
	return func(opt *Option) {
		opt.ExpectConnections = n
	}
}
