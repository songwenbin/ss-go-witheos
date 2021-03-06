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
	atype   string
}

var accountManager *AccountManager = NewAccountManager()

func init() {
	//accountManager.Add("ok")
}

func AccountMangerFactory() *AccountManager {
	return accountManager
}

func NewAccountManager() *AccountManager {
	return &AccountManager{
		mapper: map[string]*AccountInfo{},
	}
}

func (am *AccountManager) Add(publickey string, ip string, method string) {
	am.Lock()
	am.mapper[publickey] = &AccountInfo{
		ip:      ip,
		port:    RandomPort(),
		key:     GeneratePassword(),
		methods: method,
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

func (am *AccountManager) GetAccountDetail(memo string) (string, string, string, string, string) {
	am.Lock()
	value, exists := am.mapper[memo]
	am.Unlock()
	if exists == false {
		return "", "", "", "", ""
	} else {
		return value.ip, value.port, value.key, value.methods, value.atype
	}
}

func RandomPort() string {
	s := strconv.Itoa(rand.Intn(1000) + 7000)
	return s
}

func GeneratePassword() string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, 20)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
