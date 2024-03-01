package nic

import (
	"github.com/safchain/ethtool"
)

func GetNICQueueChannel(nic string) (channels *ethtool.Channels, err error) {

	var (
		et *ethtool.Ethtool
		ec ethtool.Channels
	)

	if et, err = ethtool.NewEthtool(); err != nil {
		return
	}
	defer et.Close()

	if ec, err = et.GetChannels(nic); err != nil {
		return
	}
	return &ec, err
}

func SetNICQueueChannel(nic string, desired ethtool.Channels) (channels *ethtool.Channels, err error) {

	var (
		et *ethtool.Ethtool
		ec ethtool.Channels
	)

	if et, err = ethtool.NewEthtool(); err != nil {
		return
	}
	defer et.Close()

	if ec, err = et.SetChannels(nic, desired); err != nil {
		return
	}
	return &ec, err
}
