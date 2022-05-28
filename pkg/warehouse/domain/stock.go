package domain

import "github.com/google/uuid"

type StockOperationType int

const (
	StockOperationTypeArrival = iota
	StockOperationTypeReservation
	StockOperationTypeSale
)

type StockOperation struct {
	ID           uuid.UUID
	ItemID       uuid.UUID
	Type         StockOperationType
	ItemQuantity int
	OrderID      *uuid.UUID
}

type ItemQuantity struct {
	ItemID   uuid.UUID `db:"item_id"`
	Quantity int       `db:"quantity"`
}

type Stock interface {
	NextID() uuid.UUID
	GetAvailableItemsQuantity(itemIDs []uuid.UUID) ([]ItemQuantity, error)
	GetOrderOperations(orderID uuid.UUID) ([]StockOperation, error)
	Update(op *StockOperation) error
	Delete(opIDs []uuid.UUID) error
}
