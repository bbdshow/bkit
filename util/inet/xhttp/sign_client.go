package xhttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bbdshow/bkit/auth/sign"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type SignConfig struct {
	Method    sign.Method
	Header    string
	AccessKey string
	SecretKey string
}

func (cfg SignConfig) Validate() error {
	if cfg.AccessKey == "" || cfg.SecretKey == "" {
		return fmt.Errorf("accessKey or secretKey required")
	}
	if cfg.Header == "" {
		return fmt.Errorf("header required")
	}
	switch cfg.Method {
	case sign.HmacSha256Hex, sign.HmacSha256, sign.HmacSha1Hex, sign.HmacSha1:
	default:
		return fmt.Errorf("not support sign method %s", cfg.Method)
	}
	return nil
}

type SignClient struct {
	cfg    *SignConfig
	client *http.Client
}

func NewSignClient(cfg *SignConfig) *SignClient {
	if err := cfg.Validate(); err != nil {
		panic(err.Error())
	}

	c := &SignClient{
		cfg:    cfg,
		client: &http.Client{},
	}
	return c
}

func (c *SignClient) sign(req *http.Request, body []byte) error {
	rawStr := ""
	ts := fmt.Sprintf("%d", time.Now().Unix())
	switch req.Method {
	case http.MethodGet:
		str, err := sign.SortParamForm(req)
		if err != nil {
			return err
		}
		rawStr = str + ts
	case http.MethodPost:
		if body != nil {
			reqBody := &sign.RequestBodyMap{}
			if err := json.Unmarshal(body, reqBody); err != nil {
				return err
			}
			str, err := reqBody.SortToString("&")
			if err != nil {
				return err
			}
			rawStr = str + ts
		}
	}
	signStr := sign.HmacHash(c.cfg.Method, rawStr, c.cfg.SecretKey)

	req.Header.Set(c.cfg.Header, fmt.Sprintf("%s:%s:%s", c.cfg.AccessKey, signStr, ts))

	return nil
}

func (c *SignClient) newRequest(method, _url string, params url.Values, body []byte, headers map[string]string) (*http.Request, error) {
	var (
		req *http.Request
		err error
	)
	switch method {
	case http.MethodGet:
		enc := params.Encode()
		if enc != "" {
			_url = _url + "?" + enc
		}
		req, err = http.NewRequest(method, _url, nil)
	case http.MethodPost:
		req, err = http.NewRequest(method, _url, bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
	default:
		return nil, fmt.Errorf("not support method %s", method)
	}
	for k, v := range headers {
		if req != nil {
			req.Header.Set(k, v)
		}
	}
	if err := c.sign(req, body); err != nil {
		return nil, err
	}
	return req, err
}

func (c *SignClient) do(req *http.Request, res interface{}) error {
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	byt, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(byt, res); err != nil {
		return fmt.Errorf("respBody %s  %v", string(byt), err)
	}
	return nil
}

// Get  response header Content-Type application/json
func (c *SignClient) Get(url string, params url.Values, headers map[string]string, res interface{}) error {
	req, err := c.newRequest(http.MethodGet, url, params, nil, headers)
	if err != nil {
		return err
	}
	return c.do(req, res)
}

// Post Content-Type application/json
func (c *SignClient) Post(url string, body []byte, headers map[string]string, res interface{}) error {
	req, err := c.newRequest(http.MethodPost, url, nil, body, headers)
	if err != nil {
		return err
	}
	return c.do(req, res)
}
