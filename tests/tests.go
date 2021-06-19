package tests

import (
	"encoding/json"
	"fmt"
)

// BeautifyToJSON 常用于测试环境打印结构体数据
func BeautifyToJSON(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(b)
}

// PrintBeautifyJSON
func PrintBeautifyJSON(v interface{}) {
	fmt.Println(BeautifyToJSON(v))
}
