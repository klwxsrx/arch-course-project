package query

import "github.com/google/uuid"

type ProductData struct {
	ID          uuid.UUID
	Title       string
	Description string
	Price       int
}

type ProductService interface {
	ListAll() ([]ProductData, error)
}
