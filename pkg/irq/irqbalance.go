package irq

import (
	"context"
	"usoftirqueue/pkg/system"

	"k8s.io/klog/v2"
)

func StopIRQBalance(dryrun bool) (err error) {

	klog.InfoS("stop irqbalance")
	if dryrun {
		klog.Info("dryrun mode: skip to stop irqbalance")
		return
	}

	var (
		svc = "irqbalance.service"
		bus *system.SystemCtl
	)

	if bus, err = system.NewWithContext(context.Background()); err != nil {
		return
	}
	defer bus.Close()

	if err = bus.StopService(svc); err != nil {
		return
	}
	klog.InfoS("service stopped", "service", svc)
	return
}

func StartIRQBalance(dryrun bool) (err error) {

	klog.InfoS("start irqbalance")
	if dryrun {
		klog.Info("dryrun mode: skip to start irqbalance")
		return
	}

	var (
		svc = "irqbalance.service"
		bus *system.SystemCtl
	)

	if bus, err = system.NewWithContext(context.Background()); err != nil {
		return
	}
	defer bus.Close()

	if err = bus.StartService(svc); err != nil {
		return
	}
	klog.InfoS("service started", "service", svc)
	return
}
