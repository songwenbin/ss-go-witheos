package httpserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type PriceResponse struct {
	Content     PricePayload `json:"content"`
	TsSignature string       `json:"ts_signature"`
}

type PricePayload struct {
	ContractAddress string `json:"ContractAddress"`
	Price           int    `json:"Price"`
	PublicKey       string `json:"PublicKey"`
	Ts              string `json:"ts"`
}

var se *SafeEncrypt = &SafeEncrypt{}

func init() {
	pub, pri := GenerateSigKey()
	se.SetPrivateKey(pri)
	se.SetPublicKey("{\"kid\": \"server_kid\", \"use\":\"enc\", \"e\": \"AQAB\", \"kty\": \"RSA\", \"n\": \"nxuARX905_3pDATluPJB5NMalvPgqc9FImgDQXZ3scpWiumVYC2disk2qSlnH8ZgBnTXvkQUyNKxfmMum9qkgHJXwKtxVoKdIVrQPy3hiC9U0tFGSvgGNeFp5qaEsm5SK8R7Y2kWWz4VEl9n0TTdmO-0D1P4co-hlk0eo4JLU95aJxpwuNafDoZDm4MZM04D4kh3ZxC_mXklT8WRQ8E-bOnkOYCfqQiniLXIHQvV7eSVgHYokhcnhK9GYaOe73gNwEdXuBQAabZsvBAasaWaPMrkfGOef9RFPt6wHDpgmpJBgSJRuAI19f7hAlJI5DeUT0TwzgU6xVfOC08sQlYVjQ\"}")
	fmt.Println(pub)
}

func ResultToClientForPrice() string {
	current := time.Now().Unix()
	current_str := strconv.FormatInt(current, 10)
	signed, _ := JWS_Sign(se, current_str)

	payload := PricePayload{
		ContractAddress: "0xdeadbeef",
		Price:           1,
		PublicKey:       se.GetSigPublicKey(),
		Ts:              current_str,
	}

	response := &PriceResponse{
		Content:     payload,
		TsSignature: signed,
	}

	result, err := json.Marshal(response)
	if err != nil {
		fmt.Println(err.Error())
	}

	return string(result)
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
	fmt.Fprintf(w, "handleLogin interface")
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
