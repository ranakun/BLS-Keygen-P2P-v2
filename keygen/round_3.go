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

func Round3_start(peer_list []string, protocolID protocol.ID, N int) {

	KGC_j := ReadPeerInfoFromFile("KGC_j")
	KGD_j := ReadPeerInfoFromFile("KGD_j")
	SPUB_j := ReadPeerInfoFromFile("SPUB_j")

	for i, j := range KGC_j {
		kgd, _ := hex.DecodeString(KGD_j[i])
		kgc, _ := hex.DecodeString(j)
		spk, _ := hex.DecodeString(SPUB_j[i])

		if !zkp.DecommitmentBLS(kgd, kgc, spk) {
			fmt.Println("[-] ZKP Verication Failed")
			os.Exit(0)
		}
		if rounds_interface.Round2_data.BPK_j[i] != SPUB_j[i] {
			fmt.Println("[-] Verification Failed -- INVALID PUBLIC KEY SHARE")
			os.Exit(0)
		}
	}
	log.Println("[+] Commitment Verified")

	rounds_interface.Status_struct.Phase = 7
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
			// send_data(peer_list, hex.EncodeToString(C1.ToAffineCompressed()), "C1", protocolID)
			// send_data(peer_list, C2, "C2", protocolID)
			// send_data(peer_list, hex.EncodeToString(C3), "C3", protocolID)
			rounds_interface.Round3_data.C1 = C1
			rounds_interface.Round3_data.C2 = C2
			rounds_interface.Round3_data.C3 = C3
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
			detail := fmt.Sprint(rounds_interface.My_index) + ">" + fmt.Sprint(i)
			fmt.Println(rounds_interface.My_index, "*****Sending to*****", fmt.Sprint(i))
			send_data(peer_list, hex.EncodeToString(C1.ToAffineCompressed()), "C1", protocolID, detail)
			send_data(peer_list, C2, "C2", protocolID, detail)
			send_data(peer_list, hex.EncodeToString(C3), "C3", protocolID, detail)
		}
	}
	wait_until(7)
	time.Sleep(time.Second * 5)
	vss := ReadPeerInfoFromFile("Vset")
	fOfi := make(map[string]string)
	for i := 0; i < N; i++ {
		C1_j := curve.Point
		var C2_j string
		var C3_j []byte
		var verification_set_string_j []string
		epkj := curve.Point
		if i == rounds_interface.My_index {
			// C1_j = rounds_interface.Round3_data.C1
			// C2_j = rounds_interface.Round3_data.C2
			// C3_j = rounds_interface.Round3_data.C3
			// epkj = EPK_i
			// verification_set_string_j = rounds_interface.Round2_data.Vss
			continue
		} else {
			C1j := ReadShare("C1_j", fmt.Sprint(i))
			C2j := ReadShare("C2_j", fmt.Sprint(i))
			C3j := ReadShare("C3_j", fmt.Sprint(i))
			C1_j_Temp, err := hex.DecodeString(C1j)
			// C1_j_Temp, err := hex.DecodeString(j[0])
			if err != nil {
				fmt.Println("0", err, C1j)
				// fmt.Println("0", err, C1j[i][0])
			}
			C1_j, err = C1_j.FromAffineCompressed(C1_j_Temp)
			if err != nil {
				fmt.Println("1", err, C1_j_Temp, 1)
				// fmt.Println("1", err, C1_j_Temp, rounds_interface.My_index)
			}

			C2_j = C2j
			C3_j, err = hex.DecodeString(C3j)
			// C3_j, err := hex.DecodeString(C3j[i][0])
			if err != nil {
				fmt.Println("2", err)
			}
			epkj_temp, _ := hex.DecodeString(EPK_j[rounds_interface.Sorted_peer_id[i]])
			epkj, _ = epkj.FromAffineCompressed(epkj_temp)
			verification_set_string_j = strings.Split(vss[rounds_interface.Sorted_peer_id[i]], " ")
		}

		dec, err1 := elgamal.AuthDecryption(C1_j, C2_j, C3_j, epkj, EPK_i, ESK_i)
		if !err1 {
			fmt.Println("Decryption Error -- ", rounds_interface.My_index)
			fmt.Println(C1_j, C2_j, C3_j, epkj, EPK_i, ESK_i)
		}

		suite := rounds_interface.Round2_data.Suite
		mar, _ := hex.DecodeString(dec)
		fi := suite.G2().Scalar()
		fi.UnmarshalBinary(mar)

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

		fOfi[rounds_interface.Sorted_peer_id[i]] = dec
	}
	rounds_interface.Round3_data.FOfi = fOfi
}
