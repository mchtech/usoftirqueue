package board

import (
	"strings"

	"github.com/jaypipes/ghw"
	"github.com/jaypipes/ghw/pkg/option"
	"k8s.io/klog/v2"
)

type SystemInfo struct {
	ChassisVendor string `json:"chassis"`
	ProductName   string `json:"product"`
	BIOSVendor    string `json:"bios"`
}

func GetSystemInfo() (systemInfo *SystemInfo, err error) {
	si := &SystemInfo{}
	if si.ChassisVendor, err = getChassisVendor(); err != nil {
		return
	}
	if si.ProductName, err = getProductName(); err != nil {
		return
	}
	if si.BIOSVendor, err = getBIOSVendor(); err != nil {
		return
	}
	systemInfo = si
	return
}

// 获取机器厂商
func getChassisVendor() (chassis string, err error) {
	chassises, err := ghw.Chassis(option.WithNullAlerter(), option.WithDisableTools())
	if err != nil {
		return
	}
	klog.InfoS("get chassis", "chassis", chassises.YAMLString())
	return strings.TrimSpace(chassises.Vendor), err
}

func getProductName() (product string, err error) {
	products, err := ghw.Product(option.WithNullAlerter(), option.WithDisableTools())
	if err != nil {
		return
	}
	klog.InfoS("get products", "products", products.YAMLString())
	return strings.TrimSpace(products.Name), err
}

func getBIOSVendor() (bios string, err error) {
	bioses, err := ghw.BIOS(option.WithNullAlerter(), option.WithDisableTools())
	if err != nil {
		return
	}
	klog.InfoS("get bios", "bios", bioses.YAMLString())
	return strings.TrimSpace(bioses.Vendor), err
}
