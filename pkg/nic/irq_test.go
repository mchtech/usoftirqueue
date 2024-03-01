package nic

import "testing"

// test GetNICIRQ
func TestGetNICIRQ(t *testing.T) {
	irqs := getNICIRQ("eth0")
	t.Log(irqs)
	// t.Fail()
}

// func TestTuneNICIRQ(t *testing.T) {
// 	TuneNICIRQ("eth0", 1, true)
// }
