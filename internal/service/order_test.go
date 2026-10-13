package service

import (
	"context"
	"errors"
	"testing"
)

type fakeOrderRepos struct {
	OrderRepository
}

type fakeProductRepos struct {
	ProductRepository
}

func TestGetOrderByID_INVALID_ID(t *testing.T) {
	orderService := NewOrderService(fakeOrderRepos{}, fakeProductRepos{})

	_, err := orderService.GetOrderByID(context.Background(), -32)
	if !errors.Is(err, ErrInvalidID) {
		t.Errorf("ждали ошибку errInvalidID, получили: %v", err)
	}
}
