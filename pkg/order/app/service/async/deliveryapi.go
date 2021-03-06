package async

import "github.com/google/uuid"

type DeliveryAPI interface {
	ScheduleDelivery(orderID uuid.UUID, addressID uuid.UUID) error
	CancelDeliverySchedule(orderID uuid.UUID) error
	ProcessDelivery(orderID uuid.UUID) error
}
