package query

import (
	"errors"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/payment/domain"
)

var ErrPaymentNotFound = errors.New("payment not found")

type PaymentData struct {
	OrderID     uuid.UUID
	Status      domain.PaymentStatus
	TotalAmount int
}

type PaymentQueryService interface {
	GetPayment(orderID uuid.UUID) (*PaymentData, error)
}
