package nic

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"usoftirqueue/pkg/config"
	"usoftirqueue/pkg/option"

	"k8s.io/klog/v2"
)

func getCPUBitmap(cpucount int) (mask string) {
	q := cpucount / 32
	for i := 0; i < q; i++ {
		mask = "ffffffff," + mask
	}
	if y := cpucount % 32; y > 0 {
		v := 0
		for i := 0; i < y; i++ {
			v |= (1 << i)
		}
		mask = fmt.Sprintf("%x", v) + "," + mask
	}
	return strings.TrimRight(mask, ",")
}

func roundUpToNearestMultipleOf2(n int) int {
	return (n + 1) & -2
}

func DisableRPSRFSXPS(nics []string, dryrun bool) (err error) {
	klog.InfoS("disable RPS RFS aRFS XPS")
	for _, eth := range nics {
		if ch, _ := GetNICQueueChannel(eth); ch != nil {
			for i := 0; i < int(ch.CombinedCount); i++ {
				err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/rx-%d/rps_cpus", eth, i), "0"))
				err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/rx-%d/rps_flow_cnt", eth, i), "0"))
				// err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/tx-%d/xps_cpus", eth, i), "0"))
			}
			for i := 0; i < int(ch.RxCount); i++ {
				err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/rx-%d/rps_cpus", eth, i), "0"))
				err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/rx-%d/rps_flow_cnt", eth, i), "0"))
			}
			// for i := 0; i < int(ch.TxCount); i++ {
			// 	err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/tx-%d/xps_cpus", eth, i), "0"))
			// }
		}
	}
	err = errors.Join(err, config.Write("/proc/sys/net/core/rps_sock_flow_entries", "0"))
	klog.InfoS("finish disabling RPS RFS aRFS XPS")
	return
}

func EnableRPSRFSXPS(nics []string, cpucount, conncount int, dryrun bool, opt *option.Option) (err error) {

	// cpu bitmap for RPS
	cpubm := getCPUBitmap(cpucount)
	klog.InfoS("get cpu bitmap", "bitmap", cpubm)

	// expecting connections
	if conncountn2 := roundUpToNearestMultipleOf2(conncount); conncountn2 != conncount {
		klog.InfoS("round up connection count", "from", conncount, "to", conncountn2)
		conncount = conncountn2
	}

	// group nics according to flow mode
	var nicGroups [][]string
	switch opt.FlowMode {
	case option.FlowModeAvgFromAllNIC:
		nicGroups = [][]string{nics}
	case option.FlowModeFromOneOfNIC:
		for _, nic := range nics {
			nicGroups = append(nicGroups, []string{nic})
		}
	case option.FlowModeSplitMainNIC:
		for i, nic := range nics {
			if i == 0 {
				nicGroups = append(nicGroups, []string{nic})
			} else {
				if len(nicGroups) == 1 {
					nicGroups = append(nicGroups, []string{})
				}
				nicGroups[1] = append(nicGroups[1], nic)
			}
		}
	default:
		return fmt.Errorf("unknown flow mode %d", opt.FlowMode)
	}
	klog.InfoS("get NIC groups", "groups", nicGroups)

	// apply rps_flow_cnt for each nic group
	klog.InfoS("enable RPS RFS aRFS XPS")

	// enable rfs
	cc := strconv.Itoa(conncount)
	err = errors.Join(err, config.Write("/proc/sys/net/core/rps_sock_flow_entries", cc))

	// enable rps flow
	for _, nicGroup := range nicGroups {
		queuecount := 0
		for _, eth := range nicGroup {
			if ch, _ := GetNICQueueChannel(eth); ch != nil {
				if ch.CombinedCount > 0 {
					queuecount += int(ch.CombinedCount)
				} else {
					queuecount += int(ch.RxCount)
				}
			}
		}
		klog.InfoS("get total queue count", "nicgroup", nicGroup, "queues", queuecount)

		conncountPerQueue := conncount / queuecount
		ccpq := strconv.Itoa(conncountPerQueue)
		klog.InfoS("get connection count per queue", "nicgroup", nicGroup, "conncount", conncount, "conncountPerQueue", conncountPerQueue)

		for _, eth := range nicGroup {
			if ch, _ := GetNICQueueChannel(eth); ch != nil {
				for i := 0; i < int(ch.CombinedCount); i++ {
					err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/rx-%d/rps_cpus", eth, i), cpubm))
					err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/rx-%d/rps_flow_cnt", eth, i), ccpq))
					// err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/tx-%d/xps_cpus", eth, i), cpubm))
				}
				for i := 0; i < int(ch.RxCount); i++ {
					err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/rx-%d/rps_cpus", eth, i), cpubm))
					err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/rx-%d/rps_flow_cnt", eth, i), ccpq))
				}
				// for i := 0; i < int(ch.TxCount); i++ {
				// 	err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/tx-%d/xps_cpus", eth, i), cpubm))
				// }
			}
		}
	}

	klog.InfoS("finish enabling RPS RFS aRFS XPS")
	return
}
