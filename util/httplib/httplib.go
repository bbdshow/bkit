package httplib

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	defaultUserAgent = "bbdshow.bkit HTTPLib"
	defaultCookieJar http.CookieJar
	settingMutex     sync.Mutex
	defaultSetting   = HTTPSettings{
		UserAgent:        defaultUserAgent,
		ConnectTimeout:   60 * time.Second,
		ReadWriteTimeout: 120 * time.Second,
		Gzip:             true,
		DumpBody:         true,
		KeepAlive:        false,
	}
)

func createDefaultCookie() {
	settingMutex.Lock()
	defaultCookieJar, _ = cookiejar.New(nil)
	settingMutex.Unlock()
}

func NewLibRequest(rawUrl, method string) *HTTPRequest {
	var resp http.Response
	u, err := url.Parse(rawUrl)
	if err != nil {
		log.Println("HTTPLib: ", err)
	}

	req := http.Request{
		URL:        u,
		Method:     method,
		Header:     make(http.Header),
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
	}

	return &HTTPRequest{
		url:     rawUrl,
		req:     &req,
		params:  map[string][]string{},
		files:   map[string]string{},
		setting: defaultSetting,
		resp:    &resp,
	}
}

// 每次都会重新生成 http.Client

func Get(url string) *HTTPRequest {
	return NewLibRequest(url, "GET")
}

func Post(url string) *HTTPRequest {
	return NewLibRequest(url, "POST")
}
func Put(url string) *HTTPRequest {
	return NewLibRequest(url, "PUT")
}
func Delete(url string) *HTTPRequest {
	return NewLibRequest(url, "DELETE")
}
func Head(url string) *HTTPRequest {
	return NewLibRequest(url, "HEAD")
}

type HTTPSettings struct {
	ShowDebug           bool
	UserAgent           string
	ConnectTimeout      time.Duration
	ReadWriteTimeout    time.Duration
	TLSClientConfig     *tls.Config
	Proxy               func(*http.Request) (*url.URL, error)
	Transport           http.RoundTripper
	CheckRedirect       func(req *http.Request, via []*http.Request) error
	EnableCookie        bool
	Gzip                bool
	DumpBody            bool
	Retries             int // 如果设置 -1 则一直重试
	KeepAlive           bool
	MaxIdleConnsPerHost int // 默认 2
}

type HTTPRequest struct {
	url     string
	req     *http.Request
	client  *http.Client
	params  map[string][]string
	files   map[string]string
	setting HTTPSettings
	resp    *http.Response
	body    []byte
	dump    []byte
	// 当设置了 retries 的时候， 出现 err client.Do() 会关掉 req.Body 所以这里 fork 一下
	forkReqBody []byte
}

func (h *HTTPRequest) GetRequest() *http.Request {
	return h.req
}

func (h *HTTPRequest) GetDumpRequest() []byte {
	return h.dump
}

func (h *HTTPRequest) Setting(setting HTTPSettings) *HTTPRequest {
	h.setting = setting
	return h
}

func (h *HTTPRequest) SetKeepAlive(enable bool) *HTTPRequest {
	h.setting.KeepAlive = enable
	return h
}

// SetMaxIdleConnsPerHost 可根据并发设置，一般默认即可
func (h *HTTPRequest) SetMaxIdleConnsPerHost(n int) *HTTPRequest {
	h.setting.MaxIdleConnsPerHost = n
	return h
}

func (h *HTTPRequest) SetBasicAuth(username, password string) *HTTPRequest {
	h.req.SetBasicAuth(username, password)
	return h
}

func (h *HTTPRequest) SetEnableCookie(enable bool) *HTTPRequest {
	h.setting.EnableCookie = enable
	return h
}

func (h *HTTPRequest) SetUserAgent(useragent string) *HTTPRequest {
	h.setting.UserAgent = useragent
	return h
}

func (h *HTTPRequest) SetDebug(isdebug bool) *HTTPRequest {
	h.setting.ShowDebug = isdebug
	return h
}

func (h *HTTPRequest) SetRetries(retries int) *HTTPRequest {
	h.setting.Retries = retries
	return h
}

func (h *HTTPRequest) SetDumpBody(dumpBody bool) *HTTPRequest {
	h.setting.DumpBody = dumpBody
	return h
}

func (h *HTTPRequest) SetTimeout(cTimeout, rwTimeout time.Duration) *HTTPRequest {
	h.setting.ConnectTimeout = cTimeout
	h.setting.ReadWriteTimeout = rwTimeout
	return h
}

func (h *HTTPRequest) SetTLSClientConfig(config *tls.Config) *HTTPRequest {
	h.setting.TLSClientConfig = config
	return h
}

func (h *HTTPRequest) SetHost(host string) *HTTPRequest {
	h.req.Host = host
	return h
}

func (h *HTTPRequest) SetProtocolVersion(vers string) *HTTPRequest {
	if len(vers) == 0 {
		vers = "HTTP/1.1"
	}
	majar, minor, ok := http.ParseHTTPVersion(vers)
	if ok {
		h.req.Proto = vers
		h.req.ProtoMajor = majar
		h.req.ProtoMinor = minor
	}

	return h
}

func (h *HTTPRequest) SetCookie(cookie *http.Cookie) *HTTPRequest {
	h.req.Header.Add("Cookie", cookie.String())
	return h
}

// SetTransport 自定义 transport
func (h *HTTPRequest) SetTransport(tran http.Transport) *HTTPRequest {
	h.setting.Transport = http.RoundTripper(&tran)
	return h
}

func (h *HTTPRequest) SetProxy(proxy func(*http.Request) (*url.URL, error)) *HTTPRequest {
	h.setting.Proxy = proxy
	return h
}

func (h *HTTPRequest) SetCheckRedirect(redirect func(req *http.Request, via []*http.Request) error) *HTTPRequest {
	h.setting.CheckRedirect = redirect
	return h
}

func (h *HTTPRequest) SetHeader(key, value string) *HTTPRequest {
	h.req.Header.Set(key, value)
	return h
}

type Params map[string]interface{}

// SetParams json 格式的 struct 直接转换成 k v 字符串， 作为表单提交
func (h *HTTPRequest) SetParams(v interface{}) (*HTTPRequest, error) {
	byts, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	params := Params{}
	if err := json.Unmarshal(byts, &params); err != nil {
		return nil, err
	}

	for k, v := range params {
		h.SetParam(k, fmt.Sprint(v))
	}

	return h, nil
}

func (h *HTTPRequest) SetParam(key, value string) *HTTPRequest {
	if param, ok := h.params[key]; ok {
		h.params[key] = append(param, value)
	} else {
		h.params[key] = []string{value}
	}

	return h
}

func (h *HTTPRequest) PostFile(fieldname, filename string) *HTTPRequest {
	h.files[fieldname] = filename
	return h
}

func (h *HTTPRequest) SetBody(data interface{}) *HTTPRequest {
	switch typ := data.(type) {
	case string:
		buf := bytes.NewBufferString(typ)
		h.req.Body = ioutil.NopCloser(buf)
		h.req.ContentLength = int64(len(typ))
	case []byte:
		buf := bytes.NewBuffer(typ)
		h.req.Body = ioutil.NopCloser(buf)
		h.req.ContentLength = int64(len(typ))
	}

	return h
}

func (h *HTTPRequest) SetJSONBody(obj interface{}) (*HTTPRequest, error) {
	if h.req.Body == nil && obj != nil {
		byts, err := json.Marshal(obj)
		if err != nil {
			return h, err
		}

		h.req.Body = ioutil.NopCloser(bytes.NewReader(byts))
		h.req.ContentLength = int64(len(byts))
		h.req.Header.Set("Content-Type", "application/json")
	}

	return h, nil
}

func (h *HTTPRequest) XMLBody(obj interface{}) (*HTTPRequest, error) {
	if h.req.Body == nil && obj != nil {
		byts, err := xml.Marshal(obj)
		if err != nil {
			return h, err
		}
		h.req.Body = ioutil.NopCloser(bytes.NewReader(byts))
		h.req.ContentLength = int64(len(byts))
		h.req.Header.Set("Content-Type", "application/xml")
	}
	return h, nil
}

func (h *HTTPRequest) getResponse() (*http.Response, error) {
	// 之前已经请求过
	if h.resp.StatusCode != 0 {
		return h.resp, nil
	}
	resp, err := h.DoRequest()
	if err != nil {
		return nil, err
	}
	h.resp = resp
	return resp, nil
}

func (h *HTTPRequest) Response() (resp *http.Response, err error) {
	return h.getResponse()
}

func (h *HTTPRequest) RespToByte() ([]byte, error) {
	if h.body != nil {
		return h.body, nil
	}

	resp, err := h.getResponse()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.Body == nil {
		return nil, nil
	}

	if h.setting.Gzip && resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}
		h.body, err = ioutil.ReadAll(reader)
		return h.body, err
	}

	h.body, err = ioutil.ReadAll(resp.Body)
	return h.body, err
}

func (h *HTTPRequest) RespToString() (string, error) {
	data, err := h.RespToByte()
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (h *HTTPRequest) RespToFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	resp, err := h.getResponse()
	if err != nil {
		return err
	}

	if resp.Body == nil {
		return nil
	}
	defer resp.Body.Close()
	_, err = io.Copy(file, resp.Body)
	return err
}

func (h *HTTPRequest) RespToJSON(v interface{}) error {
	data, err := h.RespToByte()
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, v)
	if err != nil {
		return fmt.Errorf("response body %s ,json.Unmarshal err %s", string(data), err.Error())
	}

	return err
}

func (h *HTTPRequest) RespToXML(v interface{}) error {
	data, err := h.RespToByte()
	if err != nil {
		return err
	}

	err = xml.Unmarshal(data, v)
	if err != nil {
		return fmt.Errorf("response body %s ,xml.Unmarshal err %s", string(data), err.Error())
	}

	return err
}

// func (h *HTTPRequest) RespToYAML(v interface{}) error {
// 	data, err := h.RespToByte()
// 	if err != nil {
// 		return err
// 	}

// 	return yaml.Unmarshal(data, v)
// }

func (h *HTTPRequest) DoRequest() (resp *http.Response, err error) {
	var paramBody string
	if len(h.params) > 0 {
		var buf bytes.Buffer
		for k, v := range h.params {
			for _, vv := range v {
				buf.WriteString(url.QueryEscape(k))
				buf.WriteByte('=')
				buf.WriteString(url.QueryEscape(vv))
				buf.WriteByte('&')
			}
			paramBody = buf.String()
			// 去掉最后一个 &
			paramBody = paramBody[:len(paramBody)-1]
		}
	}

	h.buildURL(paramBody)
	urlParsed, err := url.Parse(h.url)
	if err != nil {
		return nil, err
	}

	h.req.URL = urlParsed

	// 处理 tcp 传输设置， 如果存在链接，则复用
	trans := h.setting.Transport
	if trans == nil {
		trans = &http.Transport{
			TLSClientConfig: h.setting.TLSClientConfig,
			Proxy:           h.setting.Proxy,
			//Dial:                  TimeoutDialer(h.setting.ConnectTimeout, h.setting.ReadWriteTimeout),
			DialContext:           TimeoutDialerContext(h.setting.ConnectTimeout, h.setting.ReadWriteTimeout),
			DisableKeepAlives:     !h.setting.KeepAlive, // 默认关闭 keepAlive
			MaxIdleConnsPerHost:   h.setting.MaxIdleConnsPerHost,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
		h.setting.Transport = trans
	} else {
		t, ok := trans.(*http.Transport)
		if ok && (t != nil) {
			if t.TLSClientConfig == nil {
				t.TLSClientConfig = h.setting.TLSClientConfig
			}
			if t.Proxy == nil {
				t.Proxy = h.setting.Proxy
			}

			if t.DialContext == nil {
				t.DialContext = TimeoutDialerContext(h.setting.ConnectTimeout, h.setting.ReadWriteTimeout)
			}
		}
	}

	// 设置 cookie
	var jar http.CookieJar
	if h.setting.EnableCookie {
		if defaultCookieJar == nil {
			createDefaultCookie()
		}
		jar = defaultCookieJar
	}

	if h.client == nil {
		h.client = &http.Client{
			Transport: trans,
			Jar:       jar,
		}
	}

	if h.setting.UserAgent != "" && h.req.Header.Get("User-Agent") == "" {
		h.req.Header.Set("User-Agent", h.setting.UserAgent)
	}

	// 设置 重定向
	if h.setting.CheckRedirect != nil {
		h.client.CheckRedirect = h.setting.CheckRedirect
	}

	// dump 请求信息
	if h.setting.ShowDebug {
		dump, err := httputil.DumpRequest(h.req, h.setting.DumpBody)
		if err != nil {
			log.Println("httplib dump: ", err)
		}
		h.dump = dump
	}

	// 重试次数
	for i := 0; h.setting.Retries == -1 || i <= h.setting.Retries; i++ {
		if (h.setting.Retries == -1 || h.setting.Retries >= 1) && i == 0 {
			// solve 请求一次 http 关掉 req.Body
			if err := copyReqBody(h); err != nil {
				return resp, err
			}
		}

		if i > 0 {
			h.req.Body = ioutil.NopCloser(bytes.NewBuffer(h.forkReqBody))
		}

		resp, err = h.client.Do(h.req)
		if err == nil {
			break
		}
		// retry sleep
		time.Sleep(500 * time.Millisecond)
	}

	return resp, err
}

// 拼装请求参数
func (h *HTTPRequest) buildURL(paramBody string) {

	// GET 方法，拼接 URL
	if h.req.Method == "GET" && len(paramBody) > 0 {
		if strings.Contains(h.url, "?") {
			h.url += "&" + paramBody
		} else {
			h.url += "?" + paramBody
		}
		return
	}

	if (h.req.Method == "POST" || h.req.Method == "PUT" || h.req.Method == "PATCH" || h.req.Method == "DELETE") && h.req.Body == nil {
		// 存在文件
		if len(h.files) > 0 {
			pr, pw := io.Pipe()
			bodyWriter := multipart.NewWriter(pw)
			go func() {
				for fieldname, filename := range h.files {
					fileWriter, err := bodyWriter.CreateFormFile(fieldname, filename)
					if err != nil {
						log.Println("httplib: ", err)
					}

					file, err := os.Open(filename)
					if err != nil {
						log.Println("httplib: ", err)
					}

					if _, err := io.Copy(fileWriter, file); err != nil {
						log.Println("httplib: ", err)
					}

					file.Close()
				}

				for k, v := range h.params {
					for _, vv := range v {
						bodyWriter.WriteField(k, vv)
					}
				}

				bodyWriter.Close()

				pw.Close()
			}()

			h.SetHeader("Content-Type", bodyWriter.FormDataContentType())
			h.req.Body = ioutil.NopCloser(pr)
			return
		}

		// 存在 params
		if len(paramBody) > 0 {
			h.SetHeader("Content-Type", "application/x-www-form-urlencoded")
			h.SetBody(paramBody)
		}
	}

}

func TimeoutDialer(cTimeout, rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		return conn, conn.SetDeadline(time.Now().Add(rwTimeout))
	}
}

func TimeoutDialerContext(cTimeout, rwTimeout time.Duration) func(ctx context.Context, net, addr string) (c net.Conn, err error) {
	f := &net.Dialer{
		KeepAlive: 60 * time.Second,
	}

	if cTimeout > 0 {
		f.Timeout = cTimeout
	}
	if rwTimeout > 0 {
		f.Deadline = time.Now().Local().Add(rwTimeout)
	}
	return f.DialContext

	//return (&net.Dialer{
	//	Timeout:   cTimeout,
	//	KeepAlive: 30 * time.Second,
	//	Deadline:  time.Now().Local().Add(rwTimeout),
	//}).DialContext
}

func copyReqBody(ireq *HTTPRequest) error {
	if ireq.req.Body == nil {
		return nil
	}

	ibody, err := ioutil.ReadAll(ireq.req.Body)
	if err != nil {
		return err
	}

	ireq.forkReqBody = ibody
	ireq.req.Body = ioutil.NopCloser(bytes.NewBuffer(ibody))

	return nil
}
