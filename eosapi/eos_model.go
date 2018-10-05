package eosapi

type EosConfig struct {
	Address string `json:"address"`
	Url     string `json:"url"`
	Scope   string `json:"scope"`
	Code    string `json:"code"`
	Table   string `json:"table"`
}
