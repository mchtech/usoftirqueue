package option

// global options
var (
	DryRun bool
)

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
	"all":   CPUSelectMaskAll,
	"local": CPUSelectMaskLocalNUMA,
}

// TrafficSource
type TrafficSource int

const (
	TrafficSourceFromOneOfNIC  TrafficSource = iota // rx from one of all NICs
	TrafficSourceAvgFromAllNIC                      // rx from all NICs avg (e.g. bond)
	TrafficSourceCloudK8sENI                        // rx from main NIC and/or other NICs (k8s ecs/ebm: hostNetwork: eth0, CNI Network: eth1..ethx)
)

var StringToTrafficSource = map[string]TrafficSource{
	"one_of_all":    TrafficSourceFromOneOfNIC,
	"avg_from_all":  TrafficSourceAvgFromAllNIC,
	"cloud_k8s_eni": TrafficSourceCloudK8sENI,
}

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
