package async

import "github.com/google/uuid"

type OrderAPI interface {
	NotifyPaymentAuthorized(orderID uuid.UUID) error
	NotifyPaymentCompleted(orderID uuid.UUID) error
}
