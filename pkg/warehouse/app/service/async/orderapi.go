package async

import "github.com/google/uuid"

type OrderAPI interface {
	NotifyItemsReserved(orderID uuid.UUID) error
	NotifyItemsOutOfStock(orderID uuid.UUID) error
}
