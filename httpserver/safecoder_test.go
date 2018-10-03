package httpserver

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"testing"
)

func TestInputParam(t *testing.T) {
	serverPrivtateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println("no")
	}
	fmt.Println(serverPrivtateKey)
	serverPublicKey := &serverPrivtateKey.PublicKey
	fmt.Println(serverPublicKey)

	var buf bytes.Buffer
	err2 := json.NewEncoder(&buf).Encode(serverPublicKey)
	if err2 != nil {
		fmt.Println("no2")
	}

	result := SHA256(string(SHA256(buf.String())))
	fmt.Println(result)

	GenerateEncKey()
	GenerateSigKey()
	t.Error("test")
}

func TestEnDecrypt(t *testing.T) {
	pub, pri := GenerateEncKey()

	var se *SafeEncrypt = &SafeEncrypt{}
	se.SetPrivateKey(pri)
	se.SetPublicKey(pub)

	result := JWE_Encrypt(se, "test")
	actual := JWE_Decrypt(se, result)
	fmt.Println("========================")
	fmt.Println(actual)

	if actual != "test" {
		t.Error("test")
	}

	sign_result, err := JWS_Sign(se, "test2")
	fmt.Println(sign_result)
	if err != nil {
		fmt.Println(err)
	}
	sign_actual, err := JWS_Verify(se, sign_result)
	fmt.Println(sign_actual)
	if err != nil {
		fmt.Println(err)
	}

	if sign_actual != "test2" {
		fmt.Println("========================")
		fmt.Println(sign_actual)
		t.Error("test2")
	}
}

func TestEncrypt(t *testing.T) {
	fmt.Println("ok1")
	t.Log("sdfdsfsd")
	fmt.Print("111")

	//t.Error("test")
}
