package async

import (
	"github.com/google/uuid"
)

type PaymentAPI interface {
	AuthorizeOrder(orderID uuid.UUID, totalAmount int) error
	CompleteTransaction(orderID uuid.UUID) error
	CancelPayment(orderID uuid.UUID) error
}
