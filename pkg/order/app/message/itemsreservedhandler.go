package message

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/message"
	"github.com/klwxsrx/arch-course-project/pkg/order/app/service"
)

type itemsReservedHandler struct {
	orderService *service.OrderService
}

func (h *itemsReservedHandler) TopicName() string {
	return orderEventTopicName
}

func (h *itemsReservedHandler) Type() string {
	return "items_reserved"
}

func (h *itemsReservedHandler) Handle(msg *message.Message) error {
	var orderID uuid.UUID
	err := json.Unmarshal(msg.Body, &orderID)
	if err != nil {
		return fmt.Errorf("failed to decode message")
	}

	err = h.orderService.HandleItemsReserved(orderID)
	if err != nil {
		return fmt.Errorf("failed to handle items reserved: %w", err)
	}
	return nil
}

func NewItemsReservedHandler(orderService *service.OrderService) message.Handler {
	return &itemsReservedHandler{orderService: orderService}
}
