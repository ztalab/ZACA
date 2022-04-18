package main

import (
	"crypto/tls"
	"fmt"
)

func main() {
	//conn, err := net.Dial("tcp", "192.168.2.80:7000")
	//if true {
	//	conn, err = tls.Dial("tcp", "192.168.2.80:7443", &tls.Config{
	//		InsecureSkipVerify: true,
	//		ServerName: "tcp.test.com",
	//	})
	//}
	conn, err := tls.Dial("tcp", "dp-thompbxnvc67x-26656.gw105.oneitfarm.com:7443", &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		panic(fmt.Errorf("dial err: %v", err))
	}

	fmt.Println("send msg")
	defer conn.Close()

	n, err := conn.Write([]byte("11111111111\n"))
	if err != nil {
		fmt.Println(n, err)
		return
	}

	buf := make([]byte, 100)
	n, err = conn.Read(buf)
	if err != nil {
		fmt.Println(n, err)
		return
	}

	fmt.Println(string(buf[:n]))
}
