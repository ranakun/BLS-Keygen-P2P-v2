package main

import (
	"fmt"
	"strings"
	"time"

	rounds_interface "main.go/interface"
	"main.go/keygen"
)

func local() {
	fmt.Println("Enter N: ")
	var N int
	fmt.Scan(&N)
	fmt.Println("Enter T: ")
	var T int
	fmt.Scan(&T)
	fmt.Println("Enter all addresses seperated by ',' and no space: ")
	var inp_strings string
	fmt.Scan(&inp_strings)
	rounds_interface.Peer_details_list = strings.Split(inp_strings, ",")

	var random int
	keygen.Keygen_Stream_listener(rounds_interface.P2p.Host)
	//Start Acknowledger
	keygen.Host_acknowledge(rounds_interface.P2p.Host)
	fmt.Println("Enter int value to continue")
	fmt.Scan(&random)
	time.Sleep(time.Second * 5)

	test_conn()
	keygen.Keygen(N, T)

}
