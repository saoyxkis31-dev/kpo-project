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
	if m.createFullOrderFn != nil {
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

func TestOrderService_GetOrderByID(t *testing.T) {
	tests := []struct {
		name       string
		orderID    int
		mockRepoFn func(ctx context.Context, id int) (model.Order, error)
		wantErr    error
		wantOrder  model.Order
	}{
		{
			name:       "invalid id: negative",
			orderID:    -1,
			mockRepoFn: nil,
			wantErr:    ErrInvalidID,
			wantOrder:  model.Order{},
		},
		{
			name:    "not found in database",
			orderID: 42,
			mockRepoFn: func(ctx context.Context, id int) (model.Order, error) {
				return model.Order{}, pgx.ErrNoRows
			},
			wantErr:   ErrOrderNotFound,
			wantOrder: model.Order{},
		},
		{
			name:    "success: order found",
			orderID: 42,
			mockRepoFn: func(ctx context.Context, id int) (model.Order, error) {
				return model.Order{ID: 42, TotalPrice: 500}, nil
			},
			wantErr:   nil,
			wantOrder: model.Order{ID: 42, TotalPrice: 500},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderRepo := &mockOrderRepo{getOrderByIDFn: tt.mockRepoFn}
			svc := NewOrderService(orderRepo, &mockProductRepo{})

			order, err := svc.GetOrderByID(context.Background(), tt.orderID)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}

			if tt.wantErr == nil && order.ID != tt.wantOrder.ID {
				t.Errorf("expected order ID %d, got %d", tt.wantOrder.ID, order.ID)
			}
		})
	}
}

func TestOrderService_CreateOrder(t *testing.T) {
	tests := []struct {
		name            string
		req             model.CreateOrderRequest                                                                   // Входные данные
		mockProdRepoFn  func(ctx context.Context, id int) (model.Product, error)                                   // Поведение репозитория товаров
		mockOrderRepoFn func(ctx context.Context, order model.Order, items []model.OrderItem) (model.Order, error) // Поведение репозитория заказов
		wantErr         error
		wantOrder       model.Order
	}{
		{
			name: "not enough product",
			req: model.CreateOrderRequest{
				UserID: 1,
				Items: []model.CreateOrderItem{
					{ProductID: 2, Quantity: 5},
				},
			},
			mockProdRepoFn: func(ctx context.Context, id int) (model.Product, error) {
				return model.Product{ID: 2, Price: 100}, nil
			},
			mockOrderRepoFn: func(ctx context.Context, order model.Order, items []model.OrderItem) (model.Order, error) {
				return model.Order{}, repository.ErrNotEnoughProduct
			},
			wantErr:   repository.ErrNotEnoughProduct,
			wantOrder: model.Order{},
		},
		{
			name: "success",
			req: model.CreateOrderRequest{
				UserID: 9,
				Items: []model.CreateOrderItem{
					{ProductID: 2, Quantity: 2},
				},
			},
			mockProdRepoFn: func(ctx context.Context, id int) (model.Product, error) {
				return model.Product{ID: 2, Price: 250}, nil
			},
			mockOrderRepoFn: func(ctx context.Context, order model.Order, items []model.OrderItem) (model.Order, error) {
				return model.Order{
					ID:         1,
					UserID:     9,
					TotalPrice: 500,
					Status:     "new",
				}, nil
			},
			wantErr: nil,
			wantOrder: model.Order{
				ID:         1,
				UserID:     9,
				TotalPrice: 500,
				Status:     "new",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prodRepo := &mockProductRepo{getByIDFn: tt.mockProdRepoFn}
			orderRepo := &mockOrderRepo{createFullOrderFn: tt.mockOrderRepoFn}

			svc := NewOrderService(orderRepo, prodRepo)

			order, err := svc.CreateOrder(context.Background(), tt.req)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}

			if tt.wantErr == nil && (order.ID != tt.wantOrder.ID || order.TotalPrice != tt.wantOrder.TotalPrice) {
				t.Errorf("got order %+v, want %+v", order, tt.wantOrder)
			}
		})
	}
}
