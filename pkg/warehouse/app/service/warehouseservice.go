package service

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/idempotence"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/log"
	"github.com/klwxsrx/arch-course-project/pkg/warehouse/app/persistence"
	"github.com/klwxsrx/arch-course-project/pkg/warehouse/domain"
)

const decreaseItemBalanceLockKey = "decrease_warehouse_balance"

var (
	ErrItemAlreadyAdded = errors.New("item is already added")
	ErrInvalidQuantity  = errors.New("invalid quantity")
	ErrItemNotFound     = errors.New("item not found")
)

type WarehouseService struct {
	unitOfWork persistence.UnitOfWork
	logger     log.Logger
}

func (s *WarehouseService) GetAvailableItemsQuantity(itemIDs []uuid.UUID) ([]domain.ItemQuantity, error) {
	result := make([]domain.ItemQuantity, 0, len(itemIDs))
	err := s.unitOfWork.Execute("", func(p persistence.PersistentProvider) error {
		items, err := p.Stock().GetAvailableItemsQuantity(itemIDs)
		if err != nil {
			return err
		}

		if len(items) != len(itemIDs) {
			return ErrItemNotFound
		}

		result = items
		return nil
	})
	if err != nil && !errors.Is(err, ErrItemNotFound) {
		s.logger.WithError(err).Error("failed to get items quantity")
	}
	return result, err
}

func (s *WarehouseService) AddItems(idempotenceKey string, itemID uuid.UUID, quantity int) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}

	err := s.unitOfWork.Execute("", func(p persistence.PersistentProvider) error {
		err := p.IdempotenceKeyStore().StoreUnique(idempotenceKey)
		if errors.Is(err, idempotence.ErrKeyAlreadyExists) {
			return ErrItemAlreadyAdded
		}
		if err != nil {
			return err
		}

		op := &domain.StockOperation{
			ID:           p.Stock().NextID(),
			ItemID:       itemID,
			Type:         domain.StockOperationTypeArrival,
			ItemQuantity: quantity,
		}

		return p.Stock().Update(op)
	})
	if err != nil && !errors.Is(err, ErrItemAlreadyAdded) {
		s.logger.WithError(err).With(log.Fields{
			"idempotenceKey": idempotenceKey,
			"itemID":         itemID,
			"quantity":       quantity,
		}).Error("failed to add items to stock")
	}
	return err
}

func (s *WarehouseService) ReserveOrderItems(orderID uuid.UUID, itemsQuantity []domain.ItemQuantity) error {
	if len(itemsQuantity) == 0 {
		return nil
	}
	for _, item := range itemsQuantity {
		if item.Quantity <= 0 {
			return ErrInvalidQuantity
		}
	}

	err := s.unitOfWork.Execute(decreaseItemBalanceLockKey, func(p persistence.PersistentProvider) error {
		ops, err := p.Stock().GetOrderOperations(orderID)
		if err != nil {
			return err
		}
		if len(ops) > 0 {
			return nil
		}

		itemIDs := make([]uuid.UUID, 0, len(itemsQuantity))
		for _, item := range itemsQuantity {
			itemIDs = append(itemIDs, item.ItemID)
		}

		actualQuantity, err := p.Stock().GetAvailableItemsQuantity(itemIDs)
		if err != nil {
			return err
		}

		if len(itemsQuantity) != len(actualQuantity) || !s.checkAvailableItemsEnough(actualQuantity, itemsQuantity) {
			err := p.OrderAPI().NotifyItemsOutOfStock(orderID)
			if err != nil {
				return fmt.Errorf("failed to notify items out of stock: %w", err)
			}
			return nil
		}

		for _, item := range itemsQuantity {
			op := &domain.StockOperation{
				ID:           p.Stock().NextID(),
				ItemID:       item.ItemID,
				Type:         domain.StockOperationTypeReservation,
				ItemQuantity: -1 * item.Quantity,
				OrderID:      &orderID,
			}
			err := p.Stock().Update(op)
			if err != nil {
				return err
			}
		}

		err = p.OrderAPI().NotifyItemsReserved(orderID)
		if err != nil {
			return fmt.Errorf("failed to notify items reserved: %w", err)
		}
		return nil
	})
	if err != nil && !errors.Is(err, ErrItemNotFound) {
		s.logger.WithError(err).With(log.Fields{"orderID": orderID}).Error("failed to reserve order items")
	}
	return err
}

func (s *WarehouseService) DeleteOrderItemsReservation(orderID uuid.UUID) error {
	err := s.unitOfWork.Execute("", func(p persistence.PersistentProvider) error {
		ops, err := p.Stock().GetOrderOperations(orderID)
		if err != nil {
			return err
		}
		if ops == nil {
			return nil
		}

		ids := make([]uuid.UUID, 0, len(ops))
		for _, op := range ops {
			ids = append(ids, op.ID)
		}
		return p.Stock().Delete(ids)
	})

	if err != nil {
		s.logger.WithError(err).With(log.Fields{"orderID": orderID}).Error("failed to delete order items reservation")
	}
	return err
}

func (s *WarehouseService) checkAvailableItemsEnough(actualQuantity, expectedQuantity []domain.ItemQuantity) bool {
	findActualQuantity := func(itemID uuid.UUID) (int, bool) {
		for _, actualItem := range actualQuantity {
			if actualItem.ItemID == itemID {
				return actualItem.Quantity, true
			}
		}
		return 0, false
	}

	for _, expectedItem := range expectedQuantity {
		quantity, found := findActualQuantity(expectedItem.ItemID)
		if !found {
			return false
		}
		if expectedItem.Quantity > quantity {
			return false
		}
	}
	return true
}

func NewWarehouseService(unitOfWork persistence.UnitOfWork, logger log.Logger) *WarehouseService {
	return &WarehouseService{
		unitOfWork: unitOfWork,
		logger:     logger,
	}
}
