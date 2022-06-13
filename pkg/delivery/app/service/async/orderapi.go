package async

import "github.com/google/uuid"

type OrderAPI interface {
	NotifyDeliveryScheduled(orderID uuid.UUID) error
}
