package nic

import (
	"errors"
	"fmt"
	"net"
	"path/filepath"
	"regexp"
	"strings"
	"usoftirqueue/pkg/irq"
	"usoftirqueue/pkg/util/formatter"

	"github.com/safchain/ethtool"
	"k8s.io/klog/v2"
)

var nicQueueIRQFilter = map[string]*regexp.Regexp{
	"i40e":       regexp.MustCompile(`i40e-eth[0-9]+-TxRx-[0-9]+`),
	"mlx5_core":  regexp.MustCompile(`mlx5_comp[0-9]+@pci:.+`),
	"virtio_net": regexp.MustCompile(`virtio[0-9]+-(in|out)put[.][0-9]+`),
}

func UpdateNICQueueIRQFilter(driver, regex string) {
	nicQueueIRQFilter[driver] = regexp.MustCompile(regex)
}

type NIC struct {
	Name   string      // e.g. eth0
	Driver string      // e.g. mlx5_core
	Device string      //
	IRQs   irq.IRQList //
}

type NICList []*NIC

func (nics NICList) Names() []string {
	names := make([]string, len(nics))
	for i, n := range nics {
		names[i] = n.Name
	}
	return names
}

func NewNIC(name string) *NIC {
	n := &NIC{Name: name}
	n.Device = n.getDevice()
	n.Driver = n.getDriver()
	n.IRQs = n.getIRQs()
	return n
}

func List(regexpr string) (nics []*NIC, err error) {

	var (
		reg     *regexp.Regexp
		ifaces  []net.Interface
		phyNICs []net.Interface
	)

	if reg, err = regexp.Compile(regexpr); err != nil {
		return
	}

	if ifaces, err = net.Interfaces(); err != nil {
		return
	}

	// filter virtual NIC
	for _, iface := range ifaces {
		dev, _ := filepath.EvalSymlinks(fmt.Sprintf("/sys/class/net/%s", iface.Name))
		if strings.Contains(dev, "/virtual/net/") {
			klog.InfoS("ignore virtual NIC", "nic", iface.Name)
			continue
		}
		phyNICs = append(phyNICs, iface)
	}
	klog.InfoS("get all phy NICs", "nics", formatter.NewYAML(phyNICs))

	// filter NICs by regexpr
	for _, iface := range phyNICs {
		if reg.MatchString(iface.Name) {
			nics = append(nics, NewNIC(iface.Name))
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

	// n.Device = s
	klog.InfoS("get dev", "nic", n.Name, "device", dev)
	return dev
}

func (n *NIC) getDriver() string {
	drvPath := fmt.Sprintf("/sys/class/net/%s/device/driver", n.Name)
	s, _ := filepath.EvalSymlinks(drvPath)
	drv := filepath.Base(s)

	// n.Driver = drv
	klog.InfoS("get driver", "nic", n.Name, "driver", drv)
	return drv
}

func (n *NIC) getIRQs() (irqs irq.IRQList) {

	// get irq filter
	filter := nicQueueIRQFilter[n.Driver]

	// get irqs in dir dev
	var irqNums []string
	irqNumsPath, _ := filepath.Glob(fmt.Sprintf("%s/msi_irqs/*", n.Device))
	for _, p := range irqNumsPath {
		irqNums = append(irqNums, filepath.Base(p))
	}
	klog.InfoS("get irqs", "nic", n.Name, "irqs", irqNums)

	// get irq info
	for _, irqNum := range irqNums {
		irq := irq.New(irqNum)
		klog.InfoS("get irq", "nic", n.Name, "irq", irq)
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

	// n.IRQs = irqs
	klog.InfoS("get irqs", "nic", n.Name, "irqs", formatter.NewYAML(irqs))
	return
}

func (n *NIC) SetIRQAffinity(cpus []int) (err error) {
	for i, irq := range n.IRQs {
		if !irq.Activated {
			continue
		}
		if e := irq.SetAffinity(cpus[i]); e != nil {
			err = errors.Join(err, e)
		}
	}
	return
}

func (n *NIC) GetQueueChannel() (channels *ethtool.Channels, err error) {

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
	return &ec, err
}

func (n *NIC) GetRingBuffer() (ring *ethtool.Ring, err error) {

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
