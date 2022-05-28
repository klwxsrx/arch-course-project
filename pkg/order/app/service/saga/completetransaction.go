package saga

import (
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/log"
	"github.com/klwxsrx/arch-course-project/pkg/common/app/saga"
	"github.com/klwxsrx/arch-course-project/pkg/order/app/service/api"
)

type completePaymentTransactionOperation struct {
	paymentAPI api.PaymentAPI
	orderID    uuid.UUID
	logger     log.Logger
}

func (op *completePaymentTransactionOperation) Name() string {
	return "completePaymentTransaction"
}

func (op *completePaymentTransactionOperation) Do() error {
	err := op.paymentAPI.CompleteTransaction(op.orderID)
	if err != nil {
		op.logger.With(log.Fields{
			"orderID": op.orderID,
		}).WithError(err).Error("failed to complete transaction")
		return err
	}
	return nil
}

func (op *completePaymentTransactionOperation) Undo() error {
	// do nothing
	return nil
}

func NewCompletePaymentTransactionOperation(
	paymentAPI api.PaymentAPI,
	orderID uuid.UUID,
	logger log.Logger,
) saga.Operation {
	return &completePaymentTransactionOperation{
		paymentAPI: paymentAPI,
		orderID:    orderID,
		logger:     logger,
	}
}
