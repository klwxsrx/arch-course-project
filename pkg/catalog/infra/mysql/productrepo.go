package mysql

import (
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/klwxsrx/arch-course-project/pkg/catalog/domain"
	"github.com/klwxsrx/arch-course-project/pkg/common/infra/mysql"
)

type productRepo struct {
	client mysql.Client
}

func (r *productRepo) NextID() uuid.UUID {
	return uuid.New()
}

func (r *productRepo) GetByID(id uuid.UUID) (*domain.Product, error) {
	const query = `SELECT id, title, description, price FROM product WHERE id = ?`

	binaryID, err := id.MarshalBinary()
	if err != nil {
		return nil, err
	}

	var productSqlx sqlxProduct
	err = r.client.Get(&productSqlx, query, binaryID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrProductNotExists
	}
	if err != nil {
		return nil, err
	}

	return &domain.Product{
		ID:          productSqlx.ID,
		Title:       productSqlx.Title,
		Description: productSqlx.Description,
		Price:       productSqlx.Price,
	}, nil
}

func (r *productRepo) Store(product *domain.Product) error {
	const query = `
		INSERT INTO product (id, title, description, price, created_at)
		VALUES (?, ?, ?, ?, NOW())
		ON DUPLICATE KEY UPDATE
			title = VALUES(title), description = VALUES(description), price = VALUES(price), updated_at = NOW()
	`

	binaryID, err := product.ID.MarshalBinary()
	if err != nil {
		return err
	}

	_, err = r.client.Exec(query, binaryID, product.Title, product.Description, product.Price)
	return err
}

func NewProductRepository(client mysql.Client) domain.ProductRepository {
	return &productRepo{client: client}
}

type sqlxProduct struct {
	ID          uuid.UUID `db:"id"`
	Title       string    `db:"title"`
	Description string    `db:"description"`
	Price       int       `db:"price"`
}
