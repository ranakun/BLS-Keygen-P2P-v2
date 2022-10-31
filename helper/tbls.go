package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/sign/bls"
)

// SigShare encodes a threshold BLS signature share Si = i || v where the 2-byte
// big-endian value i corresponds to the share's index and v represents the
// share's value. The signature share Si is a point on curve G1.
type SigShare []byte

// Index returns the index i of the TBLS share Si.
func (s SigShare) Index() (int, error) {
	var index uint16
	buf := bytes.NewReader(s)
	err := binary.Read(buf, binary.BigEndian, &index)
	if err != nil {
		return -1, err
	}
	return int(index), nil
}

// Value returns the value v of the TBLS share Si.
func (s *SigShare) Value() []byte {
	return []byte(*s)[2:]
}

// Sign creates a threshold BLS signature Si = xi * H(m) on the given message m
// using the provided secret key share xi.
func Sign(suite pairing.Suite, private *share.PriShare, msg []byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, uint16(private.I)); err != nil {
		return nil, err
	}
	s, err := bls.Sign(suite, private.V, msg)
	if err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, s); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Verify checks the given threshold BLS signature Si on the message m using
// the public key share Xi that is associated to the secret key share xi. This
// public key share Xi can be computed by evaluating the public sharing
// polynonmial at the share's index i.
func Verify(suite pairing.Suite, public *share.PubPoly, msg, sig []byte) error {
	s := SigShare(sig)
	i, err := s.Index()
	if err != nil {
		return err
	}
	return bls.Verify(suite, public.Eval(i).V, msg, s.Value())
}

// Recover reconstructs the full BLS signature S = x * H(m) from a threshold t
// of signature shares Si using Lagrange interpolation. The full signature S
// can be verified through the regular BLS verification routine using the
// shared public key X. The shared public key can be computed by evaluating the
// public sharing polynomial at index 0.
func Recover(suite pairing.Suite, msg []byte, sigs [][]byte, t, n int) ([]byte, error) {
	pubShares := make([]*share.PubShare, 0)
	for _, sig := range sigs {
		s := SigShare(sig)
		i, err := s.Index()
		if err != nil {
			return nil, err
		}
		point := suite.G1().Point()
		if err := point.UnmarshalBinary(s.Value()); err != nil {
			return nil, err
		}
		pubShares = append(pubShares, &share.PubShare{I: i, V: point})
		if len(pubShares) >= t {
			break
		}
	}
	commit, err := share.RecoverCommit(suite.G1(), pubShares, t, n)
	if err != nil {
		return nil, err
	}
	sig, err := commit.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return sig, nil
}

// func computeLagrange(idx int, N int) int {
// 	lagCoef := 1
// 	for i := 1; i <= N; i++ {
// 		if i != idx {
// 			lagCoef *= i
// 		}
// 	}
// 	return lagCoef
// }

// AggregateSignatures combines signatures created using the Sign function
func AggregateSignatures(suite pairing.Suite, sigs [][]byte) ([]byte, error) {
	sig := suite.G1().Point()
	l := []int64{3, -3, 1}
	for i, sigBytes := range sigs {
		sigToAdd := suite.G1().Point()
		if err := sigToAdd.UnmarshalBinary(sigBytes); err != nil {
			fmt.Println("Test")
			return nil, err
		}
		lag := suite.G2().Scalar()
		lag = lag.SetInt64(l[i])
		sigToAdd.Mul(lag, sigToAdd)
		sig.Add(sig, sigToAdd)
	}
	return sig.MarshalBinary()
}

func FullRun() {
	msg := []byte("test")
	suite := bn256.NewSuite()
	n := 3
	t := 2
	sk1, pk1 := bls.NewKeyPair(suite, suite.RandomStream())
	sk2, pk2 := bls.NewKeyPair(suite, suite.RandomStream())
	sk3, pk3 := bls.NewKeyPair(suite, suite.RandomStream())

	pk1 = pk1.Add(pk1, pk2)
	pk1 = pk1.Add(pk1, pk3)

	priPoly1 := share.NewPriPoly(suite.G2(), t, sk1, suite.RandomStream())
	PublicPolynomial1 := priPoly1.Commit(suite.G2().Point().Base())
	shares1 := priPoly1.Shares(n)

	priPoly2 := share.NewPriPoly(suite.G2(), t, sk2, suite.RandomStream())
	PublicPolynomial2 := priPoly2.Commit(suite.G2().Point().Base())
	shares2 := priPoly2.Shares(n)

	priPoly3 := share.NewPriPoly(suite.G2(), t, sk3, suite.RandomStream())
	PublicPolynomial3 := priPoly3.Commit(suite.G2().Point().Base())
	shares3 := priPoly3.Shares(n)

	privateShare1S := shares1[0].V.Add(shares1[0].V, shares2[0].V)
	privateShare1S = privateShare1S.Add(privateShare1S, shares3[0].V)
	lag1 := suite.G2().Scalar()
	lag1 = lag1.SetInt64(3)
	privateShare1S.Mul(privateShare1S, lag1)

	privateShare2S := shares1[1].V.Add(shares1[1].V, shares2[1].V)
	privateShare2S = privateShare2S.Add(privateShare2S, shares3[1].V)
	lag2 := suite.G2().Scalar()
	lag2 = lag2.SetInt64(3) // i / i -
	privateShare1S.Mul(privateShare1S, lag2)

	privateShare3S := shares1[2].V.Add(shares1[2].V, shares2[2].V)
	privateShare3S = privateShare3S.Add(privateShare3S, shares3[2].V)
	lag3 := suite.G2().Scalar()
	lag3 = lag3.SetInt64(3)
	privateShare1S.Mul(privateShare1S, lag3)

	privateShare1 := &share.PriShare{I: 1, V: privateShare1S}
	privateShare2 := &share.PriShare{I: 2, V: privateShare2S}
	privateShare3 := &share.PriShare{I: 3, V: privateShare3S}

	SecretPolynomialT, _ := priPoly1.Add(priPoly2)
	SecretPolynomialT, _ = SecretPolynomialT.Add(priPoly2)

	privateShares := []*share.PriShare{privateShare1, privateShare2, privateShare3}
	SecretPolynomial, err := share.RecoverPriPoly(suite.G2(), privateShares, t, n)
	if err != nil {
		print("FAIL")
	}

	fmt.Println(SecretPolynomialT.Equal(SecretPolynomial))

	PublicPolynomial := SecretPolynomial.Commit(suite.G2().Point().Base())
	PublicPolynomialTT := SecretPolynomialT.Commit(suite.G2().Point().Base())

	PublicPolynomialT, _ := PublicPolynomial1.Add(PublicPolynomial2)
	PublicPolynomialT, _ = PublicPolynomialT.Add(PublicPolynomial3)

	fmt.Println(PublicPolynomial.Equal(PublicPolynomialT))
	fmt.Println(PublicPolynomialTT.Equal(PublicPolynomialT))

	sigShares := make([][]byte, 0)
	for _, x := range SecretPolynomial.Shares(n) {
		sig, _ := Sign(suite, x, msg)
		sigShares = append(sigShares, sig)
	}

	sig, err1 := Recover(suite, msg, sigShares, t, n)
	if err1 != nil {
		fmt.Println(err1)
	}
	err2 := bls.Verify(suite, pk1, msg, sig)
	if err2 != nil {
		fmt.Println(err2)
	} else {
		fmt.Println("SUCCESS")
	}

	// pv := PublicPolynomial.Eval(1)
	// fmt.Println(pv.V.MarshalBinary())
	// ps := suite.G2().Point().Mul(privateShare1.V, suite.G2().Point().Base())
	// fmt.Println(pv.V.Equal(ps))
}

func FullRun2() {
	msg := []byte("test")
	suite := bn256.NewSuite()
	n := 3
	t := 2
	sk1, pk1 := bls.NewKeyPair(suite, suite.RandomStream())
	sk2, pk2 := bls.NewKeyPair(suite, suite.RandomStream())
	sk3, pk3 := bls.NewKeyPair(suite, suite.RandomStream())

	pk1 = pk1.Add(pk1, pk2)
	pk1 = pk1.Add(pk1, pk3)

	priPoly1 := share.NewPriPoly(suite.G2(), t, sk1, suite.RandomStream())
	// PublicPolynomial1 := priPoly1.Commit(suite.G2().Point().Base())
	shares1 := priPoly1.Shares(n)

	priPoly2 := share.NewPriPoly(suite.G2(), t, sk2, suite.RandomStream())
	// PublicPolynomial2 := priPoly2.Commit(suite.G2().Point().Base())
	shares2 := priPoly2.Shares(n)

	priPoly3 := share.NewPriPoly(suite.G2(), t, sk3, suite.RandomStream())
	// PublicPolynomial3 := priPoly3.Commit(suite.G2().Point().Base())
	shares3 := priPoly3.Shares(n)

	privateShare1S := shares1[0].V.Add(shares1[0].V, shares2[0].V)
	privateShare1S = privateShare1S.Add(privateShare1S, shares3[0].V)
	// lag1 := suite.G2().Scalar()
	// lag1 = lag1.SetInt64(3)
	// privateShare1S = privateShare1S.Mul(privateShare1S, lag1)

	privateShare2S := shares1[1].V.Add(shares1[1].V, shares2[1].V)
	privateShare2S = privateShare2S.Add(privateShare2S, shares3[1].V)
	// lag2 := suite.G2().Scalar()
	// lag2 = lag2.SetInt64(-3)
	// privateShare2S = privateShare2S.Mul(privateShare2S, lag2)

	privateShare3S := shares1[2].V.Add(shares1[2].V, shares2[2].V)
	privateShare3S = privateShare3S.Add(privateShare3S, shares3[2].V)
	// lag3 := suite.G2().Scalar()
	// lag3 = lag3.SetInt64(1)
	// privateShare3S = privateShare3S.Mul(privateShare3S, lag3)

	privateShare1 := &share.PriShare{I: 1, V: privateShare1S}
	privateShare2 := &share.PriShare{I: 2, V: privateShare2S}
	privateShare3 := &share.PriShare{I: 3, V: privateShare3S}

	privateShares := []*share.PriShare{privateShare1, privateShare2, privateShare3}

	sigShares := make([][]byte, 0)
	for _, x := range privateShares {
		sig, _ := bls.Sign(suite, x.V, msg)
		sigShares = append(sigShares, sig)
	}

	sig, err1 := AggregateSignatures(suite, sigShares)
	if err1 != nil {
		fmt.Println(err1)
	}

	err2 := bls.Verify(suite, pk1, msg, sig)
	if err2 != nil {
		fmt.Println(err2)
	} else {
		fmt.Println("SUCCESS")
	}
}

func Test() {
	msg := []byte("test")
	suite := bn256.NewSuite()
	n := 3
	t := 2

	privateShare1S, _ := hex.DecodeString("0a7b3d16820429f61da1674d641d6e60e99f845b25c663ad5bbb4d96327b90cf")
	privateShare2S, _ := hex.DecodeString("3f224dd13a55adf8ebeb3587e8e65a944895dcedbff0a398aa05eae8740fe5e7")
	privateShare3S, _ := hex.DecodeString("73c95e8bf2a731fbba3503c26daf46c7a78c35805a1ae383f850883ab5a43aff")

	privateShare1P := suite.G2().Scalar()
	privateShare1P.SetBytes(privateShare1S)
	privateShare2P := suite.G2().Scalar()
	privateShare2P.SetBytes(privateShare2S)
	privateShare3P := suite.G2().Scalar()
	privateShare3P.SetBytes(privateShare3S)

	privateShare1 := &share.PriShare{I: 1, V: privateShare1P}
	privateShare2 := &share.PriShare{I: 2, V: privateShare2P}
	privateShare3 := &share.PriShare{I: 3, V: privateShare3P}

	privateShares := []*share.PriShare{privateShare1, privateShare2, privateShare3}
	SecretPolynomial, err := share.RecoverPriPoly(suite.G2(), privateShares, t, n)
	if err != nil {
		print("FAIL")
	}

	PublicPolynomial := SecretPolynomial.Commit(suite.G2().Point().Base())

	Sig1S, _ := hex.DecodeString("0001262f8d1f0c59ae53b63802938470cfc6fb2b51c361cf14b10b0b0969a64e1a3279592fa5a9a52baec70d18bfff994393d20df991940971760ee66e8a64ff458a")
	Sig2S, _ := hex.DecodeString("000239f1e44bcd1765e1432d78c9c25a006933a82bcdeb86bea5e19da6dfa8bf141318d7f0b32d42abe94fe95f87f6bed0423d0c84de847f74301889c48cf40d8c8c")
	Sig3S, _ := hex.DecodeString("000379fb5b49bec7fe3789a0c408ae4deda63ba283b8d1eb7aa83bfd94a2e611263c3f0d23d98ea34442ad3de043a05f3041e1abf9f31e23e2ac5dfce598e5ad61bf")
	sigShares := [][]byte{Sig1S, Sig2S, Sig3S}

	sig, err1 := Recover(suite, msg, sigShares, t, n)
	if err1 != nil {
		fmt.Println(err1)
	}
	err2 := bls.Verify(suite, PublicPolynomial.Commit(), msg, sig)
	if err2 != nil {
		fmt.Println(err2)
	} else {
		fmt.Println("SUCCESS")
	}
}

func Verification() {
	msg := []byte("test")
	suite := bn256.NewSuite()
	// n := 3
	// t := 2

	PublicKeyString, _ := hex.DecodeString("6b98569c089a0d3a0fba31cdf803a6b789398187bfa3e56dc06f0b13a16ab21c080d15f7aceb18fd3326f24947f7466cc7dd6c333abc0fb96b285af12dad25545765b691f53ef44350831842e4dee59741993a509c631744a642b6e64eac5b683749724ab30ca59f2527914116a1ab5490efe6baad400c22b3da6cf5045396ea")
	PublicKey := suite.G2().Point()
	PublicKey.UnmarshalBinary(PublicKeyString)

	Sig1S, _ := hex.DecodeString("736ce71c68b02fe085f963d0b7c7bfbdbdfac0fc6fcb9fa151b90f0e2e83ca4d38efa26b09ae34d707f67f83ccb308d47f1cf919a9e24ba6f515d7e813eae1af")
	Sig2S, _ := hex.DecodeString("62180066f42749134dd6c5b58f934a3ead5077136c282a1b86bc4d6a92032eab09971562ea23bfc8337912f030251510a89565ae9f37018024ea5b4ecd330254")
	Sig3S, _ := hex.DecodeString("8af8681007e9bd8cf3bbc37c51d6aaafbe20990d59c7f27defa51fbb21af7d981fee73ea6383a299df54ea42052822aa92dff65dafb933a192e2455c13052574")
	sigShares := [][]byte{Sig1S, Sig2S, Sig3S}

	sig, err1 := AggregateSignatures(suite, sigShares)
	if err1 != nil {
		fmt.Println(err1)
	}
	err2 := bls.Verify(suite, PublicKey, msg, sig)
	if err2 != nil {
		fmt.Println(err2)
	} else {
		fmt.Println("SUCCESS")
	}
}

func main() {
	// FullRun2()
	Verification()
}

/* TEST PRIVATE SHARES
51ca5ff93e0e7fc568798cfb98dbaf53754133ec5bfd47856348d89cbd134cb2
71601d8bbeb5da6136c51df1c66ed5c2d7fb73dcbe528061b7aebf709a8e215d
0140d93af4b9ad035aa0c22f927d20110c2825ba287c8019f1e5b1e9205c83a7
*/

/* TEST SIGNS
000127d0a5074d918045816b57590ed2ff11362499700958112c7df85591245cbe8c6b5b9c7e98f5d57a15a6e59189ee042cc7f7d0872606673ebabfce886bdf071b
000279e402e2261311874930dc059299083092e09e7760b8ccce7e83c1908a155f8f3608ec39aea581aab22de3553cb2ac4c32357a393b15f7669d24adb3b1833ddc
0003712cdccc0215a51e6182c8efb89026ae7f54913ce9a4c3f36633e5b8c58f238720bdde30b2290eb9fe31b7056d362d17cd212f236e579ad4fac8c4c8384d2412
*/
