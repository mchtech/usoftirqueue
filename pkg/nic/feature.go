package nic

import (
	"github.com/safchain/ethtool"
)

func GetNICFeatures(nic string) (features map[string]bool, err error) {

	var (
		et *ethtool.Ethtool
	)

	if et, err = ethtool.NewEthtool(); err != nil {
		return
	}
	defer et.Close()

	return et.Features(nic)
}

func SetNICFeatures(nic string, key string, value bool) (err error) {

	var (
		et *ethtool.Ethtool
	)

	if et, err = ethtool.NewEthtool(); err != nil {
		return
	}
	defer et.Close()

	return et.Change(nic, map[string]bool{key: value})
}
