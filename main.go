package main

import (
	"flag"
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
	logFile           = "/var/log/usoftirqueue.log"
	irqPolicy         = "auto"
	cpuStrategy       = "auto"
	cpuMask           = "auto"
	flowMode          = "auto"
	nicRegex          = `^eth[\d]+`
	queueRegexes      util.ArrayFlags
	largeNodeThresh   = 16
	expectConnections = 0
)

func main() {

	var (
		err error
		p   *profile.Profile
	)

	defer klog.Flush()
	klog.InitFlags(nil)
	flag.Set("logtostderr", "false")
	flag.Set("log_file", logFile)
	flag.Set("alsologtostderr", "true")

	flag.BoolVar(&option.DryRun, "option.dry-run", false, "dry run")
	flag.StringVar(&irqPolicy, "custom.policy", irqPolicy, `
┌──────────┬───────────────────────────────────────────────────────────────────────┐
│ static   │ set NIC hardirq affinity                                              │
├──────────┼───────────────────────────────────────────────────────────────────────┤
│ steering │ set NIC hardirq affinity and use RPS/RFS/aRFS/XPS to steering softirq │
└──────────┴───────────────────────────────────────────────────────────────────────┘
`)
	flag.StringVar(&cpuStrategy, "custom.policy.cpu.strategy", cpuStrategy, `
┌────────────┬───────────────────────────────────────────┐
│ scatter    │ scatter cores if remained NumIRQ < NumCPU │
│            │ 0 1 2 3 4 5 6 7 0 2 4 6                   │
│            │ 0 1 2 3 4 5 6 7 1 3 5 7                   │
│            │ 0 1 2 3 4 5 6 7 2 6                       │
│            │ ......                                    │
├────────────┼───────────────────────────────────────────┤
│ roundrobin │ select cores roundrobin                   │
│            │ 0 1 2 3                                   │
│            │ 4 5 6 7                                   │
│            │ 0 1 2 3 4 5                               │
│            │ 6 7 0 1 2 3                               │
│            │ ......                                    │
└────────────┴───────────────────────────────────────────┘
`)
	flag.StringVar(&cpuMask, "custom.policy.cpu.mask", cpuMask, `
┌────────────┬──────────────────────────────────────────────┐
│ all        │ use all cores                                │
├────────────┼──────────────────────────────────────────────┤
│ local_numa │ use cores of same NUMA node as NIC installed │
└────────────┴──────────────────────────────────────────────┘
`)
	flag.StringVar(&flowMode, "custom.policy.nic.flowmode", flowMode, `
┌───────────────┬──────────────────────────────────────────┐
│ one_of_all    │ RX/TX from one of all NICs               │
├───────────────┼──────────────────────────────────────────┤
│ avg_from_all  │ RX/TX from all NICs averagely            │
├───────────────┼──────────────────────────────────────────┤
│ cloud_k8s_eni │ RX/TX from eth0 for K8s HostNetwork      │
│               │ RX/TX from eth1..ethn for K8s CNINetwork │
└───────────────┴──────────────────────────────────────────┘
`)
	flag.StringVar(&nicRegex, "option.nic-filter", nicRegex, "filter nic name by regexp")
	flag.Var(&queueRegexes, "option.irq-filter", "filter nic rx/tx queue irq name by regexp (can be specified multiple times)\nformat: <driver>:<regexp>\n"+formatter.NewTableFromMap(nic.GetNICQueueIRQFilter(), []string{"driver", "regexp"}).String())
	flag.IntVar(&largeNodeThresh, "option.large-node-threshold", largeNodeThresh, "machine is considered as a large node (prefer steering policy), if cpu core count >")
	flag.IntVar(&expectConnections, "option.expect-connections", expectConnections, "required for RFS: estimated max concurrent connections that the node can process\n"+formatter.NewTableFromMap(policy.CoreToConnections, []string{"cpu", "connections"}).String())
	flag.IntVar(&policy.BaseConnectionsPerCore, "option.connections-per-core", policy.BaseConnectionsPerCore, "required for RFS: connections per cpu core (only valid when num-cpu does not exist in `num-cpu:connections` map)")
	flag.Parse()

	if irqPolicy == "auto" && (cpuStrategy != "auto" || cpuMask != "auto" || flowMode != "auto") {
		klog.Error("--custom.policy must be set together with --custom.policy.cpu.strategy --custom.policy.cpu.mask --custom.policy.nic.flowmode")
		os.Exit(1)
		return
	}

	klog.InfoS("starting", "at", time.Now().Format(time.RFC3339Nano))

	if len(queueRegexes) > 0 {
		nic.RegisterNICQueueIRQFilter(queueRegexes)
	}

	if p, err = profile.New(nicRegex, irqPolicy, cpuStrategy, cpuMask, flowMode, largeNodeThresh, expectConnections); err != nil {
		klog.ErrorS(err, "failed to setup profile")
		os.Exit(2)
		return
	}
	if err = p.Run(); err != nil {
		klog.ErrorS(err, "failed to optimize NIC")
		os.Exit(3)
		return
	}
}
