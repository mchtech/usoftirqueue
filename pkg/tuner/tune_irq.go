package tuner

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"usoftirqueue/pkg/hardware/cpu"
	"usoftirqueue/pkg/hardware/nic"
	"usoftirqueue/pkg/option"
	"usoftirqueue/pkg/util"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
)

var (
	once sync.Once
	srr  *util.ScatterRoundRobinIterator[int]
	rr   *util.RoundRobinIterator[int]
)

func (t *Tuner) TuneIRQ(eth *nic.NIC, cpus *cpu.Topology, opt *option.Option) (err error) {

	var desiredAffinity []int

	irqs := eth.IRQs
	numIRQ := len(irqs)
	numCore := cpus.NumCPU()

	if q, r := numIRQ/numCore, numIRQ%numCore; r == 0 && q > 1 && q%2 == 0 {
		klog.Warningf("sort irqs by input/ouput, then by irq number: count(IRQs:%d) / count(cores:%d) = even number", numIRQ, numCore)
		sort.Slice(irqs, func(i, j int) bool {
			if irqs[i].Prefix == irqs[j].Prefix {
				return irqs[i].Number < irqs[j].Number
			}
			return irqs[i].Prefix < irqs[j].Prefix
		})
	}

	var numa int
	switch opt.CPUSelectMask {
	case option.CPUSelectMaskLocalNUMA:
		numa = irqs[0].NUMAID
	case option.CPUSelectMaskAll:
		numa = -1
	default:
		return fmt.Errorf("unknown CPUSelectMask: %v", opt.CPUSelectMask)
	}
	if numa >= 0 {
		numCore = cpus.NumLocalCPU(uint(numa))
	}

	cores := cpus.CoreIDs(numa, false)
	once.Do(func() {
		rr = util.NewRoundRobinIterator(cores)
		srr = util.NewScatterRoundRobinIterator(cores)
	})

	switch opt.CPUSelectStrategy {
	case option.CPUSelectStrategyRoundRobin:
		desiredAffinity = rr.NextN(numIRQ)
	case option.CPUSelectStrategyScatter:
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
		return errors.Errorf("unsupported CPU select policy %q", opt.CPUSelectStrategy)
	}

	for i := 0; i < numIRQ; i++ {
		klog.InfoS("tune NIC IRQ affinity",
			"nic", eth.Name,
			"driver", eth.Driver,
			"irq", irqs[i].Number,
			"name", irqs[i].Name,
			"prefix", irqs[i].Prefix,
			"affinity", fmt.Sprintf("%v -> %v", irqs[i].Affinity, []int{desiredAffinity[i]}),
		)
	}

	// tune irq
	if !option.DryRun {
		if err = eth.SetIRQAffinity(desiredAffinity); err != nil {
			return errors.Wrap(err, "failed to set NIC IRQ affinity")
		}
		klog.InfoS("finish tuning NIC IRQ affinity", "nic", eth.Name, "irq", irqs.Numbers(), "affinity", desiredAffinity)
	} else {
		klog.InfoS("dryrun mode: skip tuning NIC IRQ affinity", "nic", eth.Name)
	}

	return
}

func (t *Tuner) StopIRQBalance() (err error) {

	klog.InfoS("stop irqbalance")
	if option.DryRun {
		klog.Info("dryrun mode: skip to stop irqbalance")
		return
	}

	var (
		svc = "irqbalance.service"
		bus *util.SystemCtl
	)

	if bus, err = util.NewWithContext(context.Background()); err != nil {
		return
	}
	defer bus.Close()

	if err = bus.StopService(svc); err != nil {
		return
	}
	klog.InfoS("service stopped", "service", svc)
	return
}

func (t *Tuner) StartIRQBalance() (err error) {

	klog.InfoS("start irqbalance")
	if option.DryRun {
		klog.Info("dryrun mode: skip to start irqbalance")
		return
	}

	var (
		svc = "irqbalance.service"
		bus *util.SystemCtl
	)

	if bus, err = util.NewWithContext(context.Background()); err != nil {
		return
	}
	defer bus.Close()

	if err = bus.StartService(svc); err != nil {
		return
	}
	klog.InfoS("service started", "service", svc)
	return
}
