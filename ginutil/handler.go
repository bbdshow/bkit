package ginutil

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var skipPaths = []string{"/health", "/admin", "/static", "/docs", "/favicon.ico"}

// AddSkipPaths 添加的Path不会被日志记录
func AddSkipPaths(paths ...string) {
	skipPaths = append(skipPaths, paths...)
}

const (
	MRelease = 1 << iota // Gin release 模式
	MSwagger             // 开启Swagger

	// 中间件启用
	MTraceId       // 请求context 加入TraceId
	MReqLogger     //请求日志
	MDumpBody      // Dump 请求参数&返回参数
	MRecoverLogger // Recover 日志写入到日志中心

	MStd = MSwagger | MTraceId | MReqLogger | MDumpBody // 默认
)

func DefaultEngine(flags int) *gin.Engine {

	if flags&MRelease != 0 {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()

	healthHandler(engine)

	// swagger
	if flags&MSwagger != 0 {
		engine.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// 中间件
	if flags&MTraceId != 0 {
		engine.Use(ContextWithTraceId())
	}
	if flags&MReqLogger != 0 {
		engine.Use(ReqLogger(skipPaths...))
	}
	if flags&MDumpBody != 0 {
		engine.Use(DumpBodyLogger(skipPaths...))
	}
	if flags&MRecoverLogger != 0 {
		engine.Use(RecoveryLogger())
	}
	return engine
}

func healthHandler(engine *gin.Engine) {
	engine.HEAD("/health", func(c *gin.Context) {
		c.Status(200)
	})
	engine.GET("/health", func(c *gin.Context) {
		c.String(200, "ok")
	})
}
