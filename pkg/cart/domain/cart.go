package domain

import "github.com/google/uuid"

type ProductQuantity struct {
	ID       uuid.UUID
	Quantity int
}

type Cart struct {
	UserID   uuid.UUID
	Products []ProductQuantity
}

type CartStorage interface {
	GetByUserID(userID uuid.UUID) (*Cart, error)
	Store(cart *Cart) error
	Delete(userID uuid.UUID) error
}
