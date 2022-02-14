package ginutil

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http/pprof"
)

var skipPaths = []string{"/health", "/admin", "/static", "/docs", "/favicon.ico", "/debug"}

// AddSkipPaths add skip Path, not write logger
func AddSkipPaths(paths ...string) {
	skipPaths = append(skipPaths, paths...)
}

const (
	MRelease = 1 << iota // Gin release Mode
	MSwagger             // enable Swagger

	// middleware enable

	MTraceId       // request context add TraceId
	MReqLogger     // request logging
	MDumpBody      // Dump req | resp body
	MRecoverLogger // Recover logging write to qelog
	MPprof         // http pprof

	MStd = MSwagger | MTraceId | MReqLogger | MDumpBody // default
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

	// middleware
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

	if flags&MPprof != 0 {
		pprofHandler(engine)
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

func pprofHandler(engine *gin.Engine) {
	engine.GET("/debug/pprof/", gin.WrapF(pprof.Index))
	engine.GET("/debug/pprof/cmdline", gin.WrapF(pprof.Cmdline))
	engine.GET("/debug/pprof/profile", gin.WrapF(pprof.Profile))
	engine.GET("/debug/pprof/symbol", gin.WrapF(pprof.Symbol))
	engine.GET("/debug/pprof/trace", gin.WrapF(pprof.Trace))
}
