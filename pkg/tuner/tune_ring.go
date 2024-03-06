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

	r := eth.Ring

	klog.InfoS("tune NIC ring buffer", "nic", eth.Name,
		"RxJumbo", fmt.Sprintf("%d -> %d", r.RxJumboPending, r.RxJumboMaxPending),
		"RxMini", fmt.Sprintf("%d -> %d", r.RxMiniPending, r.RxMiniPending),
		"Tx", fmt.Sprintf("%d -> %d", r.TxPending, r.TxPending),
		"Rx", fmt.Sprintf("%d -> %d", r.RxPending, r.RxPending),
	)

	// tune NIC ring
	if r.RxJumboPending < r.RxJumboMaxPending || r.TxPending < r.TxMaxPending || r.RxPending < r.RxMaxPending || r.RxMiniPending < r.RxMiniMaxPending {
		if !option.DryRun {
			r.RxJumboPending = r.RxJumboMaxPending
			r.RxMiniPending = r.RxMiniMaxPending
			r.TxPending = r.TxMaxPending
			r.RxPending = r.RxMaxPending

			curr, err := eth.SetRingBuffer(*r)
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
