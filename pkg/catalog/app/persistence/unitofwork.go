package persistence

import "github.com/klwxsrx/arch-course-project/pkg/catalog/domain"

type PersistentProvider interface {
	ProductRepository() domain.ProductRepository
}

type UnitOfWork interface {
	Execute(f func(p PersistentProvider) error) error
}
