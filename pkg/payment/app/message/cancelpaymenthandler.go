package message

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/message"
	"github.com/klwxsrx/arch-course-project/pkg/payment/app/service"
)

type cancelPaymentHandler struct {
	paymentService *service.PaymentService
}

func (h *cancelPaymentHandler) TopicName() string {
	return paymentEventTopicName
}

func (h *cancelPaymentHandler) Type() string {
	return "cancel_payment"
}

func (h *cancelPaymentHandler) Handle(msg *message.Message) error {
	var orderID uuid.UUID
	err := json.Unmarshal(msg.Body, &orderID)
	if err != nil {
		return fmt.Errorf("failed to decode message")
	}

	err = h.paymentService.CancelPayment(orderID)
	if err != nil {
		return fmt.Errorf("failed to cancel payment: %w", err)
	}
	return nil
}

func NewCancelPaymentHandler(paymentService *service.PaymentService) message.Handler {
	return &cancelPaymentHandler{paymentService: paymentService}
}
