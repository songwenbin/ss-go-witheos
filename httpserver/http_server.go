package httpserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type PublicKey struct {
	Kid string `json:"kid"`
	E   string `json:"e"`
	Kty string `json:"kty"`
	N   string `json:"n"`
}

type PublicKeyForLogin struct {
	Kid string `json:"kid"`
	E   string `json:"e"`
	Kty string `json:"kty"`
	N   string `json:"n"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	Enc string `json:"enc"`
}

type TsSignature struct {
	Payload   string `json:"payload"`
	Signature string `json:"signature"`
	Protected string `json:"protected"`
}

type PriceResponse struct {
	Content     PricePayload `json:"content"`
	TsSignature TsSignature  `json:"ts_signature"`
}

type PricePayload struct {
	ContractAddress string    `json:"ContractAddress"`
	Price           int       `json:"Price"`
	PublicKey       PublicKey `json:"PublicKey"`
	Ts              int64     `json:"ts"`
}

type LoginPayload struct {
	Key       PublicKeyForLogin `json:"key"`
	Signature TsSignature       `json:"signature"`
	Request   string            `json:"request"`
}

type LoginResponse struct {
	SScert         []AccountResponse `json:"ss_cert"`
	Ts             int64             `json:"ts"`
	SignatureStrTs TsSignature       `json:"signature_str_ts"`
}

type AccountResponse struct {
	Type    string `json:"Type"`
	Address string `json:"address"`
	Port    string `json:"port"`
	Key     string `json:"key"`
	Method  string `json:"method"`
}

var se *SafeEncrypt = &SafeEncrypt{}

func init() {
	pub, pri := GenerateSigKey()
	se.SetPrivateKey(pri)
	se.SetPublicKey(pub)
}

func ResultToClientForPrice() string {
	current := time.Now().Unix()
	current_str := strconv.FormatInt(current, 10)
	signed, _ := JWS_Sign(se, current_str)

	var sign TsSignature
	err := json.Unmarshal([]byte(signed), &sign)
	if err != nil {
		fmt.Println(err.Error())
	}

	var pubkey PublicKey
	err = json.Unmarshal([]byte(se.GetPublicKey()), &pubkey)
	if err != nil {
		fmt.Println(err.Error())
	}

	payload := PricePayload{
		ContractAddress: "0xdeadbeef",
		Price:           1,
		PublicKey:       pubkey,
		Ts:              current,
	}

	response := &PriceResponse{
		Content:     payload,
		TsSignature: sign,
	}

	result, err := json.Marshal(response)
	if err != nil {
		fmt.Println(err.Error())
	}

	return string(result)
}

func DecryptInputParamForLogin(se *SafeEncrypt, content string) string {
	result := JWE_Decrypt(se, content)
	return result
}

func ResponseForLogin(client *SafeEncrypt, server *SafeEncrypt, res LoginPayload) string {

	// 2 signed_payload = JWS(ts_server, server_private_key)
	ts := time.Now().Unix()
	signed, err := JWS_Sign(se, string(ts))
	if err != nil {
		fmt.Println(err.Error())
	}

	var sign TsSignature
	err = json.Unmarshal([]byte(signed), &sign)
	if err != nil {
		fmt.Println(err.Error())
	}

	//1 service_list = [{"type":"s", "address":"1.1.1.1", "port":"13345", "key":"key134555", "method":"hello"},{"type":"d", "address":"1.1.1.1", "port":"13345", "key":"key134555", "method":"world"}]
	//3 response_of_server_signed = {"ss_cert": ss_cert_list, "ts": ts_server, "signature_str_ts": signed_payload})
	var service_list LoginResponse = LoginResponse{
		SScert: []AccountResponse{AccountResponse{
			Type:    "s",
			Address: "1.1.1.1",
			Port:    "13345",
			Key:     "key13455",
			Method:  "hello",
		}, AccountResponse{
			Type:    "s",
			Address: "1.1.1.1",
			Port:    "13345",
			Key:     "key13455",
			Method:  "hello",
		}},
		Ts:             ts,
		SignatureStrTs: sign,
	}
	servicelist_json, err := json.Marshal(&service_list)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(string(servicelist_json))

	//4 response_http = JWE(response_of_server_signed, public_key_client)

	response_http := JWE_Encrypt(client, string(servicelist_json))
	fmt.Println(response_http)
	return response_http
}

func handlePrice(w http.ResponseWriter, r *http.Request) {
	// 解决跨域的参数
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")

	response := ResultToClientForPrice()
	fmt.Fprintf(w, response)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	// 解决跨域的参数
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")

	r.ParseForm()
	input := r.Form["code"][0]

	fmt.Println(input)
	result := DecryptInputParamForLogin(se, input)
	fmt.Println("============")
	fmt.Println(result)

	var res LoginPayload
	if err := json.Unmarshal([]byte(result), &res); err != nil {
		fmt.Println("数据无法解析")
	}

	keyjson, _ := json.Marshal(res.Key)
	fmt.Println(string(keyjson))
	var client SafeEncrypt
	client.SetPublicKey(string(keyjson))

	signjson, _ := json.Marshal(res.Signature)
	timestamp, _ := JWS_Verify(&client, string(signjson))
	fmt.Println(timestamp)
	fmt.Println("sfdsfsdfsdfdsfsdfsdfs")

	fmt.Fprintf(w, ResponseForLogin(&client, se, res))
}

func HttpServer() {
	// Todo 服务器的启动地址需要参数进行传入
	server := http.Server{
		Addr: "localhost:8887",
	}
	http.HandleFunc("/price.json", handlePrice)
	http.HandleFunc("/cert.info", handleLogin)

	server.ListenAndServe()
}
