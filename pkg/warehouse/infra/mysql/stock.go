package mysql

import (
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/klwxsrx/arch-course-project/pkg/common/infra/mysql"
	"github.com/klwxsrx/arch-course-project/pkg/warehouse/domain"
)

type stock struct {
	client mysql.Client
}

func (s *stock) NextID() uuid.UUID {
	return uuid.New()
}

func (s *stock) GetAvailableItemsQuantity(itemIDs []uuid.UUID) ([]domain.ItemQuantity, error) {
	if itemIDs == nil {
		return nil, nil
	}

	const query = `
		SELECT item_id, SUM(quantity) AS quantity
		FROM stock_balance
		WHERE item_id IN (?) AND deleted_at IS NULL
		GROUP BY item_id
	`

	binaryItemIDs := make([][]byte, 0, len(itemIDs))
	for _, itemID := range itemIDs {
		binaryID, err := itemID.MarshalBinary()
		if err != nil {
			return nil, err
		}
		binaryItemIDs = append(binaryItemIDs, binaryID)
	}

	resultQuery, args, err := sqlx.In(query, binaryItemIDs)
	if err != nil {
		return nil, err
	}

	var itemsQuantity []struct {
		ItemID   uuid.UUID `db:"item_id"`
		Quantity int       `db:"quantity"`
	}

	err = s.client.Select(&itemsQuantity, resultQuery, args...)
	if err != nil {
		return nil, err
	}

	result := make([]domain.ItemQuantity, 0, len(itemsQuantity))
	for _, item := range itemsQuantity {
		result = append(result, domain.ItemQuantity{
			ItemID:   item.ItemID,
			Quantity: item.Quantity,
		})
	}
	return result, nil
}

func (s *stock) GetOrderOperations(orderID uuid.UUID) ([]domain.StockOperation, error) {
	const query = `
		SELECT id, item_id, type, quantity, order_id
		FROM stock_balance
		WHERE order_id = ? AND deleted_at IS NULL
	`

	binaryID, err := orderID.MarshalBinary()
	if err != nil {
		return nil, err
	}

	var sqlxResult []sqlxOperation
	err = s.client.Select(&sqlxResult, query, binaryID)
	if err != nil {
		return nil, err
	}

	result := make([]domain.StockOperation, 0, len(sqlxResult))
	for _, sqlxItem := range sqlxResult {
		result = append(result, domain.StockOperation{
			ID:           sqlxItem.ID,
			ItemID:       sqlxItem.ItemID,
			Type:         domain.StockOperationType(sqlxItem.Type),
			ItemQuantity: sqlxItem.ItemQuantity,
			OrderID:      sqlxItem.OrderID,
		})
	}
	return result, nil
}

func (s *stock) Update(op *domain.StockOperation) error {
	const query = `
		INSERT INTO stock_balance (id, item_id, type, quantity, order_id)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			item_id = VALUES(item_id), type = VALUES(type), quantity = VALUES(quantity), order_id = VALUES(order_id),
			updated_at = NOW(), deleted_at = NULL
	`

	binaryID, err := op.ID.MarshalBinary()
	if err != nil {
		return err
	}

	binaryItemID, err := op.ItemID.MarshalBinary()
	if err != nil {
		return err
	}

	var binaryOrderID *[]byte
	if op.OrderID != nil {
		binaryID, err := op.OrderID.MarshalBinary()
		if err != nil {
			return err
		}
		binaryOrderID = &binaryID
	}

	_, err = s.client.Exec(query, binaryID, binaryItemID, op.Type, op.ItemQuantity, binaryOrderID)
	return err
}

func (s *stock) Delete(opIDs []uuid.UUID) error {
	if opIDs == nil {
		return nil
	}

	const query = `UPDATE stock_balance SET deleted_at = NOW() WHERE id IN (?)`

	binaryOpIDs := make([][]byte, 0, len(opIDs))
	for _, opID := range opIDs {
		binaryID, err := opID.MarshalBinary()
		if err != nil {
			return err
		}
		binaryOpIDs = append(binaryOpIDs, binaryID)
	}

	resultQuery, args, err := sqlx.In(query, binaryOpIDs)
	if err != nil {
		return err
	}

	_, err = s.client.Exec(resultQuery, args...)
	return err
}

func NewStock(client mysql.Client) domain.Stock {
	return &stock{client: client}
}

type sqlxOperation struct {
	ID           uuid.UUID  `db:"id"`
	ItemID       uuid.UUID  `db:"item_id"`
	Type         int        `db:"type"`
	ItemQuantity int        `db:"quantity"`
	OrderID      *uuid.UUID `db:"order_id"`
}
