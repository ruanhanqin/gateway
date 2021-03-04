package service

import (
	"context"

	"github.com/Naist4869/gateway/internal/biz"

	"github.com/Naist4869/common/api/database"

	generator "github.com/Naist4869/base/tools/generator"
	errno "github.com/Naist4869/common/api/errno"
	pb "github.com/Naist4869/common/api/gateway"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

// PayGateway is a greeter service.
type PayGatewayService struct {
	pb.UnimplementedPayGatewayServer

	uc          *biz.GreeterUsecase
	log         *log.Helper
	idGenerator *generator.Generator
}

func NewPayGatewayService(uc *biz.GreeterUsecase, logger log.Logger) *PayGatewayService {
	return &PayGatewayService{uc: uc, log: log.NewHelper("service/gateway", logger), idGenerator: generator.NewGenerator("PAY", 100_000)}
}

func (s *PayGatewayService) Pay(ctx context.Context, req *pb.PayRequest) (resp *pb.PayResponse, err error) {
	ctx = context.Background()

	s.log.Infof("Received: %v", req.GetMethod())
	if err := req.Validate(); err != nil {
		s.log.Errorf("validate error: %v", err)
		return nil, errors.InvalidArgument(errno.Errors_ErrInvalidArgument, "pay :%v", err)
	}
	gatewayOrderID := s.idGenerator.GenerateID()
	var v *database.ChannelConfig
	if req.ChannelId == "" {
		if v, err = s.uc.GetDefaultChannelConfig(ctx, req.AppId, req.Method); err != nil {
			return nil, err
		}
	} else {
		if v, err = s.uc.GetChannelConfig(ctx, req.AppId, req.Method, req.ChannelId); err != nil {
			return nil, err
		}
	}
	var channelAccount string
	req.ChannelId = v.ChannelID
	channelAccount = v.ChannelAccount
	if channelAccount == "" {
		return nil, errors.Internal(errno.Errors_ErrNotFoundChannelAccount, "can not find out channel account")
	}
	err = s.uc.SavePayOrder(ctx, req, gatewayOrderID, channelAccount)
	if err != nil {
		return nil, err
	}

	payResponse, err := s.uc.ChannelPay(ctx, req, gatewayOrderID, v)
	if err != nil {
		return nil, err
	}
	return &pb.PayResponse{
		GatewayOrderId:     gatewayOrderID,
		ChannelPayResponse: payResponse.Data,
	}, nil
}
