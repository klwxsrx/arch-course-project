package persistence

import "github.com/klwxsrx/arch-course-project/pkg/delivery/domain"

type PersistentProvider interface {
	DeliveryRepository() domain.DeliveryRepository
}

type UnitOfWork interface {
	Execute(f func(p PersistentProvider) error) error
}
