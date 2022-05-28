package mysql

import (
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/common/infra/mysql"
	"github.com/klwxsrx/arch-course-project/pkg/payment/domain"
)

type paymentRepo struct {
	client mysql.Client
}

func (r *paymentRepo) GetByID(id uuid.UUID) (*domain.Payment, error) {
	const paymentQuery = `
		SELECT order_id, status, total_amount
		FROM ` + " `payment` " + `
		WHERE order_id = ?
	`

	binaryID, err := id.MarshalBinary()
	if err != nil {
		return nil, err
	}

	var paymentSqlx sqlxPayment
	err = r.client.Get(&paymentSqlx, paymentQuery, binaryID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrPaymentNotFound
	}
	if err != nil {
		return nil, err
	}

	return &domain.Payment{
		OrderID:     paymentSqlx.OrderID,
		TotalAmount: paymentSqlx.TotalAmount,
		Status:      domain.PaymentStatus(paymentSqlx.Status),
	}, nil
}

func (r *paymentRepo) Store(payment *domain.Payment) error {
	const paymentQuery = `
		INSERT INTO` + " `payment` " + `(order_id, status, total_amount, created_at)
		VALUES (?, ?, ?, NOW())
		ON DUPLICATE KEY UPDATE
			status = VALUES(status), total_amount = VALUES(total_amount), updated_at = NOW()
	`

	binaryOrderID, err := payment.OrderID.MarshalBinary()
	if err != nil {
		return err
	}

	_, err = r.client.Exec(paymentQuery, binaryOrderID, int(payment.Status), payment.TotalAmount)
	return err
}

func NewPaymentRepository(client mysql.Client) domain.PaymentRepository {
	return &paymentRepo{client: client}
}

type sqlxPayment struct {
	OrderID     uuid.UUID `db:"order_id"`
	Status      int       `db:"status"`
	TotalAmount int       `db:"total_amount"`
}
