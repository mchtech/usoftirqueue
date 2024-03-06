package irq

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"usoftirqueue/pkg/util/config"

	"github.com/pkg/errors"
)

type IRQ struct {
	Name     string
	Names    []string
	Number   int
	NUMAID   int
	Affinity []int
}

type IRQList []IRQ

func (l IRQList) Numbers() (q []int) {
	for _, i := range l {
		q = append(q, i.Number)
	}
	return
}

func New(irq string) (IRQ, error) {

	irqInt, _ := strconv.Atoi(irq)

	// /proc/irq/$irq/
	var names []string
	f, err := os.Open(fmt.Sprintf("/proc/irq/%s/", irq))
	if err != nil {
		if os.IsNotExist(err) {
			err = errors.Wrapf(err, "IRQ %s is not activated", irq)
		}
		return IRQ{Number: irqInt}, err
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
		Name:     strings.Join(names, ","),
		Names:    names,
		Number:   irqInt,
		NUMAID:   numa,
		Affinity: affinityInt,
	}, nil
}

func (irq *IRQ) SetAffinity(cpu int) (err error) {
	return config.Write(
		fmt.Sprintf("/proc/irq/%d/smp_affinity_list", irq.Number),
		fmt.Sprintf("%d", cpu),
	)
}
