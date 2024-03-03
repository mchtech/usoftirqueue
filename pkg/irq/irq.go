package irq

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"usoftirqueue/pkg/config"
)

type IRQ struct {
	Activated bool
	Name      string
	Names     []string
	IRQ       int
	NUMA      int
	Affinity  []int
}

type IRQList []IRQ

func (l IRQList) GetActivatedIRQs() (q []int) {
	for _, i := range l {
		if i.Activated {
			q = append(q, i.IRQ)
		}
	}
	return
}

func New(irq string) IRQ {

	irqInt, _ := strconv.Atoi(irq)

	// /proc/irq/$irq/
	var names []string
	f, err := os.Open(fmt.Sprintf("/proc/irq/%s/", irq))
	if err != nil {
		return IRQ{
			IRQ:       irqInt,
			Activated: false,
		}
	}

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
		Activated: true,
		Name:      strings.Join(names, ","),
		Names:     names,
		IRQ:       irqInt,
		NUMA:      numa,
		Affinity:  affinityInt,
	}
}

func (irq *IRQ) SetAffinity(cpu int) (err error) {
	if !irq.Activated {
		return fmt.Errorf("IRQ %d is not activated", irq.IRQ)
	}
	return config.Write(
		fmt.Sprintf("/proc/irq/%d/smp_affinity_list", irq.IRQ),
		fmt.Sprintf("%d", cpu),
	)
}
