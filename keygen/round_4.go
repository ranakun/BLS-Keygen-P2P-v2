package keygen

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"time"

	"main.go/bls"

	"github.com/libp2p/go-libp2p-core/protocol"
	rounds_interface "main.go/interface"
)

func Round4_start(peer_list []string, protocolID protocol.ID) {

	suite := rounds_interface.Round2_data.Suite
	BPK_j := rounds_interface.Round2_data.BPK_j
	BPK_i := rounds_interface.Round2_data.BPK_i
	shares := rounds_interface.Round2_data.Shares
	fOfi := rounds_interface.Round3_data.FOfi
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

	private_key_share := shares[rounds_interface.My_index].V
	for _, j := range fOfi {
		mar, _ := hex.DecodeString(j)
		x := suite.G2().Scalar()
		x.UnmarshalBinary(mar)
		private_key_share.Add(private_key_share, x)
	}

	fmt.Println("\nPRIVATE KEY SHARE: ", private_key_share)
	f, _ := os.Create(strconv.Itoa(rounds_interface.My_index) + "private_share.txt")
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

	rounds_interface.Status_struct.Phase = 7
	send_data(peer_list, "FIN", "FIN", protocolID)
	wait_until(7)

	time.Sleep(time.Second * 5)
}
