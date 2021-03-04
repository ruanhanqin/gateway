package server

import (
	"github.com/Naist4869/common/api/gateway"
	"github.com/Naist4869/gateway/internal/conf"
	"github.com/Naist4869/gateway/internal/service"

	"github.com/Naist4869/common"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new a HTTP server.
func NewHTTPServer(c *conf.Server, greeter *service.PayGatewayService) *http.Server {
	var opts = []http.ServerOption{}
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
	m := http.Middleware(
		middleware.Chain(
			recovery.Recovery(),
			tracing.Server(),
			logging.Server(),
		),
	)
	srv.HanldePrefix("/", gateway.NewPayGatewayHandler(greeter, m, http.ErrorEncoder(common.EncodeErrorFunc), http.ResponseEncoder(common.EncodeResponseFunc)))
	return srv
}
