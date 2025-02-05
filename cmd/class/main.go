package main

import (
	"flag"
	"github.com/asynccnu/Muxi_ClassList/internal/classLog"
	"github.com/asynccnu/Muxi_ClassList/internal/conf"
	"github.com/asynccnu/Muxi_ClassList/internal/metrics"
	"github.com/asynccnu/Muxi_ClassList/internal/pkg/tool"
	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2"
	"github.com/prometheus/client_golang/prometheus"
	"os"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string = "MuXi_ClassList"
	// Version is the version of the compiled software.
	Version string = "v1"
	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()
)

func init() {
	prometheus.MustRegister(metrics.Counter, metrics.Summary)
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

func newApp(logger log.Logger, gs *grpc.Server, r *etcd.Registry) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			gs,
		),
		kratos.Registrar(r),
	)
}

func main() {
	flag.Parse()
	//logger := log.With(log.NewStdLogger(os.Stdout),
	//	"ts", log.DefaultTimestamp,
	//	"caller", log.DefaultCaller,
	//	"service.id", id,
	//	"service.name", Name,
	//	"service.version", Version,
	//	"trace.id", tracing.TraceID(),
	//	"span.id", tracing.SpanID(),
	//)
	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	//if err := c.Watch("schoolday.holidayTime", func(key string, value config.Value) {
	//	res, err1 := value.String()
	//	if err1 != nil {
	//		log.Info("config change failed: %s\n", key)
	//	}
	//	bc.Schoolday.HolidayTime = res
	//	fmt.Printf("config changed: %s = %v\n", key, value)
	//}); err != nil {
	//	log.Error(err)
	//}
	//if err := c.Watch("schoolday.schoolTime", func(key string, value config.Value) {
	//	res, err1 := value.String()
	//	if err1 != nil {
	//		log.Info("config change failed: %s\n", key)
	//	}
	//	bc.Schoolday.SchoolTime = res
	//	fmt.Printf("config changed: %s = %v\n", key, value)
	//}); err != nil {
	//	log.Error(err)
	//}

	logger := classLog.Logger(bc.Zaplog)
	logger = log.With(logger,
		"service.id", id,
		"service.name", Name,
	)
	//gorm的日志文件
	//在main函数中声明,程序结束执行Close
	//防止只有连接数据库的时候，才会将sql语句写入
	logfile, err := tool.OpenFile(bc.Data.Database.LogPath, bc.Data.Database.LogFileName)
	if err != nil {
		panic(err)
	}
	defer logfile.Close()
	app, cleanup, err := wireApp(bc.Server, bc.Data, bc.Registry, bc.Schoolday, logfile, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
