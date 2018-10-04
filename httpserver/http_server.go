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
	Key       string `json:"key"`
	Signature string `json:"signature"`
	Request   string `json:"request"`
}

type LoginResponse struct {
	SScert []AccountResponse `json:"ss_cert"`
	Ts     string            `json:"ts"`
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
	fmt.Println(result)
	return result
}

func ResponseForLogin(client *SafeEncrypt, server *SafeEncrypt, res LoginPayload) string {
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
		Ts: strconv.FormatInt(time.Now().Unix(), 10),
	}

	servicelist_json, err := json.Marshal(&service_list)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(string(servicelist_json))

	//2 response_of_server_signed = JWS(service_list, private_key_server)
	response_of_server_signed, err := JWS_Sign(server, string(servicelist_json))
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(response_of_server_signed)
	//3  response_http = JWE(response_of_server_signed, public_key_client)

	client.SetPublicKey(res.Key)
	response_http := JWE_Encrypt(client, response_of_server_signed)
	fmt.Println(response_http)
	return response_http
}

func handlePrice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
	w.Header().Set("content-type", "application/json")
	response := ResultToClientForPrice()
	fmt.Fprintf(w, response)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
	w.Header().Set("content-type", "application/json")

	r.ParseForm()
	input := r.Form["code"][0]

	result := DecryptInputParamForLogin(se, input)

	var res LoginPayload
	if err := json.Unmarshal([]byte(result), &res); err != nil {
		fmt.Println("数据无法解析")
	}

	var client SafeEncrypt
	client.SetPublicKey(res.Key)
	timestamp, _ := JWS_Verify(&client, res.Signature)
	fmt.Println(timestamp)

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
