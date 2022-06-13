package message

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/message"
	"github.com/klwxsrx/arch-course-project/pkg/delivery/app/service"
)

type processDeliveryHandler struct {
	service *service.DeliveryService
}

func (h *processDeliveryHandler) TopicName() string {
	return deliveryTopicName
}

func (h *processDeliveryHandler) Type() string {
	return "process_delivery"
}

func (h *processDeliveryHandler) Handle(msg *message.Message) error {
	var orderID uuid.UUID
	err := json.Unmarshal(msg.Body, &orderID)
	if err != nil {
		return fmt.Errorf("failed to decode message")
	}

	err = h.service.ProcessDelivery(orderID)
	if err != nil {
		return fmt.Errorf("failed to process delivery: %w", err)
	}
	return nil
}

func NewProcessDeliveryHandler(service *service.DeliveryService) message.Handler {
	return &processDeliveryHandler{
		service: service,
	}
}
