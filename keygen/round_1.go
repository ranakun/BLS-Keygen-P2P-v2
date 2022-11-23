package keygen

import (
	"encoding/hex"
	"fmt"

	"main.go/elgamal"

	"github.com/libp2p/go-libp2p-core/protocol"
	rounds_interface "main.go/interface"
)

func Round1_start(peer_list []string, protocolID protocol.ID) {

	// Elgamal Phase
	rounds_interface.Status_struct.Phase = 1
	curve := elgamal.Setup() // Choosen curve : ED25519
	ESK_i, EPK_i := elgamal.KeyGen(curve)
	fmt.Println("\nElgmal Private:" + string(ESK_i.BigInt().String()))
	mar := EPK_i.ToAffineCompressed()
	send_data(peer_list, hex.EncodeToString(mar), "epk_j", protocolID, "")
	wait_until(1)

	rounds_interface.Round1_data.EPK_i = EPK_i
	rounds_interface.Round1_data.ESK_i = ESK_i
	rounds_interface.Round1_data.EPK_j = ReadPeerInfoFromFile("EPK_j")
	rounds_interface.Round1_data.Curve = curve

}
