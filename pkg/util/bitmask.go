package util

import (
	"fmt"
	"strings"
)

func BitMask(removeLeadingZero bool, total int, idxs ...int) (mask string) {

	sz := total / 32
	if total%32 > 0 {
		sz++
	}
	bms := make([]uint32, sz)

	for _, idx := range idxs {
		n, r := idx/32, idx%32
		bms[n] |= 1 << r
	}

	// to string
	var masks []string
	cont := true
	o := 0
	for i := sz - 1; i >= 0; i-- {
		if sz > 1 && i != 0 && cont && bms[i] == 0 {
			continue
		}
		cont = false
		if removeLeadingZero && o == 0 {
			masks = append(masks, fmt.Sprintf("%x", bms[i]))
		} else {
			masks = append(masks, fmt.Sprintf("%08x", bms[i]))
		}
		o++
	}
	return strings.Join(masks, ",")
}

func RoundUpToPowerOfTwo(n int) int {
	if n <= 2 || n&(n-1) == 0 {
		return n
	}

	var c, v int
	for v := n; v != 0; c++ {
		v >>= 1
	}
	v = 1 << c
	if n > v {
		return n
	}
	return v
}
