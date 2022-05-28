package domain

import (
	"errors"
	"github.com/google/uuid"
)

type PaymentStatus int

const (
	PaymentStatusAuthorized PaymentStatus = iota
	PaymentStatusCancelled
	PaymentStatusCompleted
	PaymentStatusRejected
)

type Payment struct {
	OrderID     uuid.UUID
	TotalAmount int
	Status      PaymentStatus
}

var ErrPaymentNotFound = errors.New("payment not found")

type PaymentRepository interface {
	GetByID(id uuid.UUID) (*Payment, error)
	Store(payment *Payment) error
}
