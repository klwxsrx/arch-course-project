package deliveryapi

import (
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/event"
	"github.com/klwxsrx/arch-course-project/pkg/order/app/service/async"
)

type apiClient struct {
	eventDispatcher event.Dispatcher
}

func (a *apiClient) ScheduleDelivery(orderID uuid.UUID, addressID uuid.UUID) error {
	// TODO:
	return nil
}

func (a *apiClient) DeleteDeliverySchedule(orderID uuid.UUID) error {
	// TODO:
	return nil
}

func New(eventDispatcher event.Dispatcher) async.DeliveryAPI {
	return &apiClient{eventDispatcher: eventDispatcher}
}
