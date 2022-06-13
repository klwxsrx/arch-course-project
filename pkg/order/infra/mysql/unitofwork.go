package mysql

import (
	"fmt"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/event"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/idempotence"
	"github.com/klwxsrx/arch-course-project/pkg/common/infra/mysql"
	"github.com/klwxsrx/arch-course-project/pkg/order/app/persistence"
	"github.com/klwxsrx/arch-course-project/pkg/order/app/service/async"
	"github.com/klwxsrx/arch-course-project/pkg/order/domain"
	"github.com/klwxsrx/arch-course-project/pkg/order/infra/deliveryapi"
	"github.com/klwxsrx/arch-course-project/pkg/order/infra/paymentapi"
	"github.com/klwxsrx/arch-course-project/pkg/order/infra/warehouseapi"
)

type persistentProvider struct {
	db mysql.Client
}

func (p *persistentProvider) PaymentAPI() async.PaymentAPI {
	return paymentapi.New(p.eventDispatcher(p.db))
}

func (p *persistentProvider) WarehouseAPI() async.WarehouseAPI {
	return warehouseapi.New(p.eventDispatcher(p.db))
}

func (p *persistentProvider) DeliveryAPI() async.DeliveryAPI {
	return deliveryapi.New(p.eventDispatcher(p.db))
}

func (p *persistentProvider) OrderRepository() domain.OrderRepository {
	return NewOrderRepository(p.db)
}

func (p *persistentProvider) IdempotenceKeyStore() idempotence.KeyStore {
	return mysql.NewIdempotenceKeyStore(p.db)
}

func (p *persistentProvider) eventDispatcher(db mysql.Client) event.Dispatcher {
	return event.NewDispatcher(mysql.NewMessageStore(db))
}

type unitOfWork struct {
	client mysql.TransactionalClient
}

func (u *unitOfWork) Execute(f func(p persistence.PersistentProvider) error) error {
	tx, err := u.client.Begin()
	if err != nil {
		return fmt.Errorf("failed to start tx: %w", err)
	}

	pp := &persistentProvider{tx}
	err = f(pp)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit tx: %w", err)
	}
	return nil
}

func NewUnitOfWork(client mysql.TransactionalClient) persistence.UnitOfWork {
	return &unitOfWork{client: client}
}
