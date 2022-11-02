package keygen

import (
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"main.go/elgamal"
	"main.go/zkp"

	"github.com/libp2p/go-libp2p-core/protocol"
	rounds_interface "main.go/interface"
)

func Round3_start(peer_list []string, protocolID protocol.ID) {

	msgs := ReadPeerInfoFromFile()

	for i, j := range msgs {
		kgd, _ := hex.DecodeString(j[4])
		kgc, _ := hex.DecodeString(j[2])
		spk, _ := hex.DecodeString(j[3])

		if !zkp.DecommitmentBLS(kgd, kgc, spk) {
			fmt.Println("[-] ZKP Verication Failed")
			os.Exit(0)
		}
		if rounds_interface.Round2_data.BPK_j[i] != j[3] {
			fmt.Println("[-] Verification Failed -- INVALID PUBLIC KEY SHARE")
			os.Exit(0)
		}
	}
	log.Println("[+] Commitment Verified")

	rounds_interface.Status_struct.Phase = 6
	// k := 0
	shares := rounds_interface.Round2_data.Shares
	curve := rounds_interface.Round1_data.Curve
	ESK_i := rounds_interface.Round1_data.ESK_i
	EPK_i := rounds_interface.Round1_data.EPK_i
	EPK_j := rounds_interface.Round1_data.EPK_j
	for i, share := range shares {
		if i == rounds_interface.My_index {
			sh, _ := share.V.MarshalBinary()
			shareStr := hex.EncodeToString(sh)
			C1, C2, C3 := elgamal.AuthEncryption(curve, shareStr, ESK_i, EPK_i, EPK_i)
			send_data(peer_list, hex.EncodeToString(C1.ToAffineCompressed()), "C1", protocolID)
			send_data(peer_list, C2, "C2", protocolID)
			send_data(peer_list, hex.EncodeToString(C3), "C3", protocolID)
		} else {
			sh, _ := share.V.MarshalBinary()
			shareStr := hex.EncodeToString(sh)
			temp, err := hex.DecodeString(EPK_j[rounds_interface.Sorted_peer_id[i]])
			if err != nil {
				fmt.Println(err)
			}
			epk_j := curve.Point
			epk_j, err = epk_j.FromAffineCompressed(temp)
			if err != nil {
				fmt.Println(err, temp, rounds_interface.Sorted_peer_id)
			}
			C1, C2, C3 := elgamal.AuthEncryption(curve, shareStr, ESK_i, EPK_i, epk_j)
			send_data(peer_list, hex.EncodeToString(C1.ToAffineCompressed()), "C1", protocolID)
			send_data(peer_list, C2, "C2", protocolID)
			send_data(peer_list, hex.EncodeToString(C3), "C3", protocolID)
		}
	}
	wait_until(6)
	time.Sleep(time.Second * 5)
	msgs = ReadPeerInfoFromFile()
	fOfi := make(map[string]string)
	for i, j := range msgs {
		C1_j := curve.Point
		C1_j_Temp, err := hex.DecodeString(j[6+(rounds_interface.My_index*3)])
		if err != nil {
			fmt.Println("0", err, j[6+(rounds_interface.My_index*3)])
		}
		C1_j, err = C1_j.FromAffineCompressed(C1_j_Temp)
		if err != nil {
			fmt.Println("1", err, C1_j_Temp, rounds_interface.My_index)
		}

		C2_j := j[7+(rounds_interface.My_index*3)]
		C3_j, err := hex.DecodeString(j[8+(rounds_interface.My_index*3)])
		if err != nil {
			fmt.Println("2", err)
		}

		epkj := curve.Point
		epkj_temp, _ := hex.DecodeString(EPK_j[i])
		epkj, _ = epkj.FromAffineCompressed(epkj_temp)
		dec, err1 := elgamal.AuthDecryption(C1_j, C2_j, C3_j, epkj, EPK_i, ESK_i)
		if !err1 {
			fmt.Println("Decryption Error -- ", rounds_interface.My_index)
			fmt.Println(C1_j, C2_j, C3_j, epkj, EPK_i, ESK_i)
		}

		suite := rounds_interface.Round2_data.Suite
		mar, _ := hex.DecodeString(dec)
		fi := suite.G2().Scalar()
		fi.UnmarshalBinary(mar)

		verification_set_string_j := strings.Split(j[5], " ")
		lhs := suite.G2().Point().Null()
		for ix, jx := range verification_set_string_j {
			tp := suite.G2().Point().Null()
			X := math.Pow(float64(rounds_interface.My_index+1), float64(ix))
			x := suite.G2().Scalar().SetInt64(int64(X))

			v := suite.G2().Point()
			tmp, _ := hex.DecodeString(jx)
			v.UnmarshalBinary(tmp)
			lhs = lhs.Add(lhs, tp.Mul(x, v))
		}

		rhs := suite.G2().Point().Base()
		rhs = rhs.Mul(fi, rhs)

		if !lhs.Equal(rhs) {
			fmt.Println("Verification Failed")
			os.Exit(0)
		}

		fOfi[i] = dec
	}
}
