package message

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/message"
	"github.com/klwxsrx/arch-course-project/pkg/order/app/service"
)

type paymentCompletedHandler struct {
	orderService *service.OrderService
}

func (h *paymentCompletedHandler) TopicName() string {
	return orderEventTopicName
}

func (h *paymentCompletedHandler) Type() string {
	return "payment_completed"
}

func (h *paymentCompletedHandler) Handle(msg *message.Message) error {
	var orderID uuid.UUID
	err := json.Unmarshal(msg.Body, &orderID)
	if err != nil {
		return fmt.Errorf("failed to decode message")
	}

	err = h.orderService.HandlePaymentCompleted(orderID)
	if err != nil {
		return fmt.Errorf("failed to handle payment completed: %w", err)
	}
	return nil
}

func NewPaymentCompletedHandler(orderService *service.OrderService) message.Handler {
	return &paymentCompletedHandler{orderService: orderService}
}
