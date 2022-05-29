package redis

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/cart/domain"
	"github.com/klwxsrx/arch-course-project/pkg/common/infra/redis"
)

type cartRepo struct {
	client redis.Client
}

type productQuantity struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
}

func (r *cartRepo) GetByUserID(userID uuid.UUID) (*domain.Cart, error) {
	cartStr, err := r.client.Get(r.getCartKey(userID))
	if errors.Is(err, redis.ErrKeyDoesNotExist) {
		return &domain.Cart{
			UserID:   userID,
			Products: nil,
		}, nil
	}
	if err != nil {
		return nil, err
	}

	var quantity []productQuantity
	err = json.Unmarshal([]byte(cartStr), &quantity)
	if err != nil {
		return nil, err
	}

	domainQuantity := make([]domain.ProductQuantity, 0, len(quantity))
	for _, item := range quantity {
		domainQuantity = append(domainQuantity, domain.ProductQuantity{
			ID:       item.ProductID,
			Quantity: item.Quantity,
		})
	}

	return &domain.Cart{
		UserID:   userID,
		Products: domainQuantity,
	}, nil
}

func (r *cartRepo) Store(cart *domain.Cart) error {
	quantity := make([]productQuantity, 0, len(cart.Products))
	for _, item := range cart.Products {
		quantity = append(quantity, productQuantity{
			ProductID: item.ID,
			Quantity:  item.Quantity,
		})
	}

	result, err := json.Marshal(quantity)
	if err != nil {
		return err
	}

	return r.client.Set(r.getCartKey(cart.UserID), string(result), nil)
}

func (r *cartRepo) Delete(userID uuid.UUID) error {
	return r.client.Del(r.getCartKey(userID))
}

func (r *cartRepo) getCartKey(userID uuid.UUID) string {
	return fmt.Sprintf("shopping_cart:%v:products", userID)
}

func NewCartStorage(client redis.Client) domain.CartStorage {
	return &cartRepo{client: client}
}
