package mysql

import (
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/klwxsrx/arch-course-project/pkg/catalog/app/query"
	"github.com/klwxsrx/arch-course-project/pkg/common/infra/mysql"
)

type productQueryService struct {
	client mysql.Client
}

func (s *productQueryService) ListAll() ([]query.ProductData, error) {
	const selectQuery = `SELECT id, title, description, price FROM product`

	var productsSqlx []sqlxProduct
	err := s.client.Select(&productsSqlx, selectQuery)
	if err != nil {
		return nil, err
	}

	result := make([]query.ProductData, 0, len(productsSqlx))
	for _, item := range productsSqlx {
		result = append(result, query.ProductData{
			ID:          item.ID,
			Title:       item.Title,
			Description: item.Description,
			Price:       item.Price,
		})
	}

	return result, nil
}

func (s *productQueryService) GetByIDs(ids []uuid.UUID) ([]query.ProductData, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	binaryIDs := make([][]byte, 0, len(ids))
	for _, id := range ids {
		binaryID, err := id.MarshalBinary()
		if err != nil {
			return nil, err
		}
		binaryIDs = append(binaryIDs, binaryID)
	}

	selectQuery, args, err := sqlx.In(`SELECT id, title, description, price FROM product WHERE id IN (?)`, binaryIDs)
	if err != nil {
		return nil, err
	}

	var productsSqlx []sqlxProduct
	err = s.client.Select(&productsSqlx, selectQuery, args...)
	if err != nil {
		return nil, err
	}

	if len(ids) != len(productsSqlx) {
		return nil, query.ErrProductByIDNotFound
	}

	result := make([]query.ProductData, 0, len(productsSqlx))
	for _, item := range productsSqlx {
		result = append(result, query.ProductData{
			ID:          item.ID,
			Title:       item.Title,
			Description: item.Description,
			Price:       item.Price,
		})
	}

	return result, nil
}

func NewProductQueryService(client mysql.Client) query.ProductService {
	return &productQueryService{client: client}
}
