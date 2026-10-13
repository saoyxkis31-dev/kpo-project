package service

import (
	"Confetki/internal/model"
	"context"
)

type ProductRepository interface {
	GetAll(ctx context.Context) ([]model.Product, error)
	GetByID(ctx context.Context, id int) (model.Product, error)
	Create(ctx context.Context, product model.Product) (model.Product, error)
	Delete(ctx context.Context, id int) error
}

type ProductService struct {
	repository ProductRepository
}

func NewProductService(repository ProductRepository) *ProductService {
	return &ProductService{
		repository: repository,
	}
}

func (s *ProductService) GetAllProducts(ctx context.Context) ([]model.Product, error) {
	return s.repository.GetAll(ctx)
}

func (s *ProductService) GetProductByID(ctx context.Context, id int) (model.Product, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *ProductService) CreateProduct(ctx context.Context, product model.Product) (model.Product, error) {
	return s.repository.Create(ctx, product)
}

func (s *ProductService) DeleteProduct(ctx context.Context, id int) error {
	return s.repository.Delete(ctx, id)
}
