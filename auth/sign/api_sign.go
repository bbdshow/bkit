package sign

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bbdshow/gocache"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

/*
1. 签名头X-Signature=accessKey:signStr:timestamp (:分割不同元素)
2. 签名方式： HmacSha1ToBase64(rawStr+timestamp, secretKey) 最终signStr是Base64编码
3. rawStr 解释 如果是 GET请求，例子: http://example.com/hello?n=1&a=2  => /hello?a=2&n=11626167650 参数要进行字符正序
4. 非GET请求，对于Content-Type: application/json {"n":"m","a":2} 也进行字符正序 => {"a":2,"n":"m"} 最终 a=2&n=m
5. rawStr+timestamp = /hello?a=2&n=m1626167650 时间戳(秒)  最终在签名,携带时间戳，是为了验证签名时间有效性 当前有效性 -+10s
6. 最终形式都是  path?/k=v&k1=v1 + timestamp
*/
// APISign 接口签名
type APISign struct {
	cache        gocache.Cache
	getSecretKey func(accessKey string) (string, error)
}

var (
	defSignValidTime = 5 * time.Second
)

// signValidDuration sign valid time interval
func NewAPISign(signValidDuration time.Duration) *APISign {
	if signValidDuration > defSignValidTime {
		defSignValidTime = signValidDuration
	}
	sign := &APISign{
		cache:        gocache.NewRWMapCache(),
		getSecretKey: nil,
	}
	return sign
}

// Verify 验证签名
func (sign *APISign) Verify(req *http.Request, header string) error {
	val := req.Header.Get(header)
	if val == "" {
		return fmt.Errorf("sign header required")
	}
	accessKey, signStr, timestamp, err := sign.decodeHeaderVal(val)
	if err != nil {
		return err
	}
	secretKey, err := sign.secretKey(accessKey)
	if err != nil || secretKey == "" {
		return fmt.Errorf("sign access key invalid")
	}

	rawStr := ""
	switch strings.ToUpper(req.Method) {
	default:
		// 没有签名
		return nil
	case http.MethodPost, http.MethodPut, http.MethodDelete:
		ct := filterFlags(req.Header.Get("Content-Type"))
		switch strings.ToLower(ct) {
		case "application/json":
			byt, err := ioutil.ReadAll(req.Body)
			if err != nil {
				_ = req.Body.Close()
				return err
			}
			_ = req.Body.Close()

			reqBody := make(RequestBodyMap)
			if err := json.Unmarshal(byt, &reqBody); err != nil {
				return err
			}

			rawStr = fmt.Sprintf("%s?%s", req.URL.Path, reqBody.SortToString("&"))

			req.Body = ioutil.NopCloser(bytes.NewBuffer(byt))
		case "multipart/form-data":
			rawStr, err = sortParamForm(req, true)
			rawStr = fmt.Sprintf("%s?%s", req.URL.Path, rawStr)
			if err != nil {
				return err
			}
		}

	case http.MethodGet:
		rawStr, err = sortParamForm(req, true)
		if err != nil {
			return err
		}
	}
	rawStr = rawStr + timestamp
	signStr1 := HmacSha1ToBase64(rawStr, secretKey)
	if signStr1 != signStr {
		return fmt.Errorf("sign method invalid rawStr:%s", rawStr)
	}
	return nil
}

// decodeHeaderVal 签名头类容格式 accessKey:signStr(HmacSha1ToBase64(rawStr+timestamp, secretKey)):timestamp
func (sign *APISign) decodeHeaderVal(headerVal string) (accessKey, signStr, timestamp string, err error) {
	strs := strings.Split(headerVal, ":")
	if len(strs) != 3 {
		return "", "", "", fmt.Errorf("sign header invalid")
	}
	accessKey = strs[0]
	signStr = strs[1]
	timestamp = strs[2]
	i, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return "", "", "", fmt.Errorf("sign timestamp invalid")
	}
	t := time.Unix(i, 0)
	if t.Before(time.Now().Add(-defSignValidTime)) || t.After(time.Now().Add(defSignValidTime)) {
		return "", "", "", fmt.Errorf("sign timestamp invalid")
	}
	return accessKey, signStr, timestamp, nil
}

// SetGetSecretKey 设置SecretKey 获取方法
func (sign *APISign) SetGetSecretKey(f func(accessKey string) (string, error)) {
	sign.getSecretKey = f
}

// secretKey
func (sign *APISign) secretKey(accessKey string) (string, error) {
	v, exists := sign.cache.Get(accessKey)
	if !exists {

		if sign.getSecretKey == nil {
			return "", fmt.Errorf("sign getSecretKey function not init")
		}
		key, err := sign.getSecretKey(accessKey)
		if err != nil {
			return "", err
		}
		_ = sign.cache.SetWithExpire(accessKey, key, 600)
		return key, nil
	}
	if vv, ok := v.(string); ok {
		return vv, nil
	}
	return "", fmt.Errorf("accessKey invalid")
}

// sortParamForm URL 和 form-data 参数
func sortParamForm(req *http.Request, path bool) (string, error) {
	resource := req.URL.Path
	switch filterFlags(req.Header.Get("Content-Type")) {
	case "multipart/form-data":
		err := req.ParseMultipartForm(10 << 20)
		if err != nil {
			return "", err
		}
	default:
		err := req.ParseForm()
		if err != nil {
			return "", err
		}
	}

	var paramNames []string
	if req.Form != nil && len(req.Form) > 0 {
		for k := range req.Form {
			paramNames = append(paramNames, k)
		}
		sort.Strings(paramNames)

		var query []string
		for _, k := range paramNames {
			query = append(query, url.QueryEscape(k)+"="+url.QueryEscape(req.Form.Get(k)))
		}
		if path {
			resource = resource + "?" + strings.Join(query, "&")
		} else {
			resource = strings.Join(query, "&")
		}
	}
	return resource, nil
}

type RequestBodyMap map[string]interface{}

func (r RequestBodyMap) GetStringValue(key string) (string, error) {
	val, ok := r[key]
	if !ok {
		return "", fmt.Errorf("request body miss %s key", key)
	}
	v, ok := val.(string)
	if !ok {
		return "", fmt.Errorf("request body %s key not string type", key)
	}
	return v, nil
}

func (r RequestBodyMap) SortToString(separator string) string {
	if len(r) == 0 {
		return ""
	}
	kvs := make(KvSlice, 0)
	for k, v := range r {
		kvs = append(kvs, Kv{Key: k, Value: v})
	}

	sort.Sort(kvs)

	var s = make([]string, 0, len(kvs))
	for _, v := range kvs {
		switch v.Value.(type) {
		case float64:
			s = append(s, fmt.Sprintf("%s=%v", v.Key, decimal.NewFromFloat(v.Value.(float64)).String()))
		case float32:
			s = append(s, fmt.Sprintf("%s=%v", v.Key, decimal.NewFromFloat(float64(v.Value.(float32))).String()))
		case *float64:
			s = append(s, fmt.Sprintf("%s=%v", v.Key, decimal.NewFromFloat(*v.Value.(*float64)).String()))
		case *float32:
			s = append(s, fmt.Sprintf("%s=%v", v.Key, decimal.NewFromFloat(float64(*v.Value.(*float32))).String()))
		default:
			s = append(s, fmt.Sprintf("%s=%v", v.Key, v.Value))
		}
	}
	return strings.Join(s, separator)
}

type Kv struct {
	Key   string
	Value interface{}
}
type KvSlice []Kv

func (s KvSlice) Len() int           { return len(s) }
func (s KvSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s KvSlice) Less(i, j int) bool { return s[i].Key < s[j].Key }

func filterFlags(content string) string {
	for i, char := range content {
		if char == ' ' || char == ';' {
			return content[:i]
		}
	}
	return content
}
