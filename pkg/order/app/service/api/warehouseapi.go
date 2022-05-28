package api

import (
	"errors"
	"github.com/google/uuid"
)

var (
	ErrOrderItemsOutOfStock        = errors.New("items out of stock")
	ErrOrderOperationsAlreadyExist = errors.New("order items are already exist")
)

type ItemQuantity struct {
	ItemID   uuid.UUID
	Quantity int
}

type WarehouseAPI interface {
	ReserveOrderItems(orderID uuid.UUID, items []ItemQuantity) error
	RemoveOrderItemsReservation(orderID uuid.UUID) error
}
