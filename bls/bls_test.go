package bls

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/sign"

	"go.dedis.ch/kyber/v3/sign/bls" // using depreciated version only for testing
	"go.dedis.ch/kyber/v3/util/random"
)

func TestBLS(t *testing.T) {
	msg := []byte("Hello Boneh-Lynn-Shacham")
	suite := bn256.NewSuite()
	private, public := KeyGen(suite, random.New())
	sig, err := Sign(suite, private, msg)
	require.Nil(t, err)
	err = Verify(suite, public, msg, sig)
	require.Nil(t, err)
}

func TestBLSFailSig(t *testing.T) {
	msg := []byte("Hello Boneh-Lynn-Shacham")
	suite := bn256.NewSuite()
	private, public := KeyGen(suite, random.New())
	sig, err := Sign(suite, private, msg)
	require.Nil(t, err)
	sig[0] ^= 0x01
	if Verify(suite, public, msg, sig) == nil {
		t.Fatal("FATAL: bls verification succeeded - unexpectedly; EXPECTED FAIL")
	}
}

func TestBLSFailKey(t *testing.T) {
	msg := []byte("Hello Boneh-Lynn-Shacham")
	suite := bn256.NewSuite()
	private, _ := KeyGen(suite, random.New())
	sig, err := Sign(suite, private, msg)
	require.Nil(t, err)
	_, public := KeyGen(suite, random.New())
	if Verify(suite, public, msg, sig) == nil {
		t.Fatal("FATAL: bls verification succeeded - unexpectedly; EXPECTED FAIL")
	}
}

func TestBLSNullSign(t *testing.T) {
	msg := []byte("Hello Boneh-Lynn-Shacham")
	suite := bn256.NewSuite()
	private := suite.G1().Scalar()
	private.Zero()
	_, err := Sign(suite, private, msg)
	// require.Nil(t, err)
	if err == nil {
		t.Fatal("FATAL: bls signing succeeded - unexpectedly; EXPECTED FAIL")
	}
}

func TestBLSNull(t *testing.T) {
	msg := []byte("Hello Boneh-Lynn-Shacham")
	suite := bn256.NewSuite()
	private, _ := KeyGen(suite, random.New())
	sig, err := Sign(suite, private, msg)
	require.Nil(t, err)
	public := suite.G2().Point() // generates a null point
	public.Null()
	if Verify(suite, public, msg, sig) == nil {
		t.Logf("private: %v", private)
		t.Logf("public: %v", public)
		t.Fatal("FATAL: bls verification succeeded - unexpectedly; EXPECTED FAIL")
	}
}

func TestBLSP(t *testing.T) {
	msg := []byte("Hello Boneh-Lynn-Shacham")
	suite := bn256.NewSuite()
	private, pub := KeyGen(suite, random.New())
	sig, err := Sign(suite, private, msg)
	require.Nil(t, err)
	public := suite.G2().Point().Base() // generates a null point
	public.Null()
	if Verify(suite, public, msg, sig) == nil {
		t.Fatal("FATAL: bls verification succeeded - unexpectedly; EXPECTED FAIL")
	}
	if Verify(suite, pub, msg, sig) != nil {
		t.Fatal("FATAL: bls verification failed - unexpectedly; EXPECTED PASS")
	}
}

func BenchmarkBLSKeyCreation(b *testing.B) {
	suite := bn256.NewSuite()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		KeyGen(suite, random.New())
	}
}

func BenchmarkBLSSign(b *testing.B) {
	suite := bn256.NewSuite()
	private, _ := KeyGen(suite, random.New())
	msg := []byte("Hello Boneh-Lynn-Shacham")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Sign(suite, private, msg)
	}
}

func BenchmarkBLSVerify(b *testing.B) {
	msg := []byte("Hello Boneh-Lynn-Shacham")
	suite := bn256.NewSuite()
	private, public := KeyGen(suite, random.New())
	sig, _ := Sign(suite, private, msg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Verify(suite, public, msg, sig)
	}
}

// TESTING AGGREGATION

// Reference test for other languages
func TestBDN_HashPointToR_BN256(t *testing.T) {
	suite := pairing.NewSuiteBn256()
	two := suite.Scalar().Add(suite.Scalar().One(), suite.Scalar().One())
	three := suite.Scalar().Add(two, suite.Scalar().One())

	p1 := suite.Point().Base()
	p2 := suite.Point().Mul(two, suite.Point().Base())
	p3 := suite.Point().Mul(three, suite.Point().Base())

	coefs, err := hashPointToR([]kyber.Point{p1, p2, p3})

	require.NoError(t, err)
	require.Equal(t, "35b5b395f58aba3b192fb7e1e5f2abd3", coefs[0].String())
	require.Equal(t, "14dcc79d46b09b93075266e47cd4b19e", coefs[1].String())
	require.Equal(t, "933f6013eb3f654f9489d6d45ad04eaf", coefs[2].String())
	require.Equal(t, 16, coefs[0].MarshalSize())

	mask, _ := sign.NewMask(suite, []kyber.Point{p1, p2, p3}, nil)
	mask.SetBit(0, true)
	mask.SetBit(1, true)
	mask.SetBit(2, true)

	agg, err := AggregatePublicKeys(suite, mask)
	require.NoError(t, err)

	buf, err := agg.MarshalBinary()
	require.NoError(t, err)
	ref := "1432ef60379c6549f7e0dbaf289cb45487c9d7da91fc20648f319a9fbebb23164abea76cdf7b1a3d20d539d9fe096b1d6fb3ee31bf1d426cd4a0d09d603b09f55f473fde972aa27aa991c249e890c1e4a678d470592dd09782d0fb3774834f0b2e20074a49870f039848a6b1aff95e1a1f8170163c77098e1f3530744d1826ce"
	require.Equal(t, ref, fmt.Sprintf("%x", buf))
}

func TestBDN_AggregateSignatures(t *testing.T) {
	msg := []byte("Hello Boneh-Lynn-Shacham")
	suite := bn256.NewSuite()
	private1, public1 := KeyGen(suite, random.New())
	private2, public2 := KeyGen(suite, random.New())
	sig1, err := Sign(suite, private1, msg)
	require.NoError(t, err)
	sig2, err := Sign(suite, private2, msg)
	require.NoError(t, err)

	mask, _ := sign.NewMask(suite, []kyber.Point{public1, public2}, nil)
	mask.SetBit(0, true)
	mask.SetBit(1, true)

	_, err = AggregateSignatures(suite, [][]byte{sig1}, mask)
	require.Error(t, err)

	aggregatedSig, err := AggregateSignatures(suite, [][]byte{sig1, sig2}, mask)
	require.NoError(t, err)

	aggregatedKey, err := AggregatePublicKeys(suite, mask)
	require.NoError(t, err)

	sig, err := aggregatedSig.MarshalBinary()
	require.NoError(t, err)

	err = Verify(suite, aggregatedKey, msg, sig)
	require.NoError(t, err)

	mask.SetBit(1, false)
	aggregatedKey, err = AggregatePublicKeys(suite, mask)
	require.NoError(t, err)

	err = Verify(suite, aggregatedKey, msg, sig)
	require.Error(t, err)
}

func TestBDN_SubsetSignature(t *testing.T) {
	msg := []byte("Hello Boneh-Lynn-Shacham")
	suite := bn256.NewSuite()
	private1, public1 := KeyGen(suite, random.New())
	private2, public2 := KeyGen(suite, random.New())
	_, public3 := KeyGen(suite, random.New())
	sig1, err := Sign(suite, private1, msg)
	require.NoError(t, err)
	sig2, err := Sign(suite, private2, msg)
	require.NoError(t, err)

	mask, _ := sign.NewMask(suite, []kyber.Point{public1, public3, public2}, nil)
	mask.SetBit(0, true)
	mask.SetBit(2, true)

	aggregatedSig, err := AggregateSignatures(suite, [][]byte{sig1, sig2}, mask)
	require.NoError(t, err)

	aggregatedKey, err := AggregatePublicKeys(suite, mask)
	require.NoError(t, err)

	sig, err := aggregatedSig.MarshalBinary()
	require.NoError(t, err)

	err = Verify(suite, aggregatedKey, msg, sig)
	require.NoError(t, err)
}

func TestBDN_RogueAttack(t *testing.T) {
	msg := []byte("Hello Boneh-Lynn-Shacham")
	suite := bn256.NewSuite()
	// honest
	_, public1 := KeyGen(suite, random.New())
	// attacker
	private2, public2 := KeyGen(suite, random.New())

	// create a forged public-key for public1
	rogue := public1.Clone().Sub(public2, public1)

	pubs := []kyber.Point{public1, rogue}

	sig, err := Sign(suite, private2, msg)
	require.NoError(t, err)

	//  Old scheme not resistant to the attack
	agg := bls.AggregatePublicKeys(suite, pubs...)
	require.NoError(t, bls.Verify(suite, agg, msg, sig))

	// New scheme that should detect
	mask, _ := sign.NewMask(suite, pubs, nil)
	mask.SetBit(0, true)
	mask.SetBit(1, true)
	agg, err = AggregatePublicKeys(suite, mask)
	require.NoError(t, err)
	require.Error(t, Verify(suite, agg, msg, sig))
}

func Benchmark_BDN_AggregateSigs(b *testing.B) {
	suite := bn256.NewSuite()
	private1, public1 := KeyGen(suite, random.New())
	private2, public2 := KeyGen(suite, random.New())
	msg := []byte("Hello many times Boneh-Lynn-Shacham")
	sig1, err := Sign(suite, private1, msg)
	require.Nil(b, err)
	sig2, err := Sign(suite, private2, msg)
	require.Nil(b, err)

	mask, _ := sign.NewMask(suite, []kyber.Point{public1, public2}, nil)
	mask.SetBit(0, true)
	mask.SetBit(1, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AggregateSignatures(suite, [][]byte{sig1, sig2}, mask)
	}
}
