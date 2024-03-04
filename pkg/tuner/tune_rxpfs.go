package tuner

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"usoftirqueue/pkg/hardware/nic"
	"usoftirqueue/pkg/option"
	"usoftirqueue/pkg/util/config"

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

func (t *Tuner) DisableRPSRFSXPS(eths []*nic.NIC) (err error) {
	klog.InfoS("disable RPS RFS aRFS XPS")
	for _, eth := range eths {
		if ch, _ := eth.GetQueueChannel(); ch != nil {
			for i := 0; i < int(ch.CombinedCount); i++ {
				err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/rx-%d/rps_cpus", eth.Name, i), "0"))
				err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/rx-%d/rps_flow_cnt", eth.Name, i), "0"))
				// err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/tx-%d/xps_cpus", eth.Name, i), "0"))
			}
			for i := 0; i < int(ch.RxCount); i++ {
				err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/rx-%d/rps_cpus", eth.Name, i), "0"))
				err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/rx-%d/rps_flow_cnt", eth.Name, i), "0"))
			}
			// for i := 0; i < int(ch.TxCount); i++ {
			// 	err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/tx-%d/xps_cpus", eth.Name, i), "0"))
			// }
		}
	}
	err = errors.Join(err, config.Write("/proc/sys/net/core/rps_sock_flow_entries", "0"))
	klog.InfoS("finish disabling RPS RFS aRFS XPS")
	return
}

func (t *Tuner) EnableRPSRFSXPS(eths []*nic.NIC, cpucount, conncount int, opt *option.Option) (err error) {

	// cpu bitmap for RPS
	cpubm := getCPUBitmap(cpucount)
	klog.InfoS("get cpu bitmap", "bitmap", cpubm)

	// expecting connections
	if conncountn2 := roundUpToNearestMultipleOf2(conncount); conncountn2 != conncount {
		klog.InfoS("round up connection count", "from", conncount, "to", conncountn2)
		conncount = conncountn2
	}

	// group nics according to flow mode
	var nicGroups nic.NICGroup
	switch opt.FlowMode {
	case option.FlowModeAvgFromAllNIC:
		nicGroups = nic.NICGroup{eths}
	case option.FlowModeFromOneOfNIC:
		for _, eth := range eths {
			nicGroups = append(nicGroups, []*nic.NIC{eth})
		}
	case option.FlowModeCloudK8sENI:
		for i, eth := range eths {
			if i == 0 {
				nicGroups = append(nicGroups, []*nic.NIC{eth})
			} else {
				if len(nicGroups) == 1 {
					nicGroups = append(nicGroups, []*nic.NIC{})
				}
				nicGroups[1] = append(nicGroups[1], eth)
			}
		}
	default:
		return fmt.Errorf("unknown flow mode %d", opt.FlowMode)
	}
	klog.InfoS("get NIC groups", "groups", nicGroups.Names())

	// apply rps_flow_cnt for each nic group
	klog.InfoS("enable RPS RFS aRFS XPS")

	// enable rfs
	cc := strconv.Itoa(conncount)
	err = errors.Join(err, config.Write("/proc/sys/net/core/rps_sock_flow_entries", cc))

	// enable rps flow
	for _, nicGroup := range nicGroups {
		queuecount := 0
		for _, eth := range nicGroup {
			if ch, _ := eth.GetQueueChannel(); ch != nil {
				if ch.CombinedCount > 0 {
					queuecount += int(ch.CombinedCount)
				} else {
					queuecount += int(ch.RxCount)
				}
			}
		}
		klog.InfoS("get total queue count", "nicgroup", nicGroup.Names(), "queues", queuecount)

		conncountPerQueue := conncount / queuecount
		ccpq := strconv.Itoa(conncountPerQueue)
		klog.InfoS("get connection count per queue", "nicgroup", nicGroup.Names(), "conncount", conncount, "conncountPerQueue", conncountPerQueue)

		for _, eth := range nicGroup {
			if ch, _ := eth.GetQueueChannel(); ch != nil {
				for i := 0; i < int(ch.CombinedCount); i++ {
					err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/rx-%d/rps_cpus", eth.Name, i), cpubm))
					err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/rx-%d/rps_flow_cnt", eth.Name, i), ccpq))
					// err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/tx-%d/xps_cpus", eth.Name, i), cpubm))
				}
				for i := 0; i < int(ch.RxCount); i++ {
					err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/rx-%d/rps_cpus", eth.Name, i), cpubm))
					err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/rx-%d/rps_flow_cnt", eth.Name, i), ccpq))
				}
				// for i := 0; i < int(ch.TxCount); i++ {
				// 	err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/tx-%d/xps_cpus", eth.Name, i), cpubm))
				// }
			}
		}
	}

	klog.InfoS("finish enabling RPS RFS aRFS XPS")
	return
}
