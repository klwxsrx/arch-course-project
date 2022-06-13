package message

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/message"
	"github.com/klwxsrx/arch-course-project/pkg/order/app/service"
)

type paymentCompletionRejectedHandler struct {
	orderService *service.OrderService
}

func (h *paymentCompletionRejectedHandler) TopicName() string {
	return orderEventTopicName
}

func (h *paymentCompletionRejectedHandler) Type() string {
	return "payment_completion_rejected"
}

func (h *paymentCompletionRejectedHandler) Handle(msg *message.Message) error {
	var orderID uuid.UUID
	err := json.Unmarshal(msg.Body, &orderID)
	if err != nil {
		return fmt.Errorf("failed to decode message")
	}

	err = h.orderService.HandlePaymentCompletionRejected(orderID)
	if err != nil {
		return fmt.Errorf("failed to handle payment compoletion rejected: %w", err)
	}
	return nil
}

func NewPaymentCompletionRejectedHandler(orderService *service.OrderService) message.Handler {
	return &paymentCompletionRejectedHandler{orderService: orderService}
}
