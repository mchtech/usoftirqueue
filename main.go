package main

import (
	"flag"
	"fmt"
	"time"
	"usoftirqueue/pkg/option"
	"usoftirqueue/pkg/policy"
	"usoftirqueue/pkg/profile"

	"k8s.io/klog/v2"
)

var (
	logFile           = "/var/log/usoftirqueue.log"
	irqPolicy         = "auto"
	cpuStrategy       = "auto"
	cpuMask           = "auto"
	flowMode          = "auto"
	nicRegex          = `^eth[\d]+`
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
	flag.StringVar(&irqPolicy, "option.policy", irqPolicy, `static | steering
┌──────────┬───────────────────────────────────────────────────────────────────────┐
│ static   │ set NIC hardirq affinity                                              │
├──────────┼───────────────────────────────────────────────────────────────────────┤
│ steering │ set NIC hardirq affinity and use RPS/RFS/aRFS/XPS to steering softirq │
└──────────┴───────────────────────────────────────────────────────────────────────┘
`)
	flag.StringVar(&cpuStrategy, "option.policy.cpu.strategy", cpuStrategy, `scatter | roundrobin
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
	flag.StringVar(&cpuMask, "option.policy.cpu.mask", cpuMask, `all | local_numa
┌────────────┬──────────────────────────────────────────────┐
│ all        │ use all cores                                │
├────────────┼──────────────────────────────────────────────┤
│ local_numa │ use cores of same NUMA node as NIC installed │
└────────────┴──────────────────────────────────────────────┘
`)
	flag.StringVar(&flowMode, "option.policy.nic.flowmode", flowMode, `one_of_all | avg_from_all | cloud_k8s_eni
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
	flag.IntVar(&largeNodeThresh, "option.large-node-threshold", largeNodeThresh, "machine is considered as a large node (prefer steering policy), if cpu core count >")
	flag.IntVar(&expectConnections, "option.expect-connections", expectConnections, fmt.Sprintf("required for RFS: estimated max concurrent connections that the node can process (default num-cpu:connections=%v)", policy.CoreToConnections))
	flag.IntVar(&policy.BaseConnectionsPerCore, "option.connections-per-core", policy.BaseConnectionsPerCore, "required for RFS: connections per cpu core (only valid when num-cpu does not exist in `num-cpu:connections` map)")
	flag.Parse()

	klog.InfoS("starting", "at", time.Now().Format(time.RFC3339Nano))

	if p, err = profile.New(nicRegex, irqPolicy, cpuStrategy, cpuMask, flowMode, largeNodeThresh, expectConnections); err != nil {
		klog.ErrorS(err, "failed to setup profile")
		return
	}
	if err = p.Run(); err != nil {
		klog.ErrorS(err, "failed to optimize NIC")
		return
	}
}
