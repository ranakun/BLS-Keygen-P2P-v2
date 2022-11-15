package keygen

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/multiformats/go-multiaddr"
	"main.go/comm"
	rounds_interface "main.go/interface"
)

func ReadPeerInfoFromFile(name string) map[string]string {
	f, err := os.Open("peer_Data/" + strconv.Itoa(rounds_interface.My_index) + "/" + name + ".txt")
	if err != nil {
		log.Fatal(err)
	}

	var d = make(map[string]string)

	// read the file line by line using scanner
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		res := strings.Split(scanner.Text(), ">")
		d[res[0]] = res[1]
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	f.Close()
	return d
}

func Keygen(N int, T int) {

	peer_list := rounds_interface.Peer_details_list
	//current_flag = "1"
	var protocolID protocol.ID = "/keygen/0.0.1"
	//Start Listener
	Keygen_Stream_listener(rounds_interface.P2p.Host)
	//Start Acknowledger
	Host_acknowledge(rounds_interface.P2p.Host)

	os.MkdirAll("peer_Data/"+strconv.Itoa(rounds_interface.My_index), os.ModePerm)

	//Generate broadcast wait time
	time.Sleep(time.Second * 5)

	// status_struct.Phase = 0
	// send_data(peer_list, "ACK", "ACK", protocolID)
	// wait_until(0)

	rounds_interface.P2p.Round = 1
	Round1_start(peer_list, protocolID)
	rounds_interface.P2p.Round = 2
	Round2_start(peer_list, protocolID, N, T)
	rounds_interface.P2p.Round = 3
	Round3_start(peer_list, protocolID)
	rounds_interface.P2p.Round = 4
	Round4_start(peer_list, protocolID)

}

func send_data(peer_list []string, value string, name string, protocolID protocol.ID) {

	log.Println("Sending phase:", rounds_interface.Status_struct.Phase)
	for i, item := range peer_list {
		log.Println(item)
		if i == rounds_interface.My_index {
			continue
		}
		addr, _ := multiaddr.NewMultiaddr(item)
		peer_info, err := peer.AddrInfoFromP2pAddr(addr)

		if err != nil {
			panic(err)
		}

		// peer_num, _ := strconv.Atoi(item)
		// peer_num = peer_num + len(p2p.Peers)/2
		message_send := comm.Message{
			Phase: rounds_interface.Status_struct.Phase,
			Name:  name,
			Value: value,
			To:    rounds_interface.Peer_map[item],
		}

		s, err := rounds_interface.P2p.Host.NewStream(rounds_interface.P2p.Ctx, peer_info.ID, protocolID)
		if err != nil {
			log.Println(rounds_interface.Peer_map[item])
			log.Println(err, "Connecting to send message error")
			return
		}

		b_message, err := json.Marshal(message_send)
		if err != nil {
			log.Println(err, "Error in jsonifying data")
			return
		}
		_, err = s.Write(append(b_message, '\n'))

		if err != nil {
			log.Println(err, "Sending message erorr")
			return
		}

	}
}

func wait_until(phase int) {
	for {
		flag := 0
		for i, item := range rounds_interface.Peer_details_list {
			item = strings.Split(item, "/")[len(strings.Split(item, "/"))-1]
			if i == rounds_interface.My_index {
				continue
			}
			if phase > rounds_interface.Receive_peer_phase[item] {
				// if phase > receive_peer_phase[item] {
				// 	// Resend value to 'item'
				// }
				flag = 1
				// log.Println("heres why: ", receive_peer_phase[item])
			}
			if phase > rounds_interface.Sent_peer_phase[item] {
				flag = 1
				// log.Println("heres why: ", sent_peer_phase[item])
			}

		}
		if flag == 1 {

			time.Sleep(time.Microsecond * 5)
			// log.Println(flag, phase, receive_peer_phase, sent_peer_phase)
			flag = 0
			continue
		}
		fmt.Println("Returning from phase ", phase)
		// time.Sleep(time.Second)
		// log.Println(phase, receive_peer_phase, sent_peer_phase)
		return

	}
}
