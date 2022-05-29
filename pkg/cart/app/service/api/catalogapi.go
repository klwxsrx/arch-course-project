package api

import (
	"errors"
	"github.com/google/uuid"
)

type Product struct {
	ID    uuid.UUID
	Price int
}

var ErrProductsNotFound = errors.New("one or more products are not found")

type CatalogAPI interface {
	GetProducts(productIDs []uuid.UUID) ([]Product, error)
}
