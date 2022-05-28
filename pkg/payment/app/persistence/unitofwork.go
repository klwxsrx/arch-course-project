package persistence

import (
	"github.com/klwxsrx/arch-course-project/pkg/payment/domain"
)

type PersistentProvider interface {
	PaymentRepository() domain.PaymentRepository
}

type UnitOfWork interface {
	Execute(f func(p PersistentProvider) error) error
}
