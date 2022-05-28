package paymentapi

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

func (a *apiClient) AuthorizeOrder(orderID uuid.UUID, totalAmount int) error {
	createPayment := struct {
		OrderID     uuid.UUID `json:"order_id"`
		TotalAmount int       `json:"total_amount"`
	}{
		orderID, totalAmount,
	}

	createPaymentJSON, err := json.Marshal(createPayment)
	if err != nil {
		return errors.New("failed to encode json request")
	}

	url := fmt.Sprintf("%s/payments", a.serviceURL)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(createPaymentJSON))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute http request: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusCreated:
		return nil
	default:
		return fmt.Errorf("failed to createPayment, httpCode: %v", resp.StatusCode)
	}
}

func (a *apiClient) CompleteTransaction(orderID uuid.UUID) error {
	url := fmt.Sprintf("%s/payment/%v/complete", a.serviceURL, orderID)
	req, err := http.NewRequest(http.MethodPost, url, nil)
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
	case http.StatusNotFound:
		return api.ErrOrderPaymentNotFound
	case http.StatusMethodNotAllowed:
		return api.ErrOrderPaymentNotAuthorized
	case http.StatusNotAcceptable:
		return api.ErrOrderPaymentRejected
	default:
		return fmt.Errorf("failed to completeTransaction, httpCode: %v", resp.StatusCode)
	}
}

func (a *apiClient) CancelOrder(orderID uuid.UUID) error {
	url := fmt.Sprintf("%s/payment/%v/cancel", a.serviceURL, orderID)
	req, err := http.NewRequest(http.MethodPost, url, nil)
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
	case http.StatusNotFound:
		return api.ErrOrderPaymentNotFound
	case http.StatusMethodNotAllowed:
		return api.ErrOrderPaymentNotAuthorized
	default:
		return fmt.Errorf("failed to cancelOrder, httpCode: %v", resp.StatusCode)
	}
}

func New(serviceURL string) api.PaymentAPI {
	return &apiClient{
		client:     &http.Client{},
		serviceURL: serviceURL,
	}
}
