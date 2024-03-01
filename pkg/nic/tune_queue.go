package nic

import (
	"fmt"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
)

func TuneNICQueueChannelMax(localCPUCount, totalCPUCount uint32, eth string, dryrun bool) error {

	// get NIC queue
	ech, err := GetNICQueueChannel(eth)
	if err != nil {
		return errors.Wrap(err, "failed to get NIC queue channel")
	}
	klog.InfoS("current NIC queue channel", "nic", eth, "queue", ech)

	// calculate desired queue channel
	desiredCombined := ech.MaxCombined
	if desiredCombined > totalCPUCount {
		desiredCombined = totalCPUCount
	} else if desiredCombined > localCPUCount {
		desiredCombined = localCPUCount
	}

	desiredRX := ech.MaxRx
	if desiredRX > totalCPUCount {
		desiredRX = totalCPUCount
	} else if desiredRX > localCPUCount {
		desiredRX = localCPUCount
	}

	desiredTX := ech.MaxTx
	if desiredTX > totalCPUCount {
		desiredTX = totalCPUCount
	} else if desiredTX > localCPUCount {
		desiredTX = localCPUCount
	}

	klog.InfoS("tune NIC queue channel", "nic", eth,
		"Combined", fmt.Sprintf("%d -> %d", ech.CombinedCount, desiredCombined),
		"Rx", fmt.Sprintf("%d -> %d", ech.RxCount, desiredRX),
		"Tx", fmt.Sprintf("%d -> %d", ech.TxCount, desiredTX),
	)

	// tune queue channel
	if ech.CombinedCount != desiredCombined || ech.RxCount != desiredRX || ech.TxCount != desiredTX {
		if !dryrun {
			ech.CombinedCount = desiredCombined
			ech.RxCount = desiredRX
			ech.TxCount = desiredTX
			curr, err := SetNICQueueChannel(eth, *ech)
			if err != nil {
				return errors.Wrap(err, "failed to set NIC queue channel")
			}
			klog.InfoS("finish tuning NIC queue channel", "nic", eth, "queue", curr)
		} else {
			klog.InfoS("dryrun mode: skip tuning NIC queue channel", "nic", eth)
		}
	} else {
		klog.InfoS("no need to tune NIC queue channel", "nic", eth)
	}

	return nil
}
