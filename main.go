package main

import (
	"flag"
	"fmt"
	"time"
	"usoftirqueue/pkg/hardware/board"
	"usoftirqueue/pkg/option"
	"usoftirqueue/pkg/policy"
	"usoftirqueue/pkg/profile"

	"k8s.io/klog/v2"
)

const (
	Auto = "(auto detect)"
)

var (
	logFile           = "/var/log/usoftirqueue.log"
	nicPolicy         = Auto
	nicRegex          = `^eth[\d]+`
	largeNodeThresh   = 16
	expectConnections = 0
)

func main() {

	var (
		err error
		si  *board.SystemInfo
		p   *profile.Profile
	)

	defer klog.Flush()
	klog.InitFlags(nil)
	flag.Set("logtostderr", "false")
	flag.Set("log_file", logFile)
	flag.Set("alsologtostderr", "true")

	flag.BoolVar(&option.DryRun, "option.dry-run", false, "dry run")
	flag.StringVar(&nicPolicy, "option.policy", Auto, "static | steering")
	flag.StringVar(&nicRegex, "option.nic-filter", nicRegex, "filter nic name by regexp")
	flag.IntVar(&largeNodeThresh, "option.large-node-threshold", largeNodeThresh, "machine is considered as a large node (prefer steering policy), if cpu core count >")
	flag.IntVar(&expectConnections, "option.expect-connections", expectConnections, fmt.Sprintf("estimated max concurrent connections that the node can process (default num-cpu:connections=%v)", policy.CoreToConnections))
	flag.IntVar(&policy.BaseConnectionsPerCore, "option.connections-per-core", policy.BaseConnectionsPerCore, "connections per cpu core (only valid when num-cpu does not exist in `num-cpu:connections` map)")
	flag.Parse()

	klog.InfoS("starting", "at", time.Now().Format(time.RFC3339Nano))

	if si, err = board.GetSystemInfo(); err != nil {
		klog.ErrorS(err, "failed to get system info")
		return
	}
	klog.InfoS("system info", "chassis", si.ChassisVendor, "product", si.ProductName, "bios", si.BIOSVendor)

	// setup
	if p, err = profile.New(nicRegex, largeNodeThresh, expectConnections); err != nil {
		klog.ErrorS(err, "failed to setup profile")
		return
	}

	// apply user-override policy
	if nicPolicy != Auto {
		err = p.SetupUser(nicPolicy)
		goto end
	}

	// apply auto-detect policy
	switch si.ChassisVendor {
	case "Red Hat":
		err = p.SetupKVM()
	case "AlibabaCloud", "Alibaba Cloud":
		switch si.BIOSVendor {
		case "SeaBIOS":
			err = p.SetupECS()
		default:
			err = p.SetupEBM()
		}
	case "HuaweiCloud", "Huawei Cloud":
		switch si.ProductName {
		case "OpenStack Nova":
			err = p.SetupECS()
		default:
			err = p.SetupEBM()
		}
	default:
		klog.InfoS("unknown chassis", "vendor", si)
		err = p.SetupBM()
	}

end:
	if err != nil {
		klog.ErrorS(err, "failed to optimize NIC")
		return
	}
}
