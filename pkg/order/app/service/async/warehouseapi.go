package async

import (
	"github.com/google/uuid"
)

type ItemQuantity struct {
	ItemID   uuid.UUID
	Quantity int
}

type WarehouseAPI interface {
	ReserveItems(orderID uuid.UUID, items []ItemQuantity) error
	RemoveItemsReservation(orderID uuid.UUID) error
}
