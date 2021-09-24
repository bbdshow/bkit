package sign

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"testing"
	"time"
)

func testHttpServer() {
	cfg := Config{
		SignValidDuration: 10 * time.Second,
		Method:            HmacSha1,
		PathSign:          true,
	}
	sign := NewAPISign(&cfg)
	sign.SetGetSecretKey(func(accessKey string) (string, error) {
		vals := map[string]string{
			"abc": "abc_secretKey",
			"efg": "efg_secretKey",
		}
		return vals[accessKey], nil
	})

	http.HandleFunc("/sign", func(writer http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case "GET":
			msg := request.Method
			err := sign.Verify(request, "Authorization")
			if err != nil {
				msg += err.Error()
			}
			writer.WriteHeader(200)
			writer.Write([]byte(msg))
		case "POST", "PUT":
			msg := request.Method
			err := sign.Verify(request, "Authorization")
			if err != nil {
				msg += err.Error()
			}

			// 读出Body
			bodyData, _ := ioutil.ReadAll(request.Body)
			defer request.Body.Close()

			writer.WriteHeader(200)
			writer.Write([]byte(msg + string(bodyData)))
		}
	})
	http.ListenAndServe(":8080", http.DefaultServeMux)
}

func TestAPISign_Verify(t *testing.T) {
	go func() {
		testHttpServer()
	}()
	time.Sleep(time.Second)
	path := "http://localhost:8080/sign?2=2&1=1"
	pointVal := float32(40.5)
	type Add struct {
		Address string `json:"address"`
		No      int
	}
	b := struct {
		Name    string
		Age     int
		Balance float64
		Point   *float32
		Adds    []Add
	}{Name: "nice", Age: 5, Balance: 102.22222, Point: &pointVal, Adds: []Add{{Address: "XX", No: 88}, {Address: "AA", No: 77}}}
	byts, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	ts := time.Now().Unix()
	timestamp := fmt.Sprintf("%d", ts)

	//signStr := getMethodSign("/sign", timestamp, map[string]interface{}{"2": 2, "1": 1})
	_url, err := url.Parse(path)
	if err != nil {
		t.Fatal(err)
	}
	signStr := jsonSign(_url.Path, timestamp, bytes.NewBuffer(byts))
	req, err := http.NewRequest("POST", path, bytes.NewReader(byts))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("Authorization", fmt.Sprintf("%s:%s:%s", "abc", signStr, timestamp))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	respVal, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(respVal))
}

func getMethodSign(path, timestamp string, kv map[string]interface{}) string {
	keys := make([]string, 0)
	for k, _ := range kv {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	vals := make([]string, 0)
	for _, k := range keys {
		vals = append(vals, fmt.Sprintf("%s=%v", k, kv[k]))
	}
	str := path + "?" + strings.Join(vals, "&") + timestamp
	fmt.Println(str)
	return HmacSha1ToBase64(str, "abc_secretKey")
}

func jsonSign(path string, timestamp string, body io.Reader) string {
	b := make(RequestBodyMap)
	byts, err := ioutil.ReadAll(body)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(byts, &b); err != nil {
		panic(err)
	}
	str, _ := b.SortToString("&")

	str = path + "?" + str + timestamp
	fmt.Println(str)
	return HmacSha1ToBase64(str, "abc_secretKey")
}

type mockSignStruct struct {
	Txt               string         `json:"txt"`
	Integer           int            `json:"integer"`
	SliceBool         []bool         `json:"sliceBool"`
	SliceString       []string       `json:"sliceString"`
	Float             float64        `json:"float"`
	SliceFloat        []float64      `json:"float64s"`
	SliceCustomStruct []CustomStruct `json:"sliceCustomStruct"`
	CustomStruct      CustomStruct   `json:"customStruct"`
}
type CustomStruct struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestSortToString(t *testing.T) {
	obj := mockSignStruct{
		Txt:         "Mock",
		Integer:     100,
		SliceBool:   []bool{true, false},
		SliceString: []string{"A", "B"},
		Float:       9.99,
		SliceFloat:  []float64{8.88, 9.99},
		SliceCustomStruct: []CustomStruct{
			{
				Name: "SliceCustomStruct",
				Age:  10,
			},
			{
				Name: "SliceCustomStruct",
				Age:  20,
			},
		},
		CustomStruct: CustomStruct{
			Name: "CustomStruct",
			Age:  10,
		},
	}
	str, _ := json.Marshal(obj)
	body := RequestBodyMap{}

	if err := json.Unmarshal(str, &body); err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(str))

	ts := fmt.Sprintf("%d", time.Now().Unix())
	sortStr, _ := body.SortToString("&")

	rawStr := sortStr + ts
	fmt.Println(rawStr)

	signed := HmacHash(HmacSha256Hex, rawStr, "lc3pptr2g2sumgvcbt5yw5g3e0tf8oni")

	fmt.Println(signed)
}
