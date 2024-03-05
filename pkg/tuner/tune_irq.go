package tuner

import (
	"context"
	"fmt"
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

	// TODO: filter enabled
	irqs := eth.IRQs
	numIRQ := irqs.NumIRQ()
	numCore := cpus.NumCPU()

	var numa int
	switch opt.CPUSelectMask {
	case option.CPUSelectMaskLocalNUMA:
		numa = irqs[0].NUMA
	case option.CPUSelectMaskAll:
		numa = -1
	default:
		return fmt.Errorf("unknown CPUSelectMask: %s", opt.CPUSelectMask)
	}
	if numa >= 0 {
		numCore = cpus.NumLocalCPU(numa)
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
			"irq", irqs[i].IRQ,
			"name", irqs[i].Name,
			"driver", eth.Driver,
			// "bus", eth.Device,
			"affinity", fmt.Sprintf("%v -> %v", irqs[i].Affinity, []int{desiredAffinity[i]}),
		)
	}

	// tune irq
	if !option.DryRun {
		if err = eth.SetIRQAffinity(desiredAffinity); err != nil {
			return errors.Wrap(err, "failed to set NIC IRQ affinity")
		}
		klog.InfoS("finish tuning NIC IRQ affinity", "nic", eth.Name, "irq", irqs.GetActivatedIRQs(), "affinity", desiredAffinity)
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
