package httpserver

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/eoscanada/eos-go/btcsuite/btcutil/base58"
	"gopkg.in/square/go-jose.v2"
	b "gopkg.in/square/go-jose.v2"
)

type SafeEncrypt struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
}

type PayloadToServer struct {
	Key       string `json:"key"`
	Signature string `json:"signature"`
	Request   string `json:"request"`
}

func (se *SafeEncrypt) SetPrivateKey(jsonformat string) {
	privateKey := jose.JSONWebKey{}
	privateKey.UnmarshalJSON([]byte(jsonformat))

	se.PrivateKey = privateKey.Key.(*rsa.PrivateKey)
}

func (se *SafeEncrypt) SetPublicKey(jsonformat string) {
	publicKey := jose.JSONWebKey{}
	err := publicKey.UnmarshalJSON([]byte(jsonformat))
	if err != nil {
		fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
		fmt.Println(err)
	}

	se.PublicKey = publicKey.Key.(*rsa.PublicKey)
}

func (se *SafeEncrypt) GetPrivateKey() string {
	kid := base64.StdEncoding.EncodeToString([]byte("serverid"))
	//priv := jose.JSONWebKey{Key: se.PrivateKey, KeyID: kid, Algorithm: string(jose.RSA_OAEP_256), Use: "enc"}
	priv := jose.JSONWebKey{Key: se.PrivateKey, KeyID: kid, Algorithm: string(jose.RSA_OAEP), Use: "enc"}
	privJS, _ := priv.MarshalJSON()
	return string(privJS)
}

func (se *SafeEncrypt) GetPublicKey() string {
	kid := base64.StdEncoding.EncodeToString([]byte("serverid"))
	//pub := jose.JSONWebKey{Key: se.PublicKey, KeyID: kid, Algorithm: string(jose.RSA_OAEP_256), Use: "enc"}
	pub := jose.JSONWebKey{Key: se.PublicKey, KeyID: kid, Algorithm: string(jose.RSA_OAEP), Use: "enc"}
	pubJS, _ := pub.MarshalJSON()
	return string(pubJS)
}

func (se *SafeEncrypt) GetSigPrivateKey() string {
	kid := base64.StdEncoding.EncodeToString([]byte("serverid"))
	priv := jose.JSONWebKey{Key: se.PrivateKey, KeyID: kid, Algorithm: string(jose.RS256), Use: "sig"}
	privJS, _ := priv.MarshalJSON()
	return string(privJS)
}

func (se *SafeEncrypt) GetSigPublicKey() string {
	kid := base64.StdEncoding.EncodeToString([]byte("serverid"))
	pub := jose.JSONWebKey{Key: se.PublicKey, KeyID: kid, Algorithm: string(jose.RS256), Use: "sig"}
	pubJS, _ := pub.MarshalJSON()
	return string(pubJS)
}

func JWE_Encrypt(se *SafeEncrypt, content string) string {
	//kid := base64.StdEncoding.EncodeToString([]byte("serverid"))
	//pub := jose.JSONWebKey{Key: se.PublicKey, KeyID: kid, Algorithm: string(jose.RSA_OAEP_256), Use: "enc"}
	//encrypter, err := b.NewEncrypter(b.A256CBC_HS512, b.Recipient{Algorithm: b.RSA_OAEP, Key: se.PublicKey}, nil)
	opt := b.EncrypterOptions{}
	opt.WithType("JWE")
	encrypter, err := b.NewEncrypter(b.A256GCM, b.Recipient{Algorithm: b.RSA_OAEP, Key: se.PublicKey}, &opt)

	if err != nil {
		fmt.Println(err)
	}

	var plaintext = []byte(content)
	object, err := encrypter.Encrypt(plaintext)
	if err != nil {
		panic(err)
	}

	fmt.Println("JWE格式")
	fmt.Println(object.FullSerialize())
	result, _ := object.CompactSerialize()
	return result
}

func JWE_Decrypt(se *SafeEncrypt, content string) string {
	object, err := b.ParseEncrypted(content)
	if err != nil {
		panic(err)
	}
	decrypted, err := object.Decrypt(se.PrivateKey)
	if err != nil {
		panic(err)
	}

	return string(decrypted)
}

func JWS_Sign(se *SafeEncrypt, content string) (string, error) {
	kid := base64.StdEncoding.EncodeToString([]byte("serverid"))
	pri := jose.JSONWebKey{Key: se.PrivateKey, KeyID: kid, Algorithm: string(jose.RS256), Use: "sig"}
	signer, err := b.NewSigner(b.SigningKey{Algorithm: b.RS256, Key: pri}, nil)
	if err != nil {
		fmt.Println("ss")
		fmt.Println(err)
		return "", err
	}

	c := []byte(content)
	object, err := signer.Sign(c)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return object.FullSerialize(), nil
}

func JWS_Verify(se *SafeEncrypt, payload string) (string, error) {
	object, err := b.ParseSigned(payload)
	if err != nil {
		fmt.Println(err)
		fmt.Println("....")
		return "", err
	}

	output, err := object.Verify(se.PublicKey)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return string(output), nil
}

func SHA256(content string) []byte {
	h := sha256.New()
	h.Write([]byte(content))
	return h.Sum(nil)
}

func GenerateEncKey() (string, string) {
	var privKey crypto.PrivateKey
	var pubKey crypto.PublicKey
	pubKey, privKey, err = KeygenEnc(2048)
	/*
		b := make([]byte, 5)
		rand.Read(b)
		kid := base32.StdEncoding.EncodeToString(b)
	*/
	kid := base64.StdEncoding.EncodeToString([]byte("serverid"))
	priv := jose.JSONWebKey{Key: privKey, KeyID: kid, Algorithm: string(jose.RSA_OAEP), Use: "enc"}
	pub := jose.JSONWebKey{Key: pubKey, KeyID: kid, Algorithm: string(jose.RSA_OAEP), Use: "enc"}
	if priv.IsPublic() || !pub.IsPublic() || !priv.Valid() || !pub.Valid() {
		//app.Fatalf("invalid keys were generated")
	}
	//fmt.Println(pub.Key)
	//fmt.Println(priv.Key)
	privJS, _ := priv.MarshalJSON()
	pubJS, _ := pub.MarshalJSON()
	//fmt.Println(string(privJS))
	//fmt.Println(string(pubJS))
	return string(pubJS), string(privJS)
}

var err error

func GenerateSigKey() (string, string) {
	var privKey crypto.PublicKey
	var pubKey crypto.PrivateKey
	pubKey, privKey, err = KeygenSig(2048)
	//b := make([]byte, 5)
	//rand.Read(b)
	kid := base64.StdEncoding.EncodeToString([]byte("serverid"))
	priv := jose.JSONWebKey{Key: privKey, KeyID: kid, Algorithm: string(jose.RS256), Use: "sig"}
	pub := jose.JSONWebKey{Key: pubKey, KeyID: kid, Algorithm: string(jose.RS256), Use: "sig"}
	//fmt.Println(priv.Key)
	//fmt.Println(pub.Key)
	if priv.IsPublic() || !pub.IsPublic() || !priv.Valid() || !pub.Valid() {
		//app.Fatalf("invalid keys were generated")
	}
	privJS, _ := priv.MarshalJSON()
	pubJS, _ := pub.MarshalJSON()
	//fmt.Println(string(privJS))
	//fmt.Println(string(pubJS))

	return string(pubJS), string(privJS)
}

func KeygenSig(bits int) (crypto.PublicKey, crypto.PrivateKey, error) {
	if bits == 0 {
		bits = 2048
	}
	if bits < 2048 {
		return nil, nil, errors.New("too short key for RSA `alg`, 2048+ is required")
	}
	key, err := rsa.GenerateKey(rand.Reader, bits)
	return key.Public(), key, err
}

// KeygenEnc generates keypair for corresponding KeyAlgorithm.
func KeygenEnc(bits int) (crypto.PublicKey, crypto.PrivateKey, error) {
	if bits == 0 {
		bits = 2048
	}
	if bits < 2048 {
		return nil, nil, errors.New("too short key for RSA `alg`, 2048+ is required")
	}
	key, err := rsa.GenerateKey(rand.Reader, bits)
	return key.Public(), key, err
}

func Base58Encoder(payload []byte) string {
	encoded := base58.Encode(payload)
	return encoded
}

func Base58Decoder(content string) []byte {
	decoded := base58.Decode(content)
	return decoded
}
