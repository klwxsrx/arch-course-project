package domain

import (
	"errors"
	"github.com/google/uuid"
)

type OrderStatus int

const (
	OrderStatusCreated OrderStatus = iota
	OrderStatusCancelled
	OrderStatusAwaitingDelivery
	OrderStatusProcessingDelivery
	OrderStatusDelivered
)

type OrderItem struct {
	ID        uuid.UUID
	ItemPrice int
	Quantity  int
}

type Order struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	AddressID   uuid.UUID
	Items       []OrderItem
	Status      OrderStatus
	TotalAmount int
}

var ErrOrderNotFound = errors.New("order not found")

type OrderRepository interface {
	NextID() uuid.UUID
	GetByID(id uuid.UUID) (*Order, error)
	Store(order *Order) error
}
