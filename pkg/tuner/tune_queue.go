package tuner

import (
	"fmt"
	"usoftirqueue/pkg/hardware/nic"
	"usoftirqueue/pkg/option"
	"usoftirqueue/pkg/util/formatter"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
)

func (t *Tuner) TuneNICQueueChannelMax(eth *nic.NIC, totalCPUCount uint32) error {

	// get NIC queue
	ech, err := eth.GetQueueChannel()
	if err != nil {
		return errors.Wrap(err, "failed to get NIC queue channel")
	}
	klog.InfoS("current NIC queue channel", "nic", eth.Name, "queue", formatter.NewYAML(ech))

	// calculate desired queue channel
	desiredCombined := ech.MaxCombined
	if desiredCombined > totalCPUCount {
		desiredCombined = totalCPUCount
	}

	desiredRX := ech.MaxRx
	if desiredRX > totalCPUCount {
		desiredRX = totalCPUCount
	}

	desiredTX := ech.MaxTx
	if desiredTX > totalCPUCount {
		desiredTX = totalCPUCount
	}

	klog.InfoS("tune NIC queue channel", "nic", eth.Name,
		"Combined", fmt.Sprintf("%d -> %d", ech.CombinedCount, desiredCombined),
		"Rx", fmt.Sprintf("%d -> %d", ech.RxCount, desiredRX),
		"Tx", fmt.Sprintf("%d -> %d", ech.TxCount, desiredTX),
	)

	// tune queue channel
	if ech.CombinedCount != desiredCombined || ech.RxCount != desiredRX || ech.TxCount != desiredTX {
		if !option.DryRun {
			ech.CombinedCount = desiredCombined
			ech.RxCount = desiredRX
			ech.TxCount = desiredTX
			curr, err := eth.SetQueueChannel(*ech)
			if err != nil {
				return errors.Wrap(err, "failed to set NIC queue channel")
			}
			klog.InfoS("finish tuning NIC queue channel", "nic", eth.Name, "queue", curr)
			// refresh IRQs
			old, new := eth.UpdateIRQs()
			klog.InfoS("finish updating NIC IRQs", "nic", eth.Name, "old", formatter.NewYAML(old), "new", formatter.NewYAML(new))
		} else {
			klog.InfoS("dryrun mode: skip tuning NIC queue channel", "nic", eth.Name)
		}
	} else {
		klog.InfoS("no need to tune NIC queue channel", "nic", eth.Name)
	}

	return nil
}
