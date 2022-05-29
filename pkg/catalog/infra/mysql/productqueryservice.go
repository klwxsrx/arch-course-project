package mysql

import (
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

func NewProductQueryService(client mysql.Client) query.ProductService {
	return &productQueryService{client: client}
}
