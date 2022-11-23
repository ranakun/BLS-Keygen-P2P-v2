package zkp

import (
	"fmt"

	"main.go/bls"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/group/edwards25519"
	"go.dedis.ch/kyber/v3/pairing/bn256"

	"math/rand"
)

var curve = edwards25519.NewBlakeSHA256Ed25519()
var sha256 = curve.Hash()
var suite = bn256.NewSuite()

type Signature struct {
	r kyber.Point
	s kyber.Scalar
}

func Hash(s string) kyber.Scalar {
	sha256.Reset()
	sha256.Write([]byte(s))

	return curve.Scalar().SetBytes(sha256.Sum(nil))
}

func RandomString(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// m: Message
// x: Private key
func Sign(m string, x kyber.Scalar) Signature {
	// Get the base of the curve.
	g := curve.Point().Base()

	// Pick a random k from allowed set.
	k := curve.Scalar().Pick(curve.RandomStream())

	// r = k * G (a.k.a the same operation as r = g^k)
	r := curve.Point().Mul(k, g)

	// Hash(m || r)
	e := Hash(m + r.String())

	// s = k - e * x
	s := curve.Scalar().Sub(k, curve.Scalar().Mul(e, x))

	return Signature{r: r, s: s}
}

// m: Message
// S: Signature
func PublicKey(m string, S Signature) kyber.Point {
	// Create a generator.
	g := curve.Point().Base()

	// e = Hash(m || r)
	e := Hash(m + S.r.String())

	// y = (r - s * G) * (1 / e)
	y := curve.Point().Sub(S.r, curve.Point().Mul(S.s, g))
	y = curve.Point().Mul(curve.Scalar().Div(curve.Scalar().One(), e), y)

	return y
}

// m: Message
// s: Signature
// y: Public key
func Verify(m string, S Signature, y kyber.Point) bool {
	// Create a generator.
	g := curve.Point().Base()

	// e = Hash(m || r)
	e := Hash(m + S.r.String())

	// Attempt to reconstruct 's * G' with a provided signature; s * G = r - e * y
	sGv := curve.Point().Sub(S.r, curve.Point().Mul(e, y))

	// Construct the actual 's * G'
	sG := curve.Point().Mul(S.s, g)

	//fmt.Println(sG)
	//fmt.Println(sGv)
	// Equality check; ensure signature and public key outputs to s * G.
	return sG.Equal(sGv)
}

func (S Signature) String() string {
	return fmt.Sprintf("(r=%s, s=%s)", S.r, S.s)
}

func Commitment(x kyber.Scalar, m string) ([]byte, []byte, []byte) {
	publicKey := curve.Point().Mul(x, curve.Point().Base()) // x.P
	sig := Sign(m, x)                                       // x

	// return (kgd and kgc->signature)
	kgd, _ := sig.r.MarshalBinary()
	kgc, _ := sig.s.MarshalBinary()
	pk, _ := publicKey.MarshalBinary()
	return kgd, kgc, pk

	/* TODO
	x - private key
	x * P - public key
	P - generator
	- xP, random_tag -- x * Sign(xp, random_tag) || commitment xP, xHx decomm random_tag
	- xP
	- xP, hx // hash(xP, random_tag)
	-  hash(xP, random_tag) == hx
	*/

	// publicKey := curve.Point().Mul(x, curve.Point().Base()) // x.P
	// sig := Sign(m+r, x)                                     // sign kgc, pk
	// // return (kgd and kgc->signature)
	// kgd, _ := sig.r.MarshalBinary() // r
	// kgc, _ := sig.s.MarshalBinary()
	// pk, _ := publicKey.MarshalBinary()
	// return kgd, kgc, pk
}

/*
	CHANGES
	x - private key
	x * P - public key
	P - generator
	- xP, random_tag -- x * Sign(xp, random_tag) || commitment xP, xHx decomm random_tag
	- xP
	- xP, hx // hash(xP, random_tag)
	-  hash(xP, random_tag) == hx
*/

func CommitmentBLS(x kyber.Scalar, m string) ([]byte, []byte, []byte) {
	publicKey := suite.G2().Point().Mul(x, suite.G2().Point().Base()) // x.P
	r := RandomString(30)
	sig, _ := bls.Sign(suite, x, []byte(m+r))

	// return (kgd and kgc->signature)
	kgd := []byte(r)
	kgc, _ := sig.MarshalBinary()
	pk, _ := publicKey.MarshalBinary()
	return kgd, kgc, pk
}

// sending public keys

func DecommitmentBLS(kgd []byte, kgc []byte, pk []byte) bool {
	publicKey := suite.G2().Point()
	err := publicKey.UnmarshalBinary(pk)
	if err != nil {
		fmt.Println(err)
	}

	msg := "Hello World" + string(kgd)

	t := bls.Verify(suite, publicKey, []byte(msg), kgc)
	return t == nil
}

func Decommitment(kgd []byte, kgc []byte, pk []byte) bool {
	fm := curve.Scalar()
	err := fm.UnmarshalBinary(kgc)
	if err != nil {
		fmt.Println(err)
	}

	f_2 := curve.Point()
	err = f_2.UnmarshalBinary(pk)
	if err != nil {
		fmt.Println(err)
	}

	f_3 := curve.Point()
	err = f_3.UnmarshalBinary(kgd)
	if err != nil {
		fmt.Println(err)
	}

	newS := Signature{}
	newS.s = fm  // signature
	newS.r = f_3 // kgd

	//fmt.Println(newS.s)
	t := Verify("Hello World", newS, f_2)
	// f_2 -> public key
	// fmt.Println(t)
	return t
}

func Setup(privateKey kyber.Scalar) ([]byte, []byte, []byte) {
	pk := curve.Scalar()
	mar, _ := privateKey.MarshalBinary()
	pk.SetBytes(mar)

	privatekey := curve.Scalar().Zero()
	privatekey = privatekey.Add(privatekey, pk)

	kgd, kgc, pubKey := Commitment(privatekey, "Hello World")

	return kgd, kgc, pubKey
}

func SetupBLS(privateKey kyber.Scalar) ([]byte, []byte, []byte) {
	kgd, kgc, pubKey := CommitmentBLS(privateKey, "Hello World")

	return kgd, kgc, pubKey
}

func SelfTest(privateKey kyber.Scalar) ([]byte, []byte, []byte) {
	pk := curve.Scalar()
	mar, _ := privateKey.MarshalBinary()
	pk.SetBytes(mar)

	privatekey := curve.Scalar().Zero()
	privatekey = privatekey.Add(privatekey, pk)

	kgd, kgc, pubKey := Commitment(privatekey, "Hello World")
	// kgd, kgc, pubKey := Commitment(pk, "Hello World")
	T := Decommitment(kgd, kgc, pubKey)
	fmt.Println("Self Check", T)

	return kgd, kgc, pubKey
}
