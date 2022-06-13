package message

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/message"
	"github.com/klwxsrx/arch-course-project/pkg/delivery/app/service"
)

type scheduleDeliveryHandler struct {
	service *service.DeliveryService
}

func (h *scheduleDeliveryHandler) TopicName() string {
	return deliveryTopicName
}

func (h *scheduleDeliveryHandler) Type() string {
	return "schedule_delivery"
}

func (h *scheduleDeliveryHandler) Handle(msg *message.Message) error {
	body := struct {
		OrderID   uuid.UUID `json:"order_id"`
		AddressID uuid.UUID `json:"address_id"`
	}{}

	err := json.Unmarshal(msg.Body, &body)
	if err != nil {
		return fmt.Errorf("failed to decode message")
	}

	err = h.service.Schedule(body.OrderID, body.AddressID)
	if err != nil {
		return fmt.Errorf("failed to schedule delivery: %w", err)
	}
	return nil
}

func NewScheduleDeliveryHandler(service *service.DeliveryService) message.Handler {
	return &scheduleDeliveryHandler{service: service}
}
