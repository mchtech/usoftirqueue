package nic

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"usoftirqueue/pkg/config"

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

func EnableRPSRFSXPS(nics []string, cpucount, conncount int, dryrun bool) (err error) {

	cpubm := getCPUBitmap(cpucount)
	klog.InfoS("get cpu bitmap", "bitmap", cpubm)

	queuecount := 0
	for _, eth := range nics {
		if ch, _ := GetNICQueueChannel(eth); ch != nil {
			if ch.CombinedCount > 0 {
				queuecount += int(ch.CombinedCount)
			} else {
				queuecount += int(ch.RxCount)
			}
		}
	}
	klog.InfoS("get total queue count", "nics", nics, "queues", queuecount)

	if conncountn2 := roundUpToNearestMultipleOf2(conncount); conncountn2 != conncount {
		klog.InfoS("round up connection count", "from", conncount, "to", conncountn2)
		conncount = conncountn2
	}
	conncountPerQueue := conncount / queuecount
	klog.InfoS("get connection count per queue", "nics", nics, "conncount", conncount, "conncountPerQueue", conncountPerQueue)

	cc := strconv.Itoa(conncount)
	ccpq := strconv.Itoa(conncountPerQueue)

	klog.InfoS("enable RPS RFS aRFS XPS")
	err = errors.Join(err, config.Write("/proc/sys/net/core/rps_sock_flow_entries", cc))
	for _, eth := range nics {
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
	klog.InfoS("finish enabling RPS RFS aRFS XPS")

	return
}
