package keygen

import (
	"bufio"
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
	global_public_key, _ := encoding.StringHexToPoint(suite.G2(), BPK_i)
	for _, p := range BPK_j {
		hdba, err123 := encoding.StringHexToPoint(suite.G2(), p) // string -> hex decode byte array
		if err123 != nil {
			fmt.Print(err123)
		}
		global_public_key = global_public_key.Add(global_public_key, hdba)
	}

	private_key_share := shares[rounds_interface.My_index].V
	// private_key_share := suite.G2().Scalar()
	for _, j := range fOfi {
		mar, _ := hex.DecodeString(j)
		x := suite.G2().Scalar()
		x.UnmarshalBinary(mar)
		private_key_share.Add(private_key_share, x)
	}

	fmt.Println("\nPRIVATE KEY SHARE: ", private_key_share)
	f, _ := os.Create(strconv.Itoa(rounds_interface.My_index) + "private_share.txt")
	f.WriteString(fmt.Sprint(private_key_share))

	t := verify_GPK()
	if !t.Equal(global_public_key) {
		fmt.Println("[+] GPK VERIFICATION FAILED")
	} else {
		fmt.Println("[+] GPK VERIFIED")
	}

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

	////// verify signature share
	public_key_share := suite.G2().Point().Mul(private_key_share, suite.G2().Point().Base())
	errr := bls.Verify(suite, public_key_share, msg, sig)
	if errr != nil {
		fmt.Println(errr)
		fmt.Println("[+] key share verification failed")
	}

	// fmt.Println("will participate in signing? Enter(Y/N)")
	// var reply string
	// fmt.Scan(&reply)
	reply := "Y"
	if reply == "N" {
		rounds_interface.Status_struct.Phase = 8
		send_data(peer_list, "Non Participant", "SIG", protocolID, "")
		wait_until(8)
	}

	if reply == "Y" {
		// sending signature share * Lagrange's value
		rounds_interface.Status_struct.Phase = 8
		tt := []int{1, 2}
		lag_1 := Lambda(1, tt)
		lag_2 := Lambda(2, tt)
		lag_2.Neg(lag_2)
		if rounds_interface.My_index == 1 {
			lag_2 = Lambda(1, tt)
			lag_1 = Lambda(2, tt)
		}
		sig_i := suite.G1().Point()
		err123 := sig_i.UnmarshalBinary(sig)
		if err123 != nil {
			fmt.Print("ERR ")
		}
		lagxSig := suite.G1().Point().Mul(lag_1, sig_i)
		abc, _ := lagxSig.MarshalBinary()
		send_data(peer_list, hex.EncodeToString(abc), "SIG", protocolID, "")
		wait_until(8)

		// combining signatures
		msgs := ReadPeerInfoFromFile("SIG_j")
		sum := lagxSig
		for _, j := range msgs {
			temp, _ := hex.DecodeString(j) // string -> hex decode byte array
			sigj := suite.G1().Point()
			err123 := sigj.UnmarshalBinary(temp) // hex decode byte array -> unmar point: UP
			if err123 != nil {
				fmt.Print("ERR ")
			}
			sum = sum.Add(sum, sigj)
		}
		/// verify combined signature
		summ, _ := sum.MarshalBinary()
		e := bls.Verify(suite, global_public_key, msg, summ)
		if e != nil {
			fmt.Println(e)
			fmt.Println("**Failure**")
		} else {
			fmt.Println("**Success**")
		}
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

	return LagCoeff
}

func verify_GPK() kyber.Point {
	suite := rounds_interface.Round2_data.Suite
	ps := suite.G2()
	sum := ps.Point()
	tt := []int{1, 2}
	fmt.Println("[+] VERIFYING GPK")
	for i := 0; i < 2; i++ {
		file, _ := os.Open(fmt.Sprint(i) + "private_share.txt")
		// temp, _ := encoding.ReadHexScalar(rounds_interface.Round2_data.Suite.G2(), file)
		scanner := bufio.NewScanner(file)
		var res string
		for scanner.Scan() {
			res = scanner.Text()
		}
		temp, _ := encoding.StringHexToScalar(ps, res)
		lag := Lambda(int64(i+1), tt)
		tempt := suite.G2().Point().Mul(temp, suite.G2().Point().Base())
		prod := ps.Point().Mul(lag, tempt)
		sum.Add(sum, prod)
		file.Close()
	}
	return sum
}
