package repository

import (
	"Confetki/internal/model"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MasterClassRepository struct {
	db *pgxpool.Pool
}

func NewMasterClassRepository(db *pgxpool.Pool) *MasterClassRepository {
	return &MasterClassRepository{
		db: db,
	}
}

func (r *MasterClassRepository) Create(ctx context.Context, masterClass model.MasterClass) (model.MasterClass, error) {
	var createdMasterClass model.MasterClass

	err := r.db.QueryRow(
		ctx,
		`INSERT INTO  master_classes (title, description, product_id, video_url)
		VALUES ($1, $2, $3, $4)
		RETURNING id, title, description, product_id, video_url`,
		masterClass.Title,
		masterClass.Description,
		masterClass.ProductID,
		masterClass.VideoURL,
	).Scan(
		&createdMasterClass.ID,
		&createdMasterClass.Title,
		&createdMasterClass.Description,
		&createdMasterClass.ProductID,
		&createdMasterClass.VideoURL,
	)

	if err != nil {
		return model.MasterClass{}, err
	}

	return createdMasterClass, nil
}

func (r *MasterClassRepository) GetByProductID(ctx context.Context, productID int) (model.MasterClass, error) {
	var masterClass model.MasterClass

	err := r.db.QueryRow(
		ctx,
		`SELECT id, title, description, product_id, video_url
		FROM master_classes
		WHERE product_id = $1`,
		productID,
	).Scan(
		&masterClass.ID,
		&masterClass.Title,
		&masterClass.Description,
		&masterClass.ProductID,
		&masterClass.VideoURL,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.MasterClass{}, ErrProductNotFound
		}

		return model.MasterClass{}, err
	}

	return masterClass, nil
}
