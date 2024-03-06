package util

import (
	"testing"
)

func TestQueueMask(t *testing.T) {
	t.Logf("%d: %s", 12, BitMap(12, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11))
	queue := 108
	for i := 0; i < queue; i++ {
		s := BitMap(queue, i)
		t.Logf("%d: %s", i, s)
	}
	// t.Fail()
}
