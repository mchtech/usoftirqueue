package tuner

import "testing"

func TestCPUMask(t *testing.T) {
	for i := 1; i < 257; i++ {
		s := getCPUBitmap(i)
		t.Logf("%d: %s", i, s)
	}
	// t.Fail()
}
