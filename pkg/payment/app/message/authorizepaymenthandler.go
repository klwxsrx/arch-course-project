package message

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/message"
	"github.com/klwxsrx/arch-course-project/pkg/payment/app/service"
)

type authorizePaymentHandler struct {
	paymentService *service.PaymentService
}

func (h *authorizePaymentHandler) TopicName() string {
	return paymentEventTopicName
}

func (h *authorizePaymentHandler) Type() string {
	return "authorize_payment"
}

func (h *authorizePaymentHandler) Handle(msg *message.Message) error {
	var authorizePayment struct {
		OrderID     uuid.UUID `json:"order_id"`
		TotalAmount int       `json:"total_amount"`
	}

	err := json.Unmarshal(msg.Body, &authorizePayment)
	if err != nil {
		return fmt.Errorf("failed to decode message")
	}

	err = h.paymentService.AuthorizePayment(authorizePayment.OrderID, authorizePayment.TotalAmount)
	if err != nil {
		return fmt.Errorf("failed to authorize payment: %w", err)
	}
	return nil
}

func NewAuthorizePaymentHandler(paymentService *service.PaymentService) message.Handler {
	return &authorizePaymentHandler{paymentService: paymentService}
}
