package nic

import (
	"errors"
	"fmt"
	"net"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"usoftirqueue/pkg/hardware/irq"
	"usoftirqueue/pkg/util/formatter"

	"github.com/safchain/ethtool"
	"k8s.io/klog/v2"
)

var nicQueueIRQFilter = map[string]*regexp.Regexp{
	"i40e":       regexp.MustCompile(`(i40e-eth[0-9]+-TxRx)-[0-9]+`),
	"mlx4_core":  regexp.MustCompile(`(mlx4_comp)[0-9]+@pci:.+`),
	"mlx5_core":  regexp.MustCompile(`(mlx5_comp)[0-9]+@pci:.+`),
	"virtio_net": regexp.MustCompile(`(virtio[0-9]+-(in|out)put)[.][0-9]+`),
	"ena":        regexp.MustCompile(`(eth[0-9]+-Tx-Rx)-[0-9]+`),
}

func RegisterNICQueueIRQFilter(driverQueueIRQRegexes []string) (err error) {
	for _, driverQueueIRQRegex := range driverQueueIRQRegexes {
		idx := strings.Index(driverQueueIRQRegex, ":")
		if idx == -1 {
			return fmt.Errorf("register NIC queue IRQ filter: unrecognized format: %s", driverQueueIRQRegex)
		}
		d, r := driverQueueIRQRegex[:idx], driverQueueIRQRegex[idx+1:]
		if nicQueueIRQFilter[d], err = regexp.Compile(r); err != nil {
			return fmt.Errorf("register NIC queue IRQ filter: unrecognized regexp: %v", err)
		}
		klog.InfoS("register NIC queue IRQ filter", "driver", d, "regexp", r)
	}
	return
}

func GetNICQueueIRQFilter() map[string]*regexp.Regexp {
	return nicQueueIRQFilter
}

type NIC struct {
	Name   string      // e.g. eth0
	Driver string      // e.g. mlx5_core
	Device string      // PCI bus
	IRQs   irq.IRQList // activated IRQs
	Queue  *ethtool.Channels
	Ring   *ethtool.Ring
}

type NICList []*NIC

func (nics NICList) Names() []string {
	names := make([]string, len(nics))
	for i, n := range nics {
		names[i] = n.Name
	}
	return names
}

// irqses
func (nics NICList) IRQs() [][]int {
	irqs := make([][]int, len(nics))
	for i, n := range nics {
		irqs[i] = n.IRQs.Numbers()
	}
	return irqs
}

type NICGroup []NICList

func (ngs NICGroup) Names() [][]string {
	names := make([][]string, len(ngs))
	for i, ng := range ngs {
		names[i] = ng.Names()
	}
	return names
}

func NewNICInfo(name string) (*NIC, error) {
	var err error
	n := &NIC{Name: name}
	n.Device = n.getDevice()
	n.Driver = n.getDriver()
	n.IRQs = n.getIRQs()
	if n.Queue, err = n.getQueueChannel(true); err != nil {
		return nil, err
	}
	if n.Ring, err = n.getRingBuffer(true); err != nil {
		return nil, err
	}
	return n, nil
}

func List(regexpr string) (nics NICList, err error) {

	var (
		reg    *regexp.Regexp
		ifaces []net.Interface
		nic    *NIC
	)

	if reg, err = regexp.Compile(regexpr); err != nil {
		return
	}

	if ifaces, err = net.Interfaces(); err != nil {
		return
	}

	for _, iface := range ifaces {
		// filter out virtual NIC
		dev, _ := filepath.EvalSymlinks(fmt.Sprintf("/sys/class/net/%s", iface.Name))
		if strings.Contains(dev, "/virtual/net/") {
			klog.InfoS("ignore software NIC", "nic", iface.Name, "mac", iface.HardwareAddr, "flag", iface.Flags)
			continue
		}
		klog.InfoS("found hardware NIC", "nic", iface.Name, "mac", iface.HardwareAddr, "flag", iface.Flags)
		// filter by regexpr
		if reg.MatchString(iface.Name) {
			if nic, err = NewNICInfo(iface.Name); err != nil {
				klog.ErrorS(err, "failed to get NIC info", "nic", iface.Name, "mac", iface.HardwareAddr, "flag", iface.Flags)
				return
			}
			nics = append(nics, nic)
		}
	}
	return
}

func (n *NIC) getDevice() string {
	devPath := fmt.Sprintf("/sys/class/net/%s/device", n.Name)
	dev, _ := filepath.EvalSymlinks(devPath)
	// s = "/sys/devices/pci0000:16/0000:16:02.0/0000:17:00.0/0000:18:11.0/0000:2a:00.0/virtio4"
	if strings.Contains(dev, "virtio") {
		dev = filepath.Dir(dev)
	}
	klog.V(4).InfoS("get device", "nic", n.Name, "device", dev)
	return dev
}

func (n *NIC) getDriver() string {
	drvPath := fmt.Sprintf("/sys/class/net/%s/device/driver", n.Name)
	s, _ := filepath.EvalSymlinks(drvPath)
	drv := filepath.Base(s)
	klog.V(4).InfoS("get driver", "nic", n.Name, "driver", drv)
	return drv
}

func (n *NIC) getIRQs() (irqs irq.IRQList) {

	// get irq filter
	filter := nicQueueIRQFilter[n.Driver]
	if filter == nil {
		filter = nicQueueIRQFilter[""]
	}

	// get irqs in dev dir
	// TODO: no MSI-X IRQ?
	var irqNums []string
	irqNumsPath, _ := filepath.Glob(fmt.Sprintf("%s/msi_irqs/*", n.Device))
	for _, p := range irqNumsPath {
		irqNums = append(irqNums, filepath.Base(p))
	}
	klog.V(4).InfoS("get irq numbers", "nic", n.Name, "irqs", irqNums)

	// get irq info
	for _, irqNum := range irqNums {
		irq, err := irq.New(irqNum, filter)
		if err != nil {
			klog.ErrorS(err, "failed to get irq info", "nic", n.Name, "irq", irqNum)
			continue
		}
		klog.V(4).InfoS("get irq info", "nic", n.Name, "irq", formatter.NewYAML(irq))
		if irq.Name != "" {
			irqs = append(irqs, irq)
		}
	}

	// sort by irq number
	sort.Slice(irqs, func(i, j int) bool {
		return irqs[i].Number < irqs[j].Number
	})

	klog.InfoS("filter TX/RX IRQs", "nic", n.Name, "driver", n.Driver, "bus", strings.ReplaceAll(n.Device, "/sys/devices/", ""), "count", len(irqs), "irqs", irqs.Numbers())
	return
}

func (n *NIC) getQueueChannel(reload bool) (channels *ethtool.Channels, err error) {

	if !reload && n.Queue != nil {
		return n.Queue, nil
	}

	var (
		et *ethtool.Ethtool
		ec ethtool.Channels
	)

	if et, err = ethtool.NewEthtool(); err != nil {
		return
	}
	defer et.Close()

	if ec, err = et.GetChannels(n.Name); err != nil {
		return
	}

	klog.V(4).InfoS("get queue channel", "nic", n.Name, "queue", ec)
	return &ec, err
}

func (n *NIC) SetQueueChannel(desired ethtool.Channels) (channels *ethtool.Channels, err error) {

	var (
		et *ethtool.Ethtool
		ec ethtool.Channels
	)

	if et, err = ethtool.NewEthtool(); err != nil {
		return
	}
	defer et.Close()

	if ec, err = et.SetChannels(n.Name, desired); err != nil {
		return
	}

	// refresh IRQs
	if n.Queue == nil || n.Queue.CombinedCount != ec.CombinedCount || n.Queue.RxCount != ec.RxCount || n.Queue.TxCount != ec.TxCount {
		old, new := n.updateIRQs()
		klog.InfoS("refresh NIC IRQs after setting queue channel", "nic", n.Name, "old", formatter.NewYAML(old), "new", formatter.NewYAML(new))
	}
	n.Queue = &ec
	return &ec, err
}

func (n *NIC) getRingBuffer(reload bool) (ring *ethtool.Ring, err error) {

	if !reload && n.Ring != nil {
		return n.Ring, nil
	}

	var (
		et *ethtool.Ethtool
		er ethtool.Ring
	)

	if et, err = ethtool.NewEthtool(); err != nil {
		return
	}
	defer et.Close()

	if er, err = et.GetRing(n.Name); err != nil {
		return
	}

	klog.V(4).InfoS("get ring buffer", "nic", n.Name, "ring", er)
	return &er, err
}

func (n *NIC) SetRingBuffer(desired ethtool.Ring) (ring *ethtool.Ring, err error) {

	var (
		et *ethtool.Ethtool
		er ethtool.Ring
	)

	if et, err = ethtool.NewEthtool(); err != nil {
		return
	}
	defer et.Close()

	if er, err = et.SetRing(n.Name, desired); err != nil {
		return
	}

	n.Ring = &er
	return &er, err
}

func (n *NIC) GetFeatures() (features map[string]bool, err error) {

	var (
		et *ethtool.Ethtool
	)

	if et, err = ethtool.NewEthtool(); err != nil {
		return
	}
	defer et.Close()

	return et.Features(n.Name) // pkg bugs: features are not full
}

func (n *NIC) SetFeature(key string, value bool) (err error) {

	var (
		et *ethtool.Ethtool
	)

	if et, err = ethtool.NewEthtool(); err != nil {
		return
	}
	defer et.Close()

	return et.Change(n.Name, map[string]bool{key: value})
}

// updateIRQs refresh irqs after change the queue channel of NIC
func (n *NIC) updateIRQs() (old, new irq.IRQList) {
	old = n.IRQs
	new = n.getIRQs()
	n.IRQs = new
	return
}

func (n *NIC) SetIRQAffinity(cpus []int) (err error) {
	if ni, nc := len(n.IRQs), len(cpus); ni != nc {
		return fmt.Errorf("the number of cpus %d is not equal to the number of irqs %d", ni, nc)
	}
	for i, irq := range n.IRQs {
		if e := irq.SetAffinity(cpus[i]); e != nil {
			err = errors.Join(err, e)
		}
	}
	return
}
