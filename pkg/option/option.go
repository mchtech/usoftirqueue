package option

type CPUSelectPolicy int

const (
	CPUSelectPolicyFromFirstCore    CPUSelectPolicy = iota // select cpus: from first core everytime (default)
	CPUSelectPolicyFromPreviousCore                        // select cpus: continue from previous core
)

type CPUSelectMask int

const (
	CPUSelectMaskAll       CPUSelectMask = iota // select all cores
	CPUSelectMaskLocalNUMA                      // select local numa cores
)

type FlowMode int

const (
	FlowModeFromOneOfNIC  FlowMode = iota // rx from one of all NICs
	FlowModeSplitMainNIC                  // rx from main NIC and other NICs (k8s ecs/ebm: hostNetwork: eth0, CNI Network: eth1..ethx)
	FlowModeAvgFromAllNIC                 // rx from all NICs avg
)

type Option struct {
	CPUSelectPolicy   CPUSelectPolicy
	CPUSelectMask     CPUSelectMask
	FlowMode          FlowMode
	ExpectConnections int
}

type OptionFunc func(*Option)

func ParseOptions(opts ...OptionFunc) *Option {
	opt := &Option{}
	for _, o := range opts {
		o(opt)
	}
	return opt
}

func WithCPUSelectPolicy(p CPUSelectPolicy) OptionFunc {
	return func(opt *Option) {
		opt.CPUSelectPolicy = p
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
