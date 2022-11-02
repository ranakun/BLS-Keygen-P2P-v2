package main

import (
	"log"
	"sort"
	"strings"

	peer "github.com/libp2p/go-libp2p-core/peer"
	rounds_interface "main.go/interface"
)

func test_conn() {
	rounds_interface.Peer_details_list = append(rounds_interface.Peer_details_list, rounds_interface.P2p.Host_ip)

	sort.Strings(rounds_interface.Peer_details_list)
	for i, item := range rounds_interface.Peer_details_list {
		rounds_interface.Sorted_peer_id = append(rounds_interface.Sorted_peer_id, strings.Split(item, "/")[len(strings.Split(item, "/"))-1])
		rounds_interface.Peer_index[strings.Split(item, "/")[len(strings.Split(item, "/"))-1]] = i
		if item == rounds_interface.P2p.Host_ip {
			rounds_interface.My_index = i
		}
	}

	for i, peer_ip := range rounds_interface.Peer_details_list {
		rounds_interface.Peer_map[strings.Split(peer_ip, "/")[len(strings.Split(peer_ip, "/"))-1]] = peer_ip
		// fmt.Println(len(sorted_peer_id))
		if i == rounds_interface.My_index {
			continue
		}
		connect_to, err := peer.AddrInfoFromString(peer_ip)
		if err != nil {
			log.Println(err)
		}
		if err := rounds_interface.P2p.Host.Connect(rounds_interface.P2p.Ctx, *connect_to); err != nil {
			log.Println("Connection failed:", peer_ip)
			rounds_interface.All_ok = false
			return
		} else {
			log.Println("Connected to: ", peer_ip)
		}
	}
}
