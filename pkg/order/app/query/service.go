package query

import (
	"errors"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/order/domain"
)

type OrderItemData struct {
	ID        uuid.UUID
	ItemPrice int
	Quantity  int
}

type OrderData struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	AddressID   uuid.UUID
	Items       []OrderItemData
	Status      domain.OrderStatus
	TotalAmount int
}

var ErrOrderNotFound = errors.New("order not found")

type Service interface {
	GetOrderData(id uuid.UUID) (*OrderData, error)
}
