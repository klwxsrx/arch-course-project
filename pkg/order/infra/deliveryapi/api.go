package deliveryapi

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/event"
	"github.com/klwxsrx/arch-course-project/pkg/order/app/service/async"
)

const deliveryEventTopicName = "delivery_event"

type apiClient struct {
	eventDispatcher event.Dispatcher
}

func (a *apiClient) ScheduleDelivery(orderID uuid.UUID, addressID uuid.UUID) error {
	body := struct {
		OrderID   uuid.UUID `json:"order_id"`
		AddressID uuid.UUID `json:"address_id"`
	}{
		OrderID:   orderID,
		AddressID: addressID,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return errors.New("failed to encode body")
	}

	err = a.eventDispatcher.Dispatch(&event.Event{
		Type:      "schedule_delivery",
		TopicName: deliveryEventTopicName,
		Key:       orderID.String(),
		Body:      jsonBody,
	})
	if err != nil {
		return errors.New("failed to dispatch message")
	}
	return nil
}

func (a *apiClient) CancelDeliverySchedule(orderID uuid.UUID) error {
	jsonID, err := json.Marshal(orderID)
	if err != nil {
		return errors.New("failed to encode orderID")
	}

	err = a.eventDispatcher.Dispatch(&event.Event{
		Type:      "cancel_schedule",
		TopicName: deliveryEventTopicName,
		Key:       orderID.String(),
		Body:      jsonID,
	})
	if err != nil {
		return errors.New("failed to dispatch message")
	}
	return nil
}

func (a *apiClient) ProcessDelivery(orderID uuid.UUID) error {
	jsonID, err := json.Marshal(orderID)
	if err != nil {
		return errors.New("failed to encode orderID")
	}

	err = a.eventDispatcher.Dispatch(&event.Event{
		Type:      "process_delivery",
		TopicName: deliveryEventTopicName,
		Key:       orderID.String(),
		Body:      jsonID,
	})
	if err != nil {
		return errors.New("failed to dispatch message")
	}
	return nil
}

func New(eventDispatcher event.Dispatcher) async.DeliveryAPI {
	return &apiClient{eventDispatcher: eventDispatcher}
}
