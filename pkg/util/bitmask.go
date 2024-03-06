package util

import (
	"fmt"
	"strings"
)

func BitMask(removeLeadingZero bool, total int, idxs ...int) (mask string) {

	sz := total / 4
	if total%4 > 0 {
		sz++
	}
	bms := make([]uint8, sz)

	for _, idx := range idxs {
		n, r := idx/4, idx%4
		bms[n] |= 1 << r
	}

	// remove leading zero
	if removeLeadingZero {
		pos := sz - 1
		for i := sz - 1; i >= 0; i-- {
			if bms[i] != 0 {
				pos = i
				break
			}
			if i == 0 {
				pos = 0
				break
			}
		}
		bms = bms[:pos+1]
		sz = len(bms)
	}

	// convert to hex string
	var masks []string
	var s string
	for i := 0; i < sz; i++ {
		s = fmt.Sprintf("%x", bms[i]) + s
		if (i+1)%8 == 0 || i == sz-1 {
			masks = append(masks, s)
			s = ""
		}
	}

	// reverse masks
	for i, j := 0, len(masks)-1; i < j; i, j = i+1, j-1 {
		masks[i], masks[j] = masks[j], masks[i]
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
