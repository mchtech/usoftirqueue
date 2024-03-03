package nic

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"usoftirqueue/pkg/config"
	"usoftirqueue/pkg/system"
	"usoftirqueue/pkg/util/formatter"

	"k8s.io/klog/v2"
)

type IRQ struct {
	NIC      string
	Driver   string
	Device   string
	Name     string
	Names    []string
	IRQ      int
	NUMA     int
	Affinity []int
}

type IRQList []IRQ

func (l IRQList) IRQs() (q []int) {
	for _, i := range l {
		q = append(q, i.IRQ)
	}
	return
}

var IRQFilters = map[string]*regexp.Regexp{
	"mlx5_core":  regexp.MustCompile(`mlx5_comp[0-9]+@pci:.+`),
	"virtio_net": regexp.MustCompile(`virtio[0-9]+-(in|out)put[.][0-9]+`),
}

func UpdateIRQFilter(driver, regex string) {
	IRQFilters[driver] = regexp.MustCompile(regex)
}

func getDevicePath(nic string) string {
	devPath := fmt.Sprintf("/sys/class/net/%s/device", nic)
	s, _ := filepath.EvalSymlinks(devPath)
	// s = "/sys/devices/pci0000:16/0000:16:02.0/0000:17:00.0/0000:18:11.0/0000:2a:00.0/virtio4"
	if strings.Contains(s, "virtio") {
		s = filepath.Dir(s)
	}
	return s
}

func getDeviceDriver(nic string) string {
	drvPath := fmt.Sprintf("/sys/class/net/%s/device/driver", nic)
	s, _ := filepath.EvalSymlinks(drvPath)
	return filepath.Base(s)
}

func getIRQInfo(irq string) IRQ {

	irqInt, _ := strconv.Atoi(irq)

	// /proc/irq/$irq/
	var names []string
	f, _ := os.Open(fmt.Sprintf("/proc/irq/%s/", irq))
	defer f.Close()
	dirs, _ := f.ReadDir(-1)
	for _, dir := range dirs {
		if dir.IsDir() {
			names = append(names, dir.Name())
		}
	}
	// names, _ := filepath.Glob(fmt.Sprintf("/proc/irq/%s/*/", irq))

	// /proc/irq/$irq/node
	// -1 means no NUMA node
	bnuma, _ := os.ReadFile(fmt.Sprintf("/proc/irq/%s/node", irq))
	numa, _ := strconv.Atoi(string(bnuma))
	if numa == -1 {
		numa = 0
	}

	// /proc/irq/$irq/smp_affinity_list
	baffinity, _ := os.ReadFile(fmt.Sprintf("/proc/irq/%s/smp_affinity_list", irq))
	affinity := strings.Split(strings.TrimSpace(string(baffinity)), ",")
	var affinityInt []int
	for _, a := range affinity {
		i, _ := strconv.Atoi(a)
		affinityInt = append(affinityInt, i)
	}

	return IRQ{
		Name:     strings.Join(names, ","),
		Names:    names,
		IRQ:      irqInt,
		NUMA:     numa,
		Affinity: affinityInt,
	}
}

func getNICIRQ(nic string) (irqs IRQList) {
	// get device path
	dev := getDevicePath(nic)
	klog.InfoS("get dev", "nic", nic, "dev", dev)

	// get device driver
	drv := getDeviceDriver(nic)
	klog.InfoS("get drv", "nic", nic, "dev", dev, "driver", drv)

	// get irq filter
	filter := IRQFilters[drv]

	// get irqs in dir dev
	var irqNums []string
	irqNumsPath, _ := filepath.Glob(fmt.Sprintf("%s/msi_irqs/*", dev))
	for _, p := range irqNumsPath {
		irqNums = append(irqNums, filepath.Base(p))
	}
	klog.InfoS("get irqs", "nic", nic, "irqs", irqNums)

	// get irq info
	for _, irqNum := range irqNums {
		irq := getIRQInfo(irqNum)
		irq.NIC = nic
		irq.Driver = drv
		irq.Device = dev
		// klog.InfoS("get irq", "nic", nic, "irq", irq)
		if filter != nil {
			for _, name := range irq.Names {
				if filter.MatchString(name) {
					irqs = append(irqs, irq)
					break
				}
			}
		} else {
			irqs = append(irqs, irq)
		}
	}
	klog.InfoS("get irqs", "nic", nic, "irqs", formatter.NewYAML(irqs))
	return
}

func setNICIRQ(irqs IRQList, cpus []int) (err error) {
	for i, irq := range irqs {
		if e := config.Write(
			fmt.Sprintf("/proc/irq/%d/smp_affinity_list", irq.IRQ),
			fmt.Sprintf("%d", cpus[i]),
		); e != nil {
			err = errors.Join(err, e)
		}
	}
	return
}

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
