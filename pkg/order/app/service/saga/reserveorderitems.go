package saga

import (
	"errors"
	"fmt"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/log"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/saga"
	"github.com/klwxsrx/arch-course-project/pkg/order/app/service/api"
	"github.com/klwxsrx/arch-course-project/pkg/order/domain"
)

type reserveOrderItemsOperation struct {
	warehouseAPI api.WarehouseAPI
	order        *domain.Order
	logger       log.Logger
}

func (op *reserveOrderItemsOperation) Name() string {
	return "reserveOrderItems"
}

func (op *reserveOrderItemsOperation) Do() error {
	items := make([]api.ItemQuantity, 0, len(op.order.Items))
	for _, orderItem := range op.order.Items {
		items = append(items, api.ItemQuantity{
			ItemID:   orderItem.ID,
			Quantity: orderItem.Quantity,
		})
	}

	err := op.warehouseAPI.ReserveOrderItems(op.order.ID, items)
	if errors.Is(err, api.ErrOrderItemsOutOfStock) {
		return fmt.Errorf("failed to reserve items: %w", err)
	}
	if err != nil {
		op.logger.With(log.Fields{
			"orderID": op.order.ID,
		}).WithError(err).Error("failed to reserve order items")
		return err
	}
	return nil
}

func (op *reserveOrderItemsOperation) Undo() error {
	err := op.warehouseAPI.RemoveOrderItemsReservation(op.order.ID)
	if err != nil {
		op.logger.With(log.Fields{
			"orderID": op.order.ID,
		}).WithError(err).Error("failed to remove order item reservation")
		return err
	}
	return nil
}

func NewReserveOrderItemsOperation(
	warehouseAPI api.WarehouseAPI,
	order *domain.Order,
	logger log.Logger,
) saga.Operation {
	return &reserveOrderItemsOperation{
		warehouseAPI: warehouseAPI,
		order:        order,
		logger:       logger,
	}
}
