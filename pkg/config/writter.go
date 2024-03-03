package config

import (
	"fmt"
	"os"
	"strings"

	"k8s.io/klog/v2"
)

var DryRun bool

func Write(file, value string) error {

	bf, _ := os.ReadFile(file)
	curr := strings.TrimSpace(string(bf))

	if DryRun {
		klog.InfoS("dryrun mode: skip to write system config", "key", file, "value", fmt.Sprintf("%s -> %s", curr, value))
		return nil
	} else {
		klog.InfoS("write system config", "key", file, "value", fmt.Sprintf("%s -> %s", curr, value))
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
