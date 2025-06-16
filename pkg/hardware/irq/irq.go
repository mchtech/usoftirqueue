package irq

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"usoftirqueue/pkg/util/config"

	"github.com/pkg/errors"
)

type IRQ struct {
	Name     string
	Prefix   string
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

func IRQsReady(irqs []string) bool {
	for _, irq := range irqs {
		if _, err := os.Stat(fmt.Sprintf("/proc/irq/%s/", irq)); err != nil {
			return false
		}
	}
	return true
}

func New(irq string, filter *regexp.Regexp) (IRQ, error) {

	var (
		name   string
		prefix string
	)

	irqInt, _ := strconv.Atoi(irq)

	// /proc/irq/$irq/
	f, err := os.Open(fmt.Sprintf("/proc/irq/%s/", irq))
	if err != nil {
		if os.IsNotExist(err) {
			err = errors.Wrapf(err, "IRQ %s is not activated", irq)
		}
		return IRQ{Number: irqInt}, err
	}
	defer f.Close()

	// name
	dirs, _ := f.ReadDir(-1)
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		dn := dir.Name()
		if filter == nil || filter.MatchString(dn) {
			name = dn
			break
		}
	}

	// prefix: virtio0-input.0 => virtio0-input
	if filter != nil {
		if matches := filter.FindStringSubmatch(name); len(matches) > 1 {
			prefix = matches[1]
		}
	}

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
		Name:     name,
		Prefix:   prefix,
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
