package board

import (
	"strings"

	"github.com/jaypipes/ghw"
	"github.com/jaypipes/ghw/pkg/option"
	"k8s.io/klog/v2"
)

type BoardInfo struct {
	ChassisVendor string `json:"chassis"`
	ProductName   string `json:"product"`
	BIOSVendor    string `json:"bios"`
}

func Collect() (boardInfo *BoardInfo, err error) {
	bi := &BoardInfo{}
	if bi.ChassisVendor, err = getChassisVendor(); err != nil {
		return
	}
	if bi.ProductName, err = getProductName(); err != nil {
		return
	}
	if bi.BIOSVendor, err = getBIOSVendor(); err != nil {
		return
	}
	boardInfo = bi
	return
}

// 获取机器厂商
func getChassisVendor() (chassis string, err error) {
	chassises, err := ghw.Chassis(option.WithNullAlerter(), option.WithDisableTools())
	if err != nil {
		return
	}
	klog.V(4).InfoS("get chassis", "chassis", chassises.YAMLString())
	return strings.TrimSpace(chassises.Vendor), err
}

func getProductName() (product string, err error) {
	products, err := ghw.Product(option.WithNullAlerter(), option.WithDisableTools())
	if err != nil {
		return
	}
	klog.V(4).InfoS("get products", "products", products.YAMLString())
	return strings.TrimSpace(products.Name), err
}

func getBIOSVendor() (bios string, err error) {
	bioses, err := ghw.BIOS(option.WithNullAlerter(), option.WithDisableTools())
	if err != nil {
		return
	}
	klog.V(4).InfoS("get bios", "bios", bioses.YAMLString())
	return strings.TrimSpace(bioses.Vendor), err
}
