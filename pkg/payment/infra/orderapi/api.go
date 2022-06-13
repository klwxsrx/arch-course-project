package orderapi

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/event"
	"github.com/klwxsrx/arch-course-project/pkg/payment/app/service/async"
)

const orderEventTopicName = "order_event"

type api struct {
	eventDispatcher event.Dispatcher
}

func (a *api) NotifyPaymentAuthorized(orderID uuid.UUID) error {
	jsonID, err := json.Marshal(orderID)
	if err != nil {
		return errors.New("failed to encode orderID")
	}

	err = a.eventDispatcher.Dispatch(&event.Event{
		Type:      "payment_authorized",
		TopicName: orderEventTopicName,
		Key:       orderID.String(),
		Body:      jsonID,
	})
	if err != nil {
		return errors.New("failed to dispatch message")
	}
	return nil
}

func (a *api) NotifyPaymentCompleted(orderID uuid.UUID) error {
	jsonID, err := json.Marshal(orderID)
	if err != nil {
		return errors.New("failed to encode orderID")
	}

	err = a.eventDispatcher.Dispatch(&event.Event{
		Type:      "payment_completed",
		TopicName: orderEventTopicName,
		Key:       orderID.String(),
		Body:      jsonID,
	})
	if err != nil {
		return errors.New("failed to dispatch message")
	}
	return nil
}

func (a *api) NotifyPaymentCompletionRejected(orderID uuid.UUID) error {
	jsonID, err := json.Marshal(orderID)
	if err != nil {
		return errors.New("failed to encode orderID")
	}

	err = a.eventDispatcher.Dispatch(&event.Event{
		Type:      "payment_completion_rejected",
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
