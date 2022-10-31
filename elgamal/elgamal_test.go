package elgamal

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestElgamal(t *testing.T) {
	msg := "Hello World"
	curve := Setup()              // Choosen curve : ED25519
	ESK_i, EPK_i := KeyGen(curve) // sender
	ESK_j, EPK_j := KeyGen(curve) // reciever
	//KeyPub = EPK_i                // key_pub -> Recievers Public Key
	C1, C2, C3 := AuthEncryption(curve, msg, ESK_i, EPK_i, EPK_j)

	// c1 := hex.EncodeToString(C1.ToAffineCompressed())
	// c2 := C2
	// c3 := hex.EncodeToString(C3)

	c1 := hex.EncodeToString(C1.ToAffineCompressed())
	c2 := C2
	c3 := hex.EncodeToString(C3)

	C1_j := curve.Point
	C1_j_temp, _ := hex.DecodeString(c1)
	C1_j, _ = C1_j.FromAffineCompressed(C1_j_temp)

	C3_j, _ := hex.DecodeString(c3)

	if !C1.Equal(C1_j) {
		t.Fatal("wqerwer")
		t.Log(C1_j)
		fmt.Println(C1_j)
	}

	recovered_msg, _ := AuthDecryption(C1_j, c2, C3_j, EPK_i, EPK_j, ESK_j)

	if recovered_msg != msg {
		t.Fatal(recovered_msg)
	}
}
