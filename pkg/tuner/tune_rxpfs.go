package tuner

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"usoftirqueue/pkg/hardware/cpu"
	"usoftirqueue/pkg/hardware/nic"
	"usoftirqueue/pkg/option"
	"usoftirqueue/pkg/util"
	"usoftirqueue/pkg/util/config"

	"k8s.io/klog/v2"
)

var (
	oncexps sync.Once
	rxqs    bool
)

func (t *Tuner) kernelSupportXPSRXQS() bool {
	oncexps.Do(func() {
		_, e := os.Stat("/sys/class/net/lo/queues/tx-0/xps_rxqs")
		rxqs = e == nil
	})
	return rxqs
}

func (t *Tuner) DisableRPSRFSXPS(eths []*nic.NIC) (err error) {
	klog.InfoS("disable RPS RFS aRFS XPS")
	for _, eth := range eths {
		if ch, _ := eth.GetQueueChannel(); ch != nil {
			for i := 0; i < int(ch.CombinedCount); i++ {
				err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/rx-%d/rps_cpus", eth.Name, i), "0"))
				err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/rx-%d/rps_flow_cnt", eth.Name, i), "0"))
				// err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/tx-%d/xps_cpus", eth.Name, i), "0"))
				if t.kernelSupportXPSRXQS() {
					err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/tx-%d/xps_rxqs", eth.Name, i), util.BitMask(int(ch.CombinedCount))))
				}
			}
			for i := 0; i < int(ch.RxCount); i++ {
				err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/rx-%d/rps_cpus", eth.Name, i), "0"))
				err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/rx-%d/rps_flow_cnt", eth.Name, i), "0"))
			}
			for i := 0; i < int(ch.TxCount); i++ {
				// err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/tx-%d/xps_cpus", eth.Name, i), "0"))
				if t.kernelSupportXPSRXQS() {
					err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/tx-%d/xps_rxqs", eth.Name, i), util.BitMask(int(ch.TxCount))))
				}
			}
			// aRFS: if threads migrated frequently, it may take lots of cpu time to sync flow rule to NIC
			// eth.SetFeature("ntuple", false)
		}
	}
	err = errors.Join(err, config.Write("/proc/sys/net/core/rps_sock_flow_entries", "0"))
	klog.InfoS("finish disabling RPS RFS aRFS XPS")
	return
}

func (t *Tuner) EnableRPSRFSXPS(eths []*nic.NIC, cpus *cpu.Topology, opt *option.Option) (err error) {

	// expecting connections
	conncount := opt.ExpectConnections
	conncountN2 := util.RoundUpToPowerOfTwo(conncount)
	if conncountN2 != conncount {
		klog.InfoS("round up connections", "from", conncount, "to", conncountN2)
	}

	// group nics according to flow mode
	var nicGroups nic.NICGroup
	switch opt.TrafficSource {
	case option.TrafficSourceAvgFromAllNIC:
		nicGroups = nic.NICGroup{eths}
	case option.TrafficSourceFromOneOfNIC:
		for _, eth := range eths {
			nicGroups = append(nicGroups, []*nic.NIC{eth})
		}
	case option.TrafficSourceCloudK8sENI:
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
		return fmt.Errorf("unknown traffic source %v", opt.TrafficSource)
	}
	klog.InfoS("get NIC groups", "groups", nicGroups.Names())

	// apply rps_flow_cnt for each nic group
	klog.InfoS("enable RPS RFS aRFS XPS")

	// cpu bitmap for RPS
	cpubm := util.BitMask(cpus.NumCPU(), cpus.CoreIDs(-1, false)...)
	klog.InfoS("get cpu bitmask", "mask", cpubm)

	// enable rfs
	cc := strconv.Itoa(conncountN2)
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
		conncountPerQueueN2 := util.RoundUpToPowerOfTwo(conncountPerQueue)
		if conncountPerQueueN2 != conncountPerQueue {
			klog.InfoS("round up per-queue connections", "from", conncountPerQueue, "to", conncountPerQueueN2)
		}
		ccpq := strconv.Itoa(conncountPerQueueN2)
		klog.InfoS("get connections per queue", "nicgroup", nicGroup.Names(), "conncountPerQueue", conncountPerQueueN2)

		for _, eth := range nicGroup {
			if ch, _ := eth.GetQueueChannel(); ch != nil {
				for i := 0; i < int(ch.CombinedCount); i++ {
					err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/rx-%d/rps_cpus", eth.Name, i), cpubm))
					err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/rx-%d/rps_flow_cnt", eth.Name, i), ccpq))
					// err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/tx-%d/xps_cpus", eth.Name, i), "cpubm"))
					if t.kernelSupportXPSRXQS() {
						err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/tx-%d/xps_rxqs", eth.Name, i), util.BitMask(int(ch.CombinedCount), i)))
					}
				}
				for i := 0; i < int(ch.RxCount); i++ {
					err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/rx-%d/rps_cpus", eth.Name, i), cpubm))
					err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/rx-%d/rps_flow_cnt", eth.Name, i), ccpq))
				}
				if ch.TxCount == ch.RxCount {
					for i := 0; i < int(ch.TxCount); i++ {
						// err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/tx-%d/xps_cpus", eth.Name, i), "cpubm"))
						if t.kernelSupportXPSRXQS() {
							err = errors.Join(err, config.Write(fmt.Sprintf("/sys/class/net/%s/queues/tx-%d/xps_rxqs", eth.Name, i), util.BitMask(int(ch.TxCount), i)))
						}
					}
				}
				// aRFS: if threads migrated frequently, it may take lots of cpu time to sync flow rule to NIC
				// eth.SetFeature("ntuple", true)
			}
		}
	}

	klog.InfoS("finish enabling RPS RFS aRFS XPS")
	return
}
