package sign

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bbdshow/bkit/caches"
	"github.com/shopspring/decimal"
)

/*
1. Sign Header Key = X-Signature  Value = accessKey:signStr:timestamp (: split elem)
2. Sign Method： Method(rawStr+timestamp, secretKey) signStr(rawStr+timestamp) => encoding Base64 | Hex(default)
3. rawStr eg: GET http://example.com/hello?n=1&a=2  => a=2&n=11626167650 param key string attaches the methods
4. other request http method，for Content-Type: application/json {"n":"m","a":2} param key string attaches the methods => {"a":2,"n":"m"} => a=2&n=m
5. rawStr+timestamp => a=2&n=m1626167650 timestamp(sec)  verify sign time valid（default 10s）
6. if Config.PathSign = true URL.Path also need to sign, rawStr => /path?k=v&k1=v1 , default Config.PathSign = false
*/
// APISign API Sign
type APISign struct {
	cfg *Config

	cache        caches.Cacher
	getSecretKey func(accessKey string) (string, error)
}
type Method string

const (
	HmacSha256    Method = "HMAC-SHA256-BASE64"
	HmacSha1      Method = "HMAC-SHA1-BASE64"
	HmacSha1Hex   Method = "HMAC-SHA1-HEX"
	HmacSha256Hex Method = "HMAC-SHA256-HEX"
)

var (
	defSignValidDuration = 10 * time.Second
)

type Config struct {
	SignValidDuration time.Duration `defval:"10s"`
	Method            `defval:"HMAC-SHA256-HEX"`
	PathSign          bool `defval:"false"`
}

// signValidDuration sign valid time interval
func NewAPISign(cfg *Config) *APISign {

	sign := &APISign{
		cfg:          cfg,
		cache:        caches.NewLRUMemory(1000),
		getSecretKey: nil,
	}
	if sign.cfg.SignValidDuration <= 0 {
		sign.cfg.SignValidDuration = defSignValidDuration
	}

	return sign
}

// Verify sign verify
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
		// not support http method
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
			bodyStr, err := reqBody.SortToString("&")
			if err != nil {
				return fmt.Errorf("SortToString %v", err)
			}
			if sign.cfg.PathSign {
				rawStr = req.URL.Path + "?" + bodyStr
			} else {
				rawStr = bodyStr
			}
			req.Body = ioutil.NopCloser(bytes.NewBuffer(byt))
		case "multipart/form-data":
			rawStr, err = SortParamForm(req, sign.cfg.PathSign)
			if err != nil {
				return err
			}
		}

	case http.MethodGet:
		rawStr, err = SortParamForm(req, sign.cfg.PathSign)
		if err != nil {
			return err
		}
	}
	rawStr = rawStr + timestamp
	signStrDist := HmacHash(sign.cfg.Method, rawStr, secretKey)
	if signStrDist != signStr {
		return fmt.Errorf("sign method invalid rawStr:%s", rawStr)
	}
	return nil
}

// decodeHeaderVal header value decode to  accessKey:signStr:timestamp
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

	if err := sign.signValidTime(time.Unix(i, 0)); err != nil {
		return "", "", "", err
	}

	return accessKey, signStr, timestamp, nil
}

func (sign *APISign) signValidTime(t time.Time) error {
	if t.Before(time.Now().Add(-sign.cfg.SignValidDuration)) || t.After(time.Now().Add(sign.cfg.SignValidDuration)) {
		return fmt.Errorf("sign timestamp invalid")
	}
	return nil
}

// SetGetSecretKey setting SecretKey get function
func (sign *APISign) SetGetSecretKey(f func(accessKey string) (string, error)) {
	sign.getSecretKey = f
}

// secretKey
func (sign *APISign) secretKey(accessKey string) (string, error) {
	var key string
	v, err := sign.cache.Get(accessKey)
	if err != nil {
		if caches.IsNotFoundErr(err) {
			if sign.getSecretKey == nil {
				return "", fmt.Errorf("sign getSecretKey function not init")
			}
			key, err = sign.getSecretKey(accessKey)
			if err != nil {
				return "", err
			}
			_ = sign.cache.SetWithTTL(accessKey, key, 10*time.Minute)
			return key, nil
		}
		return "", err
	}
	if vv, ok := v.(string); ok {
		return vv, nil
	}
	return "", fmt.Errorf("accessKey invalid")
}

// SortParamForm URL and form-data param
func SortParamForm(req *http.Request, path bool) (string, error) {
	resource := ""
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
			resource = req.URL.Path + "?" + strings.Join(query, "&")
		} else {
			resource = strings.Join(query, "&")
		}
	}
	return resource, nil
}

func HmacHash(method Method, rawStr, secretKey string) string {
	dist := ""
	switch method {
	case HmacSha1:
		dist = HmacSha1ToBase64(rawStr, secretKey)
	case HmacSha256:
		dist = HmacSha256ToBase64(rawStr, secretKey)
	case HmacSha1Hex:
		dist = HmacSha1ToHex(rawStr, secretKey)
	case HmacSha256Hex:
		dist = HmacSha256ToHex(rawStr, secretKey)
	default:
		dist = HmacSha256ToHex(rawStr, secretKey)
	}
	return dist
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

func (r RequestBodyMap) SortToString(separator string) (string, error) {
	if len(r) == 0 {
		return "", nil
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
			s = append(s, fmt.Sprintf("%s=%s", v.Key, decimal.NewFromFloat(v.Value.(float64)).String()))
		case float32:
			s = append(s, fmt.Sprintf("%s=%s", v.Key, decimal.NewFromFloat(float64(v.Value.(float32))).String()))
		case *float64:
			s = append(s, fmt.Sprintf("%s=%s", v.Key, decimal.NewFromFloat(*v.Value.(*float64)).String()))
		case *float32:
			s = append(s, fmt.Sprintf("%s=%s", v.Key, decimal.NewFromFloat(float64(*v.Value.(*float32))).String()))
		case string:
			s = append(s, fmt.Sprintf("%s=%s", v.Key, v.Value))
		case *string:
			s = append(s, fmt.Sprintf("%s=%s", v.Key, *v.Value.(*string)))
		default:
			buf := make([]byte, 0)
			buffer := bytes.NewBuffer(buf)
			if err := json.NewEncoder(buffer).Encode(v.Value); err != nil {
				return "", err
			}
			s = append(s, fmt.Sprintf("%s=%s", v.Key, string(r.TrimNewline(buffer.Bytes()))))
		}
	}
	return strings.Join(s, separator), nil
}

func (r RequestBodyMap) TrimNewline(buf []byte) []byte {
	if i := len(buf) - 1; i >= 0 {
		if buf[i] == '\n' {
			buf = buf[:i]
		}
	}
	return buf
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
