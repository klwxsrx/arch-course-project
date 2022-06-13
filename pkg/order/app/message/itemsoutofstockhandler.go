package message

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/message"
	"github.com/klwxsrx/arch-course-project/pkg/order/app/service"
)

type itemsOutOfStockHandler struct {
	orderService *service.OrderService
}

func (h *itemsOutOfStockHandler) TopicName() string {
	return orderEventTopicName
}

func (h *itemsOutOfStockHandler) Type() string {
	return "items_out_of_stock"
}

func (h *itemsOutOfStockHandler) Handle(msg *message.Message) error {
	var orderID uuid.UUID
	err := json.Unmarshal(msg.Body, &orderID)
	if err != nil {
		return fmt.Errorf("failed to decode message")
	}

	err = h.orderService.HandleItemsOutOfStock(orderID)
	if err != nil {
		return fmt.Errorf("failed to handle items out of stock: %w", err)
	}
	return nil
}

func NewItemsOutOfStockHandler(orderService *service.OrderService) message.Handler {
	return &itemsOutOfStockHandler{orderService: orderService}
}
