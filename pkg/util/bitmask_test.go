package util

import (
	"testing"
)

func TestBitMask(t *testing.T) {

	t.Logf("%d: %s", 12, BitMask(12))
	t.Logf("%d: %s", 96, BitMask(96))
	t.Logf("%d: %s", 104, BitMask(104))

	var cpu12, cpu96, cpu104 []int
	for i := 0; i < 12; i++ {
		cpu12 = append(cpu12, i)
	}
	for i := 0; i < 96; i++ {
		cpu96 = append(cpu96, i)
	}
	for i := 0; i < 104; i++ {
		cpu104 = append(cpu104, i)
	}
	t.Logf("%d: %s", 12, BitMask(12, cpu12...))
	t.Logf("%d: %s", 96, BitMask(96, cpu96...))
	t.Logf("%d: %s", 104, BitMask(104, cpu104...))

	queue := 372
	for i := 0; i < queue; i++ {
		t.Logf("%d: %s", i, BitMask(queue, i))
	}
	t.Fail()
}

func TestRoundUpToPowerOfTwo(t *testing.T) {
	if RoundUpToPowerOfTwo(16384) != 16384 {
		t.Fail()
	}
	if RoundUpToPowerOfTwo(512*16384) != 512*16384 {
		t.Fail()
	}
	if RoundUpToPowerOfTwo(12*16384) != 16*16384 {
		t.Fail()
	}
	if RoundUpToPowerOfTwo(104*16384) != 128*16384 {
		t.Fail()
	}
	if RoundUpToPowerOfTwo(384*16384) != 512*16384 {
		t.Fail()
	}
}
