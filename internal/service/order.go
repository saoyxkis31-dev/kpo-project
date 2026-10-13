package service

import (
	"Confetki/internal/model"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

var ErrInvalidID = errors.New("invalid order id")
var ErrOrderNotFound = errors.New("order not found")
var ErrBadStatus = errors.New("bad status")
var ErrProductNotFound = errors.New("product not found")

type OrderRepository interface {
	CreateFullOrder(ctx context.Context, order model.Order, items []model.OrderItem) (model.Order, error)
	GetOrderByID(ctx context.Context, id int) (model.Order, error)
	GetAllOrders(ctx context.Context) ([]model.Order, error)
	UpdateStatus(ctx context.Context, orderID int, status string) error
	CancelOrderWithStockReturn(ctx context.Context, orderID int) error
	GetOrderByUserID(ctx context.Context, userID int) ([]model.Order, error)
}

type OrderService struct {
	orderRepository   OrderRepository
	productRepository ProductRepository
}

func NewOrderService(orderRepository OrderRepository, productRepository ProductRepository) *OrderService {
	return &OrderService{
		orderRepository:   orderRepository,
		productRepository: productRepository,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, request model.CreateOrderRequest) (model.Order, error) {

	var totalPrice float64
	var orderItems []model.OrderItem

	for _, item := range request.Items {
		product, err := s.productRepository.GetByID(ctx, item.ProductID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return model.Order{}, ErrProductNotFound
			}
			return model.Order{}, err
		}

		orderItem := model.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     product.Price,
		}

		totalPrice += orderItem.Price * float64(item.Quantity)
		orderItems = append(orderItems, orderItem)
	}

	order := model.Order{
		UserID:     request.UserID,
		Status:     "new",
		TotalPrice: totalPrice,
	}

	createdOrder, err := s.orderRepository.CreateFullOrder(ctx, order, orderItems)
	if err != nil {
		return model.Order{}, err
	}

	return createdOrder, nil
}

func (s *OrderService) GetOrderByID(ctx context.Context, id int) (model.Order, error) {
	if id <= 0 {
		return model.Order{}, ErrInvalidID
	}

	order, err := s.orderRepository.GetOrderByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Order{}, ErrOrderNotFound
		}

		return model.Order{}, err
	}

	return order, nil
}

func (s *OrderService) GetAllOrders(ctx context.Context) ([]model.Order, error) {
	return s.orderRepository.GetAllOrders(ctx)
}

func (s *OrderService) UpdateStatus(ctx context.Context, orderID int, status string) error {
	switch status {
	case "paid", "completed":
		err := s.orderRepository.UpdateStatus(ctx, orderID, status)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrOrderNotFound
			}
			return err
		}
	case "canceled":
		err := s.orderRepository.CancelOrderWithStockReturn(ctx, orderID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrOrderNotFound
			}
			return err
		}
	default:
		return ErrBadStatus
	}

	return nil
}

func (s *OrderService) GetOrderByUserID(ctx context.Context, userID int) ([]model.Order, error) {
	return s.orderRepository.GetOrderByUserID(ctx, userID)
}
