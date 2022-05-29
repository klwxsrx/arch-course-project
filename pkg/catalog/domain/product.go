package domain

import (
	"errors"
	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID
	Title       string
	Description string
	Price       int
}

var (
	ErrProductNotExists = errors.New("product is not exists")
)

type ProductRepository interface {
	NextID() uuid.UUID
	GetByID(id uuid.UUID) (*Product, error)
	Store(product *Product) error
}
