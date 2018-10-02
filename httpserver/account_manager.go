package httpserver

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
)

type AccountManager struct {
	sync.Mutex
	mapper map[string]*AccountInfo //string == publickey
}

type AccountInfo struct {
	ip      string
	port    string
	key     string
	methods string
}

func NewAccountManager() *AccountManager {
	return &AccountManager{
		mapper: map[string]*AccountInfo{},
	}
}

func (am *AccountManager) Add(publickey string) {
	am.Lock()
	am.mapper[publickey] = &AccountInfo{
		ip:      "localhost",
		port:    RandomPort(),
		key:     RandomPassword(),
		methods: "aes_cbc_cfb_256",
	}
	am.Unlock()
}

func (am *AccountManager) GetPortAndPassword(publickey string) (string, string) {
	am.Lock()
	value, exists := am.mapper[publickey]
	if exists == false {
		am.Unlock()
		return "", ""
	} else {
		am.Unlock()
		return value.port, value.key
	}

}

func (am *AccountManager) Get(publickey string) *AccountInfo {
	am.Lock()
	value, exists := am.mapper[publickey]
	if exists == false {
		fmt.Println("ok")
		am.Unlock()
		return nil
	} else {
		am.Unlock()
		return value
	}
}

func RandomPort() string {
	s := strconv.Itoa(rand.Intn(1000) + 7000)
	return s
}

func RandomPassword() string {
	return "foobar"
}
