package service

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/cart/app/service/api"
	"github.com/klwxsrx/arch-course-project/pkg/cart/domain"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/log"
)

var (
	ErrInvalidQuantity   = errors.New("invalid product quantity")
	ErrInvalidProduct    = errors.New("invalid product id")
	ErrEmptyCartCheckout = errors.New("user has empty cart to checkout")
)

type CartService struct {
	catalogAPI api.CatalogAPI
	orderAPI   api.OrderAPI
	repo       domain.CartStorage
	logger     log.Logger
}

func (s *CartService) GetCart(userID uuid.UUID) (*domain.Cart, error) {
	return s.repo.GetByUserID(userID)
}

func (s *CartService) AddProduct(userID, productID uuid.UUID, expectedQuantity int) error {
	if expectedQuantity <= 0 {
		return ErrInvalidQuantity
	}

	err := s.validateProductID(productID)
	if err != nil {
		return err
	}

	cart, err := s.repo.GetByUserID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user cart: %w", err)
	}

	for i, product := range cart.Products {
		if product.ID != productID {
			continue
		}

		cart.Products[i].Quantity = expectedQuantity
		err := s.repo.Store(cart)
		if err != nil {
			return fmt.Errorf("failed to update product quantity: %w", err)
		}
		return nil
	}

	cart.Products = append(cart.Products, domain.ProductQuantity{
		ID:       productID,
		Quantity: expectedQuantity,
	})

	err = s.repo.Store(cart)
	if err != nil {
		s.logger.WithError(err).With(log.Fields{"userID": userID, "productID": productID}).Error("failed to add product")
	}
	return err
}

func (s *CartService) DeleteProduct(userID, productID uuid.UUID) error {
	cart, err := s.repo.GetByUserID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user cart: %w", err)
	}

	for i := 0; i < len(cart.Products); i++ {
		prod := &cart.Products[i]
		if prod.ID == productID {
			cart.Products[i] = cart.Products[len(cart.Products)-1]
			cart.Products = cart.Products[:len(cart.Products)-1]
			err := s.repo.Store(cart)
			if err != nil {
				return fmt.Errorf("failed to delete product from cart: %w", err)
			}
			return nil
		}
	}
	return nil
}

func (s *CartService) Checkout(userID, addressID uuid.UUID) (uuid.UUID, error) {
	var orderID uuid.UUID
	err := func() error {
		// TODO: validate addressID in delivery service

		cart, err := s.repo.GetByUserID(userID)
		if err != nil {
			return fmt.Errorf("failed to get user cart: %w", err)
		}
		if len(cart.Products) == 0 {
			return ErrEmptyCartCheckout
		}

		orderID, err = s.createOrder(cart, userID, addressID)
		if err != nil {
			return fmt.Errorf("failed to checkout: %w", err)
		}

		_ = s.repo.Delete(userID)
		return nil
	}()
	if errors.Is(err, ErrEmptyCartCheckout) {
		return orderID, err
	}
	if err != nil {
		s.logger.WithError(err).With(log.Fields{
			"userID": userID,
		}).Error("failed to checkout cart")
		return uuid.UUID{}, err
	}
	return orderID, nil
}

func (s *CartService) validateProductID(id uuid.UUID) error {
	_, err := s.catalogAPI.GetProducts([]uuid.UUID{id})
	if errors.Is(err, api.ErrProductsNotFound) {
		return ErrInvalidProduct
	}
	return err
}

func (s *CartService) createOrder(cart *domain.Cart, userID, addressID uuid.UUID) (uuid.UUID, error) {
	productIDs := make([]uuid.UUID, 0, len(cart.Products))
	for _, product := range cart.Products {
		productIDs = append(productIDs, product.ID)
	}

	products, err := s.catalogAPI.GetProducts(productIDs)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to get products for checkout: %w", err)
	}

	orderData, err := s.createOrderData(userID, addressID, cart, products)
	if err != nil {
		return uuid.UUID{}, err
	}

	orderID, err := s.orderAPI.CreateOrder(orderData)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to create order: %w", err)
	}
	return orderID, nil
}

func (s *CartService) createOrderData(userID, addressID uuid.UUID, cart *domain.Cart, products []api.Product) (*api.CreateOrderData, error) {
	findProductPrice := func(id uuid.UUID, products []api.Product) (int, error) {
		for _, apiProduct := range products {
			if id == apiProduct.ID {
				return apiProduct.Price, nil
			}
		}
		return 0, fmt.Errorf("failed to get product price for %v", id)
	}

	orderProducts := make([]api.CreateOrderProductData, 0, len(cart.Products))
	for _, cartProduct := range cart.Products {
		price, err := findProductPrice(cartProduct.ID, products)
		if err != nil {
			return nil, err
		}

		orderProducts = append(orderProducts, api.CreateOrderProductData{
			ID:           cartProduct.ID,
			ProductPrice: price,
			Quantity:     cartProduct.Quantity,
		})
	}

	return &api.CreateOrderData{
		IdempotenceKey: uuid.New().String(),
		UserID:         userID,
		AddressID:      addressID,
		Products:       orderProducts,
	}, nil
}

func NewCartService(
	catalogAPI api.CatalogAPI,
	orderAPI api.OrderAPI,
	repo domain.CartStorage,
	logger log.Logger,
) *CartService {
	return &CartService{
		catalogAPI: catalogAPI,
		orderAPI:   orderAPI,
		repo:       repo,
		logger:     logger,
	}
}
