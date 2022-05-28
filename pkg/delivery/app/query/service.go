package query

import (
	"errors"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/delivery/domain"
)

var ErrDeliveryNotFound = errors.New("delivery not found")

type Delivery struct {
	OrderID uuid.UUID
	Status  domain.DeliveryStatus
	Address string
}

type Service interface {
	GetByID(orderID uuid.UUID) (*Delivery, error)
}
