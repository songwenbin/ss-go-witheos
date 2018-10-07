package eosapi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type TableRows struct {
	Rows []struct {
		Purchaser string `json:"purchaser"`
		Eospaid   string `json:"eospaid"`
		Paid_time int    `json:"paid_time"`
		Memo      string `json:"memo"`
	} `json:"rows"`
	More bool `json:"more"`
}

type AppError struct {
	Error   error
	Message string
	Code    int
	Custom  interface{}
}

func NewAppError(error error, message string, code int64, custom interface{}) *AppError {
	return &AppError{
		error,
		message,
		int(code),
		custom,
	}
}

type HTTPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Error   struct {
		Code    int    `json:"code"`
		Name    string `json:"name"`
		What    string `json:"what"`
		Details []struct {
			Message    string `json:"message"`
			File       string `json:"file"`
			LineNumber int    `json:"line_number"`
			Method     string `json:"method"`
		} `json:"details"`
	} `json:"error"`
}

func HTTPErrorTOJSON(error HTTPError) (string, *AppError) {

	json, err := json.Marshal(error)

	if err != nil {
		return "", NewAppError(err, "error trying to marshal HTTPError", -1, nil)
	}

	return string(json), nil
}

func Post(url string, keyValues map[string]interface{}, bytes []byte) ([]byte, *AppError) {

	//fmt.Println("post keyValues: ", keyValues)
	//fmt.Println("post raw: " + string(bytes))

	var err error

	if keyValues != nil {
		bytes, err = json.Marshal(&keyValues)
	}

	if err != nil {
		return nil, NewAppError(err, "error marshalling params", -1, nil)
	}

	//fmt.Println(url)
	req, _ := http.NewRequest("POST", url, strings.NewReader(string(bytes)))

	req.Header.Add("content-type", "application/json")
	//req.Header.Add("Accept", "application/json")

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, NewAppError(err, "error trying to reach nodeos API", -1, nil)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return nil, NewAppError(err, "error reading body response", -1, nil)
	}

	httpError := HTTPError{}
	json.Unmarshal(body, &httpError)

	if httpError.Code != 0 {
		errStr, _ := HTTPErrorTOJSON(httpError)
		return nil, NewAppError(nil, "nodeos API returned an error: "+errStr, int64(httpError.Code), httpError)
	}

	return body, nil
}

func ChainGetTableRows(scope string, code string, table string, toJSON bool, lowerBound int, upperBound int, limit int) (*TableRows, *AppError) {

	_toJSON := "false"

	if toJSON {
		_toJSON = "true"
	}

	params := map[string]interface{}{
		"table": table,
		"scope": scope,
		"code":  code,
		"json":  _toJSON,
		//"lower_bound": lowerBound,
		//"upper_bound": upperBound,
		//"limit":       limit,
	}

	data, err := Post("http://ec2-54-95-158-74.ap-northeast-1.compute.amazonaws.com:8888/v1/chain/get_table_rows", params, nil)

	if err != nil {
		fmt.Println("=============")
		return nil, err
	}

	//fmt.Println("****************")
	//fmt.Println(string(data))
	//fmt.Println("****************")
	tableRows := TableRows{}
	errM := json.Unmarshal(data, &tableRows)

	if errM != nil {
		return nil, NewAppError(nil, "cannot parse result", -1, nil)
	}

	return &tableRows, nil
}

func GetContractMember(serverUrl string) []EosAccount {
	tableRows, err := ChainGetTableRows("incomering1", "incomering1", "purchase", true, -1, -1, -1)

	if err != nil {
		fmt.Println("err: ", err)
	}

	var result []EosAccount
	if tableRows != nil {
		//fmt.Println("tableRows: ", *tableRows)
		//fmt.Println("nb tableRows rows: ", len(tableRows.Rows))

		for _, v := range tableRows.Rows {
			//fmt.Println("==== ", i)
			/*
				fmt.Println(v.Eospaid)
				fmt.Println(v.Memo)
				fmt.Println(v.Paid_time)
				fmt.Println(v.Purchaser)
			*/

			result = append(result, EosAccount{
				Purchaser: v.Purchaser,
				Eospaid:   v.Eospaid,
				PaidTime:  v.Paid_time,
				Memo:      v.Memo,
			})

		}
	}

	return result
}

func TimerPullEosContract(eosAccountChan chan<- []EosAccount) {
	for {
		select {
		case <-time.After(5 * time.Second):
			eosAccountChan <- GetContractMember("")
		}
	}
}
