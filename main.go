package main

import (
	"flag"
	"fmt"
	"os"
	"time"
	"usoftirqueue/pkg/hardware/nic"
	"usoftirqueue/pkg/option"
	"usoftirqueue/pkg/policy"
	"usoftirqueue/pkg/profile"
	"usoftirqueue/pkg/util"
	"usoftirqueue/pkg/util/formatter"

	"k8s.io/klog/v2"
)

var (
	buildat string
	commit  string
)

var (
	logFile           = "/var/log/usoftirqueue.log"
	irqPolicy         = "auto"
	cpuStrategy       = "auto"
	cpuMask           = "auto"
	trafficSource     = "auto"
	nicRegex          = `^eth[\d]+`
	queueIRQRegexes   util.StringsFlags
	largeNodeThresh   = 16
	expectConnections = 0
)

func main() {

	var (
		err            error
		p              *profile.Profile
		userPolicy     policy.IRQPolicy
		userPolicyOpts []option.OptionFunc
		version        string
		v              bool
	)

	// klog flags
	defer klog.Flush()
	klog.InitFlags(nil)
	flag.Set("logtostderr", "false")
	flag.Set("log_file", logFile)
	flag.Set("alsologtostderr", "true")
	// prog flags
	flag.BoolVar(&v, "version", v, "show version")
	flag.BoolVar(&option.DryRun, "option.dry-run", false, "dry run")
	flag.StringVar(&irqPolicy, "option.policy", irqPolicy, `
┌──────────┬────────────────────────────────────────────────┐
│ auto     │ selected by usoftirqueue                       │
├──────────┼────────────────────────────────────────────────┤
│ static   │ set NIC hardirq affinity                       │
├──────────┼────────────────────────────────────────────────┤
│ steering │ set NIC hardirq affinity                       │
│          │ use RPS/RFS/aRFS/XPS to steering softirq       │
└──────────┴────────────────────────────────────────────────┘
`)
	flag.StringVar(&cpuStrategy, "option.policy.cpu.strategy", cpuStrategy, `
┌────────────┬──────────────────────────────────────────────┐
│ auto       │ selected by usoftirqueue                     │
├────────────┼──────────────────────────────────────────────┤
│ scatter    │ scatter cores if remained NumIRQ < NumCPU    │
│            │ 0 1 2 3 4 5 6 7 0 2 4 6                      │
│            │ 0 1 2 3 4 5 6 7 1 3 5 7                      │
│            │ 0 1 2 3 4 5 6 7 2 6                          │
│            │ ......                                       │
├────────────┼──────────────────────────────────────────────┤
│ roundrobin │ select cores roundrobin                      │
│            │ 0 1 2 3                                      │
│            │ 4 5 6 7                                      │
│            │ 0 1 2 3 4 5                                  │
│            │ 6 7 0 1 2 3                                  │
│            │ ......                                       │
└────────────┴──────────────────────────────────────────────┘
`)
	flag.StringVar(&cpuMask, "option.policy.cpu.mask", cpuMask, `
┌───────┬──────────────────────────────────────────────┐
│ auto  │ selected by usoftirqueue                     │
├───────┼──────────────────────────────────────────────┤
│ all   │ use all cores                                │
├───────┼──────────────────────────────────────────────┤
│ local │ use cores of same NUMA node as NIC installed │
└───────┴──────────────────────────────────────────────┘
`)
	flag.StringVar(&trafficSource, "option.policy.nic.trafficsource", trafficSource, `
┌───────────────┬───────────────────────────────────────────┐
│ auto          │ selected by usoftirqueue                  │
├───────────────┼───────────────────────────────────────────┤
│ one_of_all    │ RX/TX from one of all NICs                │
├───────────────┼───────────────────────────────────────────┤
│ avg_from_all  │ RX/TX from all NICs averagely             │
├───────────────┼───────────────────────────────────────────┤
│ cloud_k8s_eni │ RX/TX from eth0 for K8s HostNetwork       │
│               │ RX/TX from eth1..ethn for K8s CNINetwork  │
└───────────────┴───────────────────────────────────────────┘
`)
	flag.StringVar(&nicRegex, "option.nic-filter", nicRegex, "filter nic name by regexp")
	flag.Var(&queueIRQRegexes, "option.irq-filter", "filter nic rx/tx queue irq name by regexp (can be specified multiple times)\nformat: <driver>:<regexp> (regexp: first group used as IRQ input/ouput prefix)\n"+formatter.NewTableFromMap(nic.GetNICQueueIRQFilter(), []string{"driver", "regexp"}).String())
	flag.IntVar(&largeNodeThresh, "option.large-node-threshold", largeNodeThresh, "machine is considered as a large node (prefer steering policy), if cpu core count >")
	flag.IntVar(&expectConnections, "option.expect-connections", expectConnections, "required for RFS: estimated max concurrent connections that the node can process\n"+formatter.NewTableFromMap(policy.CoreToConnections, []string{"cpu", "connections"}).String())
	flag.IntVar(&policy.BaseConnectionsPerCore, "option.connections-per-core", policy.BaseConnectionsPerCore, "required for RFS: connections per cpu core (only valid when num-cpu does not exist in `num-cpu:connections` map)")
	flag.Parse()

	version = fmt.Sprintf("%s-%s", buildat, commit)
	if v {
		fmt.Println(version)
		os.Exit(0)
	}

	klog.InfoS("starting", "at", time.Now().Format(time.RFC3339Nano), "version", version)

	// user-override options
	if irqPolicy == "auto" && (cpuStrategy != "auto" || cpuMask != "auto" || trafficSource != "auto") {
		klog.Error("--option.policy must be set together with --option.policy.cpu.strategy --option.policy.cpu.mask --option.policy.nic.trafficsource")
		os.Exit(1)
		return
	}
	if irqPolicy != "auto" {
		userPolicy = policy.IRQPolicy(irqPolicy)
	}
	if cpuStrategy != "auto" {
		userPolicyOpts = append(userPolicyOpts, option.WithCPUSelectStrategy(option.StringToCPUSelectStrategy[cpuStrategy]))
	}
	if cpuMask != "auto" {
		userPolicyOpts = append(userPolicyOpts, option.WithCPUSelectMask(option.StringToCPUSelectMask[cpuMask]))
	}
	if trafficSource != "auto" {
		userPolicyOpts = append(userPolicyOpts, option.WithTrafficSource(option.StringToTrafficSource[trafficSource]))
	}

	// reg queue irq filter
	if err = nic.RegisterNICQueueIRQFilter(queueIRQRegexes); err != nil {
		klog.ErrorS(err, "failed to register NIC queue IRQ filter")
		os.Exit(2)
		return
	}

	// run
	if p, err = profile.New(nicRegex, largeNodeThresh, expectConnections, userPolicy, userPolicyOpts...); err != nil {
		klog.ErrorS(err, "failed to setup profile")
		os.Exit(3)
		os.Exit(253)
		return
	}
	if err = p.Run(); err != nil {
		klog.ErrorS(err, "failed to optimize NIC")
		os.Exit(255)
		os.Exit(254)
		return
	}
	klog.InfoS("NIC optimization finished")
}
