package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"time"

	"main.go/bls"

	"github.com/libp2p/go-libp2p-core/protocol"
)

func round4_start(peer_list []string, protocolID protocol.ID) {

	suite := round2_data.suite
	BPK_j := round2_data.BPK_j
	BPK_i := round2_data.BPK_i
	shares := round2_data.shares
	fOfi := round3_data.fOfi
	// Generate Global Public Key
	global_public_key := BPK_i
	for _, p := range BPK_j {
		hdba, _ := hex.DecodeString(p) // string -> hex decode byte array
		up := suite.G2().Point()
		err123 := up.UnmarshalBinary(hdba) // hex decode byte array -> unmar point: UP
		if err123 != nil {
			fmt.Print("ERR ")
		}
		global_public_key = global_public_key.Add(global_public_key, up)
	}

	private_key_share := shares[my_index].V
	for _, j := range fOfi {
		mar, _ := hex.DecodeString(j)
		x := suite.G2().Scalar()
		x.UnmarshalBinary(mar)
		private_key_share.Add(private_key_share, x)
	}

	fmt.Println("\nPRIVATE KEY SHARE: ", private_key_share)
	f, _ := os.Create(strconv.Itoa(my_index) + "private_share.txt")
	f.WriteString(fmt.Sprint(private_key_share))

	mar, _ := global_public_key.MarshalBinary()
	fmt.Println("\n+++ GLOBAL PUBLIC KEY +++\n", hex.EncodeToString(mar))

	msg := []byte("test")
	sig, err := bls.Sign(suite, private_key_share, msg)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("\n[*] TESTING")
		fmt.Println("Message: ", string(msg))
		fmt.Println("Signature Share: ", hex.EncodeToString(sig))
	}
	time.Sleep(time.Second * 5)

	status_struct.Phase = 7
	send_data(peer_list, "FIN", "FIN", protocolID)
	wait_until(7)

	time.Sleep(time.Second * 5)
}
