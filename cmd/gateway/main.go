package main

import (
	"flag"
	"os"

	"github.com/Naist4869/gateway/internal/service"

	"github.com/Naist4869/base/decodehook"
	"github.com/spf13/viper"

	"github.com/Naist4869/gateway/internal/conf"

	pb "github.com/Naist4869/common/api/gateway"

	"github.com/Naist4869/base/config"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../conf", "config path, eg: -conf config.yaml")
}

func newApp(logger log.Logger, hs *http.Server, gs *grpc.Server, srv *service.PayGatewayService) *kratos.App {
	pb.RegisterPayGatewayServer(gs, srv)
	pb.RegisterPayGatewayHTTPServer(hs, srv)
	return kratos.New(
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			hs,
			gs,
		),
	)
}

func main() {
	flag.Parse()
	logger := log.NewStdLogger(os.Stdout)

	if err := config.Init("", "gateway"); err != nil {
		panic(err)
	}
	// build transport server
	hc := new(conf.Server_HTTP)
	gc := new(conf.Server_GRPC)

	if err := viper.UnmarshalKey("http", hc, viper.DecodeHook(decodehook.StringToTimeDurationHookFunc())); err != nil {
		panic(err)
	}

	if err := viper.UnmarshalKey("grpc", gc, viper.DecodeHook(decodehook.StringToTimeDurationHookFunc())); err != nil {
		panic(err)
	}
	data := new(conf.Data)
	if err := viper.UnmarshalKey("data", data, viper.DecodeHook(decodehook.StringToTimeDurationHookFunc())); err != nil {
		panic(err)
	}
	app, err := initApp(&conf.Server{
		Http: hc,
		Grpc: gc,
	}, data, logger)
	if err != nil {
		panic(err)
	}

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
