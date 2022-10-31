package main

// import (
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"strings"
// 	"time"

// 	"github.com/libp2p/go-libp2p-core/peer"
// 	"github.com/libp2p/go-libp2p-core/protocol"
// 	"github.com/multiformats/go-multiaddr"
// )

// //Rework to common message sending function
// func keygen() {

// 	peer_list := peer_details_list
// 	//current_flag = "1"
// 	// status_struct.Phase = 1
// 	var protocolID protocol.ID = "/keygen/0.0.1"
// 	//Start Listener
// 	keygen_Stream_listener(p2p.Host)
// 	//Start Acknowledger
// 	host_acknowledge(p2p.Host)

// 	//Generate broadcast wait time
// 	time.Sleep(time.Second * 5)

// 	status_struct.Phase = 1
// 	send_data(peer_list, "value 1", "ppk_j", protocolID)
// 	wait_until(1)

// 	status_struct.Phase = 2
// 	send_data(peer_list, "test kgc_i value", "kgc_j", protocolID)
// 	wait_until(2)
// 	// time.Sleep(2 * time.Second)

// 	//Send kgd_i values
// 	status_struct.Phase = 3
// 	send_data(peer_list, "test kgd_i value", "kgd_j", protocolID)
// 	wait_until(3)
// 	// time.Sleep(2 * time.Second)

// 	//?!?!VSS??!?!

// }

// func send_data(peer_list []string, value string, name string, protocolID protocol.ID) {

// 	log.Println("Sending phase:", status_struct.Phase)
// 	for i, item := range peer_list {
// 		log.Println(item)
// 		if i == my_index {
// 			continue
// 		}
// 		addr, _ := multiaddr.NewMultiaddr(item)
// 		peer_info, err := peer.AddrInfoFromP2pAddr(addr)

// 		if err != nil {
// 			panic(err)
// 		}

// 		// peer_num, _ := strconv.Atoi(item)
// 		// peer_num = peer_num + len(p2p.Peers)/2
// 		message_send := message{
// 			Phase: status_struct.Phase,
// 			Name:  name,
// 			Value: value,
// 			To:    peer_map[item],
// 		}

// 		s, err := p2p.Host.NewStream(p2p.Ctx, peer_info.ID, protocolID)
// 		if err != nil {
// 			log.Println(peer_map[item])
// 			log.Println(err, "Connecting to send message error")
// 			return
// 		}

// 		b_message, err := json.Marshal(message_send)
// 		if err != nil {
// 			log.Println(err, "Error in jsonifying data")
// 			return
// 		}
// 		_, err = s.Write(append(b_message, '\n'))

// 		if err != nil {
// 			log.Println(err, "Sending message erorr")
// 			return
// 		}

// 	}

// }

// func wait_until(phase int) {
// 	for {
// 		flag := 0
// 		for i, item := range peer_details_list {
// 			item = strings.Split(item, "/")[len(strings.Split(item, "/"))-1]
// 			if i == my_index {
// 				continue
// 			}
// 			if phase != receive_peer_phase[item] {
// 				// if phase > receive_peer_phase[item] {
// 				// 	// Resend value to 'item'
// 				// }
// 				flag = 1
// 				// log.Println("heres why: ", receive_peer_phase[item])
// 			}
// 			if phase != sent_peer_phase[item] {
// 				flag = 1
// 				// log.Println("heres why: ", sent_peer_phase[item])
// 			}

// 		}
// 		if flag == 1 {

// 			time.Sleep(time.Microsecond * 5)
// 			// log.Println(flag, phase, receive_peer_phase, sent_peer_phase)
// 			flag = 0
// 			continue
// 		}
// 		fmt.Println("Returning from phase ", phase)
// 		// time.Sleep(time.Second)
// 		// log.Println(phase, receive_peer_phase, sent_peer_phase)
// 		return

// 	}
// }
