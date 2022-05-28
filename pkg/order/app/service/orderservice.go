package service

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/idempotence"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/log"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/saga"
	"github.com/klwxsrx/arch-course-project/pkg/order/app/persistence"
	"github.com/klwxsrx/arch-course-project/pkg/order/app/service/api"
	orderSaga "github.com/klwxsrx/arch-course-project/pkg/order/app/service/saga"
	"github.com/klwxsrx/arch-course-project/pkg/order/domain"
)

var (
	ErrOrderAlreadyCreated = errors.New("order with key is already created")
	ErrEmptyOrder          = errors.New("empty or completely free order")
	ErrOrderRejected       = errors.New("order rejected")
)

type OrderService struct {
	paymentAPI   api.PaymentAPI
	warehouseAPI api.WarehouseAPI
	deliveryAPI  api.DeliveryAPI

	ufw    persistence.UnitOfWork
	logger log.Logger
}

func (s *OrderService) Create(
	idempotenceKey string,
	userID uuid.UUID,
	addressID uuid.UUID,
	items []domain.OrderItem,
) (uuid.UUID, error) {
	order, err := s.createNewOrder(idempotenceKey, userID, addressID, items)
	if err != nil {
		return uuid.UUID{}, err
	}

	processOrderSaga := saga.New(fmt.Sprintf("ProcessOrder_%v", order.ID), []saga.Operation{
		orderSaga.NewAuthorizeOrderPaymentOperation(s.paymentAPI, order, s.logger),
		orderSaga.NewReserveOrderItemsOperation(s.warehouseAPI, order, s.logger),
		orderSaga.NewScheduleDeliveryOperation(s.deliveryAPI, order, s.logger),
		orderSaga.NewCompletePaymentTransactionOperation(s.paymentAPI, order.ID, s.logger),
	}, s.logger)

	err = processOrderSaga.Execute()
	s.logger.With(log.Fields{
		"order":  order.ID,
		"result": err,
	}).Info("order creation completed")

	// order set for delivery
	if err == nil {
		// TODO: sent event to process delivery
		return order.ID, s.updateOrderStatus(order.ID, domain.OrderStatusAwaitingDelivery)
	}

	// order cancelled
	err = s.updateOrderStatus(order.ID, domain.OrderStatusCancelled)
	if err != nil {
		return uuid.UUID{}, err
	}
	return order.ID, ErrOrderRejected
}

func (s *OrderService) createNewOrder(
	idempotenceKey string,
	userID uuid.UUID,
	addressID uuid.UUID,
	items []domain.OrderItem,
) (*domain.Order, error) {
	var result *domain.Order
	err := s.ufw.Execute(func(p persistence.PersistentProvider) error {
		err := p.IdempotenceKeyStore().StoreUnique(idempotenceKey)
		if errors.Is(err, idempotence.ErrKeyAlreadyExists) {
			return ErrOrderAlreadyCreated
		}
		if err != nil {
			return err
		}

		totalAmount := s.calculateTotalAmount(items)
		if totalAmount == 0 {
			return ErrEmptyOrder
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
			return err
		}
		result = order
		return nil
	})
	if err == nil || errors.Is(err, ErrOrderAlreadyCreated) || errors.Is(err, ErrEmptyOrder) {
		return result, err
	}

	s.logger.WithError(err).Error("failed to create order")
	return result, err
}

func (s *OrderService) updateOrderStatus(id uuid.UUID, status domain.OrderStatus) error {
	err := s.ufw.Execute(func(p persistence.PersistentProvider) error {
		order, err := p.OrderRepository().GetByID(id)
		if err != nil {
			return err
		}

		if order.Status == status {
			return nil
		}

		order.Status = status
		return p.OrderRepository().Store(order)
	})
	if err != nil {
		s.logger.WithError(err).Error(fmt.Errorf("failed to update order %v status to %v", id, status))
		return err
	}
	return nil
}

func (s *OrderService) calculateTotalAmount(items []domain.OrderItem) int {
	var result int
	for _, item := range items {
		result += item.ItemPrice * item.Quantity
	}
	return result
}

func NewOrderService(
	paymentAPI api.PaymentAPI,
	warehouseAPI api.WarehouseAPI,
	deliveryAPI api.DeliveryAPI,
	ufw persistence.UnitOfWork,
	logger log.Logger,
) *OrderService {
	return &OrderService{
		paymentAPI:   paymentAPI,
		warehouseAPI: warehouseAPI,
		deliveryAPI:  deliveryAPI,
		ufw:          ufw,
		logger:       logger,
	}
}
