package deliveryapi

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

func (a *apiClient) ScheduleDelivery(orderID uuid.UUID, addressID uuid.UUID) error {
	type addressSchema struct {
		Address uuid.UUID `json:"address"`
	}

	addressJSON, err := json.Marshal(addressSchema{
		Address: addressID,
	})
	if err != nil {
		return errors.New("failed to encode json request")
	}

	url := fmt.Sprintf("%s/delivery/%v/schedule", a.serviceURL, orderID)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(addressJSON))
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
		return fmt.Errorf("failed to scheduleDelivery, httpCode: %v", resp.StatusCode)
	}
}

func (a *apiClient) DeleteDeliverySchedule(orderID uuid.UUID) error {
	url := fmt.Sprintf("%s/delivery/%v/schedule", a.serviceURL, orderID)
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
		return fmt.Errorf("failed to scheduleDelivery, httpCode: %v", resp.StatusCode)
	}
}

func New(serviceURL string) api.DeliveryAPI {
	return &apiClient{
		client:     &http.Client{},
		serviceURL: serviceURL,
	}
}
