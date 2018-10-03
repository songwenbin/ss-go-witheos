package httpserver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func TestPriceInterface(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		log.Fatal(err)
	}
	greeting, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s", greeting)
	t.Error("")
}

func TestLoginInterface(t *testing.T) {
	client_pub, client_pri := GenerateSigKey()
	var client SafeEncrypt
	client.SetPrivateKey(client_pri)
	client.SetPublicKey(client_pub)

	server_pub, server_pri := GenerateSigKey()
	var server SafeEncrypt
	server.SetPrivateKey(server_pri)
	server.SetPublicKey(server_pub)

	//	1 kid_of_client_public_key = base58_encoding(SHA256(SHA256(public_key_client["n"]))
	//  An example of kid_of_client_public_key can be 'HareBAjopJi7GabM5HEX1VWsTHEmu8cFfAQFfSpbMsvT'

	//	2 my_signature = JWS(string(current_time_stamp_in_seconds), private_key_client)
	current := time.Now().Unix()
	current_str := strconv.FormatInt(current, 10)
	my_signature, _ := JWS_Sign(&client, current_str)
	fmt.Println(my_signature)

	//	3 clear_payload_to_server = {"key":JWK(public_key_client), "signature": my_signature, "request":"ss_cert"}
	clear_payload_to_server := &LoginPayload{
		Key:       client.GetPublicKey(),
		Signature: my_signature,
		Request:   "ss_cert",
	}
	// 4 encrypted_payload_server = JWE(clear_payload_to_server, public_key_server)
	content, err := json.Marshal(clear_payload_to_server)
	if err != nil {
		fmt.Println(err.Error())
	}

	input := JWE_Encrypt(&server, string(content))

	// 5 parameter to http: ?code=encrypted_payload_server
	fmt.Println("?code=", input)

	// 正式流程
	result := DecryptInputParamForLogin(&server, input)

	var res LoginPayload
	if err := json.Unmarshal([]byte(result), &res); err != nil {
		fmt.Println("数据无法解析")
	}

	client.SetPublicKey(res.Key)
	timestamp, _ := JWS_Verify(&client, res.Signature)
	if timestamp != current_str {
		t.Error("timestamp")
	}

	// 结果数据
	// 1 service_list = {"ss_cert":[{"type":"s", "address":"1.1.1.1", "port":"13345", "key":"key134555", "method":"hello"},{"type":"d", "address":"1.1.1.1", "port":"13345", "key":"key134555", "method":"world"}], "ts":12345678}
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
	response_of_server_signed, err := JWS_Sign(&server, string(servicelist_json))
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(response_of_server_signed)
	//3  response_http = JWE(response_of_server_signed, public_key_client)

	client.SetPublicKey(res.Key)
	response_http := JWE_Encrypt(&client, response_of_server_signed)
	fmt.Println(response_http)
	//t.Error("")
}

func TestResultToClient(t *testing.T) {
	response := ResultToClientForPrice()
	fmt.Println(response)

	var res PriceResponse
	if err := json.Unmarshal([]byte(response), &res); err != nil {
		t.Error("数据无法解析")
	}

	actual, _ := JWS_Verify(se, res.TsSignature)

	if actual != res.Content.Ts {
		t.Error("")
	}

	fmt.Println(res.Content.Ts)
	fmt.Println(actual)

	t.Error("")

}

func TestTemp(t *testing.T) {

	//sig := "{\"payload\":\"MTUzODU2MjE2OA\",\"protected\":\"eyJhbGciOiJSUzI1NiIsImtpZCI6IkR0eVJTaG5pUGJrdVh1RW5Obm1vNzRFR2RISERGNEpKMW9yZXNkcFBhYnlWIn0\",\"signature\":\"CxU7nRUGuwIOsmNr9hUPXKuJJgTc5tMcYQqnVrND7DaCNnU0rL5vySipoMH4HLCpW581W2aGXo3QB4o8c7QEFfKLGdR8bWWZdi7Q75wzNI4N9qRWsD82S2noWI6xuh9Dlh8nGUAC_ZfpnHqg3aMx6PWCPpYrxxdP1tWwpr2yAprBhMcsEADzINiWvUkyneFkuljt-UywU1U3Cgb5_jPj-IKcWC8QaaQkUHEcqpu8qzsWhKqKITdBvDr6fyIZ4ENzk0Gr0Bni7S9bXcFDvzaqwhOrcw2vnNXRoiq58illbA1C7DvoTz8_7BMCeVcsrCqE31HEWsA6Em5rAIQEn40Vkg\"}"
	//publc_key := "{\"kid\":\"DtyRShniPbkuXuEnNnmo74EGdHHDF4JJ1oresdpPabyV\",\"e\":\"AQAB\",\"kty\":\"RSA\",\"n\":\"nxuARX905_3pDATluPJB5NMalvPgqc9FImgDQXZ3scpWiumVYC2disk2qSlnH8ZgBnTXvkQUyNKxfmMum9qkgHJXwKtxVoKdIVrQPy3hiC9U0tFGSvgGNeFp5qaEsm5SK8R7Y2kWWz4VEl9n0TTdmO-0D1P4co-hlk0eo4JLU95aJxpwuNafDoZDm4MZM04D4kh3ZxC_mXklT8WRQ8E-bOnkOYCfqQiniLXIHQvV7eSVgHYokhcnhK9GYaOe73gNwEdXuBQAabZsvBAasaWaPMrkfGOef9RFPt6wHDpgmpJBgSJRuAI19f7hAlJI5DeUT0TwzgU6xVfOC08sQlYVjQ\"}"

	publc_key := "{\"kty\":\"RSA\",\"kid\":\"test\",\"n\":\"1O2MblLzriTX06nsrcLy6diuEHl4s3eswSzAigyf7-zKFOo-ttbNeQTQ29X2iHSeLdYbHZc_l-olyPWOhpQSVDY9ikkluml6W_tkOgSwcPZNvXOfTXWSw7m_oJmLU1FBZJhKeMorHF-62wLC5xYMe7OmXT_SOiEzP5VvVJuFmt0czexUO9wB1gSTxBvBFUbhkb__OoHL2ArSp56acfkhZJzIRCz19JQSXBn9mMgSJPOO1GjVwrhVWsddrVkaZ4ZzchM6E_wbtSQza0JUWIIX6HLOgDMuvnikAxf3j9QWtO6uGHxVdddUtWxRnlruVmQ0gZ_mkKTxK-yUCIs4qwEBUw\",\"e\":\"AQAB\"}"
	sig := "{\"payload\":\"MTUzODU2OTEwNg\",\"protected\":\"eyJhbGciOiJSUzI1NiIsImtpZCI6IiJ9\",\"signature\":\"YHftpfx2UvroU-UiHm1W37AMgpIov7mXS2dPun0dNl6cu6hHzWhicGu9-z_aF3cyGXdamJbenU7zEsOkmDE_UTCyUI5J0AXd9RE3AFKSmrkltJ7UD3P640ejd3qBKPGpoakGAs1wa2X0v896gIWnq4gCDMaaGDN6FQ1Z6Ilgry6m0TsNz_Xwcm6rZEAw2tMFgIssjg-GRWnS7kogYBhtTIgwxg3G0Qxoc_-NY_VVMij8m7nU2Lu1CMyp594jPr3lUAXzY_2kkE9fccDKfXONGsdDEU1r8J3FK7B0fpfKK7QlHxeug2G6QyGCkxvyOp5DAzCd5-dVtC43spzyydmVSg\"}"
	var se SafeEncrypt
	se.SetPublicKey(publc_key)
	//result, err := JWS_Verify(&se, string(ret))
	result, err := JWS_Verify(&se, sig)
	if err != nil {
		fmt.Println("ok", err.Error())
	}
	t.Error(result)
}
