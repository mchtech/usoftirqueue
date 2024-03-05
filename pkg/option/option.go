package option

// global options
var (
	DryRun bool
)

// CPUSelectStrategy
type CPUSelectStrategy string

const (
	CPUSelectStrategyScatter    CPUSelectStrategy = "scatter"    // select cpus: scatter from cores with round offset (default)
	CPUSelectStrategyRoundRobin CPUSelectStrategy = "roundrobin" // select cpus: continue from previous core
)

// CPUSelectMask
type CPUSelectMask string

const (
	CPUSelectMaskAll       CPUSelectMask = "all"   // select all cores
	CPUSelectMaskLocalNUMA CPUSelectMask = "local" // select local numa cores
)

// TrafficSource
type TrafficSource string

const (
	TrafficSourceFromOneOfNIC  TrafficSource = "one_of_all"    // rx from one of all NICs
	TrafficSourceAvgFromAllNIC TrafficSource = "avg_from_all"  // rx from all NICs avg (e.g. bond)
	TrafficSourceCloudK8sENI   TrafficSource = "cloud_k8s_eni" // rx from main NIC and/or other NICs (k8s ecs/ebm: hostNetwork: eth0, CNI Network: eth1..ethx)
)

type Option struct {
	CPUSelectStrategy CPUSelectStrategy
	CPUSelectMask     CPUSelectMask
	TrafficSource     TrafficSource // for RFS
	ExpectConnections int           // for RFS
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

func WithTrafficSource(m TrafficSource) OptionFunc {
	return func(opt *Option) {
		opt.TrafficSource = m
	}
}

func WithExpectConnections(n int) OptionFunc {
	return func(opt *Option) {
		opt.ExpectConnections = n
	}
}
