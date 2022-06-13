package persistence

import (
	"github.com/klwxsrx/arch-course-project/pkg/delivery/app/service/async"
	"github.com/klwxsrx/arch-course-project/pkg/delivery/domain"
)

type PersistentProvider interface {
	DeliveryRepository() domain.DeliveryRepository
	OrderAPI() async.OrderAPI
}

type UnitOfWork interface {
	Execute(f func(p PersistentProvider) error) error
}
