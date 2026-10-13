package service

import (
	"Confetki/internal/model"
	"context"
)

type MasterClassRepository interface {
	GetByProductID(ctx context.Context, productID int) (model.MasterClass, error)
	Create(ctx context.Context, masterClass model.MasterClass) (model.MasterClass, error)
}

type MasterClassService struct {
	repository MasterClassRepository
}

func NewMasterClassService(repository MasterClassRepository) *MasterClassService {
	return &MasterClassService{
		repository: repository,
	}
}

func (s *MasterClassService) GetMasterClassByProductID(ctx context.Context, product_id int) (model.MasterClass, error) {
	return s.repository.GetByProductID(ctx, product_id)
}

func (s *MasterClassService) CreateMasterClass(ctx context.Context, masterClass model.MasterClass) (model.MasterClass, error) {
	return s.repository.Create(ctx, masterClass)
}
