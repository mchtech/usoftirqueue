package config

import (
	"os"

	"k8s.io/klog/v2"
)

func Write(file, value string) error {
	klog.InfoS("write system config", "key", file, "value", value)
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
