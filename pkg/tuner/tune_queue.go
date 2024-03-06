package tuner

import (
	"fmt"
	"usoftirqueue/pkg/hardware/nic"
	"usoftirqueue/pkg/option"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
)

func (t *Tuner) TuneNICQueueChannelMax(eth *nic.NIC, totalCPUCount uint32) error {

	q := eth.Queue

	// calculate desired queue channel
	desiredCombined := q.MaxCombined
	if desiredCombined > totalCPUCount {
		desiredCombined = totalCPUCount
	}

	desiredRX := q.MaxRx
	if desiredRX > totalCPUCount {
		desiredRX = totalCPUCount
	}

	desiredTX := q.MaxTx
	if desiredTX > totalCPUCount {
		desiredTX = totalCPUCount
	}

	klog.InfoS("tune NIC queue channel", "nic", eth.Name,
		"Combined", fmt.Sprintf("%d -> %d", q.CombinedCount, desiredCombined),
		"Rx", fmt.Sprintf("%d -> %d", q.RxCount, desiredRX),
		"Tx", fmt.Sprintf("%d -> %d", q.TxCount, desiredTX),
	)

	// tune queue channel
	if q.CombinedCount != desiredCombined || q.RxCount != desiredRX || q.TxCount != desiredTX {
		if !option.DryRun {
			q.CombinedCount = desiredCombined
			q.RxCount = desiredRX
			q.TxCount = desiredTX
			curr, err := eth.SetQueueChannel(*q)
			if err != nil {
				return errors.Wrap(err, "failed to set NIC queue channel")
			}
			klog.InfoS("finish tuning NIC queue channel", "nic", eth.Name, "queue", curr)
		} else {
			klog.InfoS("dryrun mode: skip tuning NIC queue channel", "nic", eth.Name)
		}
	} else {
		klog.InfoS("no need to tune NIC queue channel", "nic", eth.Name)
	}

	return nil
}
