package saga

import (
	"github.com/klwxsrx/arch-course-project/pkg/common/app/log"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/saga"
	"github.com/klwxsrx/arch-course-project/pkg/order/app/service/api"
	"github.com/klwxsrx/arch-course-project/pkg/order/domain"
)

type scheduleDeliveryOperation struct {
	deliveryAPI api.DeliveryAPI
	order       *domain.Order
	logger      log.Logger
}

func (op *scheduleDeliveryOperation) Name() string {
	return "scheduleDelivery"
}

func (op *scheduleDeliveryOperation) Do() error {
	err := op.deliveryAPI.ScheduleDelivery(op.order.ID, op.order.AddressID)
	if err != nil {
		op.logger.With(log.Fields{
			"orderID": op.order.ID,
		}).WithError(err).Error("failed to schedule order delivery")
		return err
	}
	return nil
}

func (op *scheduleDeliveryOperation) Undo() error {
	err := op.deliveryAPI.DeleteDeliverySchedule(op.order.ID)
	if err != nil {
		op.logger.With(log.Fields{
			"orderID": op.order.ID,
		}).WithError(err).Error("failed to delete order delivery schedule")
		return err
	}
	return nil
}

func NewScheduleDeliveryOperation(
	deliveryAPI api.DeliveryAPI,
	order *domain.Order,
	logger log.Logger,
) saga.Operation {
	return &scheduleDeliveryOperation{
		deliveryAPI: deliveryAPI,
		order:       order,
		logger:      logger,
	}
}
