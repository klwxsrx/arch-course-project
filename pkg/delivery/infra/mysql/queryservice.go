package mysql

import (
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/common/infra/mysql"
	"github.com/klwxsrx/arch-course-project/pkg/delivery/app/query"
	"github.com/klwxsrx/arch-course-project/pkg/delivery/domain"
)

type queryService struct {
	client mysql.Client
}

func (q *queryService) GetByID(orderID uuid.UUID) (*query.Delivery, error) {
	const selectQuery = `SELECT order_id, status, address FROM delivery WHERE order_id = ?`

	binaryOrderID, err := orderID.MarshalBinary()
	if err != nil {
		return nil, err
	}

	var deliverySqlx sqlxDelivery
	err = q.client.Get(&deliverySqlx, selectQuery, binaryOrderID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, query.ErrDeliveryNotFound
	}
	if err != nil {
		return nil, err
	}

	return &query.Delivery{
		OrderID: deliverySqlx.OrderID,
		Status:  domain.DeliveryStatus(deliverySqlx.Status),
		Address: deliverySqlx.Address,
	}, nil
}

func NewQueryService(client mysql.Client) query.Service {
	return &queryService{client: client}
}
