package service

import (
	"Confetki/internal/model"
	"Confetki/internal/repository"
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
)

// Универсальный Mock-репозиторий заказов
type mockOrderRepo struct {
	OrderRepository   // Встраиваем интерфейс
	getOrderByIDFn    func(ctx context.Context, id int) (model.Order, error)
	createFullOrderFn func(ctx context.Context, order model.Order, items []model.OrderItem) (model.Order, error)
}

func (m *mockOrderRepo) GetOrderByID(ctx context.Context, id int) (model.Order, error) {
	if m.getOrderByIDFn != nil {
		return m.getOrderByIDFn(ctx, id)
	}
	return model.Order{}, nil
}

func (m *mockOrderRepo) CreateFullOrder(ctx context.Context, order model.Order, items []model.OrderItem) (model.Order, error) {
	if m.createFullOrderFn != nil { // ИСПРАВЛЕНО
		return m.createFullOrderFn(ctx, order, items)
	}
	return model.Order{}, nil
}

// Универсальный Mock-репозиторий товаров
type mockProductRepo struct {
	ProductRepository
	getByIDFn func(ctx context.Context, id int) (model.Product, error)
}

func (mp *mockProductRepo) GetByID(ctx context.Context, id int) (model.Product, error) {
	if mp.getByIDFn != nil {
		return mp.getByIDFn(ctx, id)
	}
	return model.Product{}, nil
}

// 2. Тесты через t.Run
func TestOrderService_GetOrderByID(t *testing.T) {
	t.Run("invalid id: should return ErrInvalidID without touching repo", func(t *testing.T) {
		svc := NewOrderService(&mockOrderRepo{}, &mockProductRepo{})

		_, err := svc.GetOrderByID(context.Background(), -1)

		if !errors.Is(err, ErrInvalidID) {
			t.Errorf("got %v, want %v", err, ErrInvalidID)
		}
	})

	t.Run("not found: repo returns pgx.ErrNoRows -> service returns ErrOrderNotFound", func(t *testing.T) {
		repo := &mockOrderRepo{
			getOrderByIDFn: func(ctx context.Context, id int) (model.Order, error) {
				return model.Order{}, pgx.ErrNoRows
			},
		}
		svc := NewOrderService(repo, &mockProductRepo{})

		_, err := svc.GetOrderByID(context.Background(), 42)

		if !errors.Is(err, ErrOrderNotFound) {
			t.Errorf("got %v, want %v", err, ErrOrderNotFound)
		}
	})

	t.Run("success: returns order from repo", func(t *testing.T) {
		expectedOrder := model.Order{
			ID:         42,
			UserID:     1,
			Status:     "new",
			TotalPrice: 500,
		}

		repo := &mockOrderRepo{
			getOrderByIDFn: func(ctx context.Context, id int) (model.Order, error) {
				return expectedOrder, nil
			},
		}
		svc := NewOrderService(repo, &mockProductRepo{})

		order, err := svc.GetOrderByID(context.Background(), 42)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if order.ID != expectedOrder.ID || order.TotalPrice != expectedOrder.TotalPrice {
			t.Errorf("got %+v, want %+v", order, expectedOrder)
		}
	})
}

func TestOrderService_CreateOrder(t *testing.T) {
	t.Run("not enough product: repo returns ErrNotEnoughProduct", func(t *testing.T) {
		prodRepo := &mockProductRepo{
			getByIDFn: func(ctx context.Context, id int) (model.Product, error) {
				return model.Product{ID: 2, Price: 100}, nil
			},
		}

		orderRepo := &mockOrderRepo{
			createFullOrderFn: func(ctx context.Context, order model.Order, items []model.OrderItem) (model.Order, error) {
				return model.Order{}, repository.ErrNotEnoughProduct
			},
		}

		req := model.CreateOrderRequest{
			UserID: 1,
			Items: []model.CreateOrderItem{
				{ProductID: 2, Quantity: 5},
			},
		}
		svc := NewOrderService(orderRepo, prodRepo)
		_, err := svc.CreateOrder(context.Background(), req)

		if !errors.Is(err, repository.ErrNotEnoughProduct) {
			t.Errorf("got %v, want %v", err, repository.ErrNotEnoughProduct)
		}
	})

	t.Run("success", func(t *testing.T) {
		expectedOrder := model.Order{
			ID:         1,
			UserID:     9,
			TotalPrice: 500,
			Status:     "new",
		}
		prodRepo := &mockProductRepo{
			getByIDFn: func(ctx context.Context, id int) (model.Product, error) {
				return model.Product{ID: 2, Price: 250}, nil
			},
		}

		orderRepo := &mockOrderRepo{
			createFullOrderFn: func(ctx context.Context, order model.Order, items []model.OrderItem) (model.Order, error) {
				return expectedOrder, nil
			},
		}

		req := model.CreateOrderRequest{
			UserID: 2,
			Items: []model.CreateOrderItem{
				{ProductID: 2, Quantity: 2},
			},
		}

		svc := NewOrderService(orderRepo, prodRepo)
		order, err := svc.CreateOrder(context.Background(), req)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if order.ID != expectedOrder.ID || order.TotalPrice != expectedOrder.TotalPrice {
			t.Errorf("got %+v, want %+v", order, expectedOrder)
		}
	})
}
