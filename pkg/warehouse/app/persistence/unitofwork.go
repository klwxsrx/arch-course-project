package persistence

import (
	"github.com/klwxsrx/arch-course-project/pkg/common/app/idempotence"
	"github.com/klwxsrx/arch-course-project/pkg/warehouse/app/service/async"
	"github.com/klwxsrx/arch-course-project/pkg/warehouse/domain"
)

type PersistentProvider interface {
	Stock() domain.Stock
	IdempotenceKeyStore() idempotence.KeyStore
	OrderAPI() async.OrderAPI
}

type UnitOfWork interface {
	Execute(lockName string, f func(p PersistentProvider) error) error
}
