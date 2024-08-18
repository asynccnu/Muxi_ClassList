package client

import (
	"class/internal/errcode"
	"context"
	v1 "github.com/asynccnu/ccnu-service/api/ccnu_service/v1"
	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

const (
	CCNU_SERVICE = "discovery:///ccnu_service"
)

type CCNUService struct {
	Cs v1.CCNUServiceClient
}

func NewCCNUService(cs v1.CCNUServiceClient) *CCNUService {
	return &CCNUService{Cs: cs}
}
func (c *CCNUService) GetCookie(ctx context.Context, stu string) (string, error) {
	resp, err := c.Cs.GetCookie(ctx, &v1.GetCookieRequest{
		Userid: stu,
	})
	if err != nil {
		return "", errcode.ErrCCNULogin
	}
	cookie := resp.Cookie
	return cookie, nil
}
func NewClient(r *etcd.Registry, logger log.Logger) (v1.CCNUServiceClient, error) {
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint(CCNU_SERVICE), // 需要发现的服务，如果是k8s部署可以直接用服务器本地地址:9001，9001端口是需要调用的服务的端口
		grpc.WithDiscovery(r),
		grpc.WithMiddleware(
			tracing.Client(),
			recovery.Recovery(),
		),
	)
	if err != nil {
		log.NewHelper(logger).WithContext(context.Background()).Errorw("kind", "grpc-client", "reason", "GRPC_CLIENT_INIT_ERROR", "err", err)
		return nil, err
	}
	return v1.NewCCNUServiceClient(conn), nil
}
