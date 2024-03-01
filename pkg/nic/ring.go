package nic

import (
	"github.com/safchain/ethtool"
)

func GetNICRingSize(nic string) (ring *ethtool.Ring, err error) {

	var (
		et *ethtool.Ethtool
		er ethtool.Ring
	)

	if et, err = ethtool.NewEthtool(); err != nil {
		return
	}
	defer et.Close()

	if er, err = et.GetRing(nic); err != nil {
		return
	}
	return &er, err
}

func SetNICRingSize(nic string, desired ethtool.Ring) (ring *ethtool.Ring, err error) {

	var (
		et *ethtool.Ethtool
		er ethtool.Ring
	)

	if et, err = ethtool.NewEthtool(); err != nil {
		return
	}
	defer et.Close()

	if er, err = et.SetRing(nic, desired); err != nil {
		return
	}
	return &er, err
}
