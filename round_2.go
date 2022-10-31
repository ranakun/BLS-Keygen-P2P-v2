package main

// package keygen

import (
	"encoding/hex"
	"fmt"

	"github.com/libp2p/go-libp2p-core/protocol"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/util/random"
	"main.go/bls"
	"main.go/zkp"
)

func round2_start(peer_list []string, protocolID protocol.ID) {
	// BLS Phase
	status_struct.Phase = 2
	suite := bn256.NewSuite()
	BSK_i, BPK_i := bls.KeyGen(suite, random.New())
	fmt.Println("\n[+] BLS Setup Done")
	mar, _ := BPK_i.MarshalBinary()
	dst2 := make([]byte, hex.EncodedLen(len(mar)))
	hex.Encode(dst2, mar[:])
	fmt.Println("\nBLS PUBLIC KEY:", string(dst2))
	fmt.Println("BLS PRIVATE KEY:", BSK_i)
	send_data(peer_list, string(dst2), "bpk_j", protocolID)
	wait_until(2)

	//Send kgc_i values
	status_struct.Phase = 3

	msgs := ReadPeerInfoFromFile()
	// fmt.Println(msgs)
	BPK_j := make(map[string]string)
	for i, j := range msgs {
		BPK_j[i] = j[1]
	}

	round2_data.BPK_j = BPK_j

	SecretPolynomial := share.NewPriPoly(suite.G2(), T, BSK_i, suite.RandomStream())
	shares := SecretPolynomial.Shares(N)
	round2_data.shares = shares

	// Verification Set Generation
	coefs := SecretPolynomial.Coefficients()
	verification_set := []kyber.Point{}
	for _, coef := range coefs {
		tp := suite.G2().Point().Base()
		verification_set = append(verification_set, tp.Mul(coef, nil)) // add exp
	}

	verificationSetString := []string{}
	for _, v := range verification_set {
		mar, _ = v.MarshalBinary()
		verificationSetString = append(verificationSetString, hex.EncodeToString(mar))
	}
	vssArray := fmt.Sprint(verificationSetString)
	send_data(peer_list, vssArray[1:len(vssArray)-1], "vset", protocolID)

	kgd, kgc, SpubKey := zkp.SetupBLS(BSK_i)

	send_data(peer_list, hex.EncodeToString(kgc), "kgc_j", protocolID)
	wait_until(3)

	status_struct.Phase = 4
	send_data(peer_list, hex.EncodeToString(SpubKey), "sPubKey_j", protocolID)
	wait_until(4)

	status_struct.Phase = 5
	send_data(peer_list, hex.EncodeToString(kgd), "kgd_j", protocolID)
	wait_until(5)

}
