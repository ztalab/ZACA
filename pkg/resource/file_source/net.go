package file_source

import (
	"bytes"
	"fmt"
	"gitlab.oneitfarm.com/bifrost/capitalizone/pkg/resource"
	"os"
	"os/exec"
	"strings"
	"time"
)

var ethInterface string

var ErrDefaultEthInterfaceNotfound = fmt.Errorf("default EthInterface notfound")

func init() {
	// todo network
	ethInterface = "eth0"
	if val := os.Getenv("MSP_ETH_INTERFACE_NAME"); len(val) > 0 {
		ethInterface = val
	}
}

func (*FileSource) CurrentNetworkStat(interval time.Duration, callback resource.NetStatCallback) {
	var rxbytesOld, txbytesOld uint64
	var err error

	folder := "/sys/class/net/" + ethInterface + "/statistics/"
	rxbytesOld, err = resource.ReadNumberFromFile(folder + "rx_bytes")
	if err != nil {
		callback(nil, err)
		return
	}
	txbytesOld, err = resource.ReadNumberFromFile(folder + "tx_bytes")
	if err != nil {
		callback(nil, err)
		return
	}
	go func() {
		time.Sleep(interval)
		rxbytesNew, err := resource.ReadNumberFromFile(folder + "rx_bytes")
		if err != nil {
			callback(nil, err)
			return
		}
		txbytesNew, err := resource.ReadNumberFromFile(folder + "tx_bytes")
		if err != nil {
			callback(nil, err)
			return
		}
		stat := &resource.NetworkStat{
			RxBytes: rxbytesNew - rxbytesOld,
			TxBytes: txbytesNew - txbytesOld,
		}
		callback(stat, nil)
	}()
}

func init() {
	// $ ip -o -4 route show to default
	// default via 172.17.0.1 dev eth0
	cmd := exec.Command("ip", "-o", "-4", "route", "show", "to", "default")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		// fmt.Println("ip cmd err: " + err.Error())
		// fmt.Println("ip cmd err result: " + stderr.String())
		return
	}
	parts := strings.Split(strings.TrimSpace(out.String()), " ")
	if len(parts) < 5 {
		fmt.Println(fmt.Errorf("invalid result from \"ip -o -4 route show to default\": %s", out.String()))
		return
	}
	ethInterface = strings.TrimSpace(parts[4])
}
