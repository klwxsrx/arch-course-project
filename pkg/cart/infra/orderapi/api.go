package orderapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/cart/app/service/api"
	"net/http"
)

type apiClient struct {
	client     *http.Client
	serviceURL string
}

type createOrderItemSchema struct {
	ID        uuid.UUID `json:"id"`
	ItemPrice int       `json:"item_price"`
	Quantity  int       `json:"quantity"`
}

type createOrderDataSchema struct {
	UserID    uuid.UUID               `json:"user_id"`
	AddressID uuid.UUID               `json:"address_id"`
	Items     []createOrderItemSchema `json:"items"`
}

func (c *apiClient) CreateOrder(data *api.CreateOrderData) (err error) {
	itemData := make([]createOrderItemSchema, 0, len(data.Products))
	for _, item := range data.Products {
		itemData = append(itemData, createOrderItemSchema{
			ID:        item.ID,
			ItemPrice: item.ProductPrice,
			Quantity:  item.Quantity,
		})
	}
	orderData := createOrderDataSchema{
		UserID:    data.UserID,
		AddressID: data.AddressID,
		Items:     itemData,
	}

	orderJSON, err := json.Marshal(orderData)
	if err != nil {
		return fmt.Errorf("failed to encode order data for request: %w", err)
	}

	url := fmt.Sprintf("%s/orders", c.serviceURL)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(orderJSON))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute http request: %w", err)
	}
	if resp.StatusCode == http.StatusConflict {
		return nil
	}
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to createOrder, httpCode: %v", resp.StatusCode)
	}

	var result struct {
		Success bool `json:"success"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return fmt.Errorf("failed to decode createOrder response: %w", err)
	}

	if !result.Success {
		return api.ErrOrderRejected
	}
	return nil
}

func New(serviceURL string) api.OrderAPI {
	return &apiClient{
		client:     &http.Client{},
		serviceURL: serviceURL,
	}
}
