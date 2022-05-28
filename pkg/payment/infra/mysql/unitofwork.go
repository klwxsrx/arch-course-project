package mysql

import (
	"fmt"
	"github.com/klwxsrx/arch-course-project/pkg/common/infra/mysql"
	"github.com/klwxsrx/arch-course-project/pkg/payment/app/persistence"
	"github.com/klwxsrx/arch-course-project/pkg/payment/domain"
)

type persistentProvider struct {
	db mysql.Client
}

func (p *persistentProvider) PaymentRepository() domain.PaymentRepository {
	return NewPaymentRepository(p.db)
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
