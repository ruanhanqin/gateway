package biz

import (
	"time"

	"github.com/Naist4869/common/api/common"

	"github.com/Naist4869/common/api/channel"
	"github.com/Naist4869/common/api/database"
	pb "github.com/Naist4869/common/api/gateway"
	"github.com/Naist4869/common/constant"
	"github.com/google/wire"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// RequestContext context of request
type RequestContext struct {
	GatewayOrderID     string
	ChannelAccount     string
	PayRequest         *pb.PayRequest
	PayOrder           *database.PayOrder
	ChannelPayRequest  *channel.ChannelPayRequest
	ChannelPayResponse *channel.ChannelPayResponse
	err                error
}

func newPayOrder(r *pb.PayRequest, gatewayOrderID, channelAccount string) *database.PayOrder {
	return &database.PayOrder{
		BasePayOrder: &database.BasePayOrder{
			Version:             r.Version,
			OutTradeNo:          r.OutTradeNo,
			ChannelAccount:      channelAccount,
			ChannelOrderId:      "",
			GatewayOrderId:      gatewayOrderID,
			PayAmount:           r.PayAmount,
			Currency:            r.Currency,
			NotifyUrl:           r.NotifyUrl,
			ReturnUrl:           r.ReturnUrl,
			AppId:               r.AppId,
			SignType:            r.SignType,
			OrderTime:           r.OrderTime,
			RequestTime:         timestamppb.New(time.Now()),
			CreateDate:          time.Now().Format("2006-01-02"),
			UserIp:              r.UserIp,
			UserId:              r.UserId,
			PayerAccount:        r.PayerAccount,
			ProductId:           r.ProductId,
			ProductName:         r.ProductName,
			ProductDescribe:     r.ProductDescribe,
			CallbackJson:        r.CallbackJson,
			ExtJson:             r.ExtJson,
			ChannelResponseJson: "",
			ErrorMessage:        "",
			ChannelId:           r.ChannelId,
			Method:              r.Method,
			Remark:              r.Remark,
			ProductQuantity:     r.ProductQuantity,
			ShippingAmount:      r.ShippingAmount,
			DiscountAmount:      r.DiscountAmount,
		},
		OrderStatus: constant.OrderStatusWaiting,
	}
}

func newChannelPayRequest(req *pb.PayRequest, gatewayOrderID, channelAccount string) *channel.ChannelPayRequest {
	return &channel.ChannelPayRequest{
		GatewayOrderId: gatewayOrderID,
		ChannelAccount: channelAccount,
		PayAmount:      req.PayAmount,
		ShippingAmount: req.ShippingAmount,
		DiscountAmount: req.DiscountAmount,
		Product: &common.Product{
			Id:          req.ProductId,
			Name:        req.ProductName,
			Description: req.ProductDescribe,
			Quantity:    req.ProductQuantity,
			Price:       req.ProductPrice,
			Total:       req.ProductTotal,
		},
		UserIp: req.UserIp,
		Method: req.Method,
	}
}

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(NewGreeterUsecase)
