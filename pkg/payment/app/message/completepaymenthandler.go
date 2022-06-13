package message

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/message"
	"github.com/klwxsrx/arch-course-project/pkg/payment/app/service"
)

type completePaymentHandler struct {
	paymentService *service.PaymentService
}

func (h *completePaymentHandler) TopicName() string {
	return paymentEventTopicName
}

func (h *completePaymentHandler) Type() string {
	return "complete_payment"
}

func (h *completePaymentHandler) Handle(msg *message.Message) error {
	var orderID uuid.UUID
	err := json.Unmarshal(msg.Body, &orderID)
	if err != nil {
		return fmt.Errorf("failed to decode message")
	}

	err = h.paymentService.CompletePayment(orderID)
	if err != nil {
		return fmt.Errorf("failed to complete payment: %w", err)
	}
	return nil
}

func NewCompletePaymentHandler(paymentService *service.PaymentService) message.Handler {
	return &completePaymentHandler{paymentService: paymentService}
}
