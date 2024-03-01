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
	biterator *util.BiasIterator[int]
)

func TuneNICIRQ(eth string, cpus *cpu.Info, dryrun bool, opt *option.Option) (err error) {

	var desiredAffinity []int

	irqs := getNICIRQ(eth)
	numIRQ := len(irqs)
	numCPU := int(cpus.TotalThreads)

	var numa int
	switch opt.CPUSelectMask {
	case option.CPUSelectMaskLocalNUMA:
		numa = irqs[0].NUMA
	case option.CPUSelectMaskAll:
		numa = -1
	default:
		return fmt.Errorf("unknown CPUSelectMask: %d", opt.CPUSelectMask)
	}

	if opt.CPUSelectPolicy == option.CPUSelectPolicyFromIterator {
		desiredAffinity = util.GetCoreListByIterator(cpus, numa, false, numIRQ)
	} else {

		numCore := numCPU
		if numa >= 0 {
			numCore = int(cpus.Processors[0].NumThreads)
		}

		cores := util.GetCoreList(cpus, numa, false)
		once.Do(func() {
			biterator = util.NewBiasIterator(cores)
		})

		if numIRQ == numCore {
			desiredAffinity = cores
		} else if numIRQ < numCore {
			if numCore%numIRQ == 0 {
				desiredAffinity = append(desiredAffinity, biterator.Interval(numCore/numIRQ).Generate(numIRQ)...)
			} else {
				desiredAffinity = append(desiredAffinity, biterator.Interval(1).Generate(numIRQ)...)
			}
		} else {
			desiredAffinity := make([]int, 0)
			for {
				diff := numIRQ - len(desiredAffinity)
				if diff > len(cores) {
					desiredAffinity = append(desiredAffinity, cores...)
				} else if diff > 0 {
					if numCore%diff == 0 {
						desiredAffinity = append(desiredAffinity, biterator.Interval(numCore/diff).Generate(diff)...)
					} else {
						desiredAffinity = append(desiredAffinity, biterator.Interval(1).Generate(diff)...)
					}
				} else {
					break
				}
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
