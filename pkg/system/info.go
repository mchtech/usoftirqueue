package system

import (
	"fmt"
	"net"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"usoftirqueue/pkg/formatter"

	"github.com/jaypipes/ghw"
	"github.com/jaypipes/ghw/pkg/cpu"
	"github.com/jaypipes/ghw/pkg/option"
	"k8s.io/klog/v2"
)

func GetCPUInfo() (cpus *cpu.Info, err error) {
	if cpus, err = ghw.CPU(option.WithNullAlerter(), option.WithDisableTools()); err != nil {
		return
	}
	for p := range cpus.Processors {
		cpus.Processors[p].Capabilities = nil
		sort.Slice(cpus.Processors[p].Cores, func(j, k int) bool {
			return cpus.Processors[p].Cores[j].ID < cpus.Processors[p].Cores[k].ID
		})
		for c := range cpus.Processors[p].Cores {
			sort.Slice(cpus.Processors[p].Cores[c].LogicalProcessors, func(j, k int) bool {
				return cpus.Processors[p].Cores[c].LogicalProcessors[j] < cpus.Processors[p].Cores[c].LogicalProcessors[k]
			})
		}
	}
	klog.InfoS("get cpus", "cpus", cpus.YAMLString())
	return
}

// 获取网卡列表
func GetNICs(regexpr string) (nics []string, err error) {

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
			continue
		}
		phyNICs = append(phyNICs, iface)
	}
	klog.InfoS("get all phy NICs", "nics", formatter.NewYAML(phyNICs))

	// filter NICs by regexpr
	for _, iface := range phyNICs {
		if reg.MatchString(iface.Name) {
			nics = append(nics, iface.Name)
		}
	}
	return
}
