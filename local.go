package main

import (
	"fmt"
	"strings"
	"time"

	rounds_interface "main.go/interface"
	"main.go/keygen"
)

func local() {
	fmt.Println("Enter number of peers: ")
	var N int
	fmt.Scan(&N)
	fmt.Println("Enter Threshold: ")
	var T int
	fmt.Scan(&T)
	fmt.Println("Enter all addresses seperated by ',' and no space: ")
	var inp_strings string
	fmt.Scan(&inp_strings)
	var d time.Duration = time.Duration(N * 5)
	time.Sleep(time.Second * d)
	rounds_interface.Peer_details_list = strings.Split(inp_strings, ",")

	var random int
	keygen.Keygen_Stream_listener(rounds_interface.P2p.Host)
	//Start Acknowledger
	keygen.Host_acknowledge(rounds_interface.P2p.Host)
	fmt.Println("Enter int value to continue")
	fmt.Scan(&random)
	d = time.Duration(T * 2)
	time.Sleep(time.Second * d)

	test_conn()
	keygen.Keygen(N, T)

}
