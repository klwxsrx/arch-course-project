package persistence

import (
	"github.com/klwxsrx/arch-course-project/pkg/common/app/idempotence"
	"github.com/klwxsrx/arch-course-project/pkg/order/domain"
)

type PersistentProvider interface {
	OrderRepository() domain.OrderRepository
	IdempotenceKeyStore() idempotence.KeyStore
}

type UnitOfWork interface {
	Execute(f func(p PersistentProvider) error) error
}
