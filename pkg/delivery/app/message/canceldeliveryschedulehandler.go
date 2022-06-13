package message

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/message"
	"github.com/klwxsrx/arch-course-project/pkg/delivery/app/service"
)

type cancelDeliveryScheduleHandler struct {
	service *service.DeliveryService
}

func (h *cancelDeliveryScheduleHandler) TopicName() string {
	return deliveryTopicName
}

func (h *cancelDeliveryScheduleHandler) Type() string {
	return "cancel_schedule"
}

func (h *cancelDeliveryScheduleHandler) Handle(msg *message.Message) error {
	var orderID uuid.UUID
	err := json.Unmarshal(msg.Body, &orderID)
	if err != nil {
		return fmt.Errorf("failed to decode message")
	}

	err = h.service.CancelSchedule(orderID)
	if err != nil {
		return fmt.Errorf("failed to cancel delivery schedule: %w", err)
	}
	return nil
}

func NewCancelDeliveryScheduleHandler(service *service.DeliveryService) message.Handler {
	return &cancelDeliveryScheduleHandler{
		service: service,
	}
}
