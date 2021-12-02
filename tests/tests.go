package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

// BeautifyToJSON  v 编码成有缩进的JSON
func BeautifyToJSON(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(b)
}

// PrintBeautifyJSON 打印 BeautifyToJSON
func PrintBeautifyJSON(v interface{}) {
	fmt.Println(BeautifyToJSON(v))
}

func EncodeJSONToReader(v interface{}) io.Reader {
	byt, _ := json.Marshal(v)
	return bytes.NewReader(byt)
}

// Post  Only support Request/Response Content-Type application/json
func Post(url string, headers map[string]string, in, out interface{}) error {
	if headers == nil {
		headers = map[string]string{}
	}
	if _, ok := headers["Content-Type"]; !ok {
		headers["Content-Type"] = "application/json"
	}
	resp, err := request(http.MethodPost, url, headers, EncodeJSONToReader(in))
	if err != nil {
		return err
	}
	err = json.Unmarshal(resp, out)
	return err
}

// Get  Only support Request url values, Response Content-Type application/json
func Get(url string, headers map[string]string, out interface{}) error {
	resp, err := request(http.MethodGet, url, headers, nil)
	if err != nil {
		return err
	}
	err = json.Unmarshal(resp, out)
	return err
}

func request(method, url string, headers map[string]string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	byt, err := ioutil.ReadAll(resp.Body)
	return byt, err
}
