package saga

import (
	"errors"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/log"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/saga"
	"github.com/klwxsrx/arch-course-project/pkg/order/app/service/api"
	"github.com/klwxsrx/arch-course-project/pkg/order/domain"
)

type authorizeOrderPaymentOperation struct {
	paymentAPI api.PaymentAPI
	order      *domain.Order
	logger     log.Logger
}

func (op *authorizeOrderPaymentOperation) Name() string {
	return "authorizeOrderPayment"
}

func (op *authorizeOrderPaymentOperation) Do() error {
	err := op.paymentAPI.AuthorizeOrder(op.order.ID, op.order.TotalAmount)
	if err != nil {
		op.logger.With(log.Fields{
			"orderID": op.order.ID,
		}).WithError(err).Error("failed to authorize order")
		return err
	}
	return nil
}

func (op *authorizeOrderPaymentOperation) Undo() error {
	err := op.paymentAPI.CancelOrder(op.order.ID)
	if errors.Is(err, api.ErrOrderPaymentNotAuthorized) {
		return nil
	}
	if err != nil {
		op.logger.With(log.Fields{
			"orderID": op.order.ID,
		}).WithError(err).Error("failed to cancel order")
		return err
	}
	return nil
}

func NewAuthorizeOrderPaymentOperation(
	paymentAPI api.PaymentAPI,
	order *domain.Order,
	logger log.Logger,
) saga.Operation {
	return &authorizeOrderPaymentOperation{
		paymentAPI: paymentAPI,
		order:      order,
		logger:     logger,
	}
}
