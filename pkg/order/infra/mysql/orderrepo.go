package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/common/infra/mysql"
	"github.com/klwxsrx/arch-course-project/pkg/order/domain"
	"strings"
)

type orderRepo struct {
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

	const itemsQuery = `
		SELECT id, price, quantity
		FROM order_item
		WHERE order_id = ?
	`

	var sqlxItems []sqlxOrderItem
	err = r.client.Select(&sqlxItems, itemsQuery, binaryID)
	if err != nil {
		return nil, err
	}

	orderItems := make([]domain.OrderItem, 0, len(sqlxItems))
	for _, sqlxItem := range sqlxItems {
		orderItems = append(orderItems, domain.OrderItem{
			ID:        sqlxItem.ID,
			ItemPrice: sqlxItem.Price,
			Quantity:  sqlxItem.Quantity,
		})
	}

	return &domain.Order{
		ID:          orderSqlx.ID,
		UserID:      orderSqlx.UserID,
		AddressID:   orderSqlx.AddressID,
		Items:       orderItems,
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

	binaryOrderID, err := order.ID.MarshalBinary()
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

	_, err = r.client.Exec(orderQuery, binaryOrderID, binaryUserID, binaryAddressID, int(order.Status), order.TotalAmount)
	if err != nil {
		return err
	}

	_, err = r.client.Exec(`DELETE FROM order_item WHERE order_id = ?`, binaryOrderID)
	if err != nil {
		return err
	}

	if len(order.Items) == 0 {
		return nil
	}

	insertQuery := fmt.Sprintf(`
		INSERT INTO order_item (id, order_id, price, quantity)
		VALUES %s%s
	`, "(?, ?, ?, ?)", strings.Repeat(", (?, ?, ?, ?)", len(order.Items)-1))
	args := make([]any, 0, len(order.Items)*4) // arguments count
	for _, item := range order.Items {
		binaryItemID, err := item.ID.MarshalBinary()
		if err != nil {
			return err
		}
		args = append(args, binaryItemID, binaryOrderID, item.ItemPrice, item.Quantity)
	}

	_, err = r.client.Exec(insertQuery, args...)
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

type sqlxOrderItem struct {
	ID       uuid.UUID `db:"id"`
	OrderID  uuid.UUID `db:"order_id"`
	Price    int       `db:"price"`
	Quantity int       `db:"quantity"`
}
