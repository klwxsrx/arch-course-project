package mysql

import (
	"fmt"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/event"
	"github.com/klwxsrx/arch-course-project/pkg/common/infra/mysql"
	"github.com/klwxsrx/arch-course-project/pkg/delivery/app/persistence"
	"github.com/klwxsrx/arch-course-project/pkg/delivery/app/service/async"
	"github.com/klwxsrx/arch-course-project/pkg/delivery/domain"
	"github.com/klwxsrx/arch-course-project/pkg/delivery/infra/orderapi"
)

type persistentProvider struct {
	db mysql.Client
}

func (p *persistentProvider) DeliveryRepository() domain.DeliveryRepository {
	return NewDeliveryRepository(p.db)
}

func (p *persistentProvider) OrderAPI() async.OrderAPI {
	return orderapi.New(p.eventDispatcher(p.db))
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
