package message

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/message"
	"github.com/klwxsrx/arch-course-project/pkg/warehouse/app/service"
	"github.com/klwxsrx/arch-course-project/pkg/warehouse/domain"
)

type reserveItemsHandler struct {
	service *service.WarehouseService
}

func (h *reserveItemsHandler) TopicName() string {
	return warehouseEventTopicName
}

func (h *reserveItemsHandler) Type() string {
	return "reserve_items"
}

func (h *reserveItemsHandler) Handle(msg *message.Message) error {
	var body struct {
		OrderID uuid.UUID `json:"order_id"`
		Items   []struct {
			ItemID   uuid.UUID `json:"item_id"`
			Quantity int       `json:"quantity"`
		} `json:"items"`
	}

	err := json.Unmarshal(msg.Body, &body)
	if err != nil {
		return fmt.Errorf("failed to decode message")
	}

	items := make([]domain.ItemQuantity, 0, len(body.Items))
	for _, bodyItem := range body.Items {
		items = append(items, domain.ItemQuantity{
			ItemID:   bodyItem.ItemID,
			Quantity: bodyItem.Quantity,
		})
	}

	err = h.service.ReserveOrderItems(body.OrderID, items)
	if err != nil {
		return fmt.Errorf("failed to reserve order items: %w", err)
	}
	return nil
}

func NewReserveItemsHandler(service *service.WarehouseService) message.Handler {
	return &reserveItemsHandler{service: service}
}
