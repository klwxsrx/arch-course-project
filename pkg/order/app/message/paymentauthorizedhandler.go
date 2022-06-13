package message

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/message"
	"github.com/klwxsrx/arch-course-project/pkg/order/app/service"
)

type paymentAuthorizedHandler struct {
	orderService *service.OrderService
}

func (h *paymentAuthorizedHandler) TopicName() string {
	return orderEventTopicName
}

func (h *paymentAuthorizedHandler) Type() string {
	return "payment_authorized"
}

func (h *paymentAuthorizedHandler) Handle(msg *message.Message) error {
	var orderID uuid.UUID
	err := json.Unmarshal(msg.Body, &orderID)
	if err != nil {
		return fmt.Errorf("failed to decode message")
	}

	err = h.orderService.HandlePaymentAuthorized(orderID)
	if err != nil {
		return fmt.Errorf("failed to handle payment authorized: %w", err)
	}
	return nil
}

func NewPaymentAuthorizedHandler(orderService *service.OrderService) message.Handler {
	return &paymentAuthorizedHandler{orderService: orderService}
}
