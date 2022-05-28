package mysql

import (
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/common/infra/mysql"
	"github.com/klwxsrx/arch-course-project/pkg/delivery/domain"
)

type deliveryRepo struct {
	client mysql.Client
}

func (r *deliveryRepo) GetByID(orderID uuid.UUID) (*domain.Delivery, error) {
	const deliveryQuery = `
		SELECT order_id, status, address
		FROM delivery
		WHERE order_id = ?
	`

	binaryOrderID, err := orderID.MarshalBinary()
	if err != nil {
		return nil, err
	}

	var deliverySqlx sqlxDelivery
	err = r.client.Get(&deliverySqlx, deliveryQuery, binaryOrderID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrItemNotFound
	}
	if err != nil {
		return nil, err
	}

	return &domain.Delivery{
		OrderID: deliverySqlx.OrderID,
		Status:  domain.DeliveryStatus(deliverySqlx.Status),
		Address: deliverySqlx.Address,
	}, nil
}

func (r *deliveryRepo) Store(d *domain.Delivery) error {
	const deliveryQuery = `
		INSERT INTO delivery (order_id, status, address, created_at)
		VALUES (?, ?, ?, NOW())
		ON DUPLICATE KEY UPDATE
			status = VALUES(status), address = VALUES(address), updated_at = NOW()
	`

	binaryOrderID, err := d.OrderID.MarshalBinary()
	if err != nil {
		return err
	}

	_, err = r.client.Exec(deliveryQuery, binaryOrderID, d.Status, d.Address)
	return err
}

func NewDeliveryRepository(client mysql.Client) domain.DeliveryRepository {
	return &deliveryRepo{client}
}

type sqlxDelivery struct {
	OrderID uuid.UUID `db:"order_id"`
	Status  int       `db:"status"`
	Address string    `db:"address"`
}
