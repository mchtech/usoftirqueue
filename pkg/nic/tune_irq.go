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
	once      sync.Once
	biterator *util.AccumulateOffsetIterator[int]
	citerator *util.CircularIterator[int]
)

func TuneNICIRQ(eth string, cpus *cpu.Info, dryrun bool, opt *option.Option) (err error) {

	var desiredAffinity []int

	irqs := getNICIRQ(eth)
	numIRQ := len(irqs)

	var numa int
	switch opt.CPUSelectMask {
	case option.CPUSelectMaskLocalNUMA:
		numa = irqs[0].NUMA
	case option.CPUSelectMaskAll:
		numa = -1
	default:
		return fmt.Errorf("unknown CPUSelectMask: %d", opt.CPUSelectMask)
	}

	cores := util.GetCoreList(cpus, numa, false)
	once.Do(func() {
		citerator = util.NewCircularIterator(cores)
		biterator = util.NewAccumulateOffsetIterator(cores)
	})

	if opt.CPUSelectPolicy == option.CPUSelectPolicyFromPreviousCore {
		desiredAffinity = citerator.NextN(numIRQ)
	} else {
		// update num core
		numCore := int(cpus.TotalThreads)
		if numa >= 0 {
			numCore = int(cpus.Processors[0].NumThreads)
		}
		// try balance IRQs across cores
		if numIRQ == numCore {
			desiredAffinity = cores
		} else if numIRQ < numCore {
			if numCore%numIRQ == 0 {
				desiredAffinity = append(desiredAffinity, biterator.NewInterval(numCore/numIRQ).NextN(numIRQ)...)
			} else {
				desiredAffinity = append(desiredAffinity, biterator.NewInterval(1).NextN(numIRQ)...)
			}
		} else { // numIRQ > numCore
			for i := 0; i < numIRQ/numCore; i++ {
				desiredAffinity = append(desiredAffinity, cores...)
			}
			if remain := numIRQ % numCore; numCore%remain == 0 {
				desiredAffinity = append(desiredAffinity, biterator.NewInterval(numCore/remain).NextN(remain)...)
			} else {
				desiredAffinity = append(desiredAffinity, biterator.NewInterval(1).NextN(remain)...)
			}
		}
	}

	for i := 0; i < numIRQ; i++ {
		klog.InfoS("tune NIC IRQ affinity",
			"nic", eth,
			"irq", irqs[i].IRQ,
			"name", irqs[i].Name,
			"driver", irqs[i].Driver,
			// "bus", irqs[i].Device,
			"affinity", fmt.Sprintf("%v -> %v", irqs[i].Affinity, []int{desiredAffinity[i]}),
		)
	}

	// tune irq
	if !dryrun {
		if err = setNICIRQ(irqs, desiredAffinity); err != nil {
			return errors.Wrap(err, "failed to set NIC IRQ affinity")
		}
		klog.InfoS("finish tuning NIC IRQ affinity", "nic", eth, "irq", irqs.IRQs(), "affinity", desiredAffinity)
	} else {
		klog.InfoS("dryrun mode: skip tuning NIC IRQ affinity", "nic", eth)
	}

	return
}
