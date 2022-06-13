package warehouseapi

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/event"
	"github.com/klwxsrx/arch-course-project/pkg/order/app/service/async"
)

const warehouseEventTopicName = "warehouse_event"

type apiClient struct {
	eventDispatcher event.Dispatcher
}

func (a *apiClient) ReserveItems(orderID uuid.UUID, items []async.ItemQuantity) error {
	type itemsQuantitySchema struct {
		ItemID   uuid.UUID `json:"item_id"`
		Quantity int       `json:"quantity"`
	}

	itemsQuantity := make([]itemsQuantitySchema, 0, len(items))
	for _, item := range items {
		itemsQuantity = append(itemsQuantity, itemsQuantitySchema{
			ItemID:   item.ItemID,
			Quantity: item.Quantity,
		})
	}
	body := struct {
		OrderID uuid.UUID             `json:"order_id"`
		Items   []itemsQuantitySchema `json:"items"`
	}{
		OrderID: orderID,
		Items:   itemsQuantity,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return errors.New("failed to encode orderItems")
	}

	err = a.eventDispatcher.Dispatch(&event.Event{
		Type:      "reserve_items",
		TopicName: warehouseEventTopicName,
		Key:       orderID.String(),
		Body:      jsonBody,
	})
	if err != nil {
		return errors.New("failed to dispatch message")
	}
	return nil
}

func (a *apiClient) RemoveItemsReservation(orderID uuid.UUID) error {
	jsonID, err := json.Marshal(orderID)
	if err != nil {
		return errors.New("failed to encode orderID")
	}

	err = a.eventDispatcher.Dispatch(&event.Event{
		Type:      "remove_items_reservation",
		TopicName: warehouseEventTopicName,
		Key:       orderID.String(),
		Body:      jsonID,
	})
	if err != nil {
		return errors.New("failed to dispatch message")
	}
	return nil
}

func New(eventDispatcher event.Dispatcher) async.WarehouseAPI {
	return &apiClient{eventDispatcher: eventDispatcher}
}
