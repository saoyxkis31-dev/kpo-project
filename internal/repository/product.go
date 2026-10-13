package repository

import (
	"Confetki/internal/model"
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrProductNotFound = errors.New("product not found")

type ProductRepository struct {
	db *pgxpool.Pool
}

func NewProductRepository(db *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{
		db: db,
	}
}

func (r *ProductRepository) GetAll(ctx context.Context) ([]model.Product, error) {
	rows, err := r.db.Query(
		ctx,
		`SELECT id, name, description, price, image_url, stock
	FROM products
	ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var products []model.Product
	for rows.Next() {
		var product model.Product
		err := rows.Scan(
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.ImageURL,
			&product.Stock,
		)

		if err != nil {
			return nil, err
		}

		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return products, nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id int) (model.Product, error) {
	var product model.Product

	err := r.db.QueryRow(
		ctx,
		`SELECT id, name, description, price, image_url, stock
		FROM products
		WHERE id = $1;`,
		id,
	).Scan(
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.ImageURL,
		&product.Stock,
	)
	if err != nil {
		return model.Product{}, err
	}

	return product, nil
}

func (r *ProductRepository) Create(ctx context.Context, product model.Product) (model.Product, error) {
	var createdProduct model.Product

	err := r.db.QueryRow(
		ctx,
		`INSERT INTO products (name, description, price, image_url, stock)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, description, price, image_url, stock`,
		product.Name,
		product.Description,
		product.Price,
		product.ImageURL,
		product.Stock,
	).Scan(
		&createdProduct.ID,
		&createdProduct.Name,
		&createdProduct.Description,
		&createdProduct.Price,
		&createdProduct.ImageURL,
		&createdProduct.Stock,
	)

	if err != nil {
		return model.Product{}, err
	}

	return createdProduct, nil
}

func (r *ProductRepository) Delete(ctx context.Context, id int) error {
	sqlQuery := `
	DELETE FROM products
	WHERE id = $1`

	result, err := r.db.Exec(ctx, sqlQuery, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrProductNotFound
	}

	return nil
}
