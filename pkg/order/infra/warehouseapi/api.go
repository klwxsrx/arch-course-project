package warehouseapi

import (
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/event"
	"github.com/klwxsrx/arch-course-project/pkg/order/app/service/async"
)

const WarehouseMessageTopicName = "warehouse-event"

type apiClient struct {
	messageDispatcher event.Dispatcher
}

func (a *apiClient) ReserveOrderItems(orderID uuid.UUID, items []async.ItemQuantity) error {
	// TODO:
	return nil
}

func (a *apiClient) RemoveOrderItemsReservation(orderID uuid.UUID) error {
	// TODO:
	return nil
}

func New(messageDispatcher event.Dispatcher) async.WarehouseAPI {
	return &apiClient{messageDispatcher: messageDispatcher}
}
