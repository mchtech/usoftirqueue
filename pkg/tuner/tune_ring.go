package tuner

import (
	"fmt"
	"usoftirqueue/pkg/hardware/nic"
	"usoftirqueue/pkg/option"
	"usoftirqueue/pkg/util/formatter"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
)

func (t *Tuner) TuneNICRingBufferMax(eth *nic.NIC) error {

	// get NIC ring
	ring, err := eth.GetRingBuffer()
	if err != nil {
		return errors.Wrap(err, "failed to get NIC ring")
	}
	klog.InfoS("get NIC ring buffer", "nic", eth.Name, "ring", ring)

	klog.InfoS("tune NIC ring buffer", "nic", eth.Name,
		"RxJumbo", fmt.Sprintf("%d -> %d", ring.RxJumboPending, ring.RxJumboMaxPending),
		"RxMini", fmt.Sprintf("%d -> %d", ring.RxMiniPending, ring.RxMiniPending),
		"Tx", fmt.Sprintf("%d -> %d", ring.TxPending, ring.TxPending),
		"Rx", fmt.Sprintf("%d -> %d", ring.RxPending, ring.RxPending),
	)

	// tune NIC ring
	if ring.RxJumboPending < ring.RxJumboMaxPending || ring.TxPending < ring.TxMaxPending || ring.RxPending < ring.RxMaxPending || ring.RxMiniPending < ring.RxMiniMaxPending {
		if !option.DryRun {
			ring.RxJumboPending = ring.RxJumboMaxPending
			ring.RxMiniPending = ring.RxMiniMaxPending
			ring.TxPending = ring.TxMaxPending
			ring.RxPending = ring.RxMaxPending

			curr, err := eth.SetRingBuffer(*ring)
			if err != nil {
				return errors.Wrap(err, "failed to set NIC ring buffer")
			}
			klog.InfoS("finish setting NIC ring buffer", "nic", eth.Name, "ring", formatter.NewYAML(curr))
		} else {
			klog.InfoS("dryrun mode: skip setting NIC ring buffer", "nic", eth.Name)
		}
	} else {
		klog.InfoS("no need to tune NIC ring buffer", "nic", eth.Name)
	}

	return nil
}
