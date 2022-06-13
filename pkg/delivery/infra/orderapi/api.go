package orderapi

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/event"
	"github.com/klwxsrx/arch-course-project/pkg/delivery/app/service/async"
)

const orderEventTopicName = "order_event"

type api struct {
	eventDispatcher event.Dispatcher
}

func (a *api) NotifyDeliveryScheduled(orderID uuid.UUID) error {
	jsonID, err := json.Marshal(orderID)
	if err != nil {
		return errors.New("failed to encode orderID")
	}

	err = a.eventDispatcher.Dispatch(&event.Event{
		Type:      "delivery_scheduled",
		TopicName: orderEventTopicName,
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
