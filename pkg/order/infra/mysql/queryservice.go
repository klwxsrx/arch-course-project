package mysql

import (
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/common/infra/mysql"
	"github.com/klwxsrx/arch-course-project/pkg/order/app/query"
	"github.com/klwxsrx/arch-course-project/pkg/order/domain"
)

type orderQueryService struct {
	client mysql.Client
}

func (s *orderQueryService) GetOrderData(id uuid.UUID) (*query.OrderData, error) {
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
	err = s.client.Get(&orderSqlx, orderQuery, binaryID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, query.ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}

	const itemsQuery = `
		SELECT id, price, quantity
		FROM order_item
		WHERE order_id = ?
	`

	var sqlxItems []sqlxOrderItem
	err = s.client.Select(&sqlxItems, itemsQuery, binaryID)
	if err != nil {
		return nil, err
	}

	orderItems := make([]query.OrderItemData, 0, len(sqlxItems))
	for _, sqlxItem := range sqlxItems {
		orderItems = append(orderItems, query.OrderItemData{
			ID:        sqlxItem.ID,
			ItemPrice: sqlxItem.Price,
			Quantity:  sqlxItem.Quantity,
		})
	}

	return &query.OrderData{
		ID:          orderSqlx.ID,
		UserID:      orderSqlx.UserID,
		AddressID:   orderSqlx.AddressID,
		Items:       orderItems,
		Status:      domain.OrderStatus(orderSqlx.Status),
		TotalAmount: orderSqlx.TotalAmount,
	}, nil
}

func NewOrderQueryService(client mysql.Client) query.Service {
	return &orderQueryService{client: client}
}
