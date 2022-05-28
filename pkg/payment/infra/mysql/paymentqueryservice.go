package mysql

import (
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/common/infra/mysql"
	"github.com/klwxsrx/arch-course-project/pkg/payment/app/query"
	"github.com/klwxsrx/arch-course-project/pkg/payment/domain"
)

type paymentQueryService struct {
	client mysql.Client
}

func (s *paymentQueryService) GetPayment(orderID uuid.UUID) (*query.PaymentData, error) {
	const paymentQuery = `
		SELECT order_id, status, total_amount
		FROM payment
		WHERE order_id = ?
	`

	binaryOrderID, err := orderID.MarshalBinary()
	if err != nil {
		return nil, err
	}

	var paymentSqlx sqlxPayment
	err = s.client.Get(&paymentSqlx, paymentQuery, binaryOrderID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, query.ErrPaymentNotFound
	}
	if err != nil {
		return nil, err
	}

	return &query.PaymentData{
		OrderID:     paymentSqlx.OrderID,
		Status:      domain.PaymentStatus(paymentSqlx.Status),
		TotalAmount: paymentSqlx.TotalAmount,
	}, nil
}

func NewPaymentQueryService(client mysql.Client) query.PaymentQueryService {
	return &paymentQueryService{client: client}
}
