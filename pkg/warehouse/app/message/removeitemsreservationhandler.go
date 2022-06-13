package message

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/message"
	"github.com/klwxsrx/arch-course-project/pkg/warehouse/app/service"
)

type removeItemsReservationHandler struct {
	service *service.WarehouseService
}

func (h *removeItemsReservationHandler) TopicName() string {
	return warehouseEventTopicName
}

func (h *removeItemsReservationHandler) Type() string {
	return "remove_items_reservation"
}

func (h *removeItemsReservationHandler) Handle(msg *message.Message) error {
	var orderID uuid.UUID
	err := json.Unmarshal(msg.Body, &orderID)
	if err != nil {
		return fmt.Errorf("failed to decode message")
	}

	err = h.service.DeleteOrderItemsReservation(orderID)
	if err != nil {
		return fmt.Errorf("failed to delete order items reservation: %w", err)
	}
	return nil
}

func NewRemoveItemsReservationHandler(service *service.WarehouseService) message.Handler {
	return &removeItemsReservationHandler{service: service}
}
