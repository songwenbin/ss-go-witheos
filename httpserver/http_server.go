package httpserver

import (
	"fmt"
	"net/http"
)

func handlePrice(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "handlePrice interface")
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
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
