package persistence

import (
	"github.com/klwxsrx/arch-course-project/pkg/payment/app/service/async"
	"github.com/klwxsrx/arch-course-project/pkg/payment/domain"
)

type PersistentProvider interface {
	PaymentRepository() domain.PaymentRepository
	OrderAPI() async.OrderAPI
}

type UnitOfWork interface {
	Execute(f func(p PersistentProvider) error) error
}
