package nic

import (
	"fmt"
	"sync"
	"usoftirqueue/pkg/option"
	"usoftirqueue/pkg/util"

	"github.com/jaypipes/ghw/pkg/cpu"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
)

var (
	once sync.Once
	srr  *util.ScatterRoundRobinIterator[int]
	rr   *util.RoundRobinIterator[int]
)

func (n *NIC) TuneIRQ(cpus *cpu.Info, dryrun bool, opt *option.Option) (err error) {

	var desiredAffinity []int

	// TODO: filter enabled
	irqs := n.IRQs
	numIRQ := len(irqs)
	numCore := int(cpus.TotalThreads)

	var numa int
	switch opt.CPUSelectMask {
	case option.CPUSelectMaskLocalNUMA:
		numa = irqs[0].NUMA
	case option.CPUSelectMaskAll:
		numa = -1
	default:
		return fmt.Errorf("unknown CPUSelectMask: %d", opt.CPUSelectMask)
	}
	if numa >= 0 {
		numCore = int(cpus.Processors[0].NumThreads)
	}

	cores := util.GetCoreList(cpus, numa, false)
	once.Do(func() {
		rr = util.NewRoundRobinIterator(cores)
		srr = util.NewScatterRoundRobinIterator(cores)
	})

	switch opt.CPUSelectPolicy {
	case option.CPUSelectPolicyRoundRobin:
		desiredAffinity = rr.NextN(numIRQ)
	case option.CPUSelectPolicyScatter:
		for i := 0; i < numIRQ/numCore; i++ {
			desiredAffinity = append(desiredAffinity, cores...)
		}
		// try to scatter IRQs across cores
		if remain := numIRQ % numCore; remain != 0 {
			if numCore%remain == 0 {
				desiredAffinity = append(desiredAffinity, srr.NewInterval(numCore/remain).NextN(remain)...)
			} else {
				desiredAffinity = append(desiredAffinity, srr.NewInterval(1).NextN(remain)...)
			}
		}
	default:
		return errors.Errorf("unsupported CPU select policy %q", opt.CPUSelectPolicy)
	}

	for i := 0; i < numIRQ; i++ {
		klog.InfoS("tune NIC IRQ affinity",
			"nic", n.Name,
			"irq", irqs[i].IRQ,
			"name", irqs[i].Name,
			"driver", n.Driver,
			// "bus", n.Device,
			"affinity", fmt.Sprintf("%v -> %v", irqs[i].Affinity, []int{desiredAffinity[i]}),
		)
	}

	// tune irq
	if !dryrun {
		if err = n.setIRQAffinity(desiredAffinity); err != nil {
			return errors.Wrap(err, "failed to set NIC IRQ affinity")
		}
		klog.InfoS("finish tuning NIC IRQ affinity", "nic", n.Name, "irq", irqs.GetActivatedIRQs(), "affinity", desiredAffinity)
	} else {
		klog.InfoS("dryrun mode: skip tuning NIC IRQ affinity", "nic", n.Name)
	}

	return
}
