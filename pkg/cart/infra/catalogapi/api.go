package catalogapi

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

func (c *apiClient) GetProducts(productIDs []uuid.UUID) ([]api.Product, error) {
	productsJSON, err := json.Marshal(productIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to encode products for request: %w", err)
	}

	url := fmt.Sprintf("%s/products", c.serviceURL)
	req, err := http.NewRequest(http.MethodGet, url, bytes.NewBuffer(productsJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute http request: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, api.ErrProductsNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to getProducts, httpCode: %v", resp.StatusCode)
	}

	var productPrices []struct {
		ID    uuid.UUID `json:"id"`
		Price int       `json:"price"`
	}
	err = json.NewDecoder(resp.Body).Decode(&productPrices)
	if err != nil {
		return nil, fmt.Errorf("failed to decode getProducts response: %w", err)
	}

	result := make([]api.Product, 0, len(productPrices))
	for _, item := range productPrices {
		result = append(result, api.Product{
			ID:    item.ID,
			Price: item.Price,
		})
	}
	return result, nil
}

func New(serviceURL string) api.CatalogAPI {
	return &apiClient{
		client:     &http.Client{},
		serviceURL: serviceURL,
	}
}
