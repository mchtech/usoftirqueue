package config

import (
	"fmt"
	"os"
	"usoftirqueue/pkg/option"

	"k8s.io/klog/v2"
)

func Write(file, value string) error {
	curr := Read(file)
	if option.DryRun {
		klog.InfoS("dryrun mode: skip to write kernel config", "key", file, "value", fmt.Sprintf("%s -> %s", curr, value))
		return nil
	} else {
		klog.InfoS("write kernel config", "key", file, "value", fmt.Sprintf("%s -> %s", curr, value))
	}

	h, err := os.OpenFile(file, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer h.Close()
	if _, err = h.WriteString(value); err != nil {
		return err
	}
	return nil
}
