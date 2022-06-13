package message

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/message"
	"github.com/klwxsrx/arch-course-project/pkg/order/app/service"
)

type deliveryScheduledHandler struct {
	service *service.OrderService
}

func (h *deliveryScheduledHandler) TopicName() string {
	return orderEventTopicName
}

func (h *deliveryScheduledHandler) Type() string {
	return "delivery_scheduled"
}

func (h *deliveryScheduledHandler) Handle(msg *message.Message) error {
	var orderID uuid.UUID
	err := json.Unmarshal(msg.Body, &orderID)
	if err != nil {
		return fmt.Errorf("failed to decode message")
	}

	err = h.service.HandleDeliveryScheduled(orderID)
	if err != nil {
		return fmt.Errorf("failed to handle delivery scheduled: %w", err)
	}
	return nil
}

func NewDeliveryScheduledHandler(service *service.OrderService) message.Handler {
	return &deliveryScheduledHandler{service: service}
}
