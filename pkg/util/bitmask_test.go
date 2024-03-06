package util

import (
	"math"
	"regexp"
	"strconv"
	"testing"
)

func atoi(a string) (i int) {
	i, _ = strconv.Atoi(a)
	return
}

func atois(as []string) (is []int) {
	for _, a := range as {
		is = append(is, atoi(a))
	}
	return
}

func atob(a int) (s []string) {
	for i := 0; i < a; i++ {
		s = append(s, strconv.Itoa(i))
	}
	return
}

func TestBitMask(t *testing.T) {

	var data = [][]string{
		// all zero
		{"6", "00", "0"},
		{},
		{"12", "000", "0"},
		{},
		{"96", "00000000,00000000,00000000", "0"},
		{},
		{"104", "00,00000000,00000000,00000000", "0"},
		{},
		// all one
		{"6", "3f", "3f"},
		atob(6),
		{"12", "fff", "fff"},
		atob(12),
		{"96", "ffffffff,ffffffff,ffffffff", "ffffffff,ffffffff,ffffffff"},
		atob(96),
		{"104", "ff,ffffffff,ffffffff,ffffffff", "ff,ffffffff,ffffffff,ffffffff"},
		atob(104),
	}

	for i := 0; i < len(data); i += 2 {
		td := data[i]
		fa := BitMask(false, atoi(td[0]), atois(data[i+1])...)
		tr := BitMask(true, atoi(td[0]), atois(data[i+1])...)
		t.Logf("BitMask(false, %s) = %s, want %s", td[0], fa, td[1])
		t.Logf("BitMask(true, %s) = %s, want %s", td[0], tr, td[2])
		if fa != td[1] || tr != td[2] {
			t.Fail()
		}
	}

	// single bit
	reg := regexp.MustCompile(`[1-9].*`)
	queue := 244
	for i := 0; i < queue; i++ {
		a := BitMask(false, queue, i)
		b := BitMask(true, queue, i)
		c := reg.FindString(a)
		t.Logf("%d: %s -> %s / %s", i, a, c, b)
		if c != b {
			t.Fail()
		}
	}
	// t.Fail()
}

func TestRoundUpToPowerOfTwo(t *testing.T) {
	var data = [][]int{
		{16384, 16384},
		{512 * 16384, 512 * 16384},
		{12 * 16384, 16 * 16384},
		{104 * 16384, 128 * 16384},
		{384 * 16384, 512 * 16384},
		{math.MaxInt64, math.MaxInt64},
	}
	for _, d := range data {
		if RoundUpToPowerOfTwo(d[0]) != d[1] {
			t.Fail()
		}
	}

}
