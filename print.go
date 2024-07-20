package bkit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
