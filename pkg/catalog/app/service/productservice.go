package service

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/catalog/app/persistence"
	"github.com/klwxsrx/arch-course-project/pkg/catalog/domain"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/log"
)

var (
	ErrInvalidProperty  = errors.New("invalid property")
	ErrProductNotExists = errors.New("product not exists")
)

type ProductService struct {
	ufw    persistence.UnitOfWork
	logger log.Logger
}

func (s *ProductService) Add(title, description string, price int) error {
	err := s.validateProductProperties(title, price)
	if err != nil {
		return err
	}

	err = s.ufw.Execute(func(p persistence.PersistentProvider) error {
		product := &domain.Product{
			ID:          p.ProductRepository().NextID(),
			Title:       title,
			Description: description,
			Price:       price,
		}

		return p.ProductRepository().Store(product)
	})
	if err != nil {
		s.logger.WithError(err).With(log.Fields{"title": title}).Error("failed to add product")
	}
	return err
}

func (s *ProductService) Update(id uuid.UUID, title, description string, price int) error {
	err := s.validateProductProperties(title, price)
	if err != nil {
		return err
	}

	err = s.ufw.Execute(func(p persistence.PersistentProvider) error {
		product, err := p.ProductRepository().GetByID(id)
		if errors.Is(err, domain.ErrProductNotExists) {
			return ErrProductNotExists
		}

		product.Title = title
		product.Description = description
		product.Price = price

		return p.ProductRepository().Store(product)
	})
	if err != nil {
		s.logger.WithError(err).With(log.Fields{"id": id}).Error("failed to update product")
	}
	return err
}

func (s *ProductService) validateProductProperties(title string, price int) error {
	if title == "" {
		return fmt.Errorf("%w title: %s", ErrInvalidProperty, title)
	}
	if price <= 0 {
		return fmt.Errorf("%w price: %d", ErrInvalidProperty, price)
	}
	return nil
}

func NewProductService(ufw persistence.UnitOfWork, logger log.Logger) *ProductService {
	return &ProductService{ufw: ufw, logger: logger}
}
