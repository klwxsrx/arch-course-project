package orderapi

import (
	"errors"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/event"
	"github.com/klwxsrx/arch-course-project/pkg/payment/app/service/async"
)

const OrderEventTopicName = "order_event"

type api struct {
	eventDispatcher event.Dispatcher
}

func (a *api) NotifyPaymentAuthorized(orderID uuid.UUID) error {
	err := a.eventDispatcher.Dispatch(&event.Event{
		Type:      "payment_authorized",
		TopicName: OrderEventTopicName,
		Key:       orderID.String(),
		Body:      []byte(orderID.String()),
	})
	if err != nil {
		return errors.New("failed to dispatch message")
	}
	return nil
}

func (a *api) NotifyPaymentCompleted(orderID uuid.UUID) error {
	err := a.eventDispatcher.Dispatch(&event.Event{
		Type:      "payment_completed",
		TopicName: OrderEventTopicName,
		Key:       orderID.String(),
		Body:      []byte(orderID.String()),
	})
	if err != nil {
		return errors.New("failed to dispatch message")
	}
	return nil
}

func New(eventDispatcher event.Dispatcher) async.OrderAPI {
	return &api{eventDispatcher: eventDispatcher}
}
