package api

import "github.com/google/uuid"

type DeliveryAPI interface {
	ScheduleDelivery(orderID uuid.UUID, addressID uuid.UUID) error
	DeleteDeliverySchedule(orderID uuid.UUID) error
}
