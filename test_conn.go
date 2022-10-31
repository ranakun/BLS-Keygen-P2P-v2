package main

import (
	"log"
	"sort"
	"strings"

	peer "github.com/libp2p/go-libp2p-core/peer"
)

func test_conn() {
	peer_details_list = append(peer_details_list, p2p.Host_ip)

	sort.Strings(peer_details_list)
	for i, item := range peer_details_list {
		sorted_peer_id = append(sorted_peer_id, strings.Split(item, "/")[len(strings.Split(item, "/"))-1])
		peer_index[strings.Split(item, "/")[len(strings.Split(item, "/"))-1]] = i
		if item == p2p.Host_ip {
			my_index = i
		}
	}

	for i, peer_ip := range peer_details_list {
		peer_map[strings.Split(peer_ip, "/")[len(strings.Split(peer_ip, "/"))-1]] = peer_ip
		// fmt.Println(len(sorted_peer_id))
		if i == my_index {
			continue
		}
		connect_to, err := peer.AddrInfoFromString(peer_ip)
		if err != nil {
			log.Println(err)
		}
		if err := p2p.Host.Connect(p2p.Ctx, *connect_to); err != nil {
			log.Println("Connection failed:", peer_ip)
			all_ok = false
			return
		} else {
			log.Println("Connected to: ", peer_ip)
		}
	}
}
