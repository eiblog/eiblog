package main

import (
	"fmt"
	"os"

	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/cdn"
)

var (
	accessKey = os.Getenv("QINIU_ACCESS_KEY")
	secretKey = os.Getenv("QINIU_SECRET_KEY")
	domain    = os.Getenv("QINIU_TEST_DOMAIN")
)

func main() {
	mac := qbox.NewMac(accessKey, secretKey)
	cdnManager := cdn.NewCdnManager(mac)

	startDate := "2017-07-20"
	endDate := "2017-07-30"
	g := "day"
	data, err := cdnManager.GetBandwidthData(startDate, endDate, g, []string{domain})
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%v\n", data)
}
