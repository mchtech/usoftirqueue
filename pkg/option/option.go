package option

type CPUSelectPolicy int

const (
	CPUSelectPolicyFromFirstCore CPUSelectPolicy = iota // select cpus: from first core everytime (default)
	CPUSelectPolicyFromIterator                         // select cpus: continue from previous core
)

type CPUSelectMask int

const (
	CPUSelectMaskAuto      CPUSelectMask = 0x0 // select cores according to num irq (default)
	CPUSelectMaskAll       CPUSelectMask = 0x1 // select all cores
	CPUSelectMaskLocalNUMA CPUSelectMask = 0x2 // select local numa cores
)

type Option struct {
	CPUSelectPolicy   CPUSelectPolicy
	CPUSelectMask     CPUSelectMask
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

func WithExpectConnections(n int) OptionFunc {
	return func(opt *Option) {
		opt.ExpectConnections = n
	}
}
