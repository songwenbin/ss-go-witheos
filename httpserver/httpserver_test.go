package httpserver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
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
