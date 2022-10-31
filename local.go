package main

// package keygen

import (
	"fmt"
	"strings"
	"time"
)

func local() {
	fmt.Println("Enter all addresses seperated by ',' and no space: ")
	var inp_strings string
	fmt.Scan(&inp_strings)
	peer_details_list = strings.Split(inp_strings, ",")

	var random int
	keygen_Stream_listener(p2p.Host)
	//Start Acknowledger
	host_acknowledge(p2p.Host)
	fmt.Println("Enter int value to continue")
	fmt.Scan(&random)
	time.Sleep(time.Second * 5)

	test_conn()
	keygen()

}
