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

var AuthorizationHeader = "X-Authorization"
var SignatureHeader = "X-Signature"
var SignValidDuration = 5 * time.Second

func ContextWithTraceId() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request = c.Request.WithContext(logs.Qezap.WithTraceID(c.Request.Context()))
		c.Next()
	}
}

// DumpBodyLogger Dump 请求和返回Body便于排查问题
func DumpBodyLogger(skipPaths ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求
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
			// 这里已经读取第一次，就关掉
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

		// 为了Dump返回值
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

// ReqLogger GIN 请求日志拦截到日志系统中
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

// RecoveryLogger GIN Recovery错误日志
func RecoveryLogger() gin.HandlerFunc {
	if logs.Qezap != nil {
		return gin.RecoveryWithWriter(logs.Qezap.NewWriter(zap.ErrorLevel, "GIN-ERROR"))
	}
	return gin.Recovery()
}

// JWTAuthVerify JWT权限验证 signingKey 自定义的 加密Key, 如果没有就使用全局默认的
func JWTAuthVerify(enable bool, signingKey ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !enable {
			c.Next()
			return
		}
		token := c.GetHeader(AuthorizationHeader)
		if token == "" {
			logs.Qezap.Warn("JWT权限异常", zap.String("Authorization", "头未传入"), logs.Qezap.FieldTraceID(c.Request.Context()))
			RespErr(c, errc.ErrAuthRequired, http.StatusUnauthorized)
			c.Abort()
			return
		}

		ok, err := jwt.VerifyJWTToken(token, signingKey...)
		if err != nil {
			logs.Qezap.Warn("JWT权限异常", zap.String("JWT", err.Error()), logs.Qezap.FieldTraceID(c.Request.Context()))
			if strings.Contains(err.Error(), "expired") {
				RespErr(c, errc.ErrAuthExpired, http.StatusUnauthorized)
			} else {
				RespErr(c, errc.ErrAuthInvalid, http.StatusUnauthorized)
			}
			c.Abort()
			return
		}
		if !ok {
			logs.Qezap.Warn("JWT权限异常", zap.String("JWT", "验证无效"), logs.Qezap.FieldTraceID(c.Request.Context()))
			RespErr(c, errc.ErrAuthInvalid, http.StatusUnauthorized)
			c.Abort()
			return
		}

		if err := SetJWTDataToContext(c, token, signingKey...); err != nil {
			logs.Qezap.Warn("JWT权限异常", zap.String("设置JWT数据", err.Error()), logs.Qezap.FieldTraceID(c.Request.Context()))
			RespErr(c, errc.ErrAuthInternalErr, http.StatusUnauthorized)
			c.Abort()
			return
		}
		c.Next()
	}
}

// ApiSignVerify API hmacSha1 接口签名 必需设置获取密钥的方法
func ApiSignVerify(enable bool, method sign.Method, supportMethods []string, getSecretKey func(accessKey string) (string, error)) gin.HandlerFunc {
	apiSign := sign.NewAPISign(SignValidDuration, method)
	apiSign.SetGetSecretKey(getSecretKey)

	return func(c *gin.Context) {
		if !enable {
			c.Next()
			return
		}
		isSupport := false
		for _, v := range supportMethods {
			if c.Request.Method == strings.ToUpper(v) {
				isSupport = true
				break
			}
		}
		if isSupport {
			if err := apiSign.Verify(c.Request, SignatureHeader); err != nil {
				logs.Qezap.Warn("API签名错误", zap.Error(err), logs.Qezap.ConditionOne(c.Request.URL.Path), logs.Qezap.FieldTraceID(c.Request.Context()))
				RespErr(c, errc.ErrAuthSignatureInvalid, http.StatusUnauthorized)
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
