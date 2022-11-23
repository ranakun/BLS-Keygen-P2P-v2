package keygen

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"time"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/util/encoding"
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
		fmt.Println("Signature Share: ", sig)
	}
	time.Sleep(time.Second * 5)

	////// verify signature
	sigg, _ := sig.MarshalBinary()
	errr := bls.Verify(suite, rounds_interface.Round2_data.BPK_i, msg, sigg)
	if errr != nil {
		fmt.Println(errr)
	}

	rounds_interface.Status_struct.Phase = 10
	fmt.Println("will participate in signing? Enter(Y/N)")
	var reply string
	fmt.Scan(&reply)
	wait_until(10)
	if reply == "N" {
		rounds_interface.Status_struct.Phase = 12
		send_data(peer_list, "Non Participant", "FIN", protocolID, "")
		wait_until(12)
		time.Sleep(time.Second * 5)
	}
	if reply == "Y" {
		rounds_interface.Status_struct.Phase = 11
		rounds_interface.T_array = append(rounds_interface.T_array, rounds_interface.My_index+1)
		lag := Lambda(int64(rounds_interface.My_index+1), rounds_interface.T_array)
		value := suite.Point()
		value = value.Mul(lag, sig)
		send_data(peer_list, value.String(), "LagXSIG", protocolID, "")
		wait_until(11)
		sum := value
		rounds_interface.Status_struct.Phase = 12
		msgs := ReadPeerInfoFromFile("LagXSIG_j")
		for j := range msgs {
			// convert string to kyber.point
			a, _ := encoding.StringHexToPoint(suite.G1(), j)
			sum = sum.Add(sum, a)
		}
		//// verify combined signature
		summ, _ := sum.MarshalBinary()
		e := bls.Verify(suite, global_public_key, msg, summ)
		if e != nil {
			fmt.Println(e)
		}

		send_data(peer_list, sum.String(), "FIN", protocolID, "")
		wait_until(12)
		time.Sleep(time.Second * 5)
	}
}

func Lambda(j int64, T_array []int) kyber.Scalar {
	var i int64
	curve := rounds_interface.Round2_data.Suite.G2()
	den := curve.Scalar().One()
	var LagCoeff = curve.Scalar().One()        //
	var J kyber.Scalar = curve.Scalar().Zero() //Converting j to kyber scalar from int64
	J.SetInt64(j)

	for i = 0; i < int64(len(T_array)); i++ {
		if int64(T_array[i]) == j {
			continue
		}

		var I kyber.Scalar = curve.Scalar().Zero()
		I.SetInt64(int64(T_array[i]))
		den.Sub(I, J)               //den=(i-j)
		den.Inv(den)                //1/(i-j)
		den.Mul(den, I)             //i/(i-j)
		LagCoeff.Mul(LagCoeff, den) // product (i/(i-j)) for each i from 1 to t such that i!=j
	}
	// fmt.Println(LagCoeff.String())
	return LagCoeff
}
