package api

import (
	"github.com/google/uuid"
)

type CreateOrderProductData struct {
	ID           uuid.UUID
	ProductPrice int
	Quantity     int
}

type CreateOrderData struct {
	IdempotenceKey string
	UserID         uuid.UUID
	AddressID      uuid.UUID
	Products       []CreateOrderProductData
}

type OrderAPI interface {
	CreateOrder(data *CreateOrderData) (orderID uuid.UUID, err error)
}
