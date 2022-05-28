package mysql

import (
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/common/infra/mysql"
	"github.com/klwxsrx/arch-course-project/pkg/order/domain"
)

type orderRepo struct { // TODO: store order items
	client mysql.Client
}

func (r *orderRepo) NextID() uuid.UUID {
	return uuid.New()
}

func (r *orderRepo) GetByID(id uuid.UUID) (*domain.Order, error) {
	const orderQuery = `
		SELECT id, user_id, address_id, status, total_amount
		FROM ` + " `order` " + `
		WHERE id = ?
	`

	binaryID, err := id.MarshalBinary()
	if err != nil {
		return nil, err
	}

	var orderSqlx sqlxOrder
	err = r.client.Get(&orderSqlx, orderQuery, binaryID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}

	return &domain.Order{
		ID:          orderSqlx.ID,
		UserID:      orderSqlx.UserID,
		AddressID:   orderSqlx.AddressID,
		Items:       nil,
		Status:      domain.OrderStatus(orderSqlx.Status),
		TotalAmount: orderSqlx.TotalAmount,
	}, nil
}

func (r *orderRepo) Store(order *domain.Order) error {
	const orderQuery = `
		INSERT INTO` + " `order` " + `(id, user_id, address_id, status, total_amount, created_at)
		VALUES (?, ?, ?, ?, ?, NOW())
		ON DUPLICATE KEY UPDATE
			user_id = VALUES(user_id), address_id = VALUES(address_id), status = VALUES(status),
			total_amount = VALUES(total_amount), updated_at = NOW()
	`

	binaryID, err := order.ID.MarshalBinary()
	if err != nil {
		return err
	}

	binaryUserID, err := order.UserID.MarshalBinary()
	if err != nil {
		return err
	}

	binaryAddressID, err := order.AddressID.MarshalBinary()
	if err != nil {
		return err
	}

	_, err = r.client.Exec(orderQuery, binaryID, binaryUserID, binaryAddressID, int(order.Status), order.TotalAmount)
	return err
}

func NewOrderRepository(client mysql.Client) domain.OrderRepository {
	return &orderRepo{client: client}
}

type sqlxOrder struct {
	ID          uuid.UUID `db:"id"`
	UserID      uuid.UUID `db:"user_id"`
	AddressID   uuid.UUID `db:"address_id"`
	Status      int       `db:"status"`
	TotalAmount int       `db:"total_amount"`
}
