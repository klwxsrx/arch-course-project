package warehouseapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/order/app/service/api"
	"net/http"
)

type apiClient struct {
	client     *http.Client
	serviceURL string
}

func (a *apiClient) ReserveOrderItems(orderID uuid.UUID, items []api.ItemQuantity) error {
	type itemQuantitySchema struct {
		ItemID   uuid.UUID `json:"item_id"`
		Quantity int       `json:"quantity"`
	}

	itemsQuantity := make([]itemQuantitySchema, 0, len(items))
	for _, item := range items {
		itemsQuantity = append(itemsQuantity, itemQuantitySchema{
			ItemID:   item.ItemID,
			Quantity: item.Quantity,
		})
	}

	itemsJSON, err := json.Marshal(itemsQuantity)
	if err != nil {
		return errors.New("failed to encode json request")
	}

	url := fmt.Sprintf("%s/warehouse/order/%v/reserve", a.serviceURL, orderID)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(itemsJSON))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute http request: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusConflict:
		return api.ErrOrderOperationsAlreadyExist
	case http.StatusMethodNotAllowed:
		return api.ErrOrderItemsOutOfStock
	case http.StatusNoContent:
		return nil
	default:
		return fmt.Errorf("failed to reserveOrderItems, httpCode: %v", resp.StatusCode)
	}
}

func (a *apiClient) RemoveOrderItemsReservation(orderID uuid.UUID) error {
	url := fmt.Sprintf("%s/warehouse/order/%v/reserve", a.serviceURL, orderID)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute http request: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusNoContent:
		return nil
	default:
		return fmt.Errorf("failed to removeOrderItemsReservation, httpCode: %v", resp.StatusCode)
	}
}

func New(serviceURL string) api.WarehouseAPI {
	return &apiClient{
		client:     &http.Client{},
		serviceURL: serviceURL,
	}
}
