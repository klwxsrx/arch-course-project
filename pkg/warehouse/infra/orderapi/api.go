package orderapi

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/event"
	"github.com/klwxsrx/arch-course-project/pkg/warehouse/app/service/async"
)

const OrderEventTopicName = "order_event"

type api struct {
	eventDispatcher event.Dispatcher
}

func (a *api) NotifyItemsReserved(orderID uuid.UUID) error {
	jsonID, err := json.Marshal(orderID)
	if err != nil {
		return errors.New("failed to encode orderID")
	}

	err = a.eventDispatcher.Dispatch(&event.Event{
		Type:      "items_reserved",
		TopicName: OrderEventTopicName,
		Key:       orderID.String(),
		Body:      jsonID,
	})
	if err != nil {
		return errors.New("failed to dispatch message")
	}
	return nil
}

func (a *api) NotifyItemsOutOfStock(orderID uuid.UUID) error {
	jsonID, err := json.Marshal(orderID)
	if err != nil {
		return errors.New("failed to encode orderID")
	}

	err = a.eventDispatcher.Dispatch(&event.Event{
		Type:      "items_out_of_stock",
		TopicName: OrderEventTopicName,
		Key:       orderID.String(),
		Body:      jsonID,
	})
	if err != nil {
		return errors.New("failed to dispatch message")
	}
	return nil
}

func New(eventDispatcher event.Dispatcher) async.OrderAPI {
	return &api{eventDispatcher: eventDispatcher}
}
