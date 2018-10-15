package httpserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ss-go-witheos/eosapi"
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
	Price           Price     `json:"Price"`
	PublicKey       PublicKey `json:"PublicKey"`
	Ts              int64     `json:"ts"`
}

type Price struct {
	Symbol string `json:"symbol"`
	Amount int    `json:"amount"`
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

func ResultToClientForPrice(contract string) string {
	// 签名
	current := time.Now().Unix()
	current_str := strconv.FormatInt(current, 10)
	fmt.Println(current_str)
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
		ContractAddress: contract,
		Price: Price{
			Symbol: "SYS",
			Amount: 1,
		},
		PublicKey: pubkey,
		Ts:        current,
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

func decryptInputParamForLogin(se *SafeEncrypt, content string) string {
	result := JWE_Decrypt(se, content)
	return result
}

func responseForLogin(client *SafeEncrypt, server *SafeEncrypt, account AccountResponse) string {

	// 2 signed_payload = JWS(ts_server, server_private_key)
	ts := time.Now().Unix()
	currentStr := strconv.FormatInt(ts, 10)
	signed, err := JWS_Sign(se, currentStr)
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
	/*
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
	*/
	var service_list LoginResponse = LoginResponse{
		SScert:         []AccountResponse{account},
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

func saveClientKey(key PublicKeyForLogin) *SafeEncrypt {
	keyjson, _ := json.Marshal(key)
	fmt.Print("客户端公钥：")
	fmt.Println(string(keyjson))
	var client SafeEncrypt
	client.SetPublicKey(string(keyjson))

	return &client
}

//resp = "eyJhbGciOiJSU0EtT0FFUCIsImVuYyI6IkEyNTZDQkMtSFM1MTIiLCJ0eXAiOiJKV0UifQ.byAh5gF451IU-P8E6jrGl5yfV-hVKXgT87XI3WolmcAO63StYSHFJNmJ2bPjw7LDc2c5qixUX0eWI3tTAFz5dw_oar4-dVMdZ_-0Fv7Uz-J_UfhHDp6mohMUxd-IA6a9vI7NUpGl2PHNiWhgE83mt5L2aHc0xGZ3gU5v_fhluwELkEhNN08DfxIG3-mz16DovOpQTTEOev7ZMZ3nsP6hq4u2TPC9GsXkgb5Y489dcuxSqnPgeQBASOQrsy9uoRQqtfZWcZ_wHib-gjX4etgEMm-_yScCpeqVp74jRrrepo8H8HQuB1uDtPfoExzom5Yueao1uNBwRFtu9fbD6imbpA.25MJxWnsg9Q5alPXWz1D_g.Yuf5yv_WniHfKM67pcorKzl06oTnJeqKbFcCDYpWL4AFce1k_ULoo4zxQoN_DfDsRbdnnVv-EjIvbwhSv_AONWtLQh1H5BPcMOG3wBo6tVU_ux-1P-VnGUqh10t9zW_YavTFKiCxknAnBr_UpQ8yhFQU_gu5Wn_fTcHx7wXxgUX83_kZXGljIBdIbnPZzfehonwrqz3BopY0IYFj-LOVXkYvaYTzOMSF1WSkT4_I-Qchletfss-BD6LQ7IYePbLz4-XNscOd4spsZx439n8fWQVtYKhBrye30pahcu5SjwT0jYZTkjJaN1os5XWt_TSjpcje-NZwWryNJzJIihlyA8lb7zgdQIFerqnZXftO_Xg6am-0hYV4YYzFUHoIYn1Lcfv__xE4gXTfwkratHquZHZEtVzoZTOp3Y9z0TB_AEG2l-RVi6t2wXcY8a9rgYG8oHSKO0YL6CPUs7uHjQtimpBjoOQwa2u6pqLRM4sS4e1Stqu-R0smN8JnnGey-mVbc_QWRuu6hcwJoxKkxTwjTd5hwdYuqgioWP4Drfkwvexq3oOxjaxXqwsyViW0SV8o21ROEXe29f1C7c3YCE59riOJlUqq__xK6B5BM81fR4o_5LJ5omyujkh5w-klPqjMpPWe0P4_dDZSw2Hwoktn3S3BVtLCi1dt5J01v3tt_BY2LTLAcP_KmppPOcFBnwyscg9gbaFUq23eViKyNOzj0X1SaKmOET29IQDk4YfmbA9K92oN8zvKS8tze1xs6wotvS1R7FA-DFddni6ArSYLwkkAk_nWsr_7TNOGSzhraFhl0ToQQtlpK-qcNOTuRYkBx6DgxNDRzZfWa1iJTIjYdSKGHymLBccENBvlC9hHIcTtptIWhqQtlV2r2VmzLiO1P4O3fOZgwOJDr_VJFTJGSCb-UwRFvq6MGcoGYjyS_zORv4g2LxRjQ4ELxAJ0AvN75MNSH-t01HdLHpjsXh0YniRaFu5iFQTbi42YMzRZ7Y76CD9V4t_kaylAvzDcJGEX.nB8JmVVibjdHnB_16V5YluXAiTDgInKKkJxOhiHlOGk"
//resp = "eyJhbGciOiJSU0EtT0FFUC0yNTYiLCJlbmMiOiJBMjU2Q0JDLUhTNTEyIn0.NGXu4j92bfGuTMQjXvSnihWV-OYbYLOuilyScVDBWf_IpWUH42SDfSH-ptEfbYt28Ur4Rv9aeXi0_uGw64vAVlZJWj6RljknPP--YcSB7E3Jk0y-Imd_v_8twD5HyA2MuOuJjz26lxZHB94kcEnUxegy0MH5dgBRUc1IYd8e_6k--gTnknwOARWxromfTNzx6Jd3RCJAOqFH-PIDdmtv_neAouiuS4KeRbDy6sdZ3cCv6KPjHd_2w3HMcVBvvvCnOLrJvEp_yskYHT1_mWpU9fzc0-os3najEMqT7ObakzQUzU2ZRbV7YJrehFRbGLJsCtpctc5nEZvK_r_MmipCQw.yZlkwhHKPwzOBPP7rU2WrA.V6cmzim_T_W2iMd4v3ZDg-XMPKOa7BzRsruquCebSqRYxqE0hneWJ7gvVuN7MeFb3KEJ7QlEZCxkrUMjm-2B7xRyPIaQ6rpIW8X6JL1esc6CriFtH8vFCm2Z_NdJda-TW0IEMs6qbx9jHDXoGXdfNVMqUjEsZH1vq2ZyZ9kOBbRf89JQzBXgDMmJyKDr6nqqV-uGWbWVZkgstBjciYQsqgwugDgEoH9NdcH0jEG_X4P_Y9bjhaeiK8MDKVykOt34pxCde4RQpTDyCj3mFEOQ9VDWr6XDV2XuVw6b2wMsnstsBeT0JNtOb8sizUS4C-kHEFSIAsWLzG0uTt0I4FX_Z2lOg0e6iEFTo9yEYcYdM7NGq0cbvbk-sk49VD_SmD7urT_XjTgz_eO8eQQYmK8rMz2FwYbYfQau-PxFuhGPkcw7Ro1TWaK04GJUN4E1uwPaPjP3q84e_OlGg4x8V3FRfV_T5hXjNsQCTUK1DzQjqBr6gehcvTjnE32jfO8AhJzP6DzyZgqengpIv00lfg26UPxHHyP9XnQ64_nQRNcKedreaKOkRYNfcyTbP5huyczddIOQqysBrqtLI_7cTyTazPYKXR5qcLeeGFJ2tseepT-P5LqYwgndeaYO25MWIAKpRCB1-zCobEvKS5huZBbbHSh-BXBQWD2NSn-gydK4MWwBebGIVBhHjM8_HRpExKQzeRp_Dz9Fowf_th3SK6UmDC03saCN84j-U1-aAhPywm_WHv3OCw5IhNBhstkJz0znE4D84fUT8TbTWlpUXhj7SAtJY7BJ4QgejS-q51tiCIXQVmXUl_sru-OSLZytE6MEYcAsqTY3k6JlQ9wLW2nOleDs0dphn3klSISt5wjAQvlCyvvdHxNJ8xMjyhnym-Yk.tQRrU7w4652eKfGlHrd2Fg"

/*
	prikkey := "{ \"kid\": \"hello\", \"d\": \"iQ5AH2OpBYqkEbjlmBbtmxRd1woFpNgCwwAa9yS74tLfoL6XEv21DBtY60-tlUGe8CpE_kC0eLlQSlQyipiKtiDffPEYROEDkW4AHvy9p2Y0I2Y2bnHgBaIrofBrCMO4kRoV_fUeAOiAU1WGIErKh5WPs6cFG2Vln-CGcOQ8tFyG-PCpwrhNlQJP5XnkbL56H9e9R-AnuKVSDm3TFM_gtKGfMpEM1g3Y7rK9LoV-l24FyEBFGth5wcr2Mb4ZxOEYslKAxhztelaU0WFnzguByMkgZAmfJDGy64XsZzdEbYqM5aPhCMEyFIQ3BVcTtr_llbo94gm6osB5klewTpD3AQ\", \"dp\": \"O6BCOMXuTCuigTJPl2-jW3udnu4kbJsoItxP2Izg914jJbEjSz6BGvezg1MXZ3eiAemhHeh5vMi0YGtw75lKIWdsHPARi2lhhjHKVd7ZPwTViK61l8L86azOL2T7O8Y244kNCXisgUmLpRyZi20QfUTIV8cI0j7hQiZEXU9M35E\", \"dq\": \"af5cgcHR9IMBZiChaHKcA4zMgzTxiLs-XcDZiLADGAdmh-sI1WJEZug38z6QrzCm84neo7wA4PjWgJV0PxP0FtfR67krlIa4YhY6J9ZzOXm0qYURwGpMhjxLTqvWlgBxz-Qj7rT0yedyxEoX0FwwTr8UDoyt6UaJhb79zaqMxkE\", \"e\": \"AQAB\", \"kty\": \"RSA\", \"n\": \"yfTdA3PJzyo2LT899lspEA8QESzDDMO-rRu7VtIKCSgBliCCI4etuVHU--2h3JXFHaNttfFTm2uA1ps1G86q90tvTol7MBqSZgD51dlBP_V5vDtdkmi_sS38Szfz_Qu6CS2-JxDmLdrBPJAXJjOW13Sz6e-cjT3TCsmV3n4zqyYbLSHOBdLpLAcH6JWtqsJE8bED68C5EZzfBtWoOXWtsNJgsGIQ2hugdh-VAm4lCPaSIOdQKGczfvSRFcsTMCKox9norysq3rfKqh0b_UBfcBDhFtFnVrCFGVtsbKV9OIf404J_QOpH8pjeyAVcocBN4TzJilrOfiDlcSvaztkaEQ\", \"p\": \"7y3gXWO794Gf3jr9jTptKvLC5kSOsY3w_UfWdZdq5dGYtRNJzpOriS3S26hL4MWz_446WBWuKjgc-VJ4sul-vd6uHpa2QOpcMJApvKRSktHZChA0v8wtc1HCNGJRLV5mgKw-FrcrjycX08EteMlAT45Es56cZYfHm5Xt0T0MnIk\", \"q\": \"2CjWpAaHqoPWxrlBxu5BTsaUnc4S-QchZuqU0d9WSiMKIKRMu7UBSLAjZbMsh18Rr7UzByih2RM99MBfgJtlm0Z9kxyLjGnkqqfVKJms5YcnmfvBltDAmAAjxpK_Exxr9bkp5IoxUTmFWeWVyHXG27u0Zfj2K18K3jpG1xa0_0k\", \"qi\": \"bx8lw3-6-LqNzlMbp4VXsHAMUQ0LdNGwFtV-2-wemd4MMVLB7BawY73AWPaaaMOpyVBuqtfvxWnCspxm2HDUTrLsfJzswDOmGCJzzaFSSoyJhdX03dS_bhALrTdyX958Y58hSv6Pt1sBdZd55KjZ4aiNV3Jatr3pbtwvFgqc1N0\" }"
		client.SetPrivateKey(prikkey)
		dec_content := JWE_Decrypt(client, resp)
		fmt.Println("解密结果")
		fmt.Println(dec_content)
		fmt.Println(resp)
*/

//resp = "eyJhbGciOiJSU0EtT0FFUCIsImVuYyI6IkEyNTZDQkMtSFM1MTIiLCJ0eXAiOiJKV0UifQ.bu0WV3GRLw-1Biy3x8fzFBZfH0k_x5XXXzZ-BnAqxiOOe_MoBfA5lQgSbbFMp0LmCmCPc7oB0hyKzdj8RwEoTYvJQxReCFE5sqm_O7PBjHB0mPgsqdArIoumvTTEoLXTCuovPi83B-YB5jgQApS5-q02yTcL8YHR4QXzIcDCzMj5skQ2oxjVjV0HY-pbTnv4F9Rds-RgLKwsyP4mkQpaMNASyhnMYhXosR9uxuaTs8zcCnPneO4PpCeNAzlo4PrHiVzOz8BziehIfJfPYLB7yVCz6DgqiJrLoobAShp36Q0NSld7Diy-bXofHlW5sL6LUc9HsndiM8N5yW7XcmEA2w.BoDqHsiIUjJWRGDbvehntQ.J7yK_cLxdGF6TX-z1_bhJfYUJLQjtAEusZ3760L31pwRR4BnYjgSO8UKI5MlG-G31kQLmkzyOBP6RCNLeOPdcfx1zE_O53d75j4cxujaFEFcDGcDbqX4zo_VdfzCGRuEdlj4x_oB0PhjvTzQYHCzeVPtqGuf79oiW8E0yNqtlLfKnYSJmLKd-t8ZI5om_QjKQGufNJ4B_96YAKN9dW6pWIxOhyp4ocxR3y3SjWHIMXYycz9JcXRp9DfRAt4v-gou2IVNxd7kUzctGReyFo5MHh9zNMMto4G0WN3lMDtTe61pMHF9ecQuNVW9Xg6aBLpxXrmOyMWrD4s_BnuomcFpCnfKjBGym3pmzouAvx0dG33qBWjMoQzYl2jLtPwd8zDkjemFCqAc28gpJMh2GurrMobVnP7k_KN9DMPpgltFBPXszX-XbD6VqI3-nxN2oV0d27xz8-Qr1S1rfrayEs4gPedC3DhCU1RdhAXHqM69-1Js96weVvMX5OdO0uyQzb6fHqLY_zZa5dzTfcYCGrFWrcvcKLlZBA_tAVHJWvRByZ6DA_QB-wJBs8SsEo38OIru6fY5PXO3b5NlnKvzTA8Wn9xqZOOf7fnelOGW1BOnWQtcFCLuSCbemB6F1-TlnMzvaMFejV8c92xGRUlfnD9AjHK1T2o7dXTUFTZ4Tohkj4s52oSMYszjECs4GUedpsj4jzL_nJDHhX5YWtHbqWllh1PwCw7yW4d2H_wJoupFYkFFiHgN8wmB1roYX0LAJ6kML81L7Snka6KqPsg2WjWQ_zd0-ZiD4QWYVkUQomhE0FMlJV8CgTT1MLtEZBqHXhtbHF1R-BIbaVVkQ7GiNZ_5ZNN3obaRPZGNFZORqDXJqB7HOA_5bH7CBcqveUnLZGfZ.0VbUNyDoqZiqsdg4n73w-Q"
/*
	var data *rsa.PublicKey
	ero := json.Unmarshal(keyjson, data)
	if ero != nil {
		fmt.Println(ero)
	}
	encrypter, err := b.NewEncrypter(b.A256CBC_HS512, b.Recipient{Algorithm: b.RSA_OAEP_256, Key: data}, nil)

	if err != nil {
		fmt.Println(err)
	}

	var plaintext = []byte(ResponseForLogin(&client, se, res))
	object, err := encrypter.Encrypt(plaintext)
	if err != nil {
		panic(err)
	}

	ret, _ := object.CompactSerialize()
	fmt.Fprintf(w, ret)
*/

func FindAccountInfo(key string) AccountResponse {
	ip, port, password, method, atype := accountManager.GetAccountDetail(key)
	fmt.Println("查找的结果:", key, ip, port, password, method, atype)
	return AccountResponse{
		Type:    atype,
		Address: ip,
		Port:    port,
		Key:     password,
		Method:  method,
	}
	/*
		return AccountResponse{
			Type:    "a",
			Address: "localhost",
			Port:    "1324",
			Key:     "8887",
			Method:  "hello",
		}
	*/
}

type PriceHandler struct {
	config eosapi.EosConfig
}

type LoginHandler struct {
	config eosapi.EosConfig
}

func (h *PriceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 解决跨域的参数
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")

	response := ResultToClientForPrice(h.config.Address)
	fmt.Fprintf(w, response)
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//func handleLogin(w http.ResponseWriter, r *http.Request) {
	// 解决跨域的参数
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")

	r.ParseForm()
	input := r.Form["code"][0]
	fmt.Println("请求的加密数据:", input)

	inputJson := decryptInputParamForLogin(se, input)
	fmt.Println("请求的解密数据:", inputJson)
	if inputJson == "" {
		http.Error(w, "解密数据失败", http.StatusBadRequest)
		return
	}

	var payload LoginPayload
	if err := json.Unmarshal([]byte(inputJson), &payload); err != nil {
		fmt.Println("解密数据无法序列化为json")
		http.Error(w, "解密数据无法序列化为json", http.StatusBadRequest)
		return
	}

	client := saveClientKey(payload.Key)

	// 校验客户端的时间戳是否正确
	signjson, _ := json.Marshal(payload.Signature)
	timestamp, _ := JWS_Verify(client, string(signjson))
	clientTimeStamp, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		fmt.Println("客户端时间戳转化失败")
	}
	currentTs := time.Now().Unix()
	fmt.Println("时间戳校验:")
	fmt.Println(currentTs)
	fmt.Println(clientTimeStamp)
	tsDiff := currentTs - clientTimeStamp
	if tsDiff >= 60 && tsDiff < 0 {
		fmt.Println("客户端时间戳校验失败")
		http.Error(w, "客户端时间戳校验失败", http.StatusBadRequest)
		return
	}

	memo := Base58Encoder(SHA256(string(SHA256(payload.Key.N))))
	fmt.Println("用户的key是:", memo)
	account := FindAccountInfo(memo)
	if account.Address == "" {
		fmt.Println("账户信息没有找到")
		http.Error(w, "账户信息没有找到", http.StatusBadRequest)
		return
	}
	resp := responseForLogin(client, se, account)
	fmt.Fprintf(w, resp)
}

func HttpServer(httpConfig HttpConfig, config eosapi.EosConfig) {
	server := http.Server{
		Addr: httpConfig.Ip + ":" + httpConfig.Port,
	}

	priceHandler := PriceHandler{}
	priceHandler.config = config
	loginHandler := LoginHandler{}
	loginHandler.config = config

	http.Handle("/price.json", &priceHandler)
	http.Handle("/cert.info", &loginHandler)

	server.ListenAndServe()
}
