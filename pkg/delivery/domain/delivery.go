package domain

import (
	"errors"
	"github.com/google/uuid"
)

type DeliveryStatus int

const (
	DeliveryStatusScheduled DeliveryStatus = iota
	DeliveryStatusAwaitingDelivery
	DeliveryStatusProcessing
	DeliveryStatusDelivered
	DeliveryStatusCancelled
)

type Delivery struct {
	OrderID uuid.UUID
	Status  DeliveryStatus
	Address string
}

var ErrItemNotFound = errors.New("item not found")

type DeliveryRepository interface {
	GetByID(orderID uuid.UUID) (*Delivery, error)
	Store(d *Delivery) error
}
