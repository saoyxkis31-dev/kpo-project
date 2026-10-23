package service

import (
	"Confetki/internal/model"
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type ProductRepository interface {
	GetAll(ctx context.Context) ([]model.Product, error)
	GetByID(ctx context.Context, id int) (model.Product, error)
	Create(ctx context.Context, product model.Product) (model.Product, error)
	Delete(ctx context.Context, id int) error
}

type ProductService struct {
	repository ProductRepository
	redis      *redis.Client
}

func NewProductService(repository ProductRepository, rdb *redis.Client) *ProductService {
	return &ProductService{
		repository: repository,
		redis:      rdb,
	}
}

func (s *ProductService) GetAllProducts(ctx context.Context) ([]model.Product, error) {
	casheKey := "products:all"

	cashedData, err := s.redis.Get(ctx, casheKey).Bytes()
	if err == nil { // Cache HIT данные есть в кэше,  декодируем и отдаем
		var products []model.Product
		if err := json.Unmarshal(cashedData, &products); err == nil {
			return products, nil
		}
	}

	products, err := s.repository.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(products)
	if err == nil {
		_ = s.redis.Set(ctx, casheKey, data, 5*time.Minute).Err()
	}

	return products, nil
}

func (s *ProductService) GetProductByID(ctx context.Context, id int) (model.Product, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *ProductService) CreateProduct(ctx context.Context, product model.Product) (model.Product, error) {
	created, err := s.repository.Create(ctx, product)
	if err != nil {
		return model.Product{}, err
	}

	_ = s.redis.Del(ctx, "products:all").Err()

	return created, nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, id int) error {
	err := s.repository.Delete(ctx, id)
	if err != nil {
		return err // если ошибка  то кэш не трогаем
	}

	//  после успеха сбрасываем кэш
	_ = s.redis.Del(ctx, "products:all").Err()

	return nil
}
