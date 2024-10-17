package classLog

import (
	"github.com/asynccnu/Muxi_ClassList/internal/conf"
	"github.com/go-kratos/kratos/v2/log"
	"testing"
)

func TestZapLogger_Log(t *testing.T) {
	c := &conf.ZapLogConfigs{
		LogLevel:          "info",
		LogFormat:         "json",
		LogPath:           "./log",
		LogFileName:       "test.log",
		LogFileMaxSize:    1,
		LogFileMaxBackups: 10,
		LogMaxAge:         100,
		LogCompress:       false,
		LogStdout:         true,
	}
	log1 := Logger(c)
	log1.Log(log.LevelInfo, "msg", "world")
	helper := log.NewHelper(log1)
	helper.Infow("name", "chen", "msg", "nihao")
	//日志初始化
	//loggers := log.With(log1,
	//	"caller", log.DefaultCaller,
	//	"service.id", 1,
	//	"service.name", "class",
	//	"service.version", "1.1",
	//	"trace_id", tracing.TraceID(),
	//	"span_id", tracing.SpanID(),
	//)
	//loggers.Log(log.LevelInfo, "hello %s", "world")

}
