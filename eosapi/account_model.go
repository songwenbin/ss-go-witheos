package eosapi

type EosAccount struct {
	Purchaser string `json:"purchaser"`
	Eospaid   string `json:"eospaid"`
	PaidTime  int    `json:"paid_time"`
	Memo      string `json:"memo"`
}
