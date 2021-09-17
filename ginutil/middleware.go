package ginutil

import (
	"bytes"
	"encoding/json"
	"github.com/bbdshow/bkit/auth/jwt"
	"github.com/bbdshow/bkit/auth/sign"
	"github.com/bbdshow/bkit/errc"
	"github.com/bbdshow/bkit/logs"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	AuthorizationHeader = "X-Authorization"
	SignatureHeader     = "X-Signature"
)

func ContextWithTraceId() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request = c.Request.WithContext(logs.Qezap.WithTraceID(c.Request.Context()))
		c.Next()
	}
}

// DumpBodyLogger Dump  req | resp body
func DumpBodyLogger(skipPaths ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		uri := c.Request.URL.RequestURI()
		path := c.FullPath()
		for _, v := range skipPaths {
			if path == v || uri == v {
				c.Next()
				return
			}
		}
		ip := ClientIP(c)
		headers := c.Request.Header.Clone()
		method := c.Request.Method

		var reqBody interface{}
		if c.Request.Body != nil {
			b, err := ioutil.ReadAll(c.Request.Body)
			if err != nil {
				_ = c.Request.Body.Close()
				RespErr(c, err)
				c.Abort()
				return
			}
			// read first time, so closed body
			_ = c.Request.Body.Close()
			body := bytes.NewBuffer(b)

			if strings.Contains(strings.ToLower(c.ContentType()), "application/json") {
				reqMap := make(map[string]interface{})
				err := json.Unmarshal(body.Bytes(), &reqMap)
				if err == nil {
					reqBody = reqMap
				}
			}
			if reqBody == nil {
				reqBody = body.String()
			}

			c.Request.Body = ioutil.NopCloser(body)
		}

		logs.Qezap.Info("DumpRequest", zap.Any("headers", headers), zap.Any("reqBody", reqBody),
			zap.String("URI", uri),
			logs.Qezap.ConditionOne(method), logs.Qezap.ConditionTwo(path), logs.Qezap.ConditionThree(ip),
			logs.Qezap.FieldTraceID(c.Request.Context()))

		// dump body
		w := &respWriter{body: bytes.NewBuffer([]byte{}), ResponseWriter: c.Writer}
		c.Writer = w

		c.Next()
		traceId := logs.Qezap.TraceID(c.Request.Context())
		baseResp := BaseResp{}
		if err := json.Unmarshal(w.body.Bytes(), &baseResp); err != nil || baseResp.Code != 0 {
			logs.Qezap.Error("DumpResponse", zap.String("respBody", w.body.String()),
				logs.Qezap.ConditionOne(method), logs.Qezap.ConditionTwo(path), logs.Qezap.ConditionThree(ip),
				logs.Qezap.FieldTraceID(c.Request.Context()),
				zap.String("latency", time.Now().Sub(traceId.Time()).String()))
		}
	}
}

// ReqLogger GIN handle request logging to qelog
func ReqLogger(skipPaths ...string) gin.HandlerFunc {
	var skip map[string]struct{}

	if length := len(skipPaths); length > 0 {
		skip = make(map[string]struct{}, length)

		for _, path := range skipPaths {
			skip[path] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log only when path is not being skipped
		if _, ok := skip[path]; !ok {
			param := gin.LogFormatterParams{
				Request: c.Request,
				Keys:    c.Keys,
			}

			// Stop timer
			param.TimeStamp = time.Now()
			param.Latency = param.TimeStamp.Sub(start)

			param.ClientIP = ClientIP(c)
			param.Method = c.Request.Method
			param.StatusCode = c.Writer.Status()
			param.ErrorMessage = c.Errors.ByType(gin.ErrorTypePrivate).String()

			param.BodySize = c.Writer.Size()

			rawPath := path
			if raw != "" {
				path = path + "?" + raw
			}

			param.Path = path

			if param.Latency > time.Minute {
				// Truncate in a golang < 1.8 safe way
				param.Latency = param.Latency - param.Latency%time.Second
			}

			logs.Qezap.Debug("GIN", zap.String("latency", param.Latency.String()),
				zap.String("method", param.Method),
				zap.String("path", param.Path),
				zap.String("error", param.ErrorMessage),
				logs.Qezap.ConditionOne(strconv.Itoa(param.StatusCode)),
				logs.Qezap.ConditionTwo(param.Method+"_"+rawPath),
				logs.Qezap.ConditionThree(param.ClientIP))
		}
	}
}

// RecoveryLogger GIN Recovery logging to qelog
func RecoveryLogger() gin.HandlerFunc {
	if logs.Qezap != nil {
		return gin.RecoveryWithWriter(logs.Qezap.NewWriter(zap.ErrorLevel, "GIN-ERROR"))
	}
	return gin.Recovery()
}

// JWTAuthVerify JWT auth verify signingKey custom signing key
func JWTAuthVerify(enable bool, signingKey ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !enable {
			c.Next()
			return
		}
		token := c.GetHeader(AuthorizationHeader)
		if token == "" {
			logs.Qezap.Warn("JWTAuthException", zap.String("Authorization", "request header required"), logs.Qezap.FieldTraceID(c.Request.Context()))
			RespErr(c, errc.ErrAuthRequired, http.StatusUnauthorized)
			c.Abort()
			return
		}

		ok, err := jwt.VerifyJWTToken(token, signingKey...)
		if err != nil {
			logs.Qezap.Warn("JWTAuthException", zap.String("JWT", err.Error()), logs.Qezap.FieldTraceID(c.Request.Context()))
			if strings.Contains(err.Error(), "expired") {
				RespErr(c, errc.ErrAuthExpired, http.StatusUnauthorized)
			} else {
				RespErr(c, errc.ErrAuthInvalid, http.StatusUnauthorized)
			}
			c.Abort()
			return
		}
		if !ok {
			logs.Qezap.Warn("JWTAuthException", zap.String("JWT", "verify invalid"), logs.Qezap.FieldTraceID(c.Request.Context()))
			RespErr(c, errc.ErrAuthInvalid, http.StatusUnauthorized)
			c.Abort()
			return
		}

		if err := SetJWTDataToContext(c, token, signingKey...); err != nil {
			logs.Qezap.Warn("JWTAuthException", zap.String("SettingJWTData", err.Error()), logs.Qezap.FieldTraceID(c.Request.Context()))
			RespErr(c, errc.ErrAuthInternalErr, http.StatusUnauthorized)
			c.Abort()
			return
		}
		c.Next()
	}
}

type SignConfig struct {
	Enable bool `defval:"false"`
	sign.Config
	SupportMethods []string `defval:"GET,POST,PUT,DELETE"`
}

// ApiSignVerify API signature, must setting get secretKey function
func ApiSignVerify(cfg *SignConfig, getSecretKey func(accessKey string) (string, error)) gin.HandlerFunc {
	apiSign := sign.NewAPISign(&cfg.Config)
	apiSign.SetGetSecretKey(getSecretKey)

	return func(c *gin.Context) {
		if !cfg.Enable {
			c.Next()
			return
		}
		isSupport := false
		for _, v := range cfg.SupportMethods {
			if c.Request.Method == strings.ToUpper(v) {
				isSupport = true
				break
			}
		}
		if isSupport {
			if err := apiSign.Verify(c.Request, SignatureHeader); err != nil {
				logs.Qezap.Warn("APISignatureException", zap.Error(err), logs.Qezap.ConditionOne(c.Request.URL.Path), logs.Qezap.FieldTraceID(c.Request.Context()))
				RespErr(c, errc.ErrAuthSignatureInvalid, http.StatusUnauthorized)
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

func GetApiSignAccessKey(c *gin.Context) string {
	h := c.GetHeader(SignatureHeader)
	if h == "" {
		return ""
	}
	accessKey, _, _, err := sign.DecodeHeaderVal(h)
	if err == nil {
		return accessKey
	}
	return ""
}
