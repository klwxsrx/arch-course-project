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
			return errors.New("failed to get order not found")
		}
		if err != nil {
			return fmt.Errorf("failed to get order: %w", err)
		}
		if order.Status != domain.OrderStatusCreated {
			return nil
		}

		err = updateOrderStatus(order, domain.OrderStatusPaymentAuthorized, p.OrderRepository())
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
		err = p.WarehouseAPI().ReserveItems(order.ID, itemQuantity)
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

func (s *OrderService) HandleItemsReserved(orderID uuid.UUID) error {
	err := s.ufw.Execute(func(p persistence.PersistentProvider) error {
		order, err := p.OrderRepository().GetByID(orderID)
		if errors.Is(err, domain.ErrOrderNotFound) {
			return errors.New("failed to get order not found")
		}
		if err != nil {
			return fmt.Errorf("failed to get order: %w", err)
		}

		if order.Status != domain.OrderStatusPaymentAuthorized {
			return nil
		}

		err = updateOrderStatus(order, domain.OrderStatusItemsReserved, p.OrderRepository())
		if err != nil {
			return fmt.Errorf("failed to update order status: %w", err)
		}

		err = p.DeliveryAPI().ScheduleDelivery(order.ID, order.AddressID)
		if err != nil {
			return fmt.Errorf("failed to schedule delivery: %w", err)
		}

		return nil
	})
	if err != nil {
		s.logger.WithError(err).With(log.Fields{"orderID": orderID}).Error("failed to handle items reserved")
	}
	return err
}

func (s *OrderService) HandleItemsOutOfStock(orderID uuid.UUID) error {
	err := s.ufw.Execute(func(p persistence.PersistentProvider) error {
		order, err := p.OrderRepository().GetByID(orderID)
		if errors.Is(err, domain.ErrOrderNotFound) {
			return errors.New("failed to get order not found")
		}
		if err != nil {
			return fmt.Errorf("failed to get order: %w", err)
		}

		if order.Status != domain.OrderStatusPaymentAuthorized {
			return nil
		}

		err = updateOrderStatus(order, domain.OrderStatusCancelled, p.OrderRepository())
		if err != nil {
			return fmt.Errorf("failed to update order status: %w", err)
		}

		err = p.PaymentAPI().CancelPayment(orderID)
		if err != nil {
			return fmt.Errorf("failed to cancel payment: %w", err)
		}
		return nil
	})
	if err != nil {
		s.logger.WithError(err).With(log.Fields{"orderID": orderID}).Error("failed to handle items out of stock")
	}
	return err
}

func (s *OrderService) HandleDeliveryScheduled(orderID uuid.UUID) error {
	err := s.ufw.Execute(func(p persistence.PersistentProvider) error {
		order, err := p.OrderRepository().GetByID(orderID)
		if errors.Is(err, domain.ErrOrderNotFound) {
			return errors.New("failed to get order not found")
		}
		if err != nil {
			return fmt.Errorf("failed to get order: %w", err)
		}

		if order.Status != domain.OrderStatusItemsReserved {
			return nil
		}

		err = updateOrderStatus(order, domain.OrderStatusDeliveryScheduled, p.OrderRepository())
		if err != nil {
			return fmt.Errorf("failed to update order status: %w", err)
		}

		err = p.PaymentAPI().CompleteTransaction(orderID)
		if err != nil {
			return fmt.Errorf("failed to complete transaction: %w", err)
		}
		return nil
	})
	if err != nil {
		s.logger.WithError(err).With(log.Fields{"orderID": orderID}).Error("failed to handle delivery scheduled")
	}
	return err
}

func (s *OrderService) HandlePaymentCompleted(orderID uuid.UUID) error {
	err := s.ufw.Execute(func(p persistence.PersistentProvider) error {
		order, err := p.OrderRepository().GetByID(orderID)
		if errors.Is(err, domain.ErrOrderNotFound) {
			return errors.New("failed to get order not found")
		}
		if err != nil {
			return fmt.Errorf("failed to get order: %w", err)
		}

		if order.Status != domain.OrderStatusDeliveryScheduled {
			return nil
		}

		err = updateOrderStatus(order, domain.OrderStatusSentToDelivery, p.OrderRepository())
		if err != nil {
			return fmt.Errorf("failed to update order status: %w", err)
		}

		err = p.DeliveryAPI().ProcessDelivery(order.ID)
		if err != nil {
			return fmt.Errorf("failed to process delivery: %w", err)
		}

		return nil
	})
	if err != nil {
		s.logger.WithError(err).With(log.Fields{"orderID": orderID}).Error("failed to handle payment completed")
	}
	return err
}

func (s *OrderService) HandlePaymentCompletionRejected(orderID uuid.UUID) error {
	err := s.ufw.Execute(func(p persistence.PersistentProvider) error {
		order, err := p.OrderRepository().GetByID(orderID)
		if errors.Is(err, domain.ErrOrderNotFound) {
			return errors.New("failed to get order not found")
		}
		if err != nil {
			return fmt.Errorf("failed to get order: %w", err)
		}

		if order.Status != domain.OrderStatusDeliveryScheduled {
			return nil
		}

		err = updateOrderStatus(order, domain.OrderStatusCancelled, p.OrderRepository())
		if err != nil {
			return fmt.Errorf("failed to update order status: %w", err)
		}

		err = p.DeliveryAPI().CancelDeliverySchedule(orderID)
		if err != nil {
			return fmt.Errorf("failed to cancel delivery schedule: %w", err)
		}
		err = p.WarehouseAPI().RemoveItemsReservation(orderID)
		if err != nil {
			return fmt.Errorf("failed to remove items reservation: %w", err)
		}

		return nil
	})
	if err != nil {
		s.logger.WithError(err).With(log.Fields{"orderID": orderID}).Error("failed to handle payment completion rejected")
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
