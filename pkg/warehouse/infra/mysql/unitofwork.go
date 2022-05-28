package mysql

import (
	"fmt"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/idempotence"
	"github.com/klwxsrx/arch-course-project/pkg/common/infra/mysql"
	mysql2 "github.com/klwxsrx/arch-course-project/pkg/order/infra/mysql"
	"github.com/klwxsrx/arch-course-project/pkg/warehouse/app/persistence"
	"github.com/klwxsrx/arch-course-project/pkg/warehouse/domain"
)

type persistentProvider struct {
	db mysql.Client
}

func (p *persistentProvider) Stock() domain.Stock {
	return NewStock(p.db)
}

func (p *persistentProvider) IdempotenceKeyStore() idempotence.KeyStore {
	return mysql2.NewIdempotenceKeyStore(p.db)
}

type unitOfWork struct {
	client mysql.TransactionalClient
}

func (u *unitOfWork) Execute(lockName string, f func(p persistence.PersistentProvider) error) error {
	if lockName != "" {
		lock := mysql.NewLock(u.client, lockName)
		err := lock.Get()
		if err != nil {
			return fmt.Errorf("failed to get lock %v: %w", lockName, err)
		}

		defer lock.Release()
	}

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
