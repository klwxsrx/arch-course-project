package query

import (
	"errors"
	"github.com/google/uuid"
)

type ProductData struct {
	ID          uuid.UUID
	Title       string
	Description string
	Price       int
}

var ErrProductByIDNotFound = errors.New("product by id is not found")

type ProductService interface {
	ListAll() ([]ProductData, error)
	GetByIDs(ids []uuid.UUID) ([]ProductData, error)
}
