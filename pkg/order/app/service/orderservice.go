package service

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/idempotence"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/log"
	"github.com/klwxsrx/arch-course-project/pkg/order/app/persistence"
	"github.com/klwxsrx/arch-course-project/pkg/order/app/service/async"
	"github.com/klwxsrx/arch-course-project/pkg/order/domain"
)

var (
	ErrOrderAlreadyCreated = errors.New("order with key is already created")
	ErrEmptyOrder          = errors.New("empty or completely free order")
)

type OrderService struct {
	ufw    persistence.UnitOfWork
	logger log.Logger
}

func (s *OrderService) Create(
	idempotenceKey string,
	userID uuid.UUID,
	addressID uuid.UUID,
	items []domain.OrderItem,
) (uuid.UUID, error) {
	var orderID uuid.UUID
	err := s.ufw.Execute(func(p persistence.PersistentProvider) error {
		order, err := createOrder(idempotenceKey, userID, addressID, items, p)
		if errors.Is(err, ErrOrderAlreadyCreated) || errors.Is(err, ErrEmptyOrder) {
			return err
		}
		if err != nil {
			return fmt.Errorf("failed to create order: %w", err)
		}

		err = p.PaymentAPI().AuthorizeOrder(order.ID, order.TotalAmount)
		if err != nil {
			return fmt.Errorf("failed to authorize order: %w", err)
		}

		orderID = order.ID
		return nil
	})

	if err == nil || errors.Is(err, ErrOrderAlreadyCreated) || errors.Is(err, ErrEmptyOrder) {
		s.logger.With(log.Fields{
			"userID": userID,
			"order":  orderID,
			"result": err,
		}).Info("Create completed")
		return orderID, nil
	}

	s.logger.WithError(err).With(log.Fields{
		"userID": userID,
	}).Error("Create failed")
	return uuid.Nil, err
}

func (s *OrderService) HandlePaymentAuthorized(orderID uuid.UUID) error {
	err := s.ufw.Execute(func(p persistence.PersistentProvider) error {
		order, err := p.OrderRepository().GetByID(orderID)
		if errors.Is(err, domain.ErrOrderNotFound) {
			return errors.New("failed to get authorized order not found")
		}
		if err != nil {
			return fmt.Errorf("failed to get authorized order: %w", err)
		}
		if order.Status != domain.OrderStatusCreated {
			return nil
		}

		err = updateOrderStatus(order, domain.OrderPaymentAuthorized, p.OrderRepository())
		if err != nil {
			return fmt.Errorf("failed to update order status: %w", err)
		}

		itemQuantity := make([]async.ItemQuantity, 0, len(order.Items))
		for _, orderItem := range order.Items {
			itemQuantity = append(itemQuantity, async.ItemQuantity{
				ItemID:   orderItem.ID,
				Quantity: orderItem.Quantity,
			})
		}
		err = p.WarehouseAPI().ReserveOrderItems(order.ID, itemQuantity)
		if err != nil {
			return fmt.Errorf("failed to reserve order items: %w", err)
		}

		return nil
	})
	if err != nil {
		s.logger.WithError(err).With(log.Fields{"orderID": orderID}).Error("failed to handle payment authorized")
	}
	return err
}

func createOrder(
	idempotenceKey string,
	userID uuid.UUID,
	addressID uuid.UUID,
	items []domain.OrderItem,
	p persistence.PersistentProvider,
) (*domain.Order, error) {
	err := p.IdempotenceKeyStore().StoreUnique(idempotenceKey)
	if errors.Is(err, idempotence.ErrKeyAlreadyExists) {
		return nil, ErrOrderAlreadyCreated
	}
	if err != nil {
		return nil, err
	}

	totalAmount := calculateTotalAmount(items)
	if totalAmount == 0 {
		return nil, ErrEmptyOrder
	}

	order := &domain.Order{
		ID:          p.OrderRepository().NextID(),
		UserID:      userID,
		AddressID:   addressID,
		Items:       items,
		Status:      domain.OrderStatusCreated,
		TotalAmount: totalAmount,
	}

	err = p.OrderRepository().Store(order)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func calculateTotalAmount(items []domain.OrderItem) int {
	var result int
	for _, item := range items {
		result += item.ItemPrice * item.Quantity
	}
	return result
}

func updateOrderStatus(
	order *domain.Order,
	new domain.OrderStatus,
	repo domain.OrderRepository,
) error {
	if order.Status == new {
		return nil
	}

	order.Status = new
	return repo.Store(order)
}

func NewOrderService(
	ufw persistence.UnitOfWork,
	logger log.Logger,
) *OrderService {
	return &OrderService{
		ufw:    ufw,
		logger: logger,
	}
}
