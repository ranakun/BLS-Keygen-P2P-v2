package elgamal

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"strings"

	"github.com/coinbase/kryptology/pkg/core/curves"
)

//to add padding to the passed string
func addBase64Padding(value string) string {
	m := len(value) % 4
	if m != 0 {
		value += strings.Repeat("=", 4-m)
	}

	return value
}

//to remove padding from the passed string
func removeBase64Padding(value string) string {
	return strings.Replace(value, "=", "", -1)
}

//to pad the bytes in src
func Pad(src []byte) []byte {
	//to calculate padding size
	padding := aes.BlockSize - len(src)%aes.BlockSize
	//padding is added to the padtext using repeat function
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

//to unpad the bytes in src
func Unpad(src []byte) ([]byte, error) {
	length := len(src)
	unpadding := int(src[length-1])

	if unpadding > length {
		return nil, errors.New("unpad error")
	}

	return src[:(length - unpadding)], nil
}

//to encrypt the text with the key passed
func AESencrypt(key []byte, text string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	msg := Pad([]byte(text))
	//to create a byte array of size blocksize plus length of msg
	ciphertext := make([]byte, aes.BlockSize+len(msg))
	iv := ciphertext[:aes.BlockSize]
	//to read the file with specified blocksize
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	//returns a Stream which encrypts with cipher feedback mode, using the given Block.//returns a Stream which encrypts with cipher feedback mode, using the given Block.
	cfb := cipher.NewCFBEncrypter(block, iv)
	//encrypts texts with xor function
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(msg))
	//padding is removed from the final msg and converted to string
	finalMsg := removeBase64Padding(base64.URLEncoding.EncodeToString(ciphertext))
	return finalMsg, nil
}

//to decrypt the text with the key passed
func AESdecrypt(key []byte, text string) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	//padding is done to the ciphertext passed
	decodedMsg, err := base64.URLEncoding.DecodeString(addBase64Padding(text))
	if err != nil {
		return "", err
	}

	if (len(decodedMsg) % aes.BlockSize) != 0 {
		return "", errors.New("blocksize must be multiple of decoded message length")
	}

	iv := decodedMsg[:aes.BlockSize]
	msg := decodedMsg[aes.BlockSize:]

	//decrypts the message
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(msg, msg)

	//message is unpadded
	unpadMsg, err := Unpad(msg)
	if err != nil {
		return "", err
	}

	//message is returned as a string
	return string(unpadMsg), nil
}

//Here we define Public Parameters ( Curve Generator , Receiver's Public Key)
var G curves.Point
var KeyPub curves.Point

func Setup() *curves.Curve {
	curve := curves.ED25519() // Choosen curve : ED25519
	G = curve.Point.Generator()
	return curve
}

func KeyGen(curve *curves.Curve) (curves.Scalar, curves.Point) { //Generates <Key_pri,KeyPub> Pair
	private := curve.Scalar.Random(rand.Reader)
	public := G.Mul(private)
	return private, public
}

func Encryption(curve *curves.Curve, msg string, key_pub_r curves.Point) (curves.Point, string) {
	r := curve.Scalar.Random(rand.Reader)
	C1 := curve.Point.Generator().Mul(r)
	// M := curve.Point.Hash([]byte(msg))
	temp := key_pub_r.Mul(r)
	aesKey := temp.ToAffineCompressed()
	//fmt.Println(string(aesKey))
	C2, _ := AESencrypt(aesKey, msg)

	return C1, C2
}

func AuthEncryption(curve *curves.Curve, msg string, key_priv curves.Scalar, key_pub_s curves.Point, key_pub_r curves.Point) (curves.Point, string, []byte) {
	//Inputs to function-> Curve used , Message to encrypt , Private key
	r := curve.Scalar.Random(rand.Reader)
	C1 := curve.Point.Generator().Mul(r)
	// M := curve.Point.Hash([]byte(msg))
	temp := key_pub_r.Mul(r)
	aesKey := temp.ToAffineCompressed()

	C2, _ := AESencrypt(aesKey, msg)

	symm_key := key_pub_r.Mul(key_priv) //Symmetric Key used by both parties

	t := []byte(C2)
	t = append(t, symm_key.ToAffineCompressed()...)
	t = append(t, key_pub_s.ToAffineCompressed()...)
	t = append(t, key_pub_r.ToAffineCompressed()...)

	h := sha256.New()
	h.Write(t)

	C3 := h.Sum(nil) // SHA256 Hash

	return C1, C2, C3
}

func Decryption(C1 curves.Point, C2 string, key_pri curves.Scalar) string {
	temp_key := C1.Mul(key_pri) //Recovering Symm. key
	aesKey := temp_key.ToAffineCompressed()

	dec, _ := AESdecrypt(aesKey, C2)

	return dec
}

func AuthDecryption(C1 curves.Point, C2 string, C3 []byte, key_pub_s curves.Point, key_pub_r curves.Point, key_pri curves.Scalar) (string, bool) {
	temp_key := C1.Mul(key_pri) //Recovering Symm. key
	aesKey := temp_key.ToAffineCompressed()

	dec, _ := AESdecrypt(aesKey, C2)
	symm_key := key_pub_s.Mul(key_pri) //Recovering Symm. key

	t := []byte(C2)
	t = append(t, symm_key.ToAffineCompressed()...)
	t = append(t, key_pub_s.ToAffineCompressed()...)
	t = append(t, key_pub_r.ToAffineCompressed()...)

	h := sha256.New()
	h.Write(t)

	temp_C3 := h.Sum(nil)

	//temp_C3 = Hash(C2, symm_key, key_pub_s, key_pub_r)

	if string(temp_C3) == string(C3) {
		return dec, true
	}
	return "INVALID", false
}

// func main() {
// 	msg := "Hello World"
// 	curve := Setup()              // Choosen curve : ED25519
// 	ESK_i, EPK_i := KeyGen(curve) // sender
// 	ESK_j, EPK_j := KeyGen(curve) // reciever
// 	//KeyPub = EPK_i                // key_pub -> Recievers Public Key
// 	C1, C2, C3 := AuthEncryption(curve, msg, ESK_i, EPK_i, EPK_j)

// 	recovered_msg, _ := AuthDecryption(C1, C2, C3, EPK_i, EPK_j, ESK_j)
// 	fmt.Println(recovered_msg)

// }
