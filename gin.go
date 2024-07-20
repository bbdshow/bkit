package bkit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/pprof"
	"strings"

	"github.com/gin-gonic/gin"
	sfiles "github.com/swaggo/files"
	gswagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

const (
	HeaderRequestID = "X-Request-Id"
)

var Gin *GinUtil

func init() {
	Gin = NewGinUtil()
}

type GinUtil struct {
	skipPaths []string
}

func NewGinUtil() *GinUtil {
	skipPaths := []string{"/health", "/admin", "/static", "/docs", "/favicon.ico", "/debug"}
	return &GinUtil{
		skipPaths: skipPaths,
	}
}

type resp struct {
	Code      ErrCode
	Message   string
	Data      interface{}
	RequestID string
}

func (g *GinUtil) respHandler(c *gin.Context, httpCode int, data interface{}, err error) resp {
	requestID := c.GetHeader(HeaderRequestID)
	if requestID == "" {
		requestID = c.GetHeader("X-Tc-Requestid")
	}
	code := ErrCodeOK
	message := GetErrMsg(code)
	if err != nil {
		if e, ok := err.(Err); ok {
			code = e.Code
			message = e.Message
		} else {
			code = ErrCodeFailed
			message = err.Error()
		}
		if code != ErrCodeOK {
			Err := &RespErrInternalLog{
				Method:   c.Request.Method,
				Path:     c.Request.URL.RequestURI(),
				IP:       c.ClientIP(),
				Form:     c.Request.Form,
				PostForm: c.Request.PostForm,
				Func:     c.HandlerName(),
				Error:    message,
			}
			if code == ErrCodeInternal {
				Zap.Error("InternalException", zap.Any("Err", Err), zap.String("RequestID", requestID))
				// 覆盖详细信息
				message = GetErrMsg(code)
			}
			if code == ErrCodeFailed {
				// Warn 日志不触发报警
				Zap.Warn("InternalFailed", zap.Any("Err", Err), zap.String("RequestID", requestID))
			}
		}
	}
	return resp{
		Code:      code,
		Message:   message,
		Data:      data,
		RequestID: requestID,
	}
}

// Resp
func (g *GinUtil) Resp(c *gin.Context, httpCode int, data interface{}, err error) {
	resp := g.respHandler(c, httpCode, data, err)
	baseResp := &BaseResp{
		Err: Err{
			Code:    resp.Code,
			Message: resp.Message,
		},
		RequestID: resp.RequestID,
	}
	out := &DataResp{
		BaseResp: baseResp,
		Data:     resp.Data,
	}
	c.JSON(httpCode, out)
}

// RespOK -
func (g *GinUtil) RespOK(c *gin.Context, status ...int) {
	g.Resp(c, g.httpStatus(status...), nil, nil)
}

// RespErr -
func (g *GinUtil) RespErr(c *gin.Context, err error, status ...int) {
	g.Resp(c, g.httpStatus(status...), nil, err)
}

// RespData -
func (g *GinUtil) RespData(c *gin.Context, data interface{}, status ...int) {
	g.Resp(c, g.httpStatus(status...), data, nil)
}

func (g *GinUtil) httpStatus(status ...int) int {
	if len(status) == 1 && status[0] > 0 {
		return status[0]
	}
	return http.StatusOK
}

// CloudResp
func (g *GinUtil) CloudResp(c *gin.Context, httpCode int, data interface{}, err error) {
	resp := g.respHandler(c, httpCode, data, err)
	baseResp := &CloudBaseResp{
		Code:      resp.Code,
		Msg:       resp.Message,
		RequestID: resp.RequestID,
	}
	out := &CloudDataResp{
		CloudBaseResp: baseResp,
		Data:          data,
	}
	c.JSON(httpCode, CloudResp{Response: out})
}

// CloudRespOK -
func (g *GinUtil) CloudRespOK(c *gin.Context, status ...int) {
	g.CloudResp(c, g.httpStatus(status...), nil, nil)
}

// CloudRespErr -
func (g *GinUtil) CloudRespErr(c *gin.Context, err error, status ...int) {
	g.CloudResp(c, g.httpStatus(status...), nil, err)
}

// CloudRespData -
func (g *GinUtil) CloudRespData(c *gin.Context, data interface{}, status ...int) {
	g.CloudResp(c, g.httpStatus(status...), data, nil)
}

// MidAPILimiter 中间件-接口限流
func (g *GinUtil) MidAPILimiter(path, method string, r, b int, isWait bool) gin.HandlerFunc {
	Limiter.Register(fmt.Sprintf("%s_%s", path, strings.ToUpper(method)), r, b)

	return func(c *gin.Context) {
		key := fmt.Sprintf("%s_%s", c.Request.URL.Path, c.Request.Method)
		if isWait {
			if err := Limiter.Wait(c, key); err != nil {
				c.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
				c.Abort()
				return
			}
		} else {
			if !Limiter.Allow(key) {
				c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
				c.Abort()
				return
			}
		}
	}
}

// MidDumpBodyLogger 中间件- Dump 请求&返回日志， 当有错误才Dump返回结构体
func (g *GinUtil) MidDumpBodyLogger(skipPaths ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		uri := c.Request.URL.RequestURI()
		path := c.FullPath()
		for _, v := range skipPaths {
			if path == v || uri == v {
				c.Next()
				return
			}
		}
		ip := c.ClientIP()
		headers := c.Request.Header.Clone()
		method := c.Request.Method

		var reqBody interface{}
		if c.Request.Body != nil {
			b, err := io.ReadAll(c.Request.Body)
			if err != nil {
				_ = c.Request.Body.Close()
				g.RespErr(c, err)
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

			c.Request.Body = io.NopCloser(body)
		}
		headers.Del("Cookie")
		headers.Del("x-tai-identity")
		headers.Del("Authorization")

		Zap.Info("DumpRequest", zap.Any("Headers", headers), zap.Any("ReqBody", reqBody),
			zap.String("URI", uri), zap.String("Method", method), zap.String("ClientIP", ip))

		// dump body
		w := &DumpRespWriter{body: bytes.NewBuffer([]byte{}), ResponseWriter: c.Writer}
		c.Writer = w

		c.Next()
		if w.body.Len() != 0 {
			baseResp := BaseResp{}
			if err := json.Unmarshal(w.body.Bytes(), &baseResp); err != nil || baseResp.Code != 0 {
				Zap.Warn("DumpResponse", zap.String("RespBody", w.body.String()),
					zap.String("URI", uri), zap.String("Method", method))
			}
		}
	}
}

// MidRecoveryLogger GIN Recovery logging to zap
func (g *GinUtil) MidRecoveryLogger() gin.HandlerFunc {
	if Zap != nil {
		return gin.RecoveryWithWriter(GetLevelLogWriter(zap.ErrorLevel, "GIN-ERROR"))
	}
	return gin.Recovery()
}

// AddSkipPaths add skip Path, not write logger
func (g *GinUtil) AddSkipPaths(paths ...string) {
	g.skipPaths = append(g.skipPaths, paths...)
}

const (
	// GIN 功能加载标记
	GinMRelease = 1 << iota // GinUtil release Mode
	GinMSwagger             // enable Swagger

	GinMDumpBody      // Dump req | resp body
	GinMRecoverLogger // Recover logging
	GinMPprof         // http pprof

	GinMStd = GinMSwagger | GinMDumpBody // default
)

func (g *GinUtil) healthHandler(engine *gin.Engine) {
	engine.HEAD("/health", func(c *gin.Context) {
		c.Status(200)
	})
	engine.GET("/health", func(c *gin.Context) {
		c.String(200, "ok")
	})
}

func (g *GinUtil) pprofHandler(engine *gin.Engine) {
	// github.com/gin-contrib/pprof
	debug := engine.Group("/debug/pprof")
	{
		debug.GET("/", gin.WrapF(pprof.Index))
		debug.GET("/cmdline", gin.WrapF(pprof.Cmdline))
		debug.GET("/profile", gin.WrapF(pprof.Profile))
		debug.POST("/symbol", gin.WrapF(pprof.Symbol))
		debug.GET("/symbol", gin.WrapF(pprof.Symbol))
		debug.GET("/trace", gin.WrapF(pprof.Trace))
		debug.GET("/allocs", gin.WrapH(pprof.Handler("allocs")))
		debug.GET("/block", gin.WrapH(pprof.Handler("block")))
		debug.GET("/goroutine", gin.WrapH(pprof.Handler("goroutine")))
		debug.GET("/heap", gin.WrapH(pprof.Handler("heap")))
		debug.GET("/mutex", gin.WrapH(pprof.Handler("mutex")))
		debug.GET("/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))
	}
}

func DefaultGinEngine(flags int, middleware ...gin.HandlerFunc) *gin.Engine {
	if flags&GinMRelease != 0 {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()

	Gin.healthHandler(engine)

	// swagger
	if flags&GinMSwagger != 0 {
		engine.GET("/docs/*any", gswagger.WrapHandler(sfiles.Handler))
	}
	if len(middleware) > 0 {
		engine.Use(middleware...)
	}

	if flags&GinMDumpBody != 0 {
		engine.Use(Gin.MidDumpBodyLogger(Gin.skipPaths...))
	}
	if flags&GinMRecoverLogger != 0 {
		engine.Use(Gin.MidRecoveryLogger())
	}

	if flags&GinMPprof != 0 {
		Gin.pprofHandler(engine)
	}

	return engine
}

func (g *GinUtil) ShouldBind(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBind(obj); err != nil {
		Zap.Warn("ParamException", zap.Any("RequestParam", obj), zap.Any("Exception", err))
		if !gin.IsDebugging() {
			// hide specific param info
			return ErrParamInvalid
		}
		return ErrParamInvalid.ErrMulti(err)
	}
	return nil
}
