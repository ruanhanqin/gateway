package biz

import (
	"context"
	"encoding/json"

	"github.com/Naist4869/common/api/common"

	pool "github.com/Naist4869/base/pool"
	"github.com/Naist4869/base/tools/util"

	_errors "errors"

	"github.com/Naist4869/common/api/channel"
	"github.com/Naist4869/common/api/database"
	"github.com/Naist4869/common/api/errno"
	pb "github.com/Naist4869/common/api/gateway"
	"github.com/Naist4869/common/model"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/spf13/viper"
)

type Greeter struct {
	Hello string
}

type GreeterRepo interface {
	CreateGreeter(*Greeter) error
	UpdateGreeter(*Greeter) error
	SavePayOrder(ctx context.Context, order *database.PayOrder) error
	UpdatePayOrder(ctx context.Context, order *database.PayOrder) error
	GetChannelConfig(ctx context.Context, in *database.FindChannelConfigReq) (*database.FindChannelConfigResp, error)
}

type GreeterUsecase struct {
	repo   GreeterRepo
	log    *log.Helper
	client *pool.ServiceClientPool
}

func NewGreeterUsecase(repo GreeterRepo, logger log.Logger) *GreeterUsecase {
	return &GreeterUsecase{repo: repo, log: log.NewHelper("usecase/greeter", logger), client: pool.NewServiceClientPool(pool.NewDefaultClientOption())}
}
func (uc *GreeterUsecase) SavePayOrder(ctx context.Context, req *pb.PayRequest, gatewayOrderID, channelAccount string) (err error) {

	order := newPayOrder(req, gatewayOrderID, channelAccount)
	return uc.repo.SavePayOrder(ctx, order)
}

func (uc *GreeterUsecase) GetDefaultChannelConfig(ctx context.Context, appID string, method common.Method) (*database.ChannelConfig, error) {
	resp, err := uc.repo.GetChannelConfig(ctx, &database.FindChannelConfigReq{
		Method:    method,
		Available: true,
		AppID:     appID,
	})
	if err != nil {
		return nil, err
	}
	return resp.ChannelConfigs[0], nil
}

func (uc *GreeterUsecase) GetChannelConfig(ctx context.Context, appID string, method common.Method, channelID string) (*database.ChannelConfig, error) {
	resp, err := uc.repo.GetChannelConfig(ctx, &database.FindChannelConfigReq{
		ChannelID: channelID,
		Method:    method,
		Available: true,
		AppID:     appID,
	})
	if err != nil {
		return nil, err
	}
	return resp.ChannelConfigs[0], nil
}

// ChannelPay
func (uc *GreeterUsecase) ChannelPay(ctx context.Context, req *pb.PayRequest, gatewayOrderID string, config *database.ChannelConfig) (payResponse *channel.ChannelPayResponse, err error) {
	payRequest := newChannelPayRequest(req, gatewayOrderID, config.ChannelAccount)
	payConfig := new(model.PayConfig)
	if err = viper.UnmarshalKey("app", payConfig); err != nil {
		uc.log.Errorf("failed viper unmarshalkey: %v", err)
		return
	}
	payRequest.NotifyUrl = util.ReplaceGatewayOrderId(payConfig.NotifyURLPattern, gatewayOrderID)
	payRequest.ReturnUrl = util.ReplaceGatewayOrderId(payConfig.ReturnURLPattern, gatewayOrderID)
	if extJSON := req.ExtJson; extJSON != "" {
		meta := make(map[string]string)
		if err = json.Unmarshal([]byte(extJSON), &meta); err != nil {
			return nil, errors.Internal(errno.Errors_ErrUnmarshal, "failed to unmarshal json: %v error: %v", extJSON, err)
		}
		payRequest.Meta = meta
	}

	conn, err := uc.client.GetClient(config.AppID + "/" + config.ChannelID)
	if err != nil {
		// 拨号失败或者未找到Client
		if _errors.Is(err, pool.ErrNotFoundClient) || _errors.Is(err, context.DeadlineExceeded) {
			uc.client.Set(config.Target, config.AppID+"/"+config.ChannelID)
			if conn, err = uc.client.GetClient(config.AppID + "/" + config.ChannelID); err != nil {
				return nil, errors.Unavailable(errno.Errors_ErrNotFindService, "not found channel service: %v", err)
			}
		} else {
			return nil, errors.Unavailable(errno.Errors_ErrNotFindService, "not found channel service: %v", err)
		}
	}
	// 更换了目标地址
	if conn.Target() != config.Target {
		uc.client.Set(config.Target, config.AppID+"/"+config.ChannelID)
		if conn, err = uc.client.GetClient(config.AppID + "/" + config.ChannelID); err != nil {
			return nil, errors.Unavailable(errno.Errors_ErrNotFindService, "not found channel service: %v", err)
		}
	}
	channelClient := channel.NewPayChannelClient(conn)
	order := newPayOrder(req, gatewayOrderID, config.ChannelAccount)
	if payResponse, err = channelClient.Pay(ctx, payRequest); err != nil {
		uc.log.Errorf("Pay channel failed! err: %s channelPayResponse: %v", err, payResponse)
		order.BasePayOrder.ErrorMessage = err.Error()
	} else {
		if data := payResponse.Data; data != nil {
			if channelResponseJSON, err := json.Marshal(data); err != nil {
				uc.log.Errorf("Failed to marshal object: %v to json! error: %v", data, err)
				order.BasePayOrder.ErrorMessage = err.Error()
			} else {
				order.BasePayOrder.ChannelResponseJson = string(channelResponseJSON)
			}
		} else {
			return nil, errors.Unimplemented(errno.Errors_ErrRespNULL, "channel response fail!")
		}
	}
	if err = uc.repo.UpdatePayOrder(ctx, order); err != nil {
		uc.log.Errorf("Failed to save order: %v returns error: %v", order, err)
		return nil, err
	}
	uc.log.Info("Save db result: %v", order)
	if payResponse == nil {
		return nil, errors.Internal(errno.Errors_ErrPayChannelReturnNull, order.BasePayOrder.ChannelResponseJson)
	}
	return payResponse, nil
}

func (uc *GreeterUsecase) Create(g *Greeter) error {
	return uc.repo.CreateGreeter(g)
}

func (uc *GreeterUsecase) Update(g *Greeter) error {
	return uc.repo.UpdateGreeter(g)
}
