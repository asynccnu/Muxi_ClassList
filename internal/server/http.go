package server

import (
	"github.com/asynccnu/Muxi_ClassList/internal/conf"
	"github.com/asynccnu/Muxi_ClassList/internal/metrics"
	"github.com/asynccnu/Muxi_ClassList/internal/pkg/encoder"
	"github.com/asynccnu/Muxi_ClassList/internal/service"
	"github.com/go-kratos/kratos/v2/middleware/validate"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, greeter *service.ClasserService, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			metrics.QPSMiddleware(),
			metrics.DelayMiddleware(),
			validate.Validator(),
		),
		http.ResponseEncoder(encoder.RespEncoder), // Notice: 将响应格式化
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	srv.Handle("/metrics", promhttp.Handler())
	//v1.RegisterClasserHTTPServer(srv, greeter)
	return srv
}
