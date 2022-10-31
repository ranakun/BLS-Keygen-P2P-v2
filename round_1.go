package main

// package keygen

import (
	"encoding/hex"
	"fmt"

	"main.go/elgamal"

	"github.com/libp2p/go-libp2p-core/protocol"
)

func round1_start(peer_list []string, protocolID protocol.ID) {

	// Elgamal Phase
	status_struct.Phase = 1
	curve := elgamal.Setup() // Choosen curve : ED25519
	ESK_i, EPK_i := elgamal.KeyGen(curve)
	fmt.Println("\nElgmal Private:" + string(ESK_i.BigInt().String()))
	mar := EPK_i.ToAffineCompressed()
	send_data(peer_list, hex.EncodeToString(mar), "epk_j", protocolID)
	wait_until(1)

	// receive and store EPK_j
	msgs := ReadPeerInfoFromFile()
	EPK_j := make(map[string]string)
	for i, j := range msgs {
		EPK_j[i] = j[0]
	}

	round1_data.EPK_i = EPK_i
	round1_data.ESK_i = ESK_i
	round1_data.EPK_j = EPK_j
	round1_data.curve = curve

}
