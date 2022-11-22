// Implementation of the BLS short signatues scheme
// Functions in modules are derivatives of :
// https://github.com/dedis/kyber/tree/master/sign/bls
// https://github.com/dedis/kyber/tree/master/sign/bdn
// supports sign, verification and aggregation

package bls

import (
	"crypto/cipher"
	"errors"
	"fmt"

	"math/big"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/group/mod"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/sign"
	"golang.org/x/crypto/blake2s"
)

type hashablePoint interface {
	Hash([]byte) kyber.Point
}

// NewKeyPair creates a new BLS signing key pair. The private key x is a scalar
// and the public key X is a point on curve G2.
func KeyGen(suite pairing.Suite, random cipher.Stream) (kyber.Scalar, kyber.Point) {
	x := suite.G2().Scalar().Pick(random)
	X := suite.G2().Point().Mul(x, nil)

	// ensures x is not null and X is not the base point
	for x.Equal(suite.G2().Scalar().Zero()) || X.Equal(suite.G2().Point().Base()) {
		x = suite.G2().Scalar().Pick(random)
		X = suite.G2().Point().Mul(x, nil)
	}

	fmt.Println("\n[+] KeyGen Success: BLS")
	return x, X
}

// Sign creates a BLS signature S = x * H(m) on a message m using the private
// key x. The signature S is a point on curve G1.
// func Sign(suite pairing.Suite, x kyber.Scalar, msg []byte) ([]byte, error) {
func Sign(suite pairing.Suite, x kyber.Scalar, msg []byte) (kyber.Point, error) {
	// checks if private key is not null
	if x.Equal(suite.G2().Scalar().Zero()) {
		return nil, errors.New("ERR: Private Key NULL error")
	}

	if len(msg) == 0 {
		return nil, errors.New("ERR: NULL message error")
	}

	hashable, ok := suite.G1().Point().(hashablePoint)
	if !ok {
		return nil, errors.New("ERR: point needs to implement hashablePoint")
	}
	HM := hashable.Hash(msg)
	xHM := HM.Mul(x, HM)

	// s, err := xHM.MarshalBinary()
	// if err != nil {
	// 	return nil, err
	// }
	return xHM, nil
}

// Verify checks the given BLS signature S on the message m using the public
// key X by verifying that the equality e(H(m), X) == e(H(m), x*B2) ==
// e(x*H(m), B2) == e(S, B2) holds where e is the pairing operation and B2 is
// the base point from curve G2.
func Verify(suite pairing.Suite, X kyber.Point, msg, sig []byte) error {
	if X.Equal(suite.G2().Point().Null()) {
		return errors.New("ERR: NULL Public Key")
	}

	if X.Equal(suite.G2().Point().Base()) {
		return errors.New("ERR: Invalid Public Key: Base Point G2.Base(); regenerate keys")
	}

	if len(msg) == 0 || len(sig) == 0 {
		return errors.New("ERR: NULL message or signature")
	}

	hashable, ok := suite.G1().Point().(hashablePoint)
	if !ok {
		return errors.New("ERR: bls point needs to implement hashablePoint")
	}
	HM := hashable.Hash(msg)
	left := suite.Pair(HM, X)
	s := suite.G1().Point()
	if err := s.UnmarshalBinary(sig); err != nil {
		return err
	}
	right := suite.Pair(s, suite.G2().Point().Base())
	if !left.Equal(right) {
		return errors.New("bls: invalid signature")
	}
	return nil
}

// AGGREGATION - BATCH VERIFICATION

// modulus128 can be provided to the big integer implementation to create numbers over 128 bits
var modulus128 = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1))

// For the choice of H, we're mostly worried about the second preimage attack. In
// other words, find m' where H(m) == H(m')
// We also use the entire roster so that the coefficient will vary for the same
// public key used in different roster
func hashPointToR(pubs []kyber.Point) ([]kyber.Scalar, error) {
	peers := make([][]byte, len(pubs))
	for i, pub := range pubs {
		peer, err := pub.MarshalBinary()
		if err != nil {
			return nil, err
		}

		peers[i] = peer
	}

	h, err := blake2s.NewXOF(blake2s.OutputLengthUnknown, nil)
	if err != nil {
		return nil, err
	}

	for _, peer := range peers {
		_, err := h.Write(peer)
		if err != nil {
			return nil, err
		}
	}

	out := make([]byte, 16*len(pubs))
	_, err = h.Read(out)
	if err != nil {
		return nil, err
	}

	coefs := make([]kyber.Scalar, len(pubs))
	for i := range coefs {
		coefs[i] = mod.NewIntBytes(out[i*16:(i+1)*16], modulus128, mod.LittleEndian)
	}

	return coefs, nil
}

// AggregateSignatures aggregates the signatures using a coefficient for each
// one of them where c = H(pk) and H: G2 -> R with R = {1, ..., 2^128}
func AggregateSignatures(suite pairing.Suite, sigs [][]byte, mask *sign.Mask) (kyber.Point, error) {
	if len(sigs) == 0 {
		return nil, errors.New("ERR: empty signature set")
	}
	if len(sigs) != mask.CountEnabled() {
		return nil, errors.New("ERR: length of signatures and public keys must match")
	}

	coefs, err := hashPointToR(mask.Publics())
	if err != nil {
		return nil, err
	}

	agg := suite.G1().Point()
	for i, buf := range sigs {
		peerIndex := mask.IndexOfNthEnabled(i)
		if peerIndex < 0 {
			// this should never happen as we check the lenths at the beginning
			// an error here is probably a bug in the mask
			return nil, errors.New("couldn't find the index")
		}

		sig := suite.G1().Point()
		err = sig.UnmarshalBinary(buf)
		if err != nil {
			return nil, err
		}
		// sigC = sig.clone().Mul(lambda2(t,i),sig)
		sigC := sig.Clone().Mul(coefs[peerIndex], sig)
		// c+1 because R is in the range [1, 2^128] and not [0, 2^128-1]
		sigC = sigC.Add(sigC, sig)
		agg = agg.Add(agg, sigC)
	}

	return agg, nil
}

// AggregatePublicKeys aggregates a set of public keys (similarly to
// AggregateSignatures for signatures) using the hash function
// H: G2 -> R with R = {1, ..., 2^128}.
func AggregatePublicKeys(suite pairing.Suite, mask *sign.Mask) (kyber.Point, error) {
	if mask.CountEnabled() == 0 {
		return nil, errors.New("ERR: invalid mask, check public key")
	}
	coefs, err := hashPointToR(mask.Publics())
	if err != nil {
		return nil, err
	}

	agg := suite.G2().Point()
	for i := 0; i < mask.CountEnabled(); i++ {
		peerIndex := mask.IndexOfNthEnabled(i)
		if peerIndex < 0 {
			// this should never happen because of the loop boundary
			// an error here is probably a bug in the mask implementation
			return nil, errors.New("FATAL ERR: couldn't find the index")
		}

		pub := mask.Publics()[peerIndex]
		pubC := pub.Clone().Mul(coefs[peerIndex], pub)
		pubC = pubC.Add(pubC, pub)
		agg = agg.Add(agg, pubC)
	}

	return agg, nil
}
