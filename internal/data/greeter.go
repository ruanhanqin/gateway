package data

import (
	"context"

	"github.com/Naist4869/gateway/internal/biz"

	pool "github.com/Naist4869/base/pool"
	"github.com/Naist4869/common/api/database"
	"github.com/Naist4869/common/api/errno"
	"github.com/go-kratos/kratos/v2/errors"

	"github.com/go-kratos/kratos/v2/log"
)

type greeterRepo struct {
	data   *Data
	log    *log.Helper
	client *pool.ServiceClientPool
}

func (r *greeterRepo) UpdatePayOrder(ctx context.Context, order *database.PayOrder) error {
	conn, err := r.client.GetClient(database.PayDatabaseService_ServiceDesc.ServiceName)
	if err != nil {
		return errors.Unavailable(errno.Errors_ErrNotFindService, "not found database service")
	}
	databaseServiceClient := database.NewPayDatabaseServiceClient(conn)
	_, err = databaseServiceClient.UpdatePayOrder(ctx, order)
	return err
}

// NewGreeterRepo .
func NewGreeterRepo(data *Data, logger log.Logger) biz.GreeterRepo {
	m := pool.NewTargetServiceNames()
	m.Set("127.0.0.1:16210", database.PayDatabaseService_ServiceDesc.ServiceName)
	serviceClientPool := pool.NewServiceClientPool(pool.NewDefaultClientOption())
	serviceClientPool.Init(*m)
	return &greeterRepo{
		data:   data,
		log:    log.NewHelper("data/greeter", logger),
		client: serviceClientPool,
	}
}

func (r *greeterRepo) SavePayOrder(ctx context.Context, order *database.PayOrder) error {
	conn, err := r.client.GetClient(database.PayDatabaseService_ServiceDesc.ServiceName)
	if err != nil {
		return errors.Unavailable(errno.Errors_ErrNotFindService, "not found database service")
	}
	databaseServiceClient := database.NewPayDatabaseServiceClient(conn)
	_, err = databaseServiceClient.SavePayOrder(ctx, order)
	return err
}

func (r *greeterRepo) GetChannelConfig(ctx context.Context, in *database.FindChannelConfigReq) (*database.FindChannelConfigResp, error) {
	conn, err := r.client.GetClient(database.PayDatabaseService_ServiceDesc.ServiceName)
	if err != nil {
		return nil, errors.Unavailable(errno.Errors_ErrNotFindService, "not found database service")
	}
	databaseServiceClient := database.NewPayDatabaseServiceClient(conn)
	return databaseServiceClient.FindChannelConfig(ctx, in)
}
func (r *greeterRepo) CreateGreeter(g *biz.Greeter) error {
	return nil
}

func (r *greeterRepo) UpdateGreeter(g *biz.Greeter) error {
	return nil
}
